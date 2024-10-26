# go-express

Simple, dependency-free, and Express-styled HTTP Router written in Go.

## Requirements

- Go 1.22 or higher

## Installation

```sh
go get github.com/ferdiebergado/go-express
```

## Usage

1. Import the router.

```go
import router "github.com/ferdiebergado/go-express"
```

2. Create a router.

```go
router := router.NewRouter()
```

3. Register global middlewares.

Global middlewares are registered by calling the Use() method on the router.

```go
router.Use(RequestLogger)
```

go-express has some commonly-used middlewares available out of the box, just import it from the middleware package.

```go
import (
	"github.com/ferdiebergado/go-express/router"
	"github.com/ferdiebergado/go-express/middleware"
)

func main() {
	router := router.NewRouter()

	router.Use(middleware.RequestLogger)
	router.Use(middleware.StripTrailingSlashes)
	router.Use(middleware.PanicRecovery)
}
```

4. Register routes.

Similar to express, the http methods are available as methods in the router. The method accepts the path and the handler as arguments.

```go
router.Get(path, handler)
```

Example:

```go
router.Get("/todos", TodoHandler)
```

The handler is an http function that accepts an http.ResponseWriter and a pointer to an http.Request as arguments.

```go
func Handler(w http.ResponseWriter, r *http.Request)
```

Example:

```go
func TodoHandler(w http.ResponseWriter, r *http.Request) {
    w.write([]byte("Todos"))
}
```

Optionally, you can register path-specific middlewares by passing them as arguments after the handler.

Example:

```go
router.Post("/todos", CreateTodoHandler, SessionMiddleware, AuthMiddleware)
```

In this example, the route has two specific middlewares: SessionMiddleware and AuthMiddleware.

You can pass any number of middlewares to a route.

5. Start an http server with the router.

```go
http.ListenAndServe(":8080", router)
```

## Serving Static Files

go-express makes it easy to serve static files from a specified directory. Simply provide the path of the static files to be served to the ServeStatic method of the router.

```go
router.ServeStatic("assets")
```

## Writing Middlewares

Middlewares are functions that accept an http.Handler and returns another http.Handler.

```go
func Middleware(next http.Handler) http.Handler
```

Example:

```go
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)
		log.Printf("%s %s %s %d", r.Method, r.URL.Path, r.Proto, duration)
	})
}
```

## Complete Usage Example

```go
package main

import (
	"net/http"

	router "github.com/ferdiebergado/go-express"
	"github.com/ferdiebergado/go-express/middleware"
)

func main() {

	// Create the router.
	router := router.NewRouter()

	// Register global middlewares.
	router.Use(middleware.RequestLogger)
	router.Use(middleware.StripTrailingSlashes)
	router.Use(middleware.PanicRecovery)

	// Serve static files.
	router.ServeStatic("assets")

	// Register routes.
	router.Get("/api/todos", ListTodos)
	router.Post("/api/todos", CreateTodo, AuthMiddleware)
	router.Put("/api/todos/{id}", UpdateTodo, AuthMiddleware)
	router.Delete("/api/todos/{id}", DeleteTodo, AuthMiddleware)

	http.ListenAndServe(":8080", router)
}

func ListTodos(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Todo list"))
}

func CreateTodo(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
}

func UpdateTodo(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	w.Write([]byte("Todo with id "+id+" updated"))
}

func DeleteTodo(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		// Check for the Authorization header
		if authHeader == "" || authHeader != "Bearer valid-token" { // replace "valid-token" with actual validation logic
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Call the next handler in the chain
		next.ServeHTTP(w, r)
	})
}
```

## License

go-express is open-sourced software licensed under [MIT License](https://github.com/ferdiebergado/go-express/blob/main/LICENSE).
