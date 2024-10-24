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
import github.com/ferdiebergado/go-express
```

2. Create a router.

```go
router := router.NewRouter()
```

3. Register global middlewares.

Global middleware are registered by calling the Use() method on the router.

```go
router.Use(middleware Middleware)
```

Example:

```go
router.Use(RequestLogger)
```

go-express has some commonly-used middlewares available out of the box, just import it from the middleware package.

```go
import github.com/ferdiebergado/go-express/middleware

router.Use(middleware.RequestLogger)
router.Use(middleware.StripTrailingSlashes)
router.Use(middleware.PanicRecovery)
```

4. Register routes.

Similar to express, the http methods are available as methods in the router. The method accepts the path as first argument and the handler as the second argument.

```go
router.Get(path string, handler Handler)
```

The Handler is an http function that accepts an http.ResponseWriter and an http.Request as arguments.

```go
func Handler(w http.ResponseWriter, r *http.Request)
```

Example:

```go
router.Get("/todos", TodoHandler)

func TodoHandler(w http.ResponseWriter, r *http.Request) {
    w.write([]byte("Todos"))
}
```

Optionally, you can register path-specific middlewares by passing them as argument next to the handler.

Example:

```go
router.Post("/todos",CreateTodoHandler, SessionMiddleware, AuthMiddleware)
```

In this example, the route has two specific middlewares: SessionMiddleware and AuthMiddleware.

You can pass any number of middlewares to every route.

## Writing Middlewares

Middlewares are functions that accepts an http.Handler and returns another http.Handler.

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

## License
go-express is open-sourced software licensed under [MIT License](https://github.com/ferdiebergado/go-express/blob/main/LICENSE).