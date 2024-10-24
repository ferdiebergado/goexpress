package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestNewRouter ensures that a new Router is created successfully.
func TestNewRouter(t *testing.T) {
	r := NewRouter()
	if r == nil {
		t.Fatal("Expected non-nil router")
	}
}

// TestRouterHandle tests route registration and response.
func TestRouterHandle(t *testing.T) {
	r := NewRouter()
	r.Get("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, world!"))
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
func TestRouterPost(t *testing.T) {
	r := NewRouter()
	r.Post("/submit", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Post submitted!"))
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
	r := NewRouter()
	r.Patch("/patch_update", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Post updated!"))
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
	r := NewRouter()
	r.Put("/put_update", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Post updated!"))
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
	r := NewRouter()

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
	r := NewRouter()

	// Add a global middleware that appends "Processed: " to the response.
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Processed: "))
			next.ServeHTTP(w, r)
		})
	})

	r.Get("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello"))
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
	r := NewRouter()

	// Add route-specific middleware that appends "Specific: " to the response.
	r.Get("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello"))
	}, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Specific: "))
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
	r := NewRouter()
	r.Get("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello"))
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
	r := NewRouter()

	r.HandleMethod("PUT", "/update", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Update successful"))
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
	r := NewRouter()
	r.Delete("/remove", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Delete successful"))
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
