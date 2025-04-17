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

func TestStatusWriterWithDefaultStatus(t *testing.T) {
	const expectedBody = "hello"
	const expectedStatus = http.StatusOK

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write([]byte(expectedBody))
		if err != nil {
			t.Fatal("failed to write response:", err)
		}
	})

	loggedHandler := goexpress.LogRequest(handler)
	loggedHandler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	if recorder.Code != expectedStatus {
		t.Errorf(errStatusFmt, expectedStatus, recorder.Code)
	}

	if recorder.Body.String() != expectedBody {
		t.Errorf(errBodyFmt, expectedBody, recorder.Body.String())
	}
}

func TestStatusWriterWithExplicitStatusOK(t *testing.T) {
	const expectedBody = "hello"
	const expectedStatus = http.StatusOK

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(expectedStatus)
		_, err := w.Write([]byte(expectedBody))
		if err != nil {
			t.Fatal("failed to write response:", err)
		}
	})

	loggedHandler := goexpress.LogRequest(handler)
	loggedHandler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	if recorder.Code != expectedStatus {
		t.Errorf(errStatusFmt, expectedStatus, recorder.Code)
	}

	if recorder.Body.String() != expectedBody {
		t.Errorf(errBodyFmt, expectedBody, recorder.Body.String())
	}
}

func TestStatusWriterWithHTTPError(t *testing.T) {
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "Not Found", http.StatusNotFound)
	})

	loggedHandler := goexpress.LogRequest(handler)
	loggedHandler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	if recorder.Code != http.StatusNotFound {
		t.Errorf(errStatusFmt, http.StatusNotFound, recorder.Code)
	}

	const expectedBody = "Not Found\n"
	if recorder.Body.String() != expectedBody {
		t.Errorf(errBodyFmt, expectedBody, recorder.Body.String())
	}
}
