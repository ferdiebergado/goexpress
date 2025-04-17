package goexpress_test

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ferdiebergado/goexpress"
)

type mockHandler struct {
	status int
	body   string
}

func (m *mockHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(m.status)
	_, err := w.Write([]byte(m.body))
	if err != nil {
		slog.Error("failed to write to the request body", "reason", err)
	}
}

// logCapture implements slog.Handler to capture log entries for assertions.
type logCapture struct {
	entries []map[string]any
}

func (l *logCapture) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

func (l *logCapture) Handle(_ context.Context, r slog.Record) error {
	entry := make(map[string]any)
	r.Attrs(func(a slog.Attr) bool {
		entry[a.Key] = a.Value.Any()
		return true
	})
	l.entries = append(l.entries, entry)
	return nil
}

func (l *logCapture) WithAttrs(_ []slog.Attr) slog.Handler {
	return l
}

func (l *logCapture) WithGroup(_ string) slog.Handler {
	return l
}

type testcase struct {
	name           string
	method         string
	path           string
	body           string
	headers        map[string]string
	handlerStatus  int
	parseBodyError bool
	userAgent      string
	remoteAddr     string
}

func TestLogRequest(t *testing.T) {
	testCases := []testcase{
		{
			name:          "GET with 200 OK",
			method:        http.MethodGet,
			path:          "/test",
			handlerStatus: http.StatusOK,
			userAgent:     "TestAgent/1.0",
			remoteAddr:    "192.0.2.1:1234",
		},
		{
			name:          "POST with 201 Created and JSON body",
			method:        http.MethodPost,
			path:          "/create",
			body:          `{"key":"value"}`,
			handlerStatus: http.StatusCreated,
			userAgent:     "PostmanRuntime/7.28.4",
			remoteAddr:    "203.0.113.5:5678",
		},
		{
			name:           "PUT with body parse error",
			method:         http.MethodPut,
			path:           "/update",
			body:           "invalid-json",
			handlerStatus:  http.StatusBadRequest,
			parseBodyError: true,
			userAgent:      "curl/7.64.1",
			remoteAddr:     "198.51.100.10:4321",
		},
		{
			name:          "DELETE with 204 No Content",
			method:        http.MethodDelete,
			path:          "/remove",
			handlerStatus: http.StatusNoContent,
			userAgent:     "Go-http-client/1.1",
			remoteAddr:    "127.0.0.1:8080",
		},
		{
			name:          "Request with custom headers",
			method:        http.MethodGet,
			path:          "/headers",
			handlerStatus: http.StatusOK,
			headers: map[string]string{
				"X-Test":       "true",
				"Content-Type": "application/json",
			},
			userAgent:  "CustomAgent/2.0",
			remoteAddr: "10.0.0.1:9999",
		},
	}

	for _, tc := range testCases {
		runTestCase(t, tc)
	}
}

func setupLogCapture() *logCapture {
	// Setup log capture
	lc := &logCapture{}
	logger := slog.New(lc)
	// Replace package-level slog default logger with our test logger temporarily
	oldLogger := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(oldLogger)

	return lc
}

func runTestCase(t *testing.T, tc testcase) {
	t.Helper()
	t.Run(tc.name, func(t *testing.T) {
		lc := setupLogCapture()
		// Setup mock handler
		handler := &mockHandler{
			status: tc.handlerStatus,
			body:   "response body",
		}

		// Create request with body and headers
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		for k, v := range tc.headers {
			req.Header.Set(k, v)
		}
		req.Header.Set("User-Agent", tc.userAgent)
		req.RemoteAddr = tc.remoteAddr

		// Create response recorder
		rr := httptest.NewRecorder()

		// Wrap handler with middleware
		middleware := goexpress.LogRequest(handler)

		// Call middleware
		middleware.ServeHTTP(rr, req)

		// Check response status code
		if rr.Code != tc.handlerStatus {
			t.Errorf("Expected status %d, got %d", tc.handlerStatus, rr.Code)
		}

		// Check logs captured
		if len(lc.entries) == 0 {
			t.Fatal("No log entries captured")
		}
		logEntry := lc.entries[len(lc.entries)-1]

		// Validate common log fields
		if got := logEntry["method"]; got != tc.method {
			t.Errorf("Logged method = %v; want %v", got, tc.method)
		}
		if got := logEntry["path"]; got != tc.path {
			t.Errorf("Logged path = %v; want %v", got, tc.path)
		}

		loggedStatus, ok := logEntry["status_code"].(int)
		if !ok {
			// Try int64, which is common for structured loggers
			if v, ok64 := logEntry["status_code"].(int64); ok64 {
				loggedStatus = int(v)
			} else {
				t.Fatalf("status_code is not int or int64, got %T", logEntry["status_code"])
			}
		}
		if loggedStatus != tc.handlerStatus {
			t.Errorf("Logged status_code = %v; want %v", loggedStatus, tc.handlerStatus)
		}
		if got := logEntry["user_agent"]; got != tc.userAgent {
			t.Errorf("Logged user_agent = %v; want %v", got, tc.userAgent)
		}
		if got := logEntry["remote_address"]; got != tc.remoteAddr {
			t.Errorf("Logged remote_address = %v; want %v", got, tc.remoteAddr)
		}

		// Validate headers logged
		headers, ok := logEntry["headers"].(http.Header)
		if !ok {
			t.Errorf("Logged headers missing or wrong type")
		} else {
			for k, v := range tc.headers {
				if hv := headers.Get(k); hv != v {
					t.Errorf("Header %s logged = %v; want %v", k, hv, v)
				}
			}
		}

		// Validate body
		bodyLogged, _ := logEntry["body"].(string)

		if !tc.parseBodyError {
			if bodyLogged != tc.body {
				t.Errorf("Logged body = %q; want %q", bodyLogged, tc.body)
			}
		}

		// Validate duration is present and non-zero
		duration, ok := logEntry["duration"].(time.Duration)
		if !ok {
			t.Error("Logged duration missing or wrong type")
		} else if duration <= 0 {
			t.Error("Logged duration should be positive")
		}
	})
}
