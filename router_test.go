package goexpress

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestNewRouter ensures that a new Router is created successfully.
func TestNewRouter(t *testing.T) {
	r := New()
	if r == nil {
		t.Fatal("Expected non-nil router")
	}
}

// TestRouterHandle tests route registration and response.
func TestRouterHandle(t *testing.T) {
	r := New()
	r.Get("/hello", func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write([]byte("Hello, world!"))

		if err != nil {
			t.Errorf("write byte: %v", err)
		}
	})

	req := httptest.NewRequest("GET", "/hello", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK; got %d", resp.StatusCode)
	}

	body := w.Body.String()
	if body != "Hello, world!" {
		t.Errorf("expected body to be 'Hello, world!'; got %s", body)
	}
}

// TestRouterPost ensures that POST routes are handled correctly.
func TestRouterGet(t *testing.T) {
	r := New()
	r.Get("/todos", func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write([]byte("Todo list."))
		if err != nil {
			t.Errorf("write byte: %v", err)
		}
	})

	req := httptest.NewRequest("GET", "/todos", nil)
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
	r := New()
	r.Post("/submit", func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write([]byte("Post submitted!"))
		if err != nil {
			t.Errorf("write byte: %v", err)
		}
	})

	req := httptest.NewRequest("POST", "/submit", nil)
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
	r := New()
	r.Patch("/patch_update", func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write([]byte("Post updated!"))
		if err != nil {
			t.Errorf("write byte: %v", err)
		}
	})

	req := httptest.NewRequest("PATCH", "/patch_update", nil)
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
	r := New()
	r.Put("/put_update", func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write([]byte("Post updated!"))
		if err != nil {
			t.Errorf("write byte: %v", err)
		}
	})

	req := httptest.NewRequest("PUT", "/put_update", nil)
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

// TestRouterNotFound ensures that the router returns error 404 for unassigned routes.
func TestRouterNotFoundHandling(t *testing.T) {
	r := New()

	req := httptest.NewRequest("GET", "/notfound", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404; got %d", resp.StatusCode)
	}
}

// TestGlobalMiddleware ensures that global middleware is applied to all routes.
func TestGlobalMiddleware(t *testing.T) {
	r := New()

	// Add a global middleware that appends "Processed: " to the response.
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte("Processed: "))
			if err != nil {
				t.Errorf("write byte: %v", err)
			}
			next.ServeHTTP(w, r)
		})
	})

	r.Get("/hello", func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write([]byte("Hello"))
		if err != nil {
			t.Errorf("write byte: %v", err)
		}
	})

	req := httptest.NewRequest("GET", "/hello", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	body := w.Body.String()
	expected := "Processed: Hello"
	if body != expected {
		t.Errorf("expected '%s', got '%s'", expected, body)
	}
}

// TestRouteSpecificMiddleware ensures route-specific middleware is applied correctly.
func TestRouteSpecificMiddleware(t *testing.T) {
	r := New()

	// Add route-specific middleware that appends "Specific: " to the response.
	r.Get("/hello", func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write([]byte("Hello"))
		if err != nil {
			t.Errorf("write byte: %v", err)
		}
	}, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte("Specific: "))
			if err != nil {
				t.Errorf("write byte: %v", err)
			}
			next.ServeHTTP(w, r)
		})
	})

	req := httptest.NewRequest("GET", "/hello", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	body := w.Body.String()
	expected := "Specific: Hello"
	if body != expected {
		t.Errorf("expected '%s', got '%s'", expected, body)
	}
}

// TestMethodNotAllowed ensures that the router returns a 405 status for unsupported methods.
func TestMethodNotAllowed(t *testing.T) {
	r := New()
	r.Get("/hello", func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write([]byte("Hello"))
		if err != nil {
			t.Errorf("write byte: %v", err)
		}
	})

	req := httptest.NewRequest("POST", "/hello", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected status MethodNotAllowed; got %d", resp.StatusCode)
	}
}

// TestRouterHandleMethod tests that a specific HTTP method and path are handled correctly.
func TestRouterHandleMethod(t *testing.T) {
	r := New()

	r.handleMethod("PUT", "/update", func(w http.ResponseWriter, _ *http.Request) {
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
	r := New()
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
	r := New()
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
	r := New()
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
	r := New()
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
	r := New()
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

// TestStaticPathHandling verifies that a request to a static file within a specified directory is correctly handled
func TestStaticPathHandling(t *testing.T) {
	const staticPath = "public"
	r := New()
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
	// Initialize the router.
	router := New()

	// Define the custom "Not Found" handler.
	notFoundHandler := func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "Custom 404 - Page Not Found", http.StatusNotFound)
	}

	// Set the custom "Not Found" handler in the router.
	router.NotFound(notFoundHandler)

	// Create a request to an undefined route.
	req := httptest.NewRequest("GET", "/undefined", nil)
	rec := httptest.NewRecorder()

	// Serve the request using the router.
	router.ServeHTTP(rec, req)

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

func TestGroup(t *testing.T) {
	// Test case 1: Basic group with no middleware
	r := New()
	r.Group("/api", func(gr *Router) *Router {
		gr.Get("/users", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("users"))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		})
		return gr
	})

	req := httptest.NewRequest("GET", "/api/users", nil)
	rec := httptest.NewRecorder()
	r.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}
	if rec.Body.String() != "users" {
		t.Errorf("Expected body %s, got %s", "users", rec.Body.String())
	}

	// Test case 2: Group with middleware
	r = New()
	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Test", "test")
			next.ServeHTTP(w, r)
		})
	}
	r.Group("/v1", func(gr *Router) *Router {
		gr.Get("/items", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("items"))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		})
		return gr
	}, middleware)

	req = httptest.NewRequest("GET", "/v1/items", nil)
	rec = httptest.NewRecorder()
	r.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}
	if rec.Body.String() != "items" {
		t.Errorf("Expected body %s, got %s", "items", rec.Body.String())
	}
	if rec.Header().Get("X-Test") != "test" {
		t.Errorf("Expected header X-Test to be %s, got %s", "test", rec.Header().Get("X-Test"))
	}

	// Test case 3: Nested groups
	r = New()
	r.Group("/admin", func(gr *Router) *Router {
		gr.Group("/users", func(gr2 *Router) *Router {
			gr2.Get("/", func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte("admin users"))
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			})
			return gr2
		})
		return gr
	})
	req = httptest.NewRequest("GET", "/admin/users/", nil)
	rec = httptest.NewRecorder()
	r.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}
	if rec.Body.String() != "admin users" {
		t.Errorf("Expected body %s, got %s", "admin users", rec.Body.String())
	}

	// Test case 4:  Empty Prefix
	r = New()
	r.Group("", func(gr *Router) *Router {
		gr.Get("/test", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("test"))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		})
		return gr
	})
	req = httptest.NewRequest("GET", "/test", nil)
	rec = httptest.NewRecorder()
	r.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}
	if rec.Body.String() != "test" {
		t.Errorf("Expected body %s, got %s", "test", rec.Body.String())
	}
}
