package webserver

import (
	"fmt"
	"net"
	"net/http"
	"time"
)

const (
	dateFormat = "2006-01-02 15:04:05.000 Z07:00"

	ContentTypeHeader = "Content-Type"

	ContentTypeFormUrlEncoded = "application/x-www-form-urlencoded"
	ContentTypeFormData       = "multipart/form-data"
	ContentTypeJson           = "application/json"
	ContentTypeEventStream    = "text/event-stream"
)

type Server struct {
	mux        *http.ServeMux
	fileSystem http.FileSystem
	routes     routesByPattern
}

type Handler func(req *Request, res *Response)

func NewServer() *Server {
	server := &Server{mux: http.NewServeMux()}

	server.routes = make(routesByPattern)
	return server
}

func NewServerWithFS(fileSystem http.FileSystem) *Server {
	router := NewServer()
	router.fileSystem = fileSystem
	return router
}

func ListenAndServe(addr string, handler Handler) error {
	return NewServer().All("/**", handler).ListenAndServe(addr)
}

func ListenAndServeTLS(addr, certFile, keyFile string, handler Handler) error {
	return NewServer().All("/**", handler).ListenAndServeTLS(addr, certFile, keyFile)
}

func Serve(l net.Listener, handler Handler) error {
	return NewServer().All("/**", handler).Serve(l)
}

func ServeTLS(l net.Listener, handler Handler, certFile string, keyFile string) error {
	return NewServer().All("/**", handler).ServeTLS(l, certFile, keyFile)
}

func (this *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, this.mux)
}

func (this *Server) ListenAndServeTLS(addr, certFile, keyFile string) error {
	return http.ListenAndServeTLS(addr, certFile, keyFile, this.mux)
}

func (this *Server) Serve(l net.Listener) error {
	return http.Serve(l, this.mux)
}

func (this *Server) ServeTLS(l net.Listener, certFile string, keyFile string) error {
	return http.ServeTLS(l, this.mux, certFile, keyFile)
}

// ================== HANDLERS ================== //

func (this *Server) HandleAll(pattern string, webserverHandler Handler) *Server {
	return this.MultiHandle(nil, pattern, webserverHandler)
}

func (this *Server) Handle(method string, pattern string, handler Handler) *Server {
	return this.MultiHandle([]string{method}, pattern, handler)
}

func (this *Server) MultiHandle(methods []string, pattern string, handler Handler) *Server {
	pattern, isNewStaticPattern := this.addRoute(methods, pattern, handler)

	if !isNewStaticPattern {
		return this
	}

	handlePattern := "/" + pattern

	handlerFunc := func(rw http.ResponseWriter, req *http.Request) {

		request := newRequest(req)
		response := newResponse(rw, this.fileSystem, request)
		request.response = response

		defer catchAllServerErrors(request, response)

		route, params := this.routes.getRoute(req.Method, pattern, request.Raw.Host, req.URL.EscapedPath())

		request.setPathParams(params)
		route.handler(request, response)
	}

	this.mux.HandleFunc(handlePattern, handlerFunc)

	if len(handlePattern) > 1 {
		this.mux.HandleFunc(handlePattern+"/", handlerFunc)
	}

	return this
}

func (this *Server) FileServerStrippingPrefix(pattern string, stripPrefix string) {
	handler := http.FileServer(this.fileSystem)

	if len(stripPrefix) > 0 {
		handler = http.StripPrefix(stripPrefix, handler)
	}

	this.mux.Handle(pattern, handler)
}

func (this *Server) FileServer(pattern string) {
	this.FileServerStrippingPrefix(pattern, "")
}

// ============== SHORCUT HANDLERS =============== //

func (this *Server) All(pattern string, webserverHandler Handler) *Server {
	return this.HandleAll(pattern, webserverHandler)
}

func (this *Server) Get(pattern string, webserverHandler Handler) *Server {
	return this.Handle(http.MethodGet, pattern, webserverHandler)
}

func (this *Server) Post(pattern string, webserverHandler Handler) *Server {
	return this.Handle(http.MethodPost, pattern, webserverHandler)
}

func (this *Server) Put(pattern string, webserverHandler Handler) *Server {
	return this.Handle(http.MethodPut, pattern, webserverHandler)
}

func (this *Server) Delete(pattern string, webserverHandler Handler) *Server {
	return this.Handle(http.MethodDelete, pattern, webserverHandler)
}

func (this *Server) Render(pattern string, filePath string) *Server {
	return this.Get(pattern, func(req *Request, res *Response) { res.Render(filePath) })
}

func (this *Server) Write(pattern string, data []byte) *Server {
	return this.Get(pattern, func(req *Request, res *Response) { res.Write(data) })
}

func (this *Server) WriteText(pattern string, text string) *Server {
	return this.Get(pattern, func(req *Request, res *Response) { res.WriteText(text) })
}

func (this *Server) WriteJSON(pattern string, filePath string) *Server {
	return this.Get(pattern, func(req *Request, res *Response) { res.WriteJSON(filePath) })
}

func (this *Server) addRoute(methods []string, pattern string, handler Handler) (rootPattern string, isNewStaticPattern bool) {
	route := this.routes.Add(methods, pattern, handler)
	return route.staticPattern, len(this.routes[route.staticPattern]) == 1
}

func catchAllServerErrors(req *Request, res *Response) {
	err := recover()
	if err == nil {
		return
	}

	var customErr *serverError

	switch err.(type) {
	case *serverError:
		customErr = err.(*serverError)
	default:
		customErr = NewError(err)
	}

	if !req.IsDone() {
		res.Status(customErr.statusCode).WriteText(customErr.message)
	}

	fmt.Println(time.Now().Format(dateFormat), "- ERROR webserver:", customErr.Error())
}
