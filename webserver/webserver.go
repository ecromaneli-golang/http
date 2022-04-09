package webserver

import (
	"fmt"
	"net/http"
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
	http.ServeMux
	fileSystem http.FileSystem
	routes     routesByPattern
}

type Handler func(req *Request, res *Response)

func NewServer() *Server {
	router := new(Server)
	router.routes = make(routesByPattern)
	return router
}

func NewServerWithFS(fileSystem http.FileSystem) *Server {
	router := NewServer()
	router.fileSystem = fileSystem
	return router
}

func ListenAndServe(addr string, handler Handler) error {
	return http.ListenAndServe(addr, NewServer().All("/**", handler))
}

func ListenAndServeTLS(addr string, certFile string, keyFile string, handler Handler) error {
	return http.ListenAndServeTLS(addr, certFile, keyFile, NewServer().All("/**", handler))
}

func (this *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, this)
}

func (this *Server) ListenAndServeTLS(addr, certFile, keyFile string) error {
	return http.ListenAndServeTLS(addr, certFile, keyFile, this)
}

// ================== HANDLERS ================== //

func (this *Server) HandleAll(pattern string, webserverHandler func(req *Request, res *Response)) *Server {
	return this.MultiHandle(nil, pattern, webserverHandler)
}

func (this *Server) Handle(method string, pattern string, handler func(req *Request, res *Response)) *Server {
	return this.MultiHandle([]string{method}, pattern, handler)
}

func (this *Server) MultiHandle(methods []string, pattern string, handler func(req *Request, res *Response)) *Server {
	pattern, isNewRootPattern := this.addRoute(methods, pattern, handler)

	if !isNewRootPattern {
		return this
	}

	handlePattern := pattern + "/"

	if len(handlePattern) > 1 {
		handlePattern = "/" + handlePattern
	}

	this.HandleFunc(handlePattern, func(rw http.ResponseWriter, req *http.Request) {

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

	this.ServeMux.Handle(pattern, handler)
}

func (this *Server) FileServer(pattern string) {
	this.FileServerStrippingPrefix(pattern, "")
}

// ============== SHORCUT HANDLERS =============== //

func (this *Server) All(pattern string, webserverHandler func(req *Request, res *Response)) *Server {
	return this.HandleAll(pattern, webserverHandler)
}

func (this *Server) Get(pattern string, webserverHandler func(req *Request, res *Response)) *Server {
	return this.Handle(http.MethodGet, pattern, webserverHandler)
}

func (this *Server) Post(pattern string, webserverHandler func(req *Request, res *Response)) *Server {
	return this.Handle(http.MethodPost, pattern, webserverHandler)
}

func (this *Server) Put(pattern string, webserverHandler func(req *Request, res *Response)) *Server {
	return this.Handle(http.MethodPut, pattern, webserverHandler)
}

func (this *Server) Delete(pattern string, webserverHandler func(req *Request, res *Response)) *Server {
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
