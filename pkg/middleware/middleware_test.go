package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestLogger(t *testing.T) {
	// Set up a dummy handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	// Wrap the handler with the RequestLogger middleware
	RequestLogger(handler).ServeHTTP(rec, req)

	// Check if the status code is still OK
	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestStripTrailingSlashes(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/trailing-slash/", nil)
	rec := httptest.NewRecorder()

	// Wrap the handler with the StripTrailingSlashes middleware
	StripTrailingSlashes(handler).ServeHTTP(rec, req)

	// Check if it redirects without the trailing slash
	if rec.Code != http.StatusMovedPermanently {
		t.Errorf("expected status %d, got %d", http.StatusMovedPermanently, rec.Code)
	}
}

func TestPanicRecovery(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()

	// Wrap the handler with the PanicRecovery middleware
	PanicRecovery(handler).ServeHTTP(rec, req)

	// Check if it returns a 500 status code
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}
