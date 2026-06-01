package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func decodeError(t *testing.T, rec *httptest.ResponseRecorder) Error {
	t.Helper()
	var got Error
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	return got
}

func TestRequestErrorHandler(t *testing.T) {
	rec := httptest.NewRecorder()
	RequestErrorHandler(rec, errors.New("bad input"))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected json content type, got %q", ct)
	}
	got := decodeError(t, rec)
	if got.Code != http.StatusBadRequest || got.Message != "bad input" {
		t.Fatalf("unexpected body: %+v", got)
	}
}

func TestInternalErrorHandler(t *testing.T) {
	rec := httptest.NewRecorder()
	InternalErrorHandler(rec)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
	got := decodeError(t, rec)
	if got.Code != http.StatusInternalServerError || got.Message != "Internal server error" {
		t.Fatalf("unexpected body: %+v", got)
	}
}

func TestNotFoundErrorHandler(t *testing.T) {
	rec := httptest.NewRecorder()
	NotFoundErrorHandler(rec)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
	got := decodeError(t, rec)
	if got.Code != http.StatusNotFound || got.Message != "Not found" {
		t.Fatalf("unexpected body: %+v", got)
	}
}
