package goexpress_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ferdiebergado/goexpress"
)

// TestNewRouter ensures that a new Router is created successfully.
func TestNewRouter(t *testing.T) {
	r := goexpress.New()
	if r == nil {
		t.Fatal("Expected non-nil router")
	}
}

// TestRouterGet ensures that GET routes are handled correctly.
func TestRouterGet(t *testing.T) {
	r := goexpress.New()
	r.Get("/todos", func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write([]byte("Todo list."))
		if err != nil {
			t.Errorf("write byte: %v", err)
		}
	})

	req := httptest.NewRequest("GET", "/todos", http.NoBody)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK; got %d", resp.StatusCode)
	}

	body := w.Body.String()
	if body != "Todo list." {
		t.Errorf("expected body to be 'Post submitted!'; got %s", body)
	}
}

// TestRouterPost ensures that POST routes are handled correctly.
func TestRouterPost(t *testing.T) {
	r := goexpress.New()
	r.Post("/submit", func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write([]byte("Post submitted!"))
		if err != nil {
			t.Errorf("write byte: %v", err)
		}
	})

	req := httptest.NewRequest("POST", "/submit", http.NoBody)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK; got %d", resp.StatusCode)
	}

	body := w.Body.String()
	if body != "Post submitted!" {
		t.Errorf("expected body to be 'Post submitted!'; got %s", body)
	}
}

// TestRouterPost ensures that POST routes are handled correctly.
func TestRouterPatch(t *testing.T) {
	r := goexpress.New()
	r.Patch("/patch_update", func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write([]byte("Post updated!"))
		if err != nil {
			t.Errorf("write byte: %v", err)
		}
	})

	req := httptest.NewRequest("PATCH", "/patch_update", http.NoBody)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK; got %d", resp.StatusCode)
	}

	body := w.Body.String()
	if body != "Post updated!" {
		t.Errorf("expected body to be 'Post updated!'; got %s", body)
	}
}

func TestRouterPut(t *testing.T) {
	r := goexpress.New()
	r.Put("/put_update", func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write([]byte("Post updated!"))
		if err != nil {
			t.Errorf("write byte: %v", err)
		}
	})

	req := httptest.NewRequest("PUT", "/put_update", http.NoBody)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK; got %d", resp.StatusCode)
	}

	body := w.Body.String()
	if body != "Post updated!" {
		t.Errorf("expected body to be 'Post updated!'; got %s", body)
	}
}

// TestRouterHandleMethod tests that a specific HTTP method and path are handled correctly.
func TestRouterHandleMethod(t *testing.T) {
	r := goexpress.New()

	r.Put("/update", func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write([]byte("Update successful"))
		if err != nil {
			t.Errorf("write byte: %v", err)
		}
	})

	req := httptest.NewRequest("PUT", "/update", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK; got %d", resp.StatusCode)
	}

	body := w.Body.String()
	if body != "Update successful" {
		t.Errorf("expected body to be 'Update successful'; got %s", body)
	}
}

// TestDeleteMethod ensures that DELETE routes are handled correctly.
func TestDeleteMethod(t *testing.T) {
	r := goexpress.New()
	r.Delete("/remove", func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write([]byte("Delete successful"))
		if err != nil {
			t.Errorf("write byte: %v", err)
		}
	})

	req := httptest.NewRequest("DELETE", "/remove", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK; got %d", resp.StatusCode)
	}

	body := w.Body.String()
	if body != "Delete successful" {
		t.Errorf("expected body to be 'Delete successful'; got %s", body)
	}
}

// TestConnect tests the Connect method of the Router.
func TestConnect(t *testing.T) {
	r := goexpress.New()
	r.Connect("/connect", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("Connect response"))
		if err != nil {
			t.Errorf("write byte: %v", err)
		}
	})

	req := httptest.NewRequest("CONNECT", "/connect", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK; got %d", resp.StatusCode)
	}

	body := w.Body.String()
	if body != "Connect response" {
		t.Errorf("expected body to be 'Connect response'; got %s", body)
	}
}

// TestOptions tests the Options method of the Router.
func TestOptions(t *testing.T) {
	r := goexpress.New()
	r.Options("/options", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	req := httptest.NewRequest("OPTIONS", "/options", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected status No Content; got %d", resp.StatusCode)
	}
}

// TestTrace tests the Trace method of the Router.
func TestTrace(t *testing.T) {
	r := goexpress.New()
	r.Trace("/trace", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("Trace response"))
		if err != nil {
			t.Errorf("write byte: %v", err)
		}
	})

	req := httptest.NewRequest("TRACE", "/trace", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK; got %d", resp.StatusCode)
	}

	body := w.Body.String()
	if body != "Trace response" {
		t.Errorf("expected body to be 'Trace response'; got %s", body)
	}
}

// TestHead tests the Head method of the Router.
func TestHead(t *testing.T) {
	r := goexpress.New()
	r.Head("/head", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Custom-Header", "CustomValue")
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("HEAD", "/head", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK; got %d", resp.StatusCode)
	}

	// Check that the custom header is present
	if headerValue := resp.Header.Get("X-Custom-Header"); headerValue != "CustomValue" {
		t.Errorf("expected header X-Custom-Header to be 'CustomValue'; got %s", headerValue)
	}

	// Ensure the body is empty for HEAD requests
	body := w.Body.String()
	if body != "" {
		t.Errorf("expected body to be empty for HEAD request; got %s", body)
	}
}

// TestStaticPathHandling verifies that a request to a static file within a specified directory is correctly handled.
func TestStaticPathHandling(t *testing.T) {
	const staticPath = "public"
	r := goexpress.New()
	r.ServeStatic(staticPath)

	req := httptest.NewRequest("GET", "/public/home.html", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK; got %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("content-type")

	if strings.Split(contentType, ";")[0] != "text/html" {
		t.Errorf("expected content-type text/html; got %s", resp.Header.Get("content-type"))
	}
}

// TestNotFound verifies that the custom "Not Found" handler is called for undefined routes.
func TestNotFound(t *testing.T) {
	// Initialize the r.
	r := goexpress.New()

	// Define the custom "Not Found" handler.
	notFoundHandler := func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "Custom 404 - Page Not Found", http.StatusNotFound)
	}

	// Set the custom "Not Found" handler in the router.
	r.NotFound(notFoundHandler)

	// Create a request to an undefined route.
	req := httptest.NewRequest("GET", "/undefined", nil)
	rec := httptest.NewRecorder()

	// Serve the request using the router.
	r.ServeHTTP(rec, req)

	// Check the response status code.
	if status := rec.Code; status != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, status)
	}

	// Check the response body.
	expectedBody := "Custom 404 - Page Not Found\n"
	if body := rec.Body.String(); body != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, body)
	}
}

func TestRouterGroup(t *testing.T) {
	router := goexpress.New()
	router.Use(goexpress.LogRequest)
	router.Use(goexpress.StripTrailingSlashes)

	// Middleware to trace execution
	var trace []string
	testMiddleware := func(name string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				trace = append(trace, name)
				t.Logf("trace: %v", trace)
				next.ServeHTTP(w, r)
			})
		}
	}

	// Define a subroute under the group
	router.Group("/api", setupGR, testMiddleware("group-middleware"))

	// t.Logf("routes: %v", router.Routes())

	// Create a test request
	req := httptest.NewRequest(http.MethodGet, "/api/hello", nil)
	rec := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(rec, req)

	// Assertions
	if got := rec.Body.String(); got != "Hello from group" {
		t.Errorf("expected response body to be 'Hello from group', got '%s'", got)
	}
	if len(trace) == 0 || trace[0] != "group-middleware" {
		t.Errorf("expected group middleware to be applied, trace: %v", trace)
	}

	req = httptest.NewRequest(http.MethodGet, "/api", nil)
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if got := rec.Body.String(); got != "Hello from api" {
		t.Errorf("expected response body to be 'Hello from api', got '%s'", got)
	}
	if len(trace) == 0 || trace[0] != "group-middleware" {
		t.Errorf("expected group middleware to be applied, trace: %v", trace)
	}
}

func setupGR(r *goexpress.Router) {
	r.Get("/", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "Hello from api")
	}))
	r.Get("/hello", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "Hello from group")
	}))
	r.Group("/users", func(r2 *goexpress.Router) {
		r2.Get("/profile", func(w http.ResponseWriter, _ *http.Request) {
			w.Write([]byte("hi from profile"))
		})
	})
}

func TestRouterString(t *testing.T) {
	m := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Log("middleware")
			next.ServeHTTP(w, r)
		})
	}

	r := goexpress.New()
	r.Use(goexpress.LogRequest)

	r.Get("/hello", helloHandler)
	r.Post("/world", worldHandler, m)
	r.Group("/users", func(router *goexpress.Router) {
		router.Get("/edit", helloHandler)
	})

	t.Logf("%s", r)
}

func helloHandler(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte("hello"))
}

func worldHandler(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte("world"))
}

func TestRouter_ServeHTTP(t *testing.T) {
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
				router.Get("/hello", func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("Hello, world!"))
				})
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

				router.Get("/", func(w http.ResponseWriter, r *http.Request) {
					val, ok := r.Context().Value(mwKey).([]string)
					if !ok {
						t.Fatalf("unable to get context value: %v", val)
					}

					w.WriteHeader(http.StatusOK)
					w.Write([]byte(strings.Join(val, ",")))
				})
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

				router.Get("/route1", func(w http.ResponseWriter, r *http.Request) {
					val, ok := r.Context().Value(mw2key).([]string)
					if !ok {
						t.Fatalf("unable to get context value: %v", val)
					}

					w.WriteHeader(http.StatusOK)
					w.Write([]byte(strings.Join(val, ",")))
				}, mw)
			},
			wantStatus: http.StatusOK,
			wantBody:   "X-Middleware1-Called",
		},
		{
			name:   "unregistered method",
			method: "POST",
			path:   "/hello",
			setup: func(router *goexpress.Router) {
				router.Get("/hello", func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("Hello, world!"))
				})
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

			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assertStatus(t, rec.Code, tt.wantStatus)

			assertBody(t, rec.Body.String(), tt.wantBody)
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

func TestRouter_Group(t *testing.T) {
	t.Parallel()

	const (
		header = "X-Middleware-Type"
	)

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
					r.Get("/hello", func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusOK)
						w.Write([]byte("hello"))
					})
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
					r.Get("/hello", func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusOK)
						w.Write([]byte("hello"))
					}, mw)
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
					r.Get("/hello", func(w http.ResponseWriter, _ *http.Request) {
						w.Write([]byte("hello"))
					}, mw)
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
						r2.Get("/hello", func(w http.ResponseWriter, _ *http.Request) {
							w.WriteHeader(http.StatusOK)
							w.Write([]byte("hello"))
						})
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
						r2.Get("/hello", func(w http.ResponseWriter, _ *http.Request) {
							w.WriteHeader(http.StatusOK)
							w.Write([]byte("hello"))
						})
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
						r2.Get("/hello", func(w http.ResponseWriter, _ *http.Request) {
							w.WriteHeader(http.StatusOK)
							w.Write([]byte("hello"))
						})
					}, grpMw)
				})
			},
			wantStatus: http.StatusOK,
			wantBody:   "hello",
			wantHeader: "inner",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := goexpress.New()
			tt.setup(router)

			req := httptest.NewRequest(tt.method, tt.path, http.NoBody)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assertStatus(t, rec.Code, tt.wantStatus)

			assertBody(t, rec.Body.String(), tt.wantBody)

			if header := rec.Header().Get(header); header != tt.wantHeader {
				t.Errorf("rec.Header().Get(%q) = %q, want: %q", header, rec.Header().Get(header), tt.wantHeader)
			}
		})
	}
}
