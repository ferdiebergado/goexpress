package middleware

import (
	"log"
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
	status     int
	headerSent bool
}

// WriteHeader sets the HTTP status code for the response and records it in the statusWriter.
// This allows middleware to track which status code was sent to the client.
func (w *statusWriter) WriteHeader(statusCode int) {
	if !w.headerSent { // check if header has already been sent
		w.status = statusCode
		w.headerSent = true // mark the header as sent
		w.ResponseWriter.WriteHeader(statusCode)
	}
}

// LogRequest logs each incoming HTTP request including the method, URL, protocol,
// status code, status text, and duration of the request. It wraps the handler to log this information.
func LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		duration := time.Since(start)
		statusCode := sw.status
		log.Printf("%s %s %s %d %s %s", r.Method, r.URL.Path, r.Proto, statusCode, http.StatusText(statusCode), duration)
	})
}

// StripTrailingSlashes is middleware that removes any trailing slashes from the URL path
// (except for the root path "/"). If a trailing slash is found, it redirects the request to the
// URL without the trailing slash using a 301 (Moved Permanently) status code.
func StripTrailingSlashes(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && strings.HasSuffix(r.URL.Path, "/") {
			// Remove the trailing slash and redirect to the new URL.
			log.Println("Removing trailing slash and redirecting...")
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
				log.Printf("Internal error: %v", err)
				log.Println(string(debug.Stack()))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
