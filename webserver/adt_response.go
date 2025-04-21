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

const sseSeparator = "\n\n"

// Response represents an HTTP response with enhanced functionality.
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

// Header adds a header to the response.
//
// The key and value parameters specify the header name and value to add.
//
// Returns the response instance for method chaining.
func (r *Response) Header(key, value string) *Response {
	r.RawWriter.Header().Add(key, value)
	return r
}

// Headers adds multiple headers to the response.
//
// The headers parameter is a map of header names to header values.
//
// Returns the response instance for method chaining.
func (r *Response) Headers(headers map[string][]string) *Response {
	for name, values := range headers {
		for _, value := range values {
			r.Header(name, value)
		}
	}
	return r
}

// View adds a view variable for template rendering.
//
// The key parameter is the variable name, and value is the content to replace in templates.
//
// Returns the response instance for method chaining.
func (r *Response) View(key string, value string) *Response {
	if r.views == nil {
		r.views = make(map[string]string)
	}

	r.views[key] = value
	return r
}

// Status sets the HTTP status code for the response.
//
// The status parameter is the HTTP status code to set.
//
// Returns the response instance for method chaining.
func (r *Response) Status(status int) *Response {
	r.RawWriter.WriteHeader(status)
	return r
}

// Render reads a file from the file system and writes it to the response with template processing.
//
// The filePath parameter specifies the path to the file to render.
//
// Panics if the file cannot be found or read.
func (r *Response) Render(filePath string) {
	file, err := r.RawFS.Open(filePath)

	var data []byte
	file.Read(data)
	file.Close()

	// TODO Analise better what status is, based on error
	panicIfNotNilUsingStatusCode(http.StatusNotFound, err)

	r.detectAndAddContentType(filePath).Write(r.replaceTokens(data))
}

// MustSupportFlusher checks if the underlying ResponseWriter supports flushing.
//
// Panics if flushing is not supported.
func (r *Response) MustSupportFlusher() {
	if !r.SupportFlusher() {
		NewHTTPError(http.StatusNotImplemented, "Streaming Not Supported").Panic()
	}
}

// SupportFlusher checks if the underlying ResponseWriter supports flushing.
//
// Returns true if flushing is supported, false otherwise.
func (r *Response) SupportFlusher() bool {
	flusher, ok := r.RawWriter.(http.Flusher)

	if !ok {
		return false
	}

	r.flusher = flusher
	return true
}

// FlushEvent writes a server-sent event to the response and flushes it immediately.
//
// The event parameter is the event to send.
//
// Returns an error if the request is done or flushing is not supported.
func (r *Response) FlushEvent(event *Event) error {
	return r.FlushText(event.ToString() + sseSeparator)
}

// FlushText writes text to the response and flushes it immediately.
//
// The text parameter is the content to write.
//
// Returns an error if the request is done or flushing is not supported.
func (r *Response) FlushText(text string) error {
	return r.Flush([]byte(text))
}

// Flush writes data to the response and flushes it immediately.
//
// The data parameter is the content to write.
//
// Returns an error if the request is done or flushing is not supported.
func (r *Response) Flush(data []byte) error {
	if r.request.IsDone() {
		return errors.New("request is no longer available")
	}

	if r.flusher == nil {
		r.MustSupportFlusher()
	}

	r.RawWriter.Write(data)
	r.flusher.Flush()
	return nil
}

// NoBody writes an empty response.
func (r *Response) NoBody() {
	r.RawWriter.Write(nil)
}

// Write writes binary data to the response.
//
// The data parameter is the content to write.
func (r *Response) Write(data []byte) {
	r.RawWriter.Write(data)
}

// WriteJSON serializes a value to JSON and writes it to the response.
//
// The value parameter is the object to serialize as JSON.
//
// Sets the Content-Type header to application/json if not already set.
func (r *Response) WriteJSON(value any) {
	if !r.hasContentType() {
		r.Header(ContentTypeHeader, ContentTypeJson)
	}
	json.NewEncoder(r.RawWriter).Encode(value)
}

// WriteText writes text to the response.
//
// The text parameter is the content to write.
func (r *Response) WriteText(text string) {
	r.Write([]byte(text))
}

func (r *Response) replaceTokens(file []byte) []byte {
	for token, value := range r.views {
		file = bytes.ReplaceAll(file, []byte("${"+token+"}"), []byte(value))
	}
	return file
}

func (r *Response) hasContentType() bool {
	return len(r.RawWriter.Header()[ContentTypeHeader]) > 0
}

func (r *Response) detectAndAddContentType(filePath string) *Response {
	for ext, ctype := range contentTypesByExtension {
		extIndex := strings.LastIndex(filePath, ext)
		if extIndex == -1 {
			continue
		}

		if extIndex+len(ext) != len(filePath) {
			continue
		}

		r.Header(ContentTypeHeader, ctype)
	}

	return r
}
