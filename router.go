// Package goexpress provides an http router implementation.
package goexpress

import (
	"fmt"
	"net/http"
	"path"
	"reflect"
	"runtime"
	"strings"
)

// Middleware defines the signature for a standard net/http middleware function.
//
// A Middleware takes an http.Handler (the 'next' handler in the chain) and returns
// a new http.Handler that wraps and executes the 'next' handler. This signature
// ensures compatibility with the standard library and all third-party Go web middleware.
//
// Example:
//
//	func MyMiddleware(next http.Handler) http.Handler {
//	    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//	        // Logic before the handler runs
//	        next.ServeHTTP(w, r)
//	        // Logic after the handler runs
//	    })
//	}
type Middleware func(http.Handler) http.Handler

// Router is a custom HTTP router built on top of http.ServeMux with support for global
// and route-specific middleware. It allows easy route registration for common HTTP methods
// (GET, POST, PATCH, PUT, DELETE) and provides a flexible middleware chain for request handling.
type Router struct {
	prefix      string         // prefix for the paths of registered routes
	mux         *http.ServeMux // underlying HTTP request multiplexer
	routes      []route        // slice to store the registered routes
	middlewares []Middleware   // slice to store global middlewares
}

// New creates and returns a custom HTTP router.
func New() *Router {
	return &Router{
		mux: http.NewServeMux(),
	}
}

// Use appends the given middleware to the router's global middleware chain. Each middleware
// added with Use will be applied to every request handled by this Router.
func (r *Router) Use(mw Middleware) {
	r.middlewares = append(r.middlewares, mw)
}

// Get registers a new GET route for the specified path and handler, applying any optional middleware.
func (r *Router) Get(p string, handler http.Handler, middlewares ...Middleware) {
	r.handle(http.MethodGet, p, handler, middlewares...)
}

// Post registers a new POST route for the specified path and handler, applying any optional middleware.
func (r *Router) Post(p string, handler http.Handler, middlewares ...Middleware) {
	r.handle(http.MethodPost, p, handler, middlewares...)
}

// Patch registers a new PATCH route for the specified path and handler, applying any optional middleware.
func (r *Router) Patch(p string, handler http.Handler, middlewares ...Middleware) {
	r.handle(http.MethodPatch, p, handler, middlewares...)
}

// Put registers a new PUT route for the specified path and handler, applying any optional middleware.
func (r *Router) Put(p string, handler http.Handler, middlewares ...Middleware) {
	r.handle(http.MethodPut, p, handler, middlewares...)
}

// Delete registers a new DELETE route for the specified path and handler, applying any optional middleware.
func (r *Router) Delete(p string, handler http.Handler, middlewares ...Middleware) {
	r.handle(http.MethodDelete, p, handler, middlewares...)
}

// Connect registers a new route that responds to HTTP CONNECT requests for the specified path.
func (r *Router) Connect(p string, handler http.Handler, middlewares ...Middleware) {
	r.handle(http.MethodConnect, p, handler, middlewares...)
}

// Options registers a new route that responds to HTTP OPTIONS requests for the specified path.
func (r *Router) Options(p string, handler http.Handler, middlewares ...Middleware) {
	r.handle(http.MethodOptions, p, handler, middlewares...)
}

// Trace registers a new route that responds to HTTP TRACE requests for the specified path.
func (r *Router) Trace(p string, handler http.Handler, middlewares ...Middleware) {
	r.handle(http.MethodTrace, p, handler, middlewares...)
}

// Head registers a new route that responds to HTTP HEAD requests for the specified path.
func (r *Router) Head(p string, handler http.Handler, middlewares ...Middleware) {
	r.handle(http.MethodHead, p, handler, middlewares...)
}

// ServeHTTP enables the Router to satisfy the http.Handler interface.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

// Group creates a new route group with a common prefix and applies the
// given function to define sub-routes within that group.
// Routes can be specified just like with the normal router.
// Middlewares for the route group can also be specified as the last arguments.
// Nested route groups are also supported.
func (r *Router) Group(prefix string, fn func(*Router), middlewares ...Middleware) {
	sub := &Router{
		mux:         r.mux,
		prefix:      r.prefix + prefix,
		middlewares: append(append([]Middleware{}, r.middlewares...), middlewares...),
	}

	fn(sub)

	r.routes = append(r.routes, sub.routes...)
}

// Static serves static files from the specified local directory path at the given url prefix.
func (r *Router) Static(prefix, dir string) {
	fullPrefix := normalizePath(prefix)
	handler := http.StripPrefix(fullPrefix, http.FileServer(http.Dir(dir)))
	wrappedHandler := r.wrap(handler, r.middlewares)

	pattern := fullPrefix
	if !strings.HasSuffix(pattern, "/") {
		pattern += "/"
	}

	r.mux.Handle(pattern, wrappedHandler)
}

// NotFound sets a custom handler for requests that don't match any registered route.
// When a request is made to an undefined route, this handler will be invoked,
// allowing a custom "Not Found" page or response to be returned.
func (r *Router) NotFound(handler http.Handler) {
	finalHandler := r.wrap(handler, r.middlewares)
	r.mux.Handle("/", finalHandler)
}

// String returns the middlewares and routes registered in the Router as a string.
func (r *Router) String() string {
	var s strings.Builder
	s.Write([]byte("\nMiddlewares:\n"))
	for _, m := range r.middlewares {
		fullFuncName := runtime.FuncForPC(reflect.ValueOf(m).Pointer()).Name()
		name := trimRepoName(fullFuncName)
		s.Write([]byte(name + "\n"))
	}

	s.Write([]byte("\nRoutes:\n"))
	for _, r := range r.routes {
		s.Write([]byte(r.String() + "\n"))
	}
	return s.String()
}

// handle registers a handle with a specified HTTP method and path, applying
// any optional middlewares to the handler.
func (r *Router) handle(method, p string, handler http.Handler, mws ...Middleware) {
	fullPath := normalizePath(r.prefix + p)
	pattern := method + " " + fullPath
	routeHandler := r.wrap(handler, mws)
	finalHandler := r.wrap(routeHandler, r.middlewares)
	r.mux.Handle(pattern, finalHandler)

	newRoute := route{
		method:      method,
		path:        fullPath,
		handler:     handler,
		middlewares: mws,
	}

	r.routes = append(r.routes, newRoute)
}

// wrap applies a series of middlewares to an http.Handler in reverse order,
// so that the first middleware is the outermost wrapper around the handler.
func (r *Router) wrap(handler http.Handler, middlewares []Middleware) http.Handler {
	finalHandler := handler
	for i := len(middlewares) - 1; i >= 0; i-- {
		finalHandler = middlewares[i](finalHandler)
	}
	return finalHandler
}

// route describes a registered route, including its HTTP method, path pattern,
// the name of the associated handler and the applied middlewares.
type route struct {
	method, path string       // HTTP method and Path
	handler      http.Handler // handler
	middlewares  []Middleware // route-specific middlewares
}

// String returns a string representation of the registered route.
func (r route) String() string {
	return fmt.Sprintf("%s %s %s %s", r.method, r.path, handlerName(r.handler), middlewareNames(r.middlewares))
}

// handlerName returns the name of the function that implements the given http.Handler.
func handlerName(h http.Handler) string {
	fullFuncName := funcName(h)
	if handlerFunc, ok := h.(http.HandlerFunc); ok {
		fullFuncName = funcName(handlerFunc)
	}
	return trimRepoName(fullFuncName)
}

func funcName(f any) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

func trimRepoName(fn string) string {
	const sep = "/"
	parts := strings.Split(fn, sep)
	r := parts
	if len(parts) > 2 {
		r = parts[2:]
	}
	name := strings.Join(r, sep)
	return strings.TrimSpace(name)
}

func middlewareNames(mws []Middleware) []string {
	names := make([]string, len(mws))
	for _, mw := range mws {
		fullFuncName := runtime.FuncForPC(reflect.ValueOf(mw).Pointer()).Name()
		name := trimRepoName(fullFuncName)
		names = append(names, name)
	}
	return names
}

func normalizePath(p string) string {
	if p == "" {
		return "/"
	}

	p = path.Clean("/" + p)
	if p == "." {
		return "/"
	}

	return p
}
