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

// NewRouter creates a new instance of Router with an underlying http.ServeMux
// and an empty slice for middlewares. This Router can be used to register routes
// and apply middleware to HTTP requests.
func NewRouter() *Router {
	return &Router{
		mux:         http.NewServeMux(),
		middlewares: make([]Middleware, 0),
	}
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
func (r *Router) Handle(pattern string, handler http.Handler, middlewares ...Middleware) {
	// Wrap the handler with route-specific middlewares first.
	finalHandler := r.wrap(handler, middlewares)

	// Wrap with global middlewares.
	finalHandler = r.wrap(finalHandler, r.middlewares)

	r.mux.Handle(pattern, finalHandler)
}

// handleMethod registers a route with a specified HTTP method, path, and handler function.
// It also allows optional middlewares to be applied to this route.
// The method and path are combined into a pattern, which is used to route the request.
func (r *Router) handleMethod(method string, path string, handler http.HandlerFunc, middlewares ...Middleware) {
	pattern := fmt.Sprintf("%s %s", method, path)
	r.Handle(pattern, handler, middlewares...)
}

// Get registers a new route that responds to HTTP GET requests for the specified path.
// It applies the provided handler and any optional middlewares to the route.
func (r *Router) Get(path string, handler http.HandlerFunc, middlewares ...Middleware) {
	r.handleMethod("GET", path, handler, middlewares...)
}

// Post registers a new route that responds to HTTP POST requests for the specified path.
// It applies the provided handler and any optional middlewares to the route.
func (r *Router) Post(path string, handler http.HandlerFunc, middlewares ...Middleware) {
	r.handleMethod("POST", path, handler, middlewares...)
}

// Patch registers a new route that responds to HTTP PATCH requests for the specified path.
// It applies the provided handler and any optional middlewares to the route.
func (r *Router) Patch(path string, handler http.HandlerFunc, middlewares ...Middleware) {
	r.handleMethod("PATCH", path, handler, middlewares...)
}

// Put registers a new route that responds to HTTP PUT requests for the specified path.
// It applies the provided handler and any optional middlewares to the route.
func (r *Router) Put(path string, handler http.HandlerFunc, middlewares ...Middleware) {
	r.handleMethod("PUT", path, handler, middlewares...)
}

// Delete registers a new route that responds to HTTP DELETE requests for the specified path.
// It applies the provided handler and any optional middlewares to the route.
func (r *Router) Delete(path string, handler http.HandlerFunc, middlewares ...Middleware) {
	r.handleMethod("DELETE", path, handler, middlewares...)
}

// ServeHTTP allows the Router to satisfy the http.Handler interface.
// It delegates the actual request handling to the underlying ServeMux
// after all middlewares have been applied to the registered handlers.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

func (r *Router) ServeStatic(path string) {
	fullPath := fmt.Sprintf("/%s/", path)
	pattern := fmt.Sprintf("GET %s", fullPath)
	r.Handle(pattern, http.StripPrefix(fullPath, http.FileServer(http.Dir(path))))
}
