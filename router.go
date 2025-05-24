// Package goexpress provides an http router implementation.
package goexpress

import (
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"strings"
)

// Route describes a registered route, including its HTTP method, path pattern,
// the name of the associated handler and the applied middlewares.
type Route struct {
	Method, Path string           // HTTP method and Path
	Handler      http.HandlerFunc // handler
	Middlewares  []func(http.Handler) http.Handler
}

// String returns a string representation of the registered route.
func (r Route) String() string {
	return fmt.Sprintf("%s %s %s %s", r.Method, r.Path, handlerName(r.Handler), middlewareNames(r.Middlewares))
}

// Router is a custom HTTP router built on top of http.ServeMux with support for global
// and route-specific middleware. It allows easy route registration for common HTTP methods
// (GET, POST, PATCH, PUT, DELETE) and provides a flexible middleware chain for request handling.
type Router struct {
	prefix      string                            // prefix for the paths of registered routes
	mux         *http.ServeMux                    // underlying HTTP request multiplexer
	routes      []Route                           // slice to store the registered routes
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

// wrap applies a series of middlewares to an http.Handler in reverse order,
// so that the first middleware is the outermost wrapper around the handler.
func (r *Router) wrap(handler http.Handler, middlewares []func(http.Handler) http.Handler) http.Handler {
	finalHandler := handler
	for i := len(middlewares) - 1; i >= 0; i-- {
		finalHandler = middlewares[i](finalHandler)
	}
	return finalHandler
}

// Handle registers a new route with a specific pattern and handler function.
// The global middlewares are applied to the route.
//
// Parameters:
//
//	pattern: URL pattern to match the route (e.g., "GET /path")
//	handler: http.Handler for handling requests to this route
//
// Example:
//
//	router.Handle("GET /static", staticFileHandler)
func (r *Router) Handle(pattern string, handler http.Handler) {
	finalHandler := r.wrap(handler, r.middlewares)
	r.mux.Handle(pattern, finalHandler)
}

// handleMethod registers a route with a specified HTTP method and path, applying
// any optional middlewares to the handler.
//
// Parameters:
//
//	method: HTTP method (e.g., "GET", "POST")
//	path: URL path for the route
//	handler: Handler function for the route
//	middlewares: Optional middlewares to apply to this specific route
//
// Example:
//
//	router.handleMethod("GET", "/info", infoHandler)
func (r *Router) handleMethod(method, path string, handler http.HandlerFunc, middlewares ...func(http.Handler) http.Handler) {
	fullPath := r.prefix + strings.TrimSuffix(path, "/")
	if fullPath == "" {
		fullPath = "/"
	}

	route := Route{
		Method:      method,
		Path:        fullPath,
		Handler:     handler,
		Middlewares: middlewares,
	}
	r.routes = append(r.routes, route)

	pattern := method + " " + fullPath
	finalHandler := r.wrap(handler, middlewares)
	r.Handle(pattern, finalHandler)
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
	r.handleMethod(http.MethodGet, path, handler, middlewares...)
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
	r.handleMethod(http.MethodPost, path, handler, middlewares...)
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
	r.handleMethod(http.MethodPatch, path, handler, middlewares...)
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
	r.handleMethod(http.MethodPut, path, handler, middlewares...)
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
	r.handleMethod(http.MethodDelete, path, handler, middlewares...)
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
	r.handleMethod(http.MethodConnect, path, handler, middlewares...)
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
	r.handleMethod(http.MethodOptions, path, handler, middlewares...)
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
	r.handleMethod(http.MethodTrace, path, handler, middlewares...)
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
	r.handleMethod(http.MethodHead, path, handler, middlewares...)
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

// ServeStatic serves static files from the specified local directory path.
// It registers a GET route to handle requests for static files and serves them
// relative to the provided path. The function applies an http.StripPrefix
// to remove the specified path prefix from incoming requests, allowing
// files to be directly accessed within the directory.
//
// Parameters:
//
//	path: The local directory path containing the static files to be served.
//
// Example:
//
//	r.ServeStatic("assets") // Serves files from the "assets" directory at "/assets/"
//
// This function constructs a GET route pattern using the specified path
// and registers it to the router, enabling clients to access static resources.
func (r *Router) ServeStatic(path string) {
	fullPath := "/" + path + "/"
	pattern := http.MethodGet + " " + fullPath
	r.Handle(pattern, http.StripPrefix(fullPath, http.FileServer(http.Dir(path))))
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
// an undefined route.
func (r *Router) NotFound(handler http.HandlerFunc) {
	r.mux.Handle("/", handler)
}

// Group creates a new route group with a common prefix and applies the
// given function to define sub-routes within that group.
//
// This method creates a new Router, passes it to the provided function
// for route definition, and then registers the grouped routes under the
// specified prefix. The routes within the group inherit the middlewares
// of the parent Router.
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
func (r *Router) Group(prefix string, fn func(router *Router), middlewares ...func(next http.Handler) http.Handler) {
	sub := r.SubRouter(prefix)
	sub.middlewares = append(sub.middlewares, middlewares...)

	fn(sub)

	r.routes = append(r.routes, sub.routes...)
}

// Mux returns the underlying http.Servemux.
func (r *Router) Mux() *http.ServeMux {
	return r.mux
}

// Routes returns a slice of all routes currently registered with the router.
func (r *Router) Routes() []Route {
	return r.routes
}

// Middlewares returns all the middlewares applied globally to the routes.
func (r *Router) Middlewares() []func(http.Handler) http.Handler {
	return r.middlewares
}

// SubRouter returns a new router which inherits the prefix, mux and middlewares of the parent router.
func (r *Router) SubRouter(prefix string) *Router {
	return &Router{
		prefix:      r.prefix + prefix,
		mux:         r.mux,
		middlewares: r.middlewares,
	}
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

// SetPrefix assigns a new prefix for the Router.
func (r *Router) SetPrefix(prefix string) {
	r.prefix = prefix
}

// SetMux assigns a new http.Servemux for the Router.
func (r *Router) SetMux(mux *http.ServeMux) {
	r.mux = mux
}

// SetMiddlewares assigns a new set of middlewares for the Router.
func (r *Router) SetMiddlewares(mws []func(http.Handler) http.Handler) {
	r.middlewares = mws
}

// handlerName returns the name of the function that implements the given http.Handler.
func handlerName(h http.Handler) string {
	fullFuncName := runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
	if handlerFunc, ok := h.(http.HandlerFunc); ok {
		fullFuncName = runtime.FuncForPC(reflect.ValueOf(handlerFunc).Pointer()).Name()
	}
	return trimRepoName(fullFuncName)
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
	sl := make([]string, len(mws))
	for _, mw := range mws {
		fullFuncName := runtime.FuncForPC(reflect.ValueOf(mw).Pointer()).Name()
		name := trimRepoName(fullFuncName)
		sl = append(sl, name)
	}
	return sl
}
