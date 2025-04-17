package goexpress

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
)

// statusWriter is a wrapper around http.ResponseWriter that tracks the status code
// written to the response. This is useful for logging or middleware that needs to
// inspect the status code after a request is handled.
type statusWriter struct {
	http.ResponseWriter
	statusCode    int
	headerWritten bool
}

// WriteHeader sets the HTTP status code for the response and records it in the statusWriter.
// This allows middleware to track which status code was written to the client.
func (w *statusWriter) WriteHeader(statusCode int) {
	if !w.headerWritten {
		w.statusCode = statusCode
		w.headerWritten = true
	}

	w.ResponseWriter.WriteHeader(statusCode)
}

// Override Write to implicitly call WriteHeader(200) if needed
func (w *statusWriter) Write(b []byte) (int, error) {
	if !w.headerWritten {
		w.WriteHeader(http.StatusOK)
	}

	return w.ResponseWriter.Write(b)
}

// LogRequest logs each incoming HTTP request including the method, URL, protocol,
// status code, status text, and duration of the request. It wraps the handler to log this information.
func LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(sw, r)

		body, err := parseRequestBody(r)
		if err != nil {
			slog.Error("failed to parse the request body", "reason", err)
		}

		duration := time.Since(start)
		slog.Info("New Request",
			"user_agent", r.UserAgent(),
			"remote_address", getIPAddress(r),
			"method", r.Method,
			"path", r.URL.Path,
			"proto", r.Proto,
			slog.Any("headers", r.Header),
			"body", string(body),
			slog.Int("status_code", sw.statusCode),
			slog.Duration("duration", duration),
		)
	})
}

// StripTrailingSlashes is middleware that removes any trailing slashes from the URL path
// (except for the root path "/"). If a trailing slash is found, it redirects the request to the
// URL without the trailing slash using a 301 (Moved Permanently) status code.
func StripTrailingSlashes(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && strings.HasSuffix(r.URL.Path, "/") {
			// Remove the trailing slash and redirect to the new URL.
			slog.Info("Removing trailing slash and redirecting...")
			http.Redirect(w, r, strings.TrimSuffix(r.URL.Path, "/"), http.StatusMovedPermanently)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RecoverFromPanic is middleware that recovers from panics that occur during the execution
// of the handler. If a panic is detected, it logs the error and stack trace, and returns
// a 500 (Internal Server Error) response to the client.
func RecoverFromPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic occurred",
					"reason", err,
					"stack_trace", string(debug.Stack()),
				)
				const status = http.StatusInternalServerError
				http.Error(w, http.StatusText(status), status)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// getIPAddress extracts the client's IP address from the request.
func getIPAddress(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	if forwardedFor := r.Header.Values("X-Forwarded-For"); len(forwardedFor) > 0 {
		firstIP := forwardedFor[0]
		ips := strings.Split(firstIP, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// ip, _, err := net.SplitHostPort(r.RemoteAddr)
	// if err != nil {
	// 	return r.RemoteAddr
	// }
	return r.RemoteAddr
}

func parseRequestBody(req *http.Request) ([]byte, error) {
	var body []byte
	if req.Body != nil {
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			return body, fmt.Errorf("error reading request body: %w", err)
		}
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		body = bodyBytes
	}

	return body, nil
}
