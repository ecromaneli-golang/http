package webserver

import (
	"net"
	"net/http"

	"github.com/ecromaneli-golang/console/logger"
)

const (
	// ContentTypeHeader is the HTTP header name for content type.
	ContentTypeHeader = "Content-Type"

	// ContentTypeFormUrlEncoded represents the URL-encoded form content type.
	ContentTypeFormUrlEncoded = "application/x-www-form-urlencoded"

	// ContentTypeFormData represents the multipart form data content type.
	ContentTypeFormData = "multipart/form-data"

	// ContentTypeJson represents the JSON content type.
	ContentTypeJson = "application/json"

	// ContentTypeEventStream represents the event stream content type.
	ContentTypeEventStream = "text/event-stream"

	// WildcardPattern is the pattern used to match all routes.
	WildcardPattern = "/**"
)

// Server represents an HTTP server with routing capabilities.
type Server struct {
	mux        *http.ServeMux
	fileSystem http.FileSystem
	routes     routesByPattern
	logger     *logger.Logger
}

// Handler defines the signature for HTTP request handlers.
type Handler func(req *Request, res *Response)

// NewServer creates a new server instance with default configuration.
func NewServer() *Server {
	server := &Server{mux: http.NewServeMux(), logger: logger.New("webserver")}
	server.routes = make(routesByPattern)
	return server
}

// NewServerWithFS creates a new server with a custom file system for serving static files.
func NewServerWithFS(fileSystem http.FileSystem) *Server {
	router := NewServer()
	router.fileSystem = fileSystem
	return router
}

// ListenAndServe creates a new server with the given handler for all routes and listens on the specified address.
//
// It blocks until the server is stopped or an error occurs.
func ListenAndServe(addr string, handler Handler) error {
	return NewServer().All(WildcardPattern, handler).ListenAndServe(addr)
}

// ListenAndServeTLS creates a new server with the given handler for all routes and
// listens on the specified address using TLS with the provided certificate and key files.
//
// It blocks until the server is stopped or an error occurs.
func ListenAndServeTLS(addr, certFile, keyFile string, handler Handler) error {
	return NewServer().All(WildcardPattern, handler).ListenAndServeTLS(addr, certFile, keyFile)
}

// Serve creates a new server with the given handler for all routes and serves requests on
// the provided listener.
//
// It blocks until the listener is closed.
func Serve(l net.Listener, handler Handler) error {
	return NewServer().All(WildcardPattern, handler).Serve(l)
}

// ServeTLS creates a new server with the given handler for all routes and serves HTTPS requests on
// the provided listener using the provided certificate and key files.
//
// It blocks until the listener is closed.
func ServeTLS(l net.Listener, handler Handler, certFile string, keyFile string) error {
	return NewServer().All(WildcardPattern, handler).ServeTLS(l, certFile, keyFile)
}

// ListenAndServe starts the server on the specified address.
//
// It blocks until the server is stopped or an error occurs.
func (s *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s.mux)
}

// ListenAndServeTLS starts the server with TLS enabled on the specified address
// using the provided certificate and key files.
//
// It blocks until the server is stopped or an error occurs.
func (s *Server) ListenAndServeTLS(addr, certFile, keyFile string) error {
	return http.ListenAndServeTLS(addr, certFile, keyFile, s.mux)
}

// Serve accepts incoming connections on the provided listener and handles
// requests using the server's handler.
//
// It blocks until the listener is closed.
func (s *Server) Serve(l net.Listener) error {
	return http.Serve(l, s.mux)
}

// ServeTLS accepts incoming connections on the provided listener and handles
// requests using TLS and the server's handler.
//
// It blocks until the listener is closed.
func (s *Server) ServeTLS(l net.Listener, certFile string, keyFile string) error {
	return http.ServeTLS(l, s.mux, certFile, keyFile)
}

// HandleAll registers a handler for all HTTP methods on the specified pattern.
//
// Returns the server instance for method chaining.
func (s *Server) HandleAll(pattern string, webserverHandler Handler) *Server {
	return s.MultiHandle(nil, pattern, webserverHandler)
}

// Handle registers a handler for a specific HTTP method and pattern.
//
// Returns the server instance for method chaining.
func (s *Server) Handle(method string, pattern string, handler Handler) *Server {
	return s.MultiHandle([]string{method}, pattern, handler)
}

// MultiHandle registers a handler for multiple HTTP methods and a specific pattern.
// If methods is nil or empty, the handler will respond to all HTTP methods.
//
// Returns the server instance for method chaining.
func (s *Server) MultiHandle(methods []string, pattern string, handler Handler) *Server {
	if s.logger.IsEnabled(logger.LevelTrace) {
		s.logger.Trace("MultiHandle(methods=\"", methods, "\", pattern=\""+pattern+"\", handler)")
	}

	pattern, isNewStaticPattern := s.addRoute(methods, pattern, handler)

	if !isNewStaticPattern {
		return s
	}

	s.RegisterRouteHandler(pattern)

	if s.logger.IsEnabled(logger.LevelDebug) {
		s.logger.Debug("Route added: \"", methods, " /", pattern, "\"")
	}

	return s
}

// RegisterRouteHandler registers an HTTP handler function for the given pattern
// with the server's internal multiplexer.
//
// Returns the server instance for method chaining.
func (s *Server) RegisterRouteHandler(pattern string) *Server {
	handlerFunc := s.createHandlerFunc(pattern)

	handlePattern := "/" + pattern
	s.mux.HandleFunc(handlePattern, handlerFunc)

	if pattern != "" {
		s.mux.HandleFunc(handlePattern+"/", handlerFunc)
	}

	return s
}

func (s *Server) createHandlerFunc(pattern string) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if s.logger.IsEnabled(logger.LevelTrace) {
			s.logger.Trace(getRemoteAddr(req), " - ", req.Method, " ", req.Host+req.URL.Path)
		}

		request := newRequest(req)
		response := newResponse(rw, s.fileSystem, request)
		request.response = response

		defer s.catchAllServerErrors(request, response)

		route, params, err := s.routes.getRoute(req.Method, pattern, request.Raw.Host, req.URL.EscapedPath())
		if err == nil {
			request.setPathParams(params)
			route.handler(request, response)
			return
		}

		response.Status(err.statusCode).NoBody()
		if s.logger.IsEnabled(logger.LevelDebug) {
			s.logger.Debug(getRemoteAddr(req), " - ", err.Error())
		}
	}
}

// FileServerStrippingPrefix serves files from the server's file system at the specified pattern.
// If stripPrefix is provided, it will be removed from the requested URL path before serving files.
func (s *Server) FileServerStrippingPrefix(pattern string, stripPrefix string) {
	handler := http.FileServer(s.fileSystem)

	if stripPrefix != "" {
		handler = http.StripPrefix(stripPrefix, handler)
	}

	s.mux.Handle(pattern, handler)
}

// FileServer serves files from the server's file system at the root path.
func (s *Server) FileServer() {
	s.FileServerStrippingPrefix("/", "")
}

// All is a shortcut for HandleAll that registers a handler for all HTTP methods on the specified pattern.
//
// Returns the server instance for method chaining.
func (s *Server) All(pattern string, webserverHandler Handler) *Server {
	return s.HandleAll(pattern, webserverHandler)
}

// Get registers a handler for HTTP GET requests on the specified pattern.
//
// Returns the server instance for method chaining.
func (s *Server) Get(pattern string, webserverHandler Handler) *Server {
	return s.Handle(http.MethodGet, pattern, webserverHandler)
}

// Post registers a handler for HTTP POST requests on the specified pattern.
//
// Returns the server instance for method chaining.
func (s *Server) Post(pattern string, webserverHandler Handler) *Server {
	return s.Handle(http.MethodPost, pattern, webserverHandler)
}

// Put registers a handler for HTTP PUT requests on the specified pattern.
//
// Returns the server instance for method chaining.
func (s *Server) Put(pattern string, webserverHandler Handler) *Server {
	return s.Handle(http.MethodPut, pattern, webserverHandler)
}

// Delete registers a handler for HTTP DELETE requests on the specified pattern.
//
// Returns the server instance for method chaining.
func (s *Server) Delete(pattern string, webserverHandler Handler) *Server {
	return s.Handle(http.MethodDelete, pattern, webserverHandler)
}

// Render registers a GET handler that renders the specified file.
//
// Returns the server instance for method chaining.
func (s *Server) Render(pattern string, filePath string) *Server {
	return s.Get(pattern, func(req *Request, res *Response) { res.Render(filePath) })
}

// Write registers a GET handler that writes the specified bytes to the response.
//
// Returns the server instance for method chaining.
func (s *Server) Write(pattern string, data []byte) *Server {
	return s.Get(pattern, func(req *Request, res *Response) { res.Write(data) })
}

// WriteText registers a GET handler that writes the specified text to the response.
//
// Returns the server instance for method chaining.
func (s *Server) WriteText(pattern string, text string) *Server {
	return s.Get(pattern, func(req *Request, res *Response) { res.WriteText(text) })
}

// WriteJSON registers a GET handler that writes the specified JSON file to the response.
//
// Returns the server instance for method chaining.
func (s *Server) WriteJSON(pattern string, filePath string) *Server {
	return s.Get(pattern, func(req *Request, res *Response) { res.WriteJSON(filePath) })
}

func (s *Server) addRoute(methods []string, pattern string, handler Handler) (rootPattern string, isNewStaticPattern bool) {
	route := s.routes.Add(methods, pattern, handler)
	return route.staticPattern, len(s.routes[route.staticPattern]) == 1
}

// SetLogLevel sets the logging level for the server.
//
// The level can be one of the following:
//   - ALL
//   - TRACE
//   - DEBUG
//   - INFO
//   - WARN
//   - ERROR
//   - FATAL
//   - OFF
func (s *Server) SetLogLevel(level string) {
	s.logger.SetLogLevelStr(level)
}

func (s *Server) catchAllServerErrors(req *Request, res *Response) {
	if err := recover(); err != nil {
		var customErr *serverError
		switch e := err.(type) {
		case *serverError:
			customErr = e
		default:
			customErr = NewError(err)
		}

		if !req.IsDone() {
			res.Status(customErr.statusCode).WriteText(customErr.message)
		}

		s.logger.Error(customErr.Error())
	}
}

// getRemoteAddr extracts the client IP address from the request.
// It checks X-Real-Ip and X-Forwarded-For headers before falling back to RemoteAddr.
func getRemoteAddr(req *http.Request) string {
	ipAddress := req.Header.Get("X-Real-Ip")

	if ipAddress == "" {
		ipAddress = req.Header.Get("X-Forwarded-For")
	}
	if ipAddress == "" {
		ipAddress = req.RemoteAddr
	}

	return ipAddress
}
