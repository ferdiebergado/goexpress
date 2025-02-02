package goexpress

import (
	"net/http"
	"strings"
)

// Router is a custom HTTP router built on top of http.ServeMux with support for global
// and route-specific middleware. It allows easy route registration for common HTTP methods
// (GET, POST, PATCH, PUT, DELETE) and provides a flexible middleware chain for request handling.
type Router struct {
	mux         *http.ServeMux // underlying HTTP request multiplexer
	middlewares []Middleware   // slice to store global middleware functions
}

// Middleware defines the signature for middleware functions.
// A Middleware function takes an http.Handler as input and returns a new http.Handler
// that wraps additional functionality around the original handler.
//
// Example:
//
//	func logMiddleware(next http.Handler) http.Handler {
//	    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//	        log.Println("Request received:", r.URL.Path)
//	        next.ServeHTTP(w, r)
//	    })
//	}
type Middleware func(http.Handler) http.Handler

// New creates and returns a new instance of Router with an initialized
// http.ServeMux and an empty slice for middlewares.
//
// Example usage:
//
//	router := New()
//	router.Get("/hello", helloHandler) // Register GET route with handler
//	http.ListenAndServe(":8080", router) // Start server with Router
func New() *Router {
	return &Router{
		mux:         http.NewServeMux(),
		middlewares: make([]Middleware, 0),
	}
}

// Use appends a middleware to the router's global middleware chain. Each middleware
// added with Use will be applied to every request handled by this Router.
//
// Parameters:
//
//	mw: Middleware function to be applied globally
//
// Example usage:
//
//	router.Use(logMiddleware)
func (r *Router) Use(mw Middleware) {
	r.middlewares = append(r.middlewares, mw)
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

// Handle registers a new route with a specific pattern and handler function, applying
// optional route-specific middleware. The route-specific middleware will wrap around
// the handler before the global middleware is applied.
//
// Parameters:
//
//	pattern: URL pattern to match the route (e.g., "GET /path")
//	handler: http.Handler for handling requests to this route
//	middlewares: Optional route-specific middleware to apply before global middleware
//
// Example usage:
//
//	router.Handle("GET /static", staticFileHandler, authMiddleware)
func (r *Router) Handle(pattern string, handler http.Handler, middlewares ...Middleware) {
	finalHandler := r.wrap(handler, middlewares)       // Wrap with route-specific middleware
	finalHandler = r.wrap(finalHandler, r.middlewares) // Wrap with global middleware
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
// Example usage:
//
//	router.handleMethod("GET", "/info", infoHandler)
func (r *Router) handleMethod(method, path string, handler http.HandlerFunc, middlewares ...Middleware) {
	if path == "" {
		path = "/"
	} else if path != "/" {
		path = strings.TrimSuffix(path, "/")
	}

	pattern := method + " " + path
	r.Handle(pattern, handler, middlewares...)
}

// Get registers a new GET route for the specified path and handler, applying any optional middleware.
//
// Parameters:
//
//	path: URL path for the GET route
//	handler: Handler function for GET requests to the specified path
//	middlewares: Optional middlewares to apply to this specific route
//
// Example usage:
//
//	router.Get("/about", aboutHandler, authMiddleware)
func (r *Router) Get(path string, handler http.HandlerFunc, middlewares ...Middleware) {
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
// Example usage:
//
//	router.Post("/submit", submitHandler, csrfMiddleware)
func (r *Router) Post(path string, handler http.HandlerFunc, middlewares ...Middleware) {
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
// Example usage:
//
//	router.Patch("/update", updateHandler, authMiddleware)
func (r *Router) Patch(path string, handler http.HandlerFunc, middlewares ...Middleware) {
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
// Example usage:
//
//	router.Put("/create", createHandler)
func (r *Router) Put(path string, handler http.HandlerFunc, middlewares ...Middleware) {
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
// Example usage:
//
//	router.Delete("/remove", removeHandler, authMiddleware)
func (r *Router) Delete(path string, handler http.HandlerFunc, middlewares ...Middleware) {
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
// Example usage:
//
//	r.Connect("/example", func(w http.ResponseWriter, r *http.Request) {
//	    // Handler implementation
//	})
func (r *Router) Connect(path string, handler http.HandlerFunc, middlewares ...Middleware) {
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
// Example usage:
//
//	r.Options("/example", func(w http.ResponseWriter, r *http.Request) {
//	    // Handler implementation
//	})
func (r *Router) Options(path string, handler http.HandlerFunc, middlewares ...Middleware) {
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
// Example usage:
//
//	r.Trace("/example", func(w http.ResponseWriter, r *http.Request) {
//	    // Handler implementation
//	})
func (r *Router) Trace(path string, handler http.HandlerFunc, middlewares ...Middleware) {
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
// Example usage:
//
//	r.Head("/example", func(w http.ResponseWriter, r *http.Request) {
//	    // Handler implementation
//	})
func (r *Router) Head(path string, handler http.HandlerFunc, middlewares ...Middleware) {
	r.handleMethod(http.MethodHead, path, handler, middlewares...)
}

// ServeHTTP enables the Router to satisfy the http.Handler interface.
// It delegates actual request handling to the underlying ServeMux after
// applying the middleware chain to registered handlers.
//
// Example usage:
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
// Example usage:
//
//	router := NewRouter()
//	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
//	    http.Error(w, "Custom 404 - Page Not Found", http.StatusNotFound)
//	})
//
// This will display "Custom 404 - Page Not Found" when a request is made to
// an undefined route.
func (r *Router) NotFound(handler http.HandlerFunc) {
	r.mux.Handle("/", handler)
}

// groupHandler is a function type that takes a pointer to a Router
// and returns a modified Router. It is used to define sub-routes
// within a group.
//
// Example:
//
//	apiRouter := func(r *Router) *Router {
//	    r.Handle("/users", usersHandler)
//	    r.Handle("/products", productsHandler)
//	    return r
//	}
type groupHandler func(*Router) *Router

// Group creates a new route group with a common prefix and applies the
// given routerFunc to define sub-routes within that group.
//
// This method creates a new Router, passes it to the provided routerFunc
// for route definition, and then registers the grouped routes under the
// specified prefix. The routes within the group inherit the middlewares
// of the parent Router.
//
// Routes can be specified just like with the normal router meaning middlewares can also be included.
//
// Middlewares for the route group can also be specified as the last arguments.
//
// Nested route groups are also supported.
//
// Parameters:
//   - prefix: The common URL path prefix for the route group. It should
//     not have a trailing slash as it will be appended automatically.
//   - h: A groupHandler that defines the routes within the group.
//
// Example:
//
//	r.Group("/api", func(r *Router) *Router {
//	    r.Get("/users", usersHandler)
//	    r.Get("/products", productsHandler)
//	    return r
//	}, authMiddleware)
//
// This will register routes:
//
//	/api/users
//	/api/products
func (r *Router) Group(prefix string, h groupHandler, m ...Middleware) {
	router := h(New())

	r.Handle(prefix+"/", http.StripPrefix(prefix, router.mux), m...)
}
