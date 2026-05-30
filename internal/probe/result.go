package probe

import "time"

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
