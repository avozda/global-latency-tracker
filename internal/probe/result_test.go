package probe

import (
	"math"
	"testing"
	"time"
)

func TestValidateTargetURL(t *testing.T) {
	tests := []struct {
		name      string
		targetURL string
		wantErr   bool
	}{
		{name: "valid https", targetURL: "https://example.com", wantErr: false},
		{name: "valid http with path", targetURL: "http://example.com/health", wantErr: false},
		{name: "trims whitespace", targetURL: "  https://example.com  ", wantErr: false},
		{name: "empty", targetURL: "", wantErr: true},
		{name: "whitespace only", targetURL: "   ", wantErr: true},
		{name: "unsupported scheme", targetURL: "ftp://example.com", wantErr: true},
		{name: "missing scheme", targetURL: "example.com", wantErr: true},
		{name: "missing host", targetURL: "https://", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTargetURL(tt.targetURL)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error for %q, got nil", tt.targetURL)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error for %q: %v", tt.targetURL, err)
			}
		})
	}
}

func validResult() Result {
	return Result{
		TargetURL:          "https://example.com",
		Region:             "US-West",
		StatusCode:         200,
		DNSLookupMS:        1,
		TCPConnectionMS:    2,
		TLSHandshakeMS:     3,
		ServerProcessingMS: 4,
		TTFBMS:             5,
		TotalRoundTripMS:   6,
		MeasuredAt:         time.Now().UTC(),
	}
}

func TestResultValidate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(r *Result)
		wantErr bool
	}{
		{name: "valid", mutate: func(*Result) {}, wantErr: false},
		{name: "status code zero allowed", mutate: func(r *Result) { r.StatusCode = 0 }, wantErr: false},
		{name: "invalid target url", mutate: func(r *Result) { r.TargetURL = "not-a-url" }, wantErr: true},
		{name: "empty region", mutate: func(r *Result) { r.Region = "  " }, wantErr: true},
		{name: "zero measured_at", mutate: func(r *Result) { r.MeasuredAt = time.Time{} }, wantErr: true},
		{name: "status code too low", mutate: func(r *Result) { r.StatusCode = 99 }, wantErr: true},
		{name: "status code too high", mutate: func(r *Result) { r.StatusCode = 600 }, wantErr: true},
		{name: "negative timing", mutate: func(r *Result) { r.DNSLookupMS = -1 }, wantErr: true},
		{name: "NaN timing", mutate: func(r *Result) { r.TTFBMS = math.NaN() }, wantErr: true},
		{name: "Inf timing", mutate: func(r *Result) { r.TotalRoundTripMS = math.Inf(1) }, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := validResult()
			tt.mutate(&r)
			err := r.Validate()
			if tt.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
