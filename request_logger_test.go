package goexpress_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/ferdiebergado/goexpress"
)

type mockHandler struct {
	body string
}

func (m *mockHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
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
	parseBodyError bool
	userAgent      string
	remoteAddr     string
}

func TestLogRequest(t *testing.T) {
	testCases := []testcase{
		{
			name:       "GET with 200 OK",
			method:     http.MethodGet,
			path:       "/test",
			userAgent:  "TestAgent/1.0",
			remoteAddr: "192.0.2.1",
		},
		// {
		// 	name:       "POST with 201 Created and JSON body",
		// 	method:     http.MethodPost,
		// 	path:       "/create",
		// 	body:       `{"key":"value"}`,
		// 	userAgent:  "PostmanRuntime/7.28.4",
		// 	remoteAddr: "203.0.113.5",
		// },
		{
			name:           "PUT with body parse error",
			method:         http.MethodPut,
			path:           "/update",
			body:           "invalid-json",
			parseBodyError: true,
			userAgent:      "curl/7.64.1",
			remoteAddr:     "198.51.100.10",
		},
		{
			name:       "DELETE with 204 No Content",
			method:     http.MethodDelete,
			path:       "/remove",
			userAgent:  "Go-http-client/1.1",
			remoteAddr: "127.0.0.1",
		},
		{
			name:   "Request with custom headers",
			method: http.MethodGet,
			path:   "/headers",
			headers: map[string]string{
				"X-Test":       "true",
				"Content-Type": goexpress.MimeJSON,
			},
			userAgent:  "CustomAgent/2.0",
			remoteAddr: "10.0.0.1",
		},
	}

	for _, tc := range testCases {
		runTestCase(t, tc)
	}
}

func runTestCase(t *testing.T, tc testcase) {
	t.Helper()
	t.Run(tc.name, func(t *testing.T) {
		// Setup log capture
		lc := &logCapture{}
		logger := slog.New(lc)
		// Replace package-level slog default logger with our test logger temporarily
		oldLogger := slog.Default()
		slog.SetDefault(logger)
		defer slog.SetDefault(oldLogger)

		// Setup mock handler
		handler := &mockHandler{
			body: "response body",
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
	})
}

func TestParseRequestBody(t *testing.T) {
	tests := []struct {
		name          string
		requestBody   io.Reader
		contentType   string
		expectedBody  []byte
		expectedError string
	}{
		{
			name:         "Non-empty body",
			requestBody:  bytes.NewBufferString("test request body"),
			expectedBody: []byte("test request body"),
		},
		{
			name:         "Empty body",
			requestBody:  nil,
			expectedBody: []byte{},
		},
		{
			name:          "Error reading body",
			requestBody:   &errorReader{},
			expectedBody:  []byte{},
			expectedError: "error reading request body: test error",
		},
		{
			name:         "JSON body",
			requestBody:  bytes.NewBufferString(`{"key": "value", "number": "123"}`),
			contentType:  goexpress.MimeJSON,
			expectedBody: []byte(`{"key": "value", "number": "123"}`),
		},
		{
			name:         "Empty JSON body",
			requestBody:  bytes.NewBufferString(`{}`),
			contentType:  goexpress.MimeJSON,
			expectedBody: []byte(`{}`),
		},
		{
			name:         "Nil request body (again)",
			requestBody:  nil,
			expectedBody: []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/test", tt.requestBody)
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			actualBody, actualError := goexpress.ReadRequestBody(req)

			if !bytes.Equal(actualBody, tt.expectedBody) {
				t.Errorf("parseRequestBody() returned body = %q, want %q", string(actualBody), string(tt.expectedBody))
			}

			if tt.expectedError != "" {
				if actualError == nil {
					t.Errorf("parseRequestBody() expected error = %q, got nil", tt.expectedError)
				} else if actualError.Error() != tt.expectedError {
					t.Errorf("parseRequestBody() error = %q, want %q", actualError.Error(), tt.expectedError)
				}
			} else if actualError != nil {
				t.Errorf("parseRequestBody() returned unexpected error: %v", actualError)
			}

			if tt.requestBody != nil {
				bodyBytes, errReadAgain := io.ReadAll(req.Body)
				if errReadAgain != nil {
					t.Errorf("error reading request body again: %v", errReadAgain)
				}
				if !bytes.Equal(bodyBytes, tt.expectedBody) {
					t.Errorf("re-read request body = %q, want %q", string(bodyBytes), string(tt.expectedBody))
				}
			}
		})
	}
}

type errorReader struct{}

func (er *errorReader) Read(_ []byte) (n int, err error) {
	return 0, fmt.Errorf("test error")
}

func (er *errorReader) Close() error {
	return nil
}

func TestMask(t *testing.T) {
	tests := []struct {
		name         string
		inputData    []byte
		fieldsToMask []string
		expectedData []byte
	}{
		{
			name:         "empty data and fields",
			inputData:    []byte("{}"),
			fieldsToMask: []string{},
			expectedData: []byte("{}"),
		},
		{
			name:         "no fields to mask",
			inputData:    []byte(`{"name": "John Doe", "age": 30}`),
			fieldsToMask: []string{},
			expectedData: []byte(`{"age":30,"name":"John Doe"}`),
		},
		{
			name:         "single field to mask",
			inputData:    []byte(`{"name": "John Doe", "age": 30}`),
			fieldsToMask: []string{"name"},
			expectedData: []byte(`{"age":30,"name":"*"}`),
		},
		{
			name:         "multiple fields to mask",
			inputData:    []byte(`{"name": "John Doe", "age": 30, "email": "john.doe@example.com"}`),
			fieldsToMask: []string{"name", "email"},
			expectedData: []byte(`{"age":30,"email":"*","name":"*"}`),
		},
		{
			name:         "field to mask not present",
			inputData:    []byte(`{"name": "John Doe", "age": 30}`),
			fieldsToMask: []string{"address"},
			expectedData: []byte(`{"age":30,"name":"John Doe"}`),
		},
		{
			name:         "empty fields to mask slice",
			inputData:    []byte(`{"name": "John Doe"}`),
			fieldsToMask: []string{},
			expectedData: []byte(`{"name":"John Doe"}`),
		},
		{
			name:         "data with different data types",
			inputData:    []byte(`{"name": "John Doe", "age": 30, "isMember": true}`),
			fieldsToMask: []string{"age"},
			expectedData: []byte(`{"age":"*","isMember":true,"name":"John Doe"}`),
		},
		{
			name:         "nested fields - should not mask (only top level)",
			inputData:    []byte(`{"user": {"name": "John Doe", "details": {"age": 30}}}`),
			fieldsToMask: []string{"name", "age"},
			expectedData: []byte(`{"user":{"details":{"age":30},"name":"John Doe"}}`),
		},
		{
			name:         "array in data - should not mask array itself",
			inputData:    []byte(`{"hobbies": ["reading", "hiking"], "name": "John Doe"}`),
			fieldsToMask: []string{"hobbies"},
			expectedData: []byte(`{"hobbies":"*","name":"John Doe"}`),
		},
		{
			name:         "invalid JSON input",
			inputData:    []byte(`{"name": "John Doe",}`),
			fieldsToMask: []string{"name"},
			expectedData: []byte(`{"name": "John Doe",}`),
		},
		{
			name:         "empty JSON input",
			inputData:    []byte(""),
			fieldsToMask: []string{"name"},
			expectedData: []byte(""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualData := goexpress.Mask(tt.inputData, tt.fieldsToMask)

			var actualMap map[string]any
			errActual := json.Unmarshal(actualData, &actualMap)

			var expectedMap map[string]any
			errExpected := json.Unmarshal(tt.expectedData, &expectedMap)

			if errActual != nil || errExpected != nil {
				if !reflect.DeepEqual(actualData, tt.expectedData) {
					t.Errorf("Test Case: %s\nInput: %s\nFields to Mask: %v\nExpected (raw): %s\nActual (raw): %s\nUnmarshal Error (actual): %v\nUnmarshal Error (expected): %v",
						tt.name, tt.inputData, tt.fieldsToMask, tt.expectedData, actualData, errActual, errExpected)
				}
				return
			}

			if !reflect.DeepEqual(actualMap, expectedMap) {
				t.Errorf("Test Case: %s\nInput: %s\nFields to Mask: %v\nExpected: %s\nActual: %s",
					tt.name, tt.inputData, tt.fieldsToMask, tt.expectedData, actualData)
			}
		})
	}
}
