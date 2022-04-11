package webserver

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

const (
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
	host       string
	port       string
}

type Handler func(req *Request, res *Response)

func NewServer(addr string) *Server {
	server := &Server{mux: http.NewServeMux()}

	server.setAddr(addr)
	server.routes = make(routesByPattern)
	return server
}

func NewServerWithFS(addr string, fileSystem http.FileSystem) *Server {
	router := NewServer(addr)
	router.fileSystem = fileSystem
	return router
}

func ListenAndServe(addr string, handler Handler) error {
	return NewServer(addr).All("/**", handler).ListenAndServe()
}

func ListenAndServeTLS(addr string, certFile string, keyFile string, handler Handler) error {
	return NewServer(addr).All("/**", handler).ListenAndServeTLS(certFile, keyFile)
}

func Serve(l net.Listener, handler Handler) error {
	return NewServer("").All("/**", handler).Serve(l)
}

func ServeTLS(l net.Listener, handler Handler, certFile string, keyFile string) error {
	return NewServer("").All("/**", handler).ServeTLS(l, certFile, keyFile)
}

func (this *Server) ListenAndServe() error {
	return http.ListenAndServe(this.getAddr(), this.mux)
}

func (this *Server) ListenAndServeTLS(certFile, keyFile string) error {
	return http.ListenAndServeTLS(this.getAddr(), certFile, keyFile, this.mux)
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
	pattern, isNewRootPattern := this.addRoute(methods, pattern, handler)

	if !isNewRootPattern {
		return this
	}

	handlePattern := pattern + "/"

	if len(handlePattern) > 1 {
		handlePattern = "/" + handlePattern
	}

	this.mux.HandleFunc(this.host+handlePattern, func(rw http.ResponseWriter, req *http.Request) {

		request := newRequest(req)
		response := newResponse(rw, this.fileSystem, request)
		request.response = response

		route, params, status := this.routes.getRoute(req.Method, pattern, req.URL.EscapedPath())

		defer catchAllServerErrors(request, response)

		if status == 0 {
			request.setPathParams(params)
			route.handler(request, response)
		} else {
			NewHTTPError(status, nil).Panic()
		}
	})

	return this
}

func (this *Server) FileServerStrippingPrefix(pattern string, stripPrefix string) {
	handler := http.FileServer(this.fileSystem)

	if len(stripPrefix) > 0 {
		handler = http.StripPrefix(stripPrefix, handler)
	}

	this.mux.Handle(this.host+pattern, handler)
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

// =================== HELPERS ================== //

func (this *Server) addRoute(methods []string, pattern string, handler Handler) (rootPattern string, isNewRootPattern bool) {
	route := this.routes.Add(methods, pattern, handler)
	return route.staticPart, len(this.routes[route.staticPart]) == 1
}

func (this *Server) getAddr() string {
	return this.host + ":" + this.port
}

func (this *Server) setAddr(addr string) {
	data := strings.Split(addr, ":")

	this.host = data[0]

	if len(data) > 1 {
		this.port = data[1]
	}
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
		customErr = newError(err)
	}

	if !req.isDone {
		res.Status(customErr.StatusCode).WriteText(customErr.Message)
	}

	fmt.Println("[WebServer]", time.Now().Format(time.RFC3339), "- ERROR", customErr.Log)
}
