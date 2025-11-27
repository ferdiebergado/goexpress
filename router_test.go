package goexpress_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/ferdiebergado/goexpress"
)

// TestNewRouter ensures that a new Router is created successfully.
func TestNewRouter(t *testing.T) {
	t.Parallel()

	r := goexpress.New()

	if r == nil {
		t.Errorf("goexpress.New() = %v, want: non-nil router", r)
	}
}

func TestHTTPVerbHelpers(t *testing.T) {
	t.Parallel()

	const wantStatus = http.StatusTeapot

	helloHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(wantStatus)
		_, err := w.Write([]byte(r.Method))
		if err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name     string
		method   string
		setup    func(*goexpress.Router)
		wantBody string
	}{
		{
			name:   "Get method",
			method: http.MethodGet,
			setup: func(router *goexpress.Router) {
				router.Get("/hello", http.HandlerFunc(helloHandler))
			},
			wantBody: "GET",
		},
		{
			name:   "Post method",
			method: http.MethodPost,
			setup: func(router *goexpress.Router) {
				router.Post("/hello", http.HandlerFunc(helloHandler))
			},
			wantBody: "POST",
		},
		{
			name:   "Put method",
			method: http.MethodPut,
			setup: func(router *goexpress.Router) {
				router.Put("/hello", http.HandlerFunc(helloHandler))
			},
			wantBody: "PUT",
		},
		{
			name:   "Patch method",
			method: http.MethodPatch,
			setup: func(router *goexpress.Router) {
				router.Patch("/hello", http.HandlerFunc(helloHandler))
			},
			wantBody: "PATCH",
		},
		{
			name:   "Delete method",
			method: http.MethodDelete,
			setup: func(router *goexpress.Router) {
				router.Delete("/hello", http.HandlerFunc(helloHandler))
			},
			wantBody: "DELETE",
		},
		{
			name:   "Options method",
			method: http.MethodOptions,
			setup: func(router *goexpress.Router) {
				router.Options("/hello", http.HandlerFunc(helloHandler))
			},
			wantBody: "OPTIONS",
		},
		{
			name:   "Head method",
			method: http.MethodHead,
			setup: func(router *goexpress.Router) {
				router.Head("/hello", http.HandlerFunc(helloHandler))
			},
			wantBody: "HEAD",
		},
		{
			name:   "Connect method",
			method: http.MethodConnect,
			setup: func(router *goexpress.Router) {
				router.Connect("/hello", http.HandlerFunc(helloHandler))
			},
			wantBody: "CONNECT",
		},
		{
			name:   "Trace method",
			method: http.MethodTrace,
			setup: func(router *goexpress.Router) {
				router.Trace("/hello", http.HandlerFunc(helloHandler))
			},
			wantBody: "TRACE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := goexpress.New()

			tt.setup(r)

			req := httptest.NewRequest(tt.method, "/hello", http.NoBody)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assertStatus(t, rec.Code, wantStatus)
			assertBody(t, rec.Body.String(), tt.wantBody)
		})
	}
}

// TestServeStatic verifies that a request to a static file within a specified directory is correctly handled.
func TestStatic(t *testing.T) {
	t.Parallel()

	const (
		staticFile = "/static/test.html"
		header     = "Content-Type"
		wantHeader = "text/html"
		wantBody   = "<h1>This is a test page</h1>"
	)

	tests := []struct {
		name       string
		setup      func(*goexpress.Router) string
		wantStatus int
		wantHeader string
		wantBody   string
	}{
		{
			name: "directory with absolute path",
			setup: func(r *goexpress.Router) string {
				tmpDir := t.TempDir()
				path := tmpDir + "/test.html"

				if err := os.WriteFile(path, []byte(wantBody), 0o600); err != nil {
					t.Fatalf("failed to write html file: %v", err)
				}

				r.Static("/static", tmpDir)

				return staticFile
			},
			wantStatus: http.StatusOK,
			wantHeader: wantHeader,
			wantBody:   wantBody,
		},
		{
			name: "directory with relative path",
			setup: func(r *goexpress.Router) string {
				r.Static("/static", "static")

				return staticFile
			},
			wantStatus: http.StatusOK,
			wantHeader: wantHeader,
			wantBody:   wantBody,
		},
		{
			name: "relative path with current directory prefix",
			setup: func(r *goexpress.Router) string {
				r.Static("/static", "./static")

				return staticFile
			},
			wantStatus: http.StatusOK,
			wantHeader: wantHeader,
			wantBody:   wantBody,
		},
		{
			name: "prefix with parent directory",
			setup: func(r *goexpress.Router) string {
				r.Static("../static", "./static")

				return staticFile
			},
			wantStatus: http.StatusOK,
			wantHeader: wantHeader,
			wantBody:   wantBody,
		},
		{
			name: "prefix with double slash",
			setup: func(r *goexpress.Router) string {
				r.Static("//static", "./static")

				return staticFile
			},
			wantStatus: http.StatusOK,
			wantHeader: wantHeader,
			wantBody:   wantBody,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := goexpress.New()

			path := tt.setup(r)

			req := httptest.NewRequest(http.MethodGet, path, http.NoBody)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			assertStatus(t, rec.Code, http.StatusOK)

			contentType := rec.Header().Get(header)
			if strings.Split(contentType, ";")[0] != wantHeader {
				t.Errorf("rec.Header().Get(%q) = %q, want: %q", header, contentType, wantHeader)
			}

			assertBody(t, rec.Body.String(), wantBody)
		})
	}
}

// TestNotFound verifies that the custom "Not Found" handler is called for undefined routes.
func TestNotFound(t *testing.T) {
	t.Parallel()

	const (
		header     = "X-Middleware"
		wantHeader = "global"
		wantStatus = http.StatusNotFound
		wantBody   = "Custom 404 - Page Not Found"
	)

	mw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(header, wantHeader)
			next.ServeHTTP(w, r)
		})
	}

	notFoundHandler := func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, wantBody, wantStatus)
	}

	r := goexpress.New()
	r.Use(mw)
	r.NotFound(http.HandlerFunc(notFoundHandler))

	req := httptest.NewRequest(http.MethodGet, "/unknown", http.NoBody)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assertStatus(t, rec.Code, wantStatus)
	assertBody(t, rec.Body.String(), wantBody)
	assertHeader(t, rec, header, wantHeader)
}

func TestRouter(t *testing.T) {
	t.Parallel()

	type ctxKey int
	const (
		mwKey ctxKey = iota + 1
		mw2key
	)

	tests := []struct {
		name       string
		method     string
		path       string
		setup      func(router *goexpress.Router)
		wantStatus int
		wantBody   string
	}{
		{
			name:   "basic route",
			method: "GET",
			path:   "/hello",
			setup: func(router *goexpress.Router) {
				router.Get("/hello", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("Hello, world!"))
				}))
			},
			wantStatus: http.StatusOK,
			wantBody:   "Hello, world!",
		},
		{
			name:       "unregistered route",
			method:     "GET",
			path:       "/notfound",
			wantStatus: http.StatusNotFound,
			wantBody:   "404 page not found",
		},
		{
			name:   "global middleware",
			method: "GET",
			path:   "/",
			setup: func(router *goexpress.Router) {
				globalMw := func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						ctxVal := []string{"X-Middleware1-Called"}
						ctx := context.WithValue(r.Context(), mwKey, ctxVal)
						r = r.WithContext(ctx)
						next.ServeHTTP(w, r)
					})
				}

				router.Use(globalMw)

				router.Get("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					val, ok := r.Context().Value(mwKey).([]string)
					if !ok {
						t.Fatalf("unable to get context value: %v", val)
					}

					w.WriteHeader(http.StatusOK)
					w.Write([]byte(strings.Join(val, ",")))
				}))
			},
			wantStatus: http.StatusOK,
			wantBody:   "X-Middleware1-Called",
		},
		{
			name:   "route-specific middleware",
			method: "GET",
			path:   "/route1",

			setup: func(router *goexpress.Router) {
				mw := func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						ctx := context.WithValue(r.Context(), mw2key, []string{"X-Middleware1-Called"})
						r = r.WithContext(ctx)
						next.ServeHTTP(w, r)
					})
				}

				router.Get("/route1", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					val, ok := r.Context().Value(mw2key).([]string)
					if !ok {
						t.Fatalf("unable to get context value: %v", val)
					}

					w.WriteHeader(http.StatusOK)
					w.Write([]byte(strings.Join(val, ",")))
				}), mw)
			},
			wantStatus: http.StatusOK,
			wantBody:   "X-Middleware1-Called",
		},
		{
			name:   "unregistered method",
			method: "POST",
			path:   "/hello",
			setup: func(router *goexpress.Router) {
				router.Get("/hello", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("Hello, world!"))
				}))
			},
			wantStatus: http.StatusMethodNotAllowed,
			wantBody:   "Method Not Allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := goexpress.New()

			if tt.setup != nil {
				tt.setup(router)
			}

			req := httptest.NewRequest(tt.method, tt.path, http.NoBody)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assertStatus(t, rec.Code, tt.wantStatus)

			assertBody(t, rec.Body.String(), tt.wantBody)
		})
	}
}

func TestGroup(t *testing.T) {
	t.Parallel()

	const header = "X-Middleware-Type"

	tests := []struct {
		name       string
		method     string
		path       string
		setup      func(*goexpress.Router)
		wantStatus int
		wantBody   string
		wantHeader string
	}{
		{
			name:   "group middleware",
			method: http.MethodGet,
			path:   "/api/hello",
			setup: func(router *goexpress.Router) {
				grpMw := func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.Header().Set(header, "group")
						next.ServeHTTP(w, r)
					})
				}

				router.Group("/api", func(r *goexpress.Router) {
					r.Get("/hello", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusOK)
						w.Write([]byte("hello"))
					}))
				}, grpMw)
			},
			wantStatus: http.StatusOK,
			wantBody:   "hello",
			wantHeader: "group",
		},
		{
			name:   "route middleware",
			method: http.MethodGet,
			path:   "/api/hello",
			setup: func(router *goexpress.Router) {
				mw := func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.Header().Set(header, "route")
						next.ServeHTTP(w, r)
					})
				}

				router.Group("/api", func(r *goexpress.Router) {
					r.Get("/hello", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusOK)
						w.Write([]byte("hello"))
					}), mw)
				})
			},
			wantStatus: http.StatusOK,
			wantBody:   "hello",
			wantHeader: "route",
		},
		{
			name:   "group and route middleware",
			method: http.MethodGet,
			path:   "/api/hello",
			setup: func(router *goexpress.Router) {
				grpMw := func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.Header().Set(header, "group")
						next.ServeHTTP(w, r)
					})
				}

				mw := func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusUnauthorized)
						next.ServeHTTP(w, r)
					})
				}

				router.Group("/api", func(r *goexpress.Router) {
					r.Get("/hello", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.Write([]byte("hello"))
					}), mw)
				}, grpMw)
			},
			wantStatus: http.StatusUnauthorized,
			wantBody:   "hello",
			wantHeader: "group",
		},
		{
			name:   "nested group",
			method: http.MethodGet,
			path:   "/api/users/hello",
			setup: func(router *goexpress.Router) {
				router.Group("/api", func(r *goexpress.Router) {
					r.Group("/users", func(r2 *goexpress.Router) {
						r2.Get("/hello", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
							w.WriteHeader(http.StatusOK)
							w.Write([]byte("hello"))
						}))
					})
				})
			},
			wantStatus: http.StatusOK,
			wantBody:   "hello",
			wantHeader: "",
		},
		{
			name:   "nested group with outer group middleware",
			method: http.MethodGet,
			path:   "/api/users/hello",
			setup: func(router *goexpress.Router) {
				grpMw := func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.Header().Set(header, "outer")
						next.ServeHTTP(w, r)
					})
				}

				router.Group("/api", func(r *goexpress.Router) {
					r.Group("/users", func(r2 *goexpress.Router) {
						r2.Get("/hello", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
							w.WriteHeader(http.StatusOK)
							w.Write([]byte("hello"))
						}))
					})
				}, grpMw)
			},
			wantStatus: http.StatusOK,
			wantBody:   "hello",
			wantHeader: "outer",
		},
		{
			name:   "nested group with inner group middleware",
			method: http.MethodGet,
			path:   "/api/users/hello",
			setup: func(router *goexpress.Router) {
				grpMw := func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.Header().Set(header, "inner")
						next.ServeHTTP(w, r)
					})
				}

				router.Group("/api", func(r *goexpress.Router) {
					r.Group("/users", func(r2 *goexpress.Router) {
						r2.Get("/hello", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
							w.WriteHeader(http.StatusOK)
							w.Write([]byte("hello"))
						}))
					}, grpMw)
				})
			},
			wantStatus: http.StatusOK,
			wantBody:   "hello",
			wantHeader: "inner",
		},
		{
			name:   "nested group with inner group and route middleware",
			method: http.MethodGet,
			path:   "/api/users/hello",
			setup: func(router *goexpress.Router) {
				grpMw := func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.Header().Set(header, "inner")
						next.ServeHTTP(w, r)
					})
				}

				mw := func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusUnauthorized)
						next.ServeHTTP(w, r)
					})
				}

				router.Group("/api", func(r *goexpress.Router) {
					r.Group("/users", func(r2 *goexpress.Router) {
						r2.Get("/hello", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
							w.WriteHeader(http.StatusOK)
							w.Write([]byte("hello"))
						}), mw)
					}, grpMw)
				})
			},
			wantStatus: http.StatusUnauthorized,
			wantBody:   "hello",
			wantHeader: "inner",
		},
		{
			name:   "base route",
			method: http.MethodPost,
			path:   "/users",
			setup: func(router *goexpress.Router) {
				router.Group("/users", func(r *goexpress.Router) {
					r.Post("/", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusCreated)
						w.Write([]byte("user created"))
					}))
				})
			},
			wantStatus: http.StatusCreated,
			wantBody:   "user created",
			wantHeader: "",
		},
		{
			name:   "route param",
			method: http.MethodGet,
			path:   "/users/1",
			setup: func(router *goexpress.Router) {
				router.Group("/users", func(r *goexpress.Router) {
					r.Get("/{id}", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
						w.WriteHeader(http.StatusOK)
						w.Write([]byte("user id: " + req.PathValue("id")))
					}))
				})
			},
			wantStatus: http.StatusOK,
			wantBody:   "user id: 1",
			wantHeader: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := goexpress.New()

			gMw := func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					t.Log("global middleware called")
					next.ServeHTTP(w, r)
				})
			}

			router.Use(gMw)

			tt.setup(router)

			req := httptest.NewRequest(tt.method, tt.path, http.NoBody)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assertStatus(t, rec.Code, tt.wantStatus)

			assertBody(t, rec.Body.String(), tt.wantBody)

			assertHeader(t, rec, header, tt.wantHeader)
		})
	}
}

func assertStatus(t *testing.T, status, wantStatus int) {
	t.Helper()

	if status != wantStatus {
		t.Errorf("got status %d, want %d", status, wantStatus)
	}
}

func assertBody(t *testing.T, body, wantBody string) {
	t.Helper()

	if strings.TrimSpace(body) != wantBody {
		t.Errorf("body = %q, want: %q", body, wantBody)
	}
}

func assertHeader(t *testing.T, rec *httptest.ResponseRecorder, header, wantHeader string) {
	t.Helper()

	gotHeader := rec.Header().Get(header)
	if gotHeader != wantHeader {
		t.Errorf("rec.Header().Get(%q) = %q, want: %q", header, gotHeader, wantHeader)
	}
}
