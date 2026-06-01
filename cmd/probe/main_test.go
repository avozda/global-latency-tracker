package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	proberesult "github.com/avozda/global-latency-tracker/internal/probe"
)

func TestValidateProbeURL(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantErr bool
	}{
		{name: "valid https", raw: "https://example.com", wantErr: false},
		{name: "trims whitespace", raw: "  http://example.com  ", wantErr: false},
		{name: "empty", raw: "", wantErr: true},
		{name: "unsupported scheme", raw: "ftp://example.com", wantErr: true},
		{name: "missing host", raw: "https://", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := validateProbeURL(tt.raw)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q", tt.raw)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if u == nil {
				t.Fatal("expected non-nil url")
			}
		})
	}
}

func TestValidateHubURL(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantErr bool
	}{
		{name: "valid", raw: "http://localhost:8080", wantErr: false},
		{name: "unsupported scheme", raw: "ftp://localhost", wantErr: true},
		{name: "missing host", raw: "http://", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateHubURL(tt.raw)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error for %q", tt.raw)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestGetDurationMS(t *testing.T) {
	start := time.Unix(0, 0)
	end := start.Add(1500 * time.Microsecond)

	if got := getDurationMS(start, end); got != 1.5 {
		t.Fatalf("expected 1.5, got %v", got)
	}
	if got := getDurationMS(time.Time{}, end); got != 0 {
		t.Fatalf("expected 0 for zero start, got %v", got)
	}
	if got := getDurationMS(start, time.Time{}); got != 0 {
		t.Fatalf("expected 0 for zero end, got %v", got)
	}
}

func TestRoundToDecimal(t *testing.T) {
	if got := roundToDecimal(1.23456, 2); got != 1.23 {
		t.Fatalf("expected 1.23, got %v", got)
	}
	if got := roundToDecimal(1.236, 2); got != 1.24 {
		t.Fatalf("expected 1.24, got %v", got)
	}
	if got := roundToDecimal(12.5, 0); got != 13 {
		t.Fatalf("expected 13, got %v", got)
	}
}

func TestLoadProbeConfig(t *testing.T) {
	base := map[string]string{
		"PROBE_TARGET_URL": "https://example.com",
		"PROBE_REGION":     "US-West",
		"PROBE_INTERVAL":   "30s",
		"HUB_URL":          "http://localhost:8080",
		"API_KEY":          "secret",
	}

	t.Run("valid", func(t *testing.T) {
		setEnv(t, base)
		cfg, err := loadProbeConfig()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.region != "US-West" {
			t.Fatalf("expected region US-West, got %q", cfg.region)
		}
		if cfg.interval != 30*time.Second {
			t.Fatalf("expected 30s interval, got %v", cfg.interval)
		}
	})

	t.Run("default interval", func(t *testing.T) {
		env := cloneEnv(base)
		delete(env, "PROBE_INTERVAL")
		setEnv(t, env)
		cfg, err := loadProbeConfig()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.interval != DEFAULT_PROBE_INTERVAL {
			t.Fatalf("expected default interval, got %v", cfg.interval)
		}
	})

	missingCases := []string{
		"PROBE_TARGET_URL",
		"PROBE_REGION",
		"HUB_URL",
		"API_KEY",
	}
	for _, key := range missingCases {
		t.Run("missing "+key, func(t *testing.T) {
			env := cloneEnv(base)
			delete(env, key)
			setEnv(t, env)
			if _, err := loadProbeConfig(); err == nil {
				t.Fatalf("expected error when %s is missing", key)
			}
		})
	}

	t.Run("invalid interval", func(t *testing.T) {
		env := cloneEnv(base)
		env["PROBE_INTERVAL"] = "nonsense"
		setEnv(t, env)
		if _, err := loadProbeConfig(); err == nil {
			t.Fatal("expected error for invalid interval")
		}
	})

	t.Run("non-positive interval", func(t *testing.T) {
		env := cloneEnv(base)
		env["PROBE_INTERVAL"] = "0s"
		setEnv(t, env)
		if _, err := loadProbeConfig(); err == nil {
			t.Fatal("expected error for non-positive interval")
		}
	})
}

func TestBuildProbeResult(t *testing.T) {
	target, _ := url.Parse("https://example.com")
	timings := &probeTimings{roundTrip: 12.5}

	resp := &http.Response{StatusCode: 200}
	result := buildProbeResult(target, "US-West", timings, resp, nil)

	if result.TargetURL != "https://example.com" {
		t.Fatalf("unexpected target url: %q", result.TargetURL)
	}
	if result.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d", result.StatusCode)
	}
	if result.TotalRoundTripMS != 12.5 {
		t.Fatalf("expected round trip 12.5, got %v", result.TotalRoundTripMS)
	}
	if result.Error != nil {
		t.Fatalf("expected nil error, got %v", *result.Error)
	}
	if result.MeasuredAt.IsZero() {
		t.Fatal("expected measured_at to be set")
	}
}

func TestBuildProbeResultWithError(t *testing.T) {
	target, _ := url.Parse("https://example.com")
	timings := &probeTimings{}

	result := buildProbeResult(target, "US-West", timings, nil, io.ErrUnexpectedEOF)

	if result.StatusCode != 0 {
		t.Fatalf("expected status 0, got %d", result.StatusCode)
	}
	if result.Error == nil {
		t.Fatal("expected error to be set")
	}
}

func TestPostProbeResult(t *testing.T) {
	const apiKey = "secret"

	t.Run("success", func(t *testing.T) {
		var gotKey, gotPath string
		var gotResult proberesult.Result
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotKey = r.Header.Get("X-API-Key")
			gotPath = r.URL.Path
			_ = json.NewDecoder(r.Body).Decode(&gotResult)
			w.WriteHeader(http.StatusCreated)
		}))
		defer srv.Close()

		result := proberesult.Result{TargetURL: "https://example.com", Region: "US-West", MeasuredAt: time.Now().UTC()}
		if err := postProbeResult(srv.URL, apiKey, result); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotKey != apiKey {
			t.Fatalf("expected api key %q, got %q", apiKey, gotKey)
		}
		if gotPath != "/api/metrics" {
			t.Fatalf("expected path /api/metrics, got %q", gotPath)
		}
		if gotResult.Region != "US-West" {
			t.Fatalf("expected region US-West, got %q", gotResult.Region)
		}
	})

	t.Run("non-2xx returns error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer srv.Close()

		err := postProbeResult(srv.URL, apiKey, proberesult.Result{})
		if err == nil {
			t.Fatal("expected error for non-2xx response")
		}
	})
}

func cloneEnv(env map[string]string) map[string]string {
	out := make(map[string]string, len(env))
	for k, v := range env {
		out[k] = v
	}
	return out
}

func setEnv(t *testing.T, env map[string]string) {
	t.Helper()
	keys := []string{"PROBE_TARGET_URL", "PROBE_REGION", "PROBE_INTERVAL", "HUB_URL", "API_KEY"}
	for _, k := range keys {
		if v, ok := env[k]; ok {
			t.Setenv(k, v)
		} else {
			t.Setenv(k, "")
		}
	}
}
