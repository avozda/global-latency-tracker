package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthenticate(t *testing.T) {
	tests := []struct {
		name       string
		envKey     string
		headerKey  string
		wantStatus int
		wantNext   bool
	}{
		{name: "missing env key", envKey: "", headerKey: "anything", wantStatus: http.StatusBadRequest, wantNext: false},
		{name: "wrong header key", envKey: "secret", headerKey: "nope", wantStatus: http.StatusBadRequest, wantNext: false},
		{name: "missing header key", envKey: "secret", headerKey: "", wantStatus: http.StatusBadRequest, wantNext: false},
		{name: "valid key", envKey: "secret", headerKey: "secret", wantStatus: http.StatusOK, wantNext: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("API_KEY", tt.envKey)

			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/api/metrics", nil)
			if tt.headerKey != "" {
				req.Header.Set("X-API-Key", tt.headerKey)
			}
			rec := httptest.NewRecorder()

			Authenticate(next).ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}
			if nextCalled != tt.wantNext {
				t.Fatalf("expected next called=%v, got %v", tt.wantNext, nextCalled)
			}
		})
	}
}
