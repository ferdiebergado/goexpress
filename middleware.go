package goexpress

import (
	"log/slog"
	"net"
	"net/http"
	"runtime/debug"
	"strings"
)

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

// RecoverPanic is middleware that recovers from panics that occur during the execution
// of the handler. If a panic is detected, it logs the error and stack trace, and returns
// a 500 (Internal Server Error) response to the client.
func RecoverPanic(next http.Handler) http.Handler {
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
