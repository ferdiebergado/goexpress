package goexpress_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ferdiebergado/goexpress"
)

const (
	errBodyFmt   = "expected body %q, got %q"
	errStatusFmt = "expected status %d, got %d"
)

func TestStripTrailingSlashes(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/trailing-slash/", nil)
	rec := httptest.NewRecorder()

	goexpress.StripTrailingSlashes(handler).ServeHTTP(rec, req)
	if rec.Code != http.StatusMovedPermanently {
		t.Errorf(errStatusFmt, http.StatusMovedPermanently, rec.Code)
	}
}

func TestPanicRecovery(t *testing.T) {
	handler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		panic("test panic")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()

	goexpress.RecoverFromPanic(handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf(errStatusFmt, http.StatusInternalServerError, rec.Code)
	}
}
