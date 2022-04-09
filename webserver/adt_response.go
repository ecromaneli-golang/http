package webserver

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

var EventStreamHeader = map[string][]string{
	ContentTypeHeader: {ContentTypeEventStream},
	"Cache-Control":   {"no-cache"},
	"Connection":      {"keep-alive"},
}

var contentTypesByExtension = map[string]string{
	".html": "text/html",
}

type Response struct {
	RawWriter http.ResponseWriter
	RawFS     http.FileSystem
	request   *Request
	flusher   http.Flusher
	views     map[string]string // TODO Implement map[string]any, use JSON serialization?
}

func newResponse(rw http.ResponseWriter, fs http.FileSystem, req *Request) *Response {
	return &Response{RawWriter: rw, RawFS: fs, request: req}
}

func (this *Response) Header(key, value string) *Response {
	this.RawWriter.Header().Add(key, value)
	return this
}

func (this *Response) Headers(headers map[string][]string) *Response {
	for name, values := range headers {
		for _, value := range values {
			this.Header(name, value)
		}
	}
	return this
}

func (this *Response) View(key string, value string) *Response {
	if this.views == nil {
		this.views = make(map[string]string)
	}

	this.views[key] = value
	return this
}

func (this *Response) Status(status int) *Response {
	this.RawWriter.WriteHeader(status)
	return this
}

func (this *Response) Render(filePath string) {
	file, err := this.RawFS.Open(filePath)

	var data []byte
	file.Read(data)
	file.Close()

	// TODO Analise better what status is, based on error
	panicIfNotNilUsingStatusCode(http.StatusNotFound, err)

	this.detectAndAddContentType(filePath).Write(this.replaceTokens(data))
}

func (this *Response) MustSupportFlusher() {
	if !this.SupportFlusher() {
		NewHTTPError(http.StatusNotImplemented, "Streaming Not Supported").Panic()
	}
}

func (this *Response) SupportFlusher() bool {
	flusher, ok := this.RawWriter.(http.Flusher)

	if !ok {
		return false
	}

	this.flusher = flusher
	return true
}

func (this *Response) FlushEvent(event *Event) error {
	return this.FlushText(event.ToString() + "\n\n")
}

func (this *Response) FlushText(text string) error {
	return this.Flush([]byte(text))
}

func (this *Response) Flush(data []byte) error {
	if this.request.IsDone() {
		return errors.New("The request is no more available")
	}

	if this.flusher == nil {
		this.MustSupportFlusher()
	}

	this.RawWriter.Write(data)
	this.flusher.Flush()
	return nil
}

func (this *Response) NoBody() {
	this.RawWriter.Write(nil)
}

func (this *Response) Write(data []byte) {
	this.RawWriter.Write(data)
}

func (this *Response) WriteJSON(value any) {
	if !this.hasContentType() {
		this.Header(ContentTypeHeader, "application/json")
	}
	json.NewEncoder(this.RawWriter).Encode(value)
}

func (this *Response) WriteText(text string) {
	this.Write([]byte(text))
}

func (this *Response) replaceTokens(file []byte) []byte {
	for token, value := range this.views {
		file = bytes.ReplaceAll(file, []byte("${"+token+"}"), []byte(value))
	}
	return file
}

func (this *Response) hasContentType() bool {
	return len(this.RawWriter.Header()[ContentTypeHeader]) > 0
}

func (this *Response) detectAndAddContentType(filePath string) *Response {
	for ext, ctype := range contentTypesByExtension {
		extIndex := strings.LastIndex(filePath, ext)
		if extIndex == -1 {
			continue
		}

		if extIndex+len(ext) != len(filePath) {
			continue
		}

		this.Header(ContentTypeHeader, ctype)
	}

	return this
}
