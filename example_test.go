package goexpress_test

import (
	"log"
	"net/http"

	"github.com/ferdiebergado/goexpress"
)

func Example() {
	router := goexpress.New()

	router.Use(goexpress.RecoverPanic)
	router.Use(goexpress.LogRequest)

	router.Get("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	})

	log.Fatal(http.ListenAndServe(":8080", router))
}

func ExampleRouter_Use() {
	router := goexpress.New()

	router.Use(goexpress.RecoverPanic)
	router.Use(goexpress.LogRequest)
}

func ExampleRouter_Get() {
	router := goexpress.New()

	router.Get("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	})
}

func ExampleRouter_Post() {
	router := goexpress.New()

	router.Post("/register", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Registered!"))
	})
}

func ExampleRouter_Group() {
	router := goexpress.New()

	authMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if token := r.Header.Get("Authorization"); token == "" {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	checkContentTypeMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ct := r.Header.Get("Content-Type"); ct != "application/json" {
				http.Error(w, "Unsupported media", http.StatusUnsupportedMediaType)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	router.Group("/users", func(usersRouter *goexpress.Router) {
		usersRouter.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("profile of user: " + r.PathValue("id")))
		})
		usersRouter.Post("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte("user created"))
		}, checkContentTypeMiddleware)
		usersRouter.Group("/posts", func(postsRouter *goexpress.Router) {
			postsRouter.Get("/", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("list of posts"))
			})
		})
	}, authMiddleware)
}

func ExampleRouter_Static() {
	router := goexpress.New()

	router.Static("/static", "./static") // Serves files from the "static" directory at "/static/"
}

func ExampleRouter_NotFound() {
	router := goexpress.New()

	router.Get("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	})

	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Custom 404 - Page Not Found", http.StatusNotFound)
	})
}
