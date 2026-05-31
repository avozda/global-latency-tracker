package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"os"
	"strings"
	"time"

	proberesult "github.com/avozda/global-latency-tracker/internal/probe"
	"github.com/joho/godotenv"
)

const DEFAULT_PROBE_INTERVAL = 60 * time.Second

type probeConfig struct {
	targetURL *url.URL
	interval  time.Duration
	hubURL    string
	apiKey    string
}

type probeTimings struct {
	start        time.Time
	dnsStart     time.Time
	dnsDone      time.Time
	connectStart time.Time
	connectDone  time.Time
	tlsStart     time.Time
	tlsDone      time.Time
	wroteRequest time.Time
	firstByte    time.Time
	roundTrip    float64
}

func main() {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, "Error loading .env:", err)
		os.Exit(1)
	}

	cfg, err := loadProbeConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	ticker := time.NewTicker(cfg.interval)
	defer ticker.Stop()

	for {
		probe(cfg)
		<-ticker.C
	}
}

func loadProbeConfig() (*probeConfig, error) {
	rawURL := strings.TrimSpace(os.Getenv("PROBE_TARGET_URL"))
	if rawURL == "" {
		return nil, errors.New("PROBE_TARGET_URL is required")
	}
	targetURL, err := validateProbeURL(rawURL)
	if err != nil {
		return nil, fmt.Errorf("PROBE_TARGET_URL: %w", err)
	}

	interval := DEFAULT_PROBE_INTERVAL
	if rawInterval := strings.TrimSpace(os.Getenv("PROBE_INTERVAL")); rawInterval != "" {
		interval, err = time.ParseDuration(rawInterval)
		if err != nil {
			return nil, fmt.Errorf("PROBE_INTERVAL: %w", err)
		}
		if interval <= 0 {
			return nil, errors.New("PROBE_INTERVAL must be greater than 0")
		}
	}

	hubURL := strings.TrimSpace(os.Getenv("HUB_URL"))
	if hubURL == "" {
		return nil, errors.New("HUB_URL is required")
	}
	if err := validateHubURL(hubURL); err != nil {
		return nil, fmt.Errorf("HUB_URL: %w", err)
	}

	apiKey := strings.TrimSpace(os.Getenv("API_KEY"))
	if apiKey == "" {
		return nil, errors.New("API_KEY is required")
	}

	return &probeConfig{
		targetURL: targetURL,
		interval:  interval,
		hubURL:    hubURL,
		apiKey:    apiKey,
	}, nil
}

func probe(cfg *probeConfig) {
	timings := &probeTimings{start: time.Now()}

	trace := newProbeTrace(timings)

	req, err := newProbeRequest(cfg.targetURL, trace)
	if err != nil {
		emitProbeResult(cfg, buildProbeResult(cfg.targetURL, timings, nil, err))
		return
	}

	resp, err := executeProbeRequest(req, timings)

	emitProbeResult(cfg, buildProbeResult(cfg.targetURL, timings, resp, err))
}

func emitProbeResult(cfg *probeConfig, result proberesult.Result) {
	writeProbeResult(result)
	if err := postProbeResult(cfg.hubURL, cfg.apiKey, result); err != nil {
		fmt.Fprintln(os.Stderr, "Error posting to hub:", err)
	}
}

func newProbeTrace(timings *probeTimings) *httptrace.ClientTrace {
	return &httptrace.ClientTrace{
		DNSStart: func(httptrace.DNSStartInfo) {
			timings.dnsStart = time.Now()
		},
		DNSDone: func(httptrace.DNSDoneInfo) {
			timings.dnsDone = time.Now()
		},
		ConnectStart: func(string, string) {
			timings.connectStart = time.Now()
		},
		ConnectDone: func(net, addr string, err error) {
			timings.connectDone = time.Now()
		},
		TLSHandshakeStart: func() {
			timings.tlsStart = time.Now()
		},
		TLSHandshakeDone: func(tls.ConnectionState, error) {
			timings.tlsDone = time.Now()
		},
		WroteRequest: func(info httptrace.WroteRequestInfo) {
			if info.Err == nil {
				timings.wroteRequest = time.Now()
			}
		},
		GotFirstResponseByte: func() {
			timings.firstByte = time.Now()
		},
	}
}

func newProbeRequest(targetURL *url.URL, trace *httptrace.ClientTrace) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, targetURL.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Close = true
	req.Header.Set("User-Agent", "Global-Latency-Tracker-Probe/1.0 (Monitoring Tool)")
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	return req, nil
}

func executeProbeRequest(req *http.Request, timings *probeTimings) (*http.Response, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)

	totalRoundTripMS := 0.0
	if err != nil {
		totalRoundTripMS = getDurationMS(timings.start, time.Now())
	} else {
		defer resp.Body.Close()
		_, readErr := io.Copy(io.Discard, resp.Body)
		totalRoundTripMS = getDurationMS(timings.start, time.Now())
		if readErr != nil {
			err = readErr
		}
	}

	timings.roundTrip = totalRoundTripMS

	return resp, err
}

func buildProbeResult(targetURL *url.URL, timings *probeTimings, resp *http.Response, err error) proberesult.Result {
	var errorPtr *string
	if err != nil {
		errStr := err.Error()
		errorPtr = &errStr
	}

	statusCode := 0
	if resp != nil {
		statusCode = resp.StatusCode
	}

	return proberesult.Result{
		TargetURL:          targetURL.String(),
		StatusCode:         statusCode,
		DNSLookupMS:        getDurationMS(timings.dnsStart, timings.dnsDone),
		TCPConnectionMS:    getDurationMS(timings.connectStart, timings.connectDone),
		TLSHandshakeMS:     getDurationMS(timings.tlsStart, timings.tlsDone),
		ServerProcessingMS: getDurationMS(timings.wroteRequest, timings.firstByte),
		MeasuredAt:         time.Now().UTC(),
		TotalRoundTripMS:   timings.roundTrip,
		TTFBMS:             getDurationMS(timings.start, timings.firstByte),
		Error:              errorPtr,
	}
}

func writeProbeResult(result proberesult.Result) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(result)
}

func postProbeResult(hubURL, apiKey string, result proberesult.Result) error {
	body, err := json.Marshal(result)
	if err != nil {
		return err
	}

	endpoint := strings.TrimRight(hubURL, "/") + "/api/metrics"
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("hub returned status %d", resp.StatusCode)
	}

	return nil
}

func getDurationMS(start, end time.Time) float64 {
	if start.IsZero() || end.IsZero() {
		return 0
	}
	return roundToDecimal(float64(end.Sub(start).Microseconds())/1000.0, 2)
}

func roundToDecimal(v float64, precision int) float64 {
	return math.Round(v*math.Pow10(precision)) / math.Pow10(precision)
}

func validateProbeURL(raw string) (*url.URL, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, errors.New("url is empty")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return nil, errors.New("invalid url")
	}
	switch strings.ToLower(u.Scheme) {
	case "http", "https":
	default:
		return nil, errors.New("unsupported scheme")
	}
	if u.Host == "" {
		return nil, errors.New("url missing host")
	}
	return u, nil
}

func validateHubURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return errors.New("invalid url")
	}
	switch strings.ToLower(u.Scheme) {
	case "http", "https":
	default:
		return errors.New("unsupported scheme")
	}
	if u.Host == "" {
		return errors.New("url missing host")
	}
	return nil
}
