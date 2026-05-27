package main

import (
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
)

const DEFAULT_PROBE_INTERVAL = 60 * time.Second

type probeResult struct {
	TargetURL          string  `json:"target_url"`
	StatusCode         int     `json:"status_code"`
	DNSLookupMS        float64 `json:"dns_lookup_ms"`
	TCPConnectionMS    float64 `json:"tcp_connection_ms"`
	TLSHandshakeMS     float64 `json:"tls_handshake_ms"`
	ServerProcessingMS float64 `json:"server_processing_ms"`
	TTFBMS             float64 `json:"ttfb_ms"`
	TotalRoundTripMS   float64 `json:"total_roundtrip_ms"`
	Timestamp          string  `json:"timestamp"`
	Error              *string `json:"error"`
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
	if len(os.Args) < 2 || len(os.Args) > 3 {
		fmt.Println("Usage: probe <url> [interval]")
		os.Exit(1)
	}
	rawURL := os.Args[1]
	validatedURL, err := validateProbeURL(rawURL)
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
	interval := DEFAULT_PROBE_INTERVAL
	if len(os.Args) == 3 {
		interval, err = time.ParseDuration(os.Args[2])
		if err != nil {
			fmt.Println("Error: ", err)
			os.Exit(1)
		}
		if interval <= 0 {
			fmt.Println("Error: interval must be greater than 0")
			os.Exit(1)
		}
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		probe(validatedURL)
		<-ticker.C
	}
}

func probe(targetURL *url.URL) {
	timings := &probeTimings{start: time.Now()}

	trace := newProbeTrace(timings)

	req, err := newProbeRequest(targetURL, trace)
	if err != nil {
		writeProbeResult(buildProbeResult(targetURL, timings, nil, err))
		return
	}

	resp, err := executeProbeRequest(req, timings)

	writeProbeResult(buildProbeResult(targetURL, timings, resp, err))
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

func buildProbeResult(targetURL *url.URL, timings *probeTimings, resp *http.Response, err error) probeResult {
	var errorPtr *string
	if err != nil {
		errStr := err.Error()
		errorPtr = &errStr
	}

	statusCode := 0
	if resp != nil {
		statusCode = resp.StatusCode
	}

	return probeResult{
		TargetURL:          targetURL.String(),
		StatusCode:         statusCode,
		DNSLookupMS:        getDurationMS(timings.dnsStart, timings.dnsDone),
		TCPConnectionMS:    getDurationMS(timings.connectStart, timings.connectDone),
		TLSHandshakeMS:     getDurationMS(timings.tlsStart, timings.tlsDone),
		ServerProcessingMS: getDurationMS(timings.wroteRequest, timings.firstByte),
		Timestamp:          time.Now().UTC().Format(time.RFC3339),
		TotalRoundTripMS:   timings.roundTrip,
		TTFBMS:             getDurationMS(timings.start, timings.firstByte),
		Error:              errorPtr,
	}
}

func writeProbeResult(result probeResult) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(result)
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
