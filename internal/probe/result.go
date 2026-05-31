package probe

import (
	"errors"
	"math"
	"net/url"
	"strings"
	"time"
)

type Result struct {
	TargetURL          string    `json:"target_url"`
	StatusCode         int       `json:"status_code"`
	DNSLookupMS        float64   `json:"dns_lookup_ms"`
	TCPConnectionMS    float64   `json:"tcp_connection_ms"`
	TLSHandshakeMS     float64   `json:"tls_handshake_ms"`
	ServerProcessingMS float64   `json:"server_processing_ms"`
	TTFBMS             float64   `json:"ttfb_ms"`
	TotalRoundTripMS   float64   `json:"total_roundtrip_ms"`
	MeasuredAt         time.Time `json:"measured_at"`
	Error              *string   `json:"error"`
}

type Record struct {
	ID                 int64     `json:"id"`
	TargetURL          string    `json:"target_url"`
	StatusCode         int       `json:"status_code"`
	DNSLookupMS        float64   `json:"dns_lookup_ms"`
	TCPConnectionMS    float64   `json:"tcp_connection_ms"`
	TLSHandshakeMS     float64   `json:"tls_handshake_ms"`
	ServerProcessingMS float64   `json:"server_processing_ms"`
	TTFBMS             float64   `json:"ttfb_ms"`
	TotalRoundTripMS   float64   `json:"total_roundtrip_ms"`
	MeasuredAt         time.Time `json:"measured_at"`
	Error              *string   `json:"error"`
	CreatedAt          time.Time `json:"created_at"`
}

func (r Result) Validate() error {
	targetURL := strings.TrimSpace(r.TargetURL)
	if targetURL == "" {
		return errors.New("target_url is required")
	}

	u, err := url.Parse(targetURL)
	if err != nil {
		return errors.New("target_url is invalid")
	}
	switch strings.ToLower(u.Scheme) {
	case "http", "https":
	default:
		return errors.New("target_url must use http or https")
	}
	if u.Host == "" {
		return errors.New("target_url must include a host")
	}

	if r.MeasuredAt.IsZero() {
		return errors.New("measured_at is required")
	}

	if r.StatusCode != 0 && (r.StatusCode < 100 || r.StatusCode > 599) {
		return errors.New("status_code must be 0 or a valid HTTP status code")
	}

	for _, ms := range []float64{
		r.DNSLookupMS,
		r.TCPConnectionMS,
		r.TLSHandshakeMS,
		r.ServerProcessingMS,
		r.TTFBMS,
		r.TotalRoundTripMS,
	} {
		if math.IsNaN(ms) || math.IsInf(ms, 0) || ms < 0 {
			return errors.New("timing values must be non-negative finite numbers")
		}
	}

	return nil
}
