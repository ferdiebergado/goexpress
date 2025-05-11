# goexpress

![Github Actions](https://github.com/ferdiebergado/goexpress/actions/workflows/go.yml/badge.svg?event=push) [![Go Report Card](https://goreportcard.com/badge/github.com/ferdiebergado/goexpress)](https://goreportcard.com/report/github.com/ferdiebergado/goexpress) [![Go Reference](https://pkg.go.dev/badge/github.com/ferdiebergado/goexpress.svg)](https://pkg.go.dev/github.com/ferdiebergado/goexpress)

Simple, dependency-free, and Express-styled HTTP Router written in Go.

## Features

- Simple and easy to use
- Expressjs-style routing
- Depends only on the Go standard library
- Static files handling
- Common middlewares available out of the box

## Requirements

- Go 1.22 or higher

## Installation

```sh
go get github.com/ferdiebergado/goexpress
```

## Usage

1. Import the router.

```go
import "github.com/ferdiebergado/goexpress"
```

2. Create a router.

```go
router := goexpress.New()
```

3. Register global middlewares.

Global middlewares are registered by calling the Use() method on the router.

```go
router.Use(goexpress.RequestLogger)
```

goexpress has some commonly-used middlewares available out of the box, just import it from the middleware package.

```go
import "github.com/ferdiebergado/goexpress"

func main() {
	router := goexpress.New()

	router.Use(goexpress.PanicRecovery)
	router.Use(goexpress.RequestLogger)
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

The handler is a function that accepts an http.ResponseWriter and a pointer to an http.Request as arguments.

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

```go
router.Post("/todos", CreateTodoHandler, SessionMiddleware, AuthMiddleware)
```

In here, the route has two specific middlewares: SessionMiddleware and AuthMiddleware.

You can pass any number of middlewares to a route.

5. Start an http server with the router.

```go
http.ListenAndServe(":8080", router)
```

## Route Groups

To simplify handling of multiple routes, a Group method is available on the Router. This makes it possible to specify multiple routes within the same prefix. Routes can be specified just like with the normal router meaning middlewares can also be included.

Middlewares for the route group can also be specified as the last arguments.

Nested route groups are also supported.

```go
router.Group("/api", func(r *Router) {
    r.Get("/users", listUsersHandler)
    r.Post("/users", createUserHandler, authorizeMiddleware)
    r.Get("/products", listProductsHandler)
    r.Group("/shipments", func(shipmentRouter *Router) {
      shipmentRouter.Get("/status", statusHandler)
    })
}, authMiddleware)
```

## Serving Static Files

goexpress makes it easy to serve static files from a specified directory. Simply provide the name of the directory containing the static files to be served to the ServeStatic method of the router.

```go
router.ServeStatic("assets")
```

This will serve files from the "assets" directory at "/assets/". Now, any request the starts with /assets/ will serve files from this directory, e.g. /assets/styles.css.

You can then add this to the head tag of your html:

```html
<head>
  <link rel="stylesheet" src="/assets/styles.css">
  <script src="/assets/js/app.js" defer>
</head>
```

## Custom 404 Error Handler

By default, goexpress returns a 404 status code and plain status text when an unregistered route is requested. To customize this behavior, pass an http handler function to the NotFound method of the router.

Example:

```go
router.NotFound(func(w http.ResponseWriter, r *http.Request){
	templates := template.Must(template.ParseGlob("templates/*.html"))

	var buf bytes.Buffer

	if err := templates.ExecuteTemplate(&buf, "404.html", nil); err != nil {
    log.Prinf("execute template: %w", err)
		return
	}

	_, err = buf.WriteTo(w)

	if err != nil {
    log.Printf("write to buffer: %w", err)
		return
	}
})
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
		log.Printf("%s %s %s", r.Method, r.URL.Path, r.Proto)
		next.ServeHTTP(w, r)
	})
}
```

## Complete Usage Example

```go
package main

import (
	"net/http"

	"github.com/ferdiebergado/goexpress"
)

func main() {

	// Create the router.
	router := goexpress.New()

	// Register global middlewares.
	router.Use(goexpress.PanicRecovery)
	router.Use(goexpress.RequestLogger)

	// Serve static files.
	router.ServeStatic("assets")

	// Register routes.
	router.Get("/api/todos", ListTodos)
	router.Post("/api/todos", CreateTodo, AuthMiddleware)
	router.Put("/api/todos/{id}", UpdateTodo, AuthMiddleware)
	router.Delete("/api/todos/{id}", DeleteTodo, AuthMiddleware)
  router.Group("/api", func(r *Router) {
      r.Get("/users", listUsersHandler)
      r.Post("/users", createUserHandler, authorizeMiddleware)
      r.Get("/products", listProductsHandler)
      r.Group("/shipments", func(shipmentRouter *Router) {
        shipmentRouter.Get("/status", statusHandler)
      })
  }, authMiddleware)

  // Start an http server with the router.
	http.ListenAndServe(":8080", router)
}

func ListTodos(w http.ResponseWriter, r *http.Request) {
	_,err:=w.Write([]byte("Todo list"))
  		if err != nil {
			t.Errorf("write byte: %v",err)
		}
}

func CreateTodo(w http.ResponseWriter, r *http.Request) {
	_, err := w.WriteHeader(http.StatusCreated)
  if err != nil {
    t.Errorf("write byte: %v",err)
  }
}

func UpdateTodo(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	_,err := w.Write([]byte("Todo with id "+id+" updated"))
  if err != nil {
    t.Errorf("write byte: %v",err)
  }
}

func DeleteTodo(w http.ResponseWriter, r *http.Request) {
	_,err := w.WriteHeader(http.StatusNoContent)
  if err != nil {
    t.Errorf("write byte: %v",err)
  }
}

func listUsersHandler(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("list of users"))
  if err != nil {
    t.Errorf("write byte: %v",err)
  }
}

func createUsersHandler(w http.ResponseWriter, r *http.Request) {
  w.WriteHeader(http.StatusCreated)
	_, err := w.Write([]byte("user created"))
  if err != nil {
    t.Errorf("write byte: %v",err)
  }
}

func listProductsHandler(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("list of products"))
  if err != nil {
    t.Errorf("write byte: %v",err)
  }
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("status of shipments"))
  if err != nil {
    t.Errorf("write byte: %v",err)
  }
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" || authHeader != "Bearer valid-token" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func AuthorizeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId := r.Context.Value("userId")

		if userId == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
```

## License

goexpress is open-sourced software licensed under [MIT License](https://github.com/ferdiebergado/goexpress/blob/main/LICENSE).
