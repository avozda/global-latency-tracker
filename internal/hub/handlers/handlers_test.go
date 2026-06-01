package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/avozda/global-latency-tracker/internal/probe"
	"github.com/go-chi/chi"
)

type mockDB struct {
	pingErr      error
	insertResult probe.Record
	insertErr    error
	getResult    probe.Record
	getErr       error
	listResult   []probe.Record
	listErr      error

	lastLimit     int
	lastOffset    int
	lastTargetURL string
}

func (m *mockDB) Close() error { return nil }

func (m *mockDB) Ping(context.Context) error { return m.pingErr }

func (m *mockDB) InsertProbeResult(probe.Result) (probe.Record, error) {
	return m.insertResult, m.insertErr
}

func (m *mockDB) GetProbeResult(int64) (probe.Record, error) {
	return m.getResult, m.getErr
}

func (m *mockDB) GetProbeResults(limit int, offset int, targetURL string) ([]probe.Record, error) {
	m.lastLimit = limit
	m.lastOffset = offset
	m.lastTargetURL = targetURL
	return m.listResult, m.listErr
}

func (m *mockDB) DeleteProbeResultsBefore(context.Context, time.Time) (int64, error) {
	return 0, nil
}

const testAPIKey = "test-key"

func newTestServer(t *testing.T, db *mockDB) *httptest.Server {
	t.Helper()
	t.Setenv("API_KEY", testAPIKey)

	r := chi.NewRouter()
	RegisterRoutes(r, db)
	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)
	return srv
}

func doRequest(t *testing.T, method, url string, body string) *http.Response {
	t.Helper()
	var reader *strings.Reader
	if body != "" {
		reader = strings.NewReader(body)
	} else {
		reader = strings.NewReader("")
	}
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("X-API-Key", testAPIKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	return resp
}

func TestHealthz(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		srv := newTestServer(t, &mockDB{})
		resp, err := http.Get(srv.URL + "/healthz")
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("db down", func(t *testing.T) {
		srv := newTestServer(t, &mockDB{pingErr: errors.New("down")})
		resp, err := http.Get(srv.URL + "/healthz")
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusServiceUnavailable {
			t.Fatalf("expected 503, got %d", resp.StatusCode)
		}
	})
}

func TestGetMetrics(t *testing.T) {
	t.Run("success with defaults", func(t *testing.T) {
		db := &mockDB{listResult: []probe.Record{{ID: 1, Region: "US-West"}}}
		srv := newTestServer(t, db)

		resp := doRequest(t, http.MethodGet, srv.URL+"/api/metrics", "")
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
		if db.lastLimit != 50 || db.lastOffset != 0 {
			t.Fatalf("expected default limit 50/offset 0, got %d/%d", db.lastLimit, db.lastOffset)
		}

		var records []probe.Record
		if err := json.NewDecoder(resp.Body).Decode(&records); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(records) != 1 {
			t.Fatalf("expected 1 record, got %d", len(records))
		}
	})

	t.Run("custom limit and offset", func(t *testing.T) {
		db := &mockDB{}
		srv := newTestServer(t, db)
		resp := doRequest(t, http.MethodGet, srv.URL+"/api/metrics?limit=10&offset=5", "")
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
		if db.lastLimit != 10 || db.lastOffset != 5 {
			t.Fatalf("expected limit 10/offset 5, got %d/%d", db.lastLimit, db.lastOffset)
		}
	})

	t.Run("invalid limit", func(t *testing.T) {
		srv := newTestServer(t, &mockDB{})
		resp := doRequest(t, http.MethodGet, srv.URL+"/api/metrics?limit=0", "")
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid offset", func(t *testing.T) {
		srv := newTestServer(t, &mockDB{})
		resp := doRequest(t, http.MethodGet, srv.URL+"/api/metrics?offset=-1", "")
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid target_url", func(t *testing.T) {
		srv := newTestServer(t, &mockDB{})
		resp := doRequest(t, http.MethodGet, srv.URL+"/api/metrics?target_url=not-a-url", "")
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", resp.StatusCode)
		}
	})

	t.Run("db error", func(t *testing.T) {
		srv := newTestServer(t, &mockDB{listErr: errors.New("boom")})
		resp := doRequest(t, http.MethodGet, srv.URL+"/api/metrics", "")
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestGetMetric(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		srv := newTestServer(t, &mockDB{getResult: probe.Record{ID: 7, Region: "Europe"}})
		resp := doRequest(t, http.MethodGet, srv.URL+"/api/metrics/7", "")
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		srv := newTestServer(t, &mockDB{})
		resp := doRequest(t, http.MethodGet, srv.URL+"/api/metrics/abc", "")
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		srv := newTestServer(t, &mockDB{getErr: sql.ErrNoRows})
		resp := doRequest(t, http.MethodGet, srv.URL+"/api/metrics/99", "")
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("db error", func(t *testing.T) {
		srv := newTestServer(t, &mockDB{getErr: errors.New("boom")})
		resp := doRequest(t, http.MethodGet, srv.URL+"/api/metrics/1", "")
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestPostMetrics(t *testing.T) {
	validBody := `{"target_url":"https://example.com","region":"US-West","status_code":200,"measured_at":"2026-06-01T09:00:00Z"}`

	t.Run("success", func(t *testing.T) {
		db := &mockDB{insertResult: probe.Record{ID: 1, TargetURL: "https://example.com"}}
		srv := newTestServer(t, db)
		resp := doRequest(t, http.MethodPost, srv.URL+"/api/metrics", validBody)
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		srv := newTestServer(t, &mockDB{})
		resp := doRequest(t, http.MethodPost, srv.URL+"/api/metrics", "{not json")
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", resp.StatusCode)
		}
	})

	t.Run("validation failure", func(t *testing.T) {
		srv := newTestServer(t, &mockDB{})
		resp := doRequest(t, http.MethodPost, srv.URL+"/api/metrics", `{"target_url":"","region":"US-West"}`)
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", resp.StatusCode)
		}
	})

	t.Run("insert error", func(t *testing.T) {
		srv := newTestServer(t, &mockDB{insertErr: errors.New("boom")})
		resp := doRequest(t, http.MethodPost, srv.URL+"/api/metrics", validBody)
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestAuthRequired(t *testing.T) {
	srv := newTestServer(t, &mockDB{})
	resp, err := http.Get(srv.URL + "/api/metrics")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 without api key, got %d", resp.StatusCode)
	}
}
