package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ferdiebergado/go-express/middleware"
)

func TestRequestLogger(t *testing.T) {
	// Set up a dummy handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	// Wrap the handler with the RequestLogger middleware
	middleware.RequestLogger(handler).ServeHTTP(rec, req)

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
	middleware.StripTrailingSlashes(handler).ServeHTTP(rec, req)

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
	middleware.PanicRecovery(handler).ServeHTTP(rec, req)

	// Check if it returns a 500 status code
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

// TestStatusWriterWithHTTPError tests the behavior of statusWriter when an HTTP handler calls http.Error.
func TestStatusWriterWithHTTPError(t *testing.T) {
	// Create a response recorder to capture the response
	recorder := httptest.NewRecorder()

	// Define a test handler that calls http.Error
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Found", http.StatusNotFound)
	})

	// Wrap the handler with the RequestLogger middleware
	loggedHandler := middleware.RequestLogger(handler)

	// Serve the HTTP request using the logged handler
	loggedHandler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	// Verify that the recorder has the correct status code
	if recorder.Code != http.StatusNotFound {
		t.Errorf("expected recorder status code %d, got %d", http.StatusNotFound, recorder.Code)
	}

	// Verify that the response body contains the expected error message
	expectedBody := "Not Found\n" // http.Error appends a newline to the error message
	if recorder.Body.String() != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, recorder.Body.String())
	}
}
