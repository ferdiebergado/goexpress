package router

import (
	"fmt"
	"net/http"
)

// Router is a custom HTTP router built on top of ServeMux with middleware support.
// It allows you to register routes and apply both global and route-specific middleware.
type Router struct {
	mux         *http.ServeMux
	middlewares []Middleware
}

// Middleware defines the signature for middleware functions.
// A Middleware takes an http.Handler as input and returns a new http.Handler
// that can wrap additional functionality around the original handler.
type Middleware func(http.Handler) http.Handler

// Handler defines a function type for handling HTTP requests.
// It takes an http.ResponseWriter to write the response and an http.Request for the request details.
type Handler func(w http.ResponseWriter, r *http.Request)

// NewRouter creates a new instance of Router with an underlying http.ServeMux
// and an empty slice for middlewares. This Router can be used to register routes
// and apply middleware to HTTP requests.
func NewRouter() *Router {
	r := &Router{
		mux:         http.NewServeMux(),
		middlewares: []Middleware{},
	}

	r.HandleNotFound()

	return r
}

// Use adds a middleware to the router's global middleware chain.
// The middleware will be applied to every request handled by this Router.
func (r *Router) Use(mw Middleware) {
	r.middlewares = append(r.middlewares, mw)
}

// wrap applies a series of middlewares to an http.Handler in reverse order.
// The handler is wrapped by each middleware so that the outermost middleware
// is the first to handle the request, and the innermost middleware is the last.
func (r *Router) wrap(handler http.Handler, middlewares []Middleware) http.Handler {
	finalHandler := handler

	for i := len(middlewares) - 1; i >= 0; i-- {
		finalHandler = middlewares[i](finalHandler)
	}

	return finalHandler
}

// Handle registers a new route with a specific pattern and handler.
// It allows optional route-specific middleware to be applied, which will be executed
// before the router's global middleware. The final handler is wrapped in both the
// route-specific and global middleware chains.
func (r *Router) Handle(pattern string, handler Handler, middlewares ...Middleware) {
	// Wrap the handler with route-specific middlewares first.
	finalHandler := r.wrap(http.HandlerFunc(handler), middlewares)

	// Wrap with global middlewares.
	finalHandler = r.wrap(finalHandler, r.middlewares)

	r.mux.Handle(pattern, finalHandler)
}

// HandleMethod registers a route with a specified HTTP method, path, and handler function.
// It also allows optional middlewares to be applied to this route.
// The method and path are combined into a pattern, which is used to route the request.
func (r *Router) HandleMethod(method string, path string, handler Handler, middlewares ...Middleware) {
	pattern := fmt.Sprintf("%s %s", method, path)
	r.Handle(pattern, handler, middlewares...)
}

// Get registers a new route that responds to HTTP GET requests for the specified path.
// It applies the provided handler and any optional middlewares to the route.
func (r *Router) Get(path string, handler Handler, middlewares ...Middleware) {
	r.HandleMethod("GET", path, handler, middlewares...)
}

// Post registers a new route that responds to HTTP POST requests for the specified path.
// It applies the provided handler and any optional middlewares to the route.
func (r *Router) Post(path string, handler Handler, middlewares ...Middleware) {
	r.HandleMethod("POST", path, handler, middlewares...)
}

// Patch registers a new route that responds to HTTP PATCH requests for the specified path.
// It applies the provided handler and any optional middlewares to the route.
func (r *Router) Patch(path string, handler Handler, middlewares ...Middleware) {
	r.HandleMethod("PATCH", path, handler, middlewares...)
}

// Put registers a new route that responds to HTTP PUT requests for the specified path.
// It applies the provided handler and any optional middlewares to the route.
func (r *Router) Put(path string, handler Handler, middlewares ...Middleware) {
	r.HandleMethod("PUT", path, handler, middlewares...)
}

// Delete registers a new route that responds to HTTP DELETE requests for the specified path.
// It applies the provided handler and any optional middlewares to the route.
func (r *Router) Delete(path string, handler Handler, middlewares ...Middleware) {
	r.HandleMethod("DELETE", path, handler, middlewares...)
}

// ServeHTTP allows the Router to satisfy the http.Handler interface.
// It delegates the actual request handling to the underlying ServeMux
// after all middlewares have been applied to the registered handlers.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

// Not found handler
func (r *Router) HandleNotFound() {
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})
}
