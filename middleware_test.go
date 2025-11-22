package goexpress_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ferdiebergado/goexpress"
)

const errStatusFmt = "expected status %d, got %d"

func TestPanicRecovery(t *testing.T) {
	handler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		panic("test panic")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", http.NoBody)
	rec := httptest.NewRecorder()

	goexpress.RecoverPanic(handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf(errStatusFmt, http.StatusInternalServerError, rec.Code)
	}
}
