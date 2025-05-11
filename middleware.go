package goexpress

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"runtime/debug"
	"strings"
)

const MimeJSON = "application/json"

// LogRequest logs each incoming HTTP request including the method, URL, protocol,
// status code, status text, and duration of the request. It wraps the handler to log this information.
func LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("New Request",
			"user_agent", r.UserAgent(),
			"remote_address", getIPAddress(r),
			"method", r.Method,
			"path", r.URL.Path,
			"proto", r.Proto,
			slog.Any("headers", r.Header),
		)
		next.ServeHTTP(w, r)
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

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

func ReadRequestBody(req *http.Request) ([]byte, error) {
	var body []byte
	if req.Body != nil {
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			req.Body = io.NopCloser(bytes.NewBuffer(nil))
			return nil, fmt.Errorf("error reading request body: %w", err)
		}
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		body = bodyBytes
	}

	return body, nil
}

func Mask(data []byte, fieldsToMask []string) []byte {
	var dataMap map[string]any
	err := json.Unmarshal(data, &dataMap)
	if err != nil {
		slog.Error("failed to unmarshal input", "reason", err)
		return data
	}

	for k := range dataMap {
		for _, f := range fieldsToMask {
			if k == f {
				dataMap[k] = "*"
			}
		}
	}

	bytes, err := json.Marshal(dataMap)
	if err != nil {
		slog.Error("failed to marshal data", "reason", err)
		return data
	}

	return bytes
}
