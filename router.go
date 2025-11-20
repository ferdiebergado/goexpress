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

// Router is a custom HTTP router built on top of http.ServeMux with support for global
// and route-specific middleware. It allows easy route registration for common HTTP methods
// (GET, POST, PATCH, PUT, DELETE) and provides a flexible middleware chain for request handling.
type Router struct {
	prefix      string                            // prefix for the paths of registered routes
	mux         *http.ServeMux                    // underlying HTTP request multiplexer
	routes      []route                           // slice to store the registered routes
	middlewares []func(http.Handler) http.Handler // slice to store global middleware functions
}

// New creates and returns a custom HTTP router that satisfies the Router interface with an initialized
// http.ServeMux and an empty slice for middlewares.
//
// Example:
//
//	router := goexpress.New()
//	router.Get("/hello", helloHandler) // Register GET route with handler
//	http.ListenAndServe(":8080", router) // Start server with Router
func New() *Router {
	return &Router{
		mux: http.NewServeMux(),
	}
}

// Use appends a middleware to the router's global middleware chain. Each middleware
// added with Use will be applied to every request handled by this Router.
//
// Parameters:
//
//	mw: Middleware function to be applied globally
//
// Example:
//
//	router.Use(logMiddleware)
func (r *Router) Use(mw func(next http.Handler) http.Handler) {
	r.middlewares = append(r.middlewares, mw)
}

// Get registers a new GET route for the specified path and handler, applying any optional middleware.
//
// Parameters:
//
//	path: URL path for the GET route
//	handler: Handler function for GET requests to the specified path
//	middlewares: Optional middlewares to apply to this specific route
//
// Example:
//
//	router.Get("/about", aboutHandler, authMiddleware)
func (r *Router) Get(path string, handler http.HandlerFunc, middlewares ...func(next http.Handler) http.Handler) {
	r.handle(http.MethodGet, path, handler, middlewares...)
}

// Post registers a new POST route for the specified path and handler, applying any optional middleware.
//
// Parameters:
//
//	path: URL path for the POST route
//	handler: Handler function for POST requests to the specified path
//	middlewares: Optional middlewares to apply to this specific route
//
// Example:
//
//	router.Post("/submit", submitHandler, csrfMiddleware)
func (r *Router) Post(path string, handler http.HandlerFunc, middlewares ...func(next http.Handler) http.Handler) {
	r.handle(http.MethodPost, path, handler, middlewares...)
}

// Patch registers a new PATCH route for the specified path and handler, applying any optional middleware.
//
// Parameters:
//
//	path: URL path for the PATCH route
//	handler: Handler function for PATCH requests to the specified path
//	middlewares: Optional middlewares to apply to this specific route
//
// Example:
//
//	router.Patch("/update", updateHandler, authMiddleware)
func (r *Router) Patch(path string, handler http.HandlerFunc, middlewares ...func(next http.Handler) http.Handler) {
	r.handle(http.MethodPatch, path, handler, middlewares...)
}

// Put registers a new PUT route for the specified path and handler, applying any optional middleware.
//
// Parameters:
//
//	path: URL path for the PUT route
//	handler: Handler function for PUT requests to the specified path
//	middlewares: Optional middlewares to apply to this specific route
//
// Example:
//
//	router.Put("/create", createHandler)
func (r *Router) Put(path string, handler http.HandlerFunc, middlewares ...func(next http.Handler) http.Handler) {
	r.handle(http.MethodPut, path, handler, middlewares...)
}

// Delete registers a new DELETE route for the specified path and handler, applying any optional middleware.
//
// Parameters:
//
//	path: URL path for the DELETE route
//	handler: Handler function for DELETE requests to the specified path
//	middlewares: Optional middlewares to apply to this specific route
//
// Example:
//
//	router.Delete("/remove", removeHandler, authMiddleware)
func (r *Router) Delete(path string, handler http.HandlerFunc, middlewares ...func(next http.Handler) http.Handler) {
	r.handle(http.MethodDelete, path, handler, middlewares...)
}

// Connect registers a new route that responds to HTTP CONNECT requests for the specified path.
// It applies the provided handler and any optional middlewares to the route.
// The CONNECT method is typically used for establishing a network connection to a web server.
//
// Parameters:
//   - path (string): The URL path for the route.
//   - handler (http.HandlerFunc): The handler function to execute when the route is matched.
//   - middlewares (...Middleware): Optional middlewares to apply to this route. These will be executed
//     before the handler function.
//
// Example:
//
//	r.Connect("/example", func(w http.ResponseWriter, r *http.Request) {
//	    // Handler implementation
//	})
func (r *Router) Connect(path string, handler http.HandlerFunc, middlewares ...func(next http.Handler) http.Handler) {
	r.handle(http.MethodConnect, path, handler, middlewares...)
}

// Options registers a new route that responds to HTTP OPTIONS requests for the specified path.
// It applies the provided handler and any optional middlewares to the route.
// The OPTIONS method is used to describe the communication options for the target resource.
//
// Parameters:
//   - path (string): The URL path for the route.
//   - handler (http.HandlerFunc): The handler function to execute when the route is matched.
//   - middlewares (...Middleware): Optional middlewares to apply to this route.
//
// Example:
//
//	r.Options("/example", func(w http.ResponseWriter, r *http.Request) {
//	    // Handler implementation
//	})
func (r *Router) Options(path string, handler http.HandlerFunc, middlewares ...func(next http.Handler) http.Handler) {
	r.handle(http.MethodOptions, path, handler, middlewares...)
}

// Trace registers a new route that responds to HTTP TRACE requests for the specified path.
// It applies the provided handler and any optional middlewares to the route.
// The TRACE method is used to perform a message loop-back test along the path to the target resource.
//
// Parameters:
//   - path (string): The URL path for the route.
//   - handler (http.HandlerFunc): The handler function to execute when the route is matched.
//   - middlewares (...Middleware): Optional middlewares to apply to this route.
//
// Example:
//
//	r.Trace("/example", func(w http.ResponseWriter, r *http.Request) {
//	    // Handler implementation
//	})
func (r *Router) Trace(path string, handler http.HandlerFunc, middlewares ...func(next http.Handler) http.Handler) {
	r.handle(http.MethodTrace, path, handler, middlewares...)
}

// Head registers a new route that responds to HTTP HEAD requests for the specified path.
// It applies the provided handler and any optional middlewares to the route.
// The HEAD method is used to retrieve the headers of a resource, without fetching the resource itself.
//
// Parameters:
//   - path (string): The URL path for the route.
//   - handler (http.HandlerFunc): The handler function to execute when the route is matched.
//   - middlewares (...Middleware): Optional middlewares to apply to this route.
//
// Example:
//
//	r.Head("/example", func(w http.ResponseWriter, r *http.Request) {
//	    // Handler implementation
//	})
func (r *Router) Head(path string, handler http.HandlerFunc, middlewares ...func(next http.Handler) http.Handler) {
	r.handle(http.MethodHead, path, handler, middlewares...)
}

// ServeHTTP enables the Router to satisfy the http.Handler interface.
// It delegates actual request handling to the underlying ServeMux after
// applying the middleware chain to registered handlers.
//
// Example:
//
//	http.ListenAndServe(":8080", router)
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

// Group creates a new route group with a common prefix and applies the
// given function to define sub-routes within that group.
//
// Routes can be specified just like with the normal router.
//
// Middlewares for the route group can also be specified as the last arguments.
//
// Nested route groups are also supported.
//
// Parameters:
//   - prefix: The common URL path prefix for the route group. It should
//     not have a trailing slash as it will be appended automatically.
//   - fn: A function that accepts a Router as an argument that defines the routes within the group.
//   - middlewares: middlewares to be applied to the route group (optional)
//
// Example:
//
//	r.Group("/api", func(r *Router) {
//	    r.Get("/users", usersHandler)
//	    r.Get("/products", productsHandler)
//	}, authMiddleware)
//
// This will register routes:
//
//	/api/users
//	/api/products
func (r *Router) Group(prefix string, fn func(router *Router), middlewares ...func(http.Handler) http.Handler) {
	sub := &Router{
		mux:         r.mux,
		prefix:      r.prefix + prefix,
		middlewares: append(r.middlewares, middlewares...),
	}

	fn(sub)

	r.routes = append(r.routes, sub.routes...)
}

// Static serves static files from the specified local directory path.
// It registers a GET route to handle requests for static files and serves them
// relative to the provided path.
//
// Parameters:
//
//	prefix: The request URL.
//	path: The local directory path containing the static files to be served.
//
// Example:
//
//	r.Static("assets", "./assets") // Serves files from the "assets" directory at "/assets/"
//
// This function constructs a GET route pattern using the specified path
// and registers it to the router, enabling clients to access static resources.
func (r *Router) Static(prefix, path string) {
	handler := http.StripPrefix(prefix, http.FileServer(http.Dir(path)))
	wrappedHandler := r.wrap(handler, r.middlewares)

	pattern := prefix
	if !strings.HasSuffix(pattern, "/") {
		pattern += "/"
	}

	r.mux.Handle(pattern, wrappedHandler)
}

// NotFound sets a custom handler for requests that don't match any registered route.
// When a request is made to an undefined route, this handler will be invoked,
// allowing a custom "Not Found" page or response to be returned.
//
// Parameters:
//   - handler: The http.HandlerFunc to handle "Not Found" responses.
//
// Example:
//
//	router := goexpress.New()
//	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
//	    http.Error(w, "Custom 404 - Page Not Found", http.StatusNotFound)
//	})
//
// This will display "Custom 404 - Page Not Found" when a request is made to
// an unregistered route.
func (r *Router) NotFound(handler http.HandlerFunc) {
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
//
// Parameters:
//
//	method: HTTP method (e.g., "GET", "POST")
//	path: URL path for the handle
//	handler: Handler function for the handle
//	middlewares: Optional middlewares to apply to this specific handle
//
// Example:
//
//	router.handle("GET", "/info", infoHandler)
func (r *Router) handle(method, path string, handler http.HandlerFunc, mws ...func(http.Handler) http.Handler) {
	fullPath := normalizePath(r.prefix + path)
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
func (r *Router) wrap(handler http.Handler, middlewares []func(http.Handler) http.Handler) http.Handler {
	finalHandler := handler
	for i := len(middlewares) - 1; i >= 0; i-- {
		finalHandler = middlewares[i](finalHandler)
	}
	return finalHandler
}

// route describes a registered route, including its HTTP method, path pattern,
// the name of the associated handler and the applied middlewares.
type route struct {
	method, path string                            // HTTP method and Path
	handler      http.HandlerFunc                  // handler
	middlewares  []func(http.Handler) http.Handler // route-specific middlewares
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

func middlewareNames(mws []func(http.Handler) http.Handler) []string {
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
