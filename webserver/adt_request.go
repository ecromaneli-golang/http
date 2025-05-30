package webserver

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Request represents an HTTP request with extended functionality.
type Request struct {
	Raw        *http.Request
	response   *Response
	params     map[string][]string
	files      map[string][]*multipart.FileHeader
	body       []byte
	readParams bool
	readBody   bool
	isDone     bool
}

func newRequest(req *http.Request) *Request {
	return &Request{Raw: req}
}

// AllHeaders returns all HTTP headers from the request.
//
// Returns the complete set of headers as an http.Header map.
func (r *Request) AllHeaders() http.Header {
	return r.Raw.Header
}

// Headers returns all values for a specific HTTP header.
//
// The name parameter specifies which header to retrieve. Returns an empty slice
// if the header doesn't exist.
func (r *Request) Headers(name string) []string {
	r.parseParams()
	return r.AllHeaders()[name]
}

// Header returns the first value for a specific HTTP header.
//
// The name parameter specifies which header to retrieve. Returns an empty string
// if the header doesn't exist.
func (r *Request) Header(name string) string {
	r.parseParams()

	header := r.Headers(name)

	if len(header) == 0 {
		return ""
	}

	return header[0]
}

// AllParams returns all request parameters combined from URL, form, and path.
//
// Returns a map of parameter names to their values.
func (r *Request) AllParams() map[string][]string {
	r.parseParams()
	return r.params
}

// Params returns all values for a specific parameter.
//
// The paramName specifies which parameter to retrieve. Returns an empty slice
// if the parameter doesn't exist.
func (r *Request) Params(paramName string) []string {
	r.parseParams()
	return r.params[paramName]
}

// Param returns the first value for a specific parameter.
//
// The paramName specifies which parameter to retrieve. Returns an empty string
// if the parameter doesn't exist.
func (r *Request) Param(paramName string) string {
	r.parseParams()

	param := r.params[paramName]

	if len(param) == 0 {
		return ""
	}

	return param[0]
}

// AllFiles returns all uploaded files from a multipart form request.
//
// Returns a map of field names to file headers.
func (r *Request) AllFiles() map[string][]*multipart.FileHeader {
	r.parseParams()
	return r.files
}

// Files returns all uploaded files for a specific form field.
//
// The paramName specifies which form field to retrieve files from. Returns an empty
// slice if no files were uploaded under that name.
func (r *Request) Files(paramName string) []*multipart.FileHeader {
	r.parseParams()
	return r.files[paramName]
}

// File returns the first uploaded file for a specific form field.
//
// The paramName specifies which form field to retrieve a file from. Returns nil
// if no files were uploaded under that name.
func (r *Request) File(paramName string) *multipart.FileHeader {
	r.parseParams()

	files := r.files[paramName]

	if len(files) == 0 {
		return nil
	}

	return files[0]
}

// UIntParam returns a parameter value converted to an unsigned integer.
//
// The paramName specifies which parameter to retrieve and convert. Returns 0
// if the parameter doesn't exist or cannot be converted.
func (r *Request) UIntParam(paramName string) uint {
	return uint(r.IntParam(paramName))
}

// IntParam returns a parameter value converted to an integer.
//
// The paramName specifies which parameter to retrieve and convert. Returns 0
// if the parameter doesn't exist. Panics if the value cannot be converted.
func (r *Request) IntParam(paramName string) int {
	strParam := r.Param(paramName)

	if len(strParam) == 0 {
		return 0
	}

	param, err := strconv.Atoi(strParam)

	panicIfNotNil(err)

	return param
}

// Float64Param returns a parameter value converted to a 64-bit floating point number.
//
// The paramName specifies which parameter to retrieve and convert. Returns 0
// if the parameter doesn't exist. Panics if the value cannot be converted.
func (r *Request) Float64Param(paramName string) float64 {
	strParam := r.Param(paramName)

	if len(strParam) == 0 {
		return 0
	}

	param, err := strconv.ParseFloat(strParam, 64)
	panicIfNotNil(err)

	return param
}

// Float32Param returns a parameter value converted to a 32-bit floating point number.
//
// The paramName specifies which parameter to retrieve and convert. Returns 0
// if the parameter doesn't exist. Panics if the value cannot be converted.
func (r *Request) Float32Param(paramName string) float32 {
	strParam := r.Param(paramName)

	if len(strParam) == 0 {
		return 0
	}

	param, err := strconv.ParseFloat(strParam, 32)
	panicIfNotNil(err)

	return float32(param)
}

// Body returns the raw body of the HTTP request as a byte slice.
//
// The body is read only once and cached for subsequent calls.
func (r *Request) Body() []byte {
	if !r.readBody {
		r.readBody = true

		body, err := io.ReadAll(r.Raw.Body)
		panicIfNotNil(err)

		r.recreateBodyReader(body)
		r.body = body
	}

	return r.body
}

// IsDone checks if the request has been completed or canceled.
//
// Returns true if the request is done or the context has been canceled.
func (r *Request) IsDone() bool {
	if r.isDone {
		return true
	}

	select {
	case <-r.Raw.Context().Done():
		r.isDone = true
		return true
	default:
		return false
	}
}

// WithContext returns a shallow copy of r with its context changed
// to ctx. The provided ctx must be non-nil.
//
// For outgoing client request, the context controls the entire
// lifetime of a request and its response: obtaining a connection,
// sending the request, and reading the response headers and body.
//
// To create a new request with a context, use [NewRequestWithContext].
// To make a deep copy of a request with a new context, use [Request.Clone].
func (r *Request) WithContext(ctx context.Context) *Request {
	newRaw := r.Raw.WithContext(ctx)
	newReq := &Request{
		Raw:        newRaw,
		response:   r.response,
		params:     r.params,
		files:      r.files,
		body:       r.body,
		readParams: r.readParams,
		readBody:   r.readBody,
		isDone:     r.isDone,
	}
	return newReq
}

// Context returns the request's context. To change the context, use
// [Request.Clone] or [Request.WithContext].
//
// The returned context is always non-nil; it defaults to the
// background context.
//
// For outgoing client requests, the context controls cancellation.
//
// For incoming server requests, the context is canceled when the
// client's connection closes, the request is canceled (with HTTP/2),
// or when the ServeHTTP method returns.
func (r *Request) Context() context.Context {
	return r.Raw.Context()
}

// Clone returns a deep copy of the request with the same context.
//
// The clone includes a copy of the URL, headers, and request body.
// If the body is an io.ReadCloser, the original body is not closed.
//
// The cloned request inherits all properties of the original request.
func (r *Request) Clone(ctx context.Context) *Request {
	if ctx == nil {
		ctx = r.Context()
	}

	rawClone := r.Raw.Clone(ctx)

	// Create a new Request instance with copied fields
	clonedReq := &Request{
		Raw:        rawClone,
		response:   r.response,
		params:     make(map[string][]string),
		files:      make(map[string][]*multipart.FileHeader),
		readParams: false, // Reset to force re-parsing
		readBody:   false, // Reset to force re-reading
		isDone:     r.isDone,
	}

	// Deep copy the params map
	if r.params != nil {
		for k, v := range r.params {
			paramsCopy := make([]string, len(v))
			copy(paramsCopy, v)
			clonedReq.params[k] = paramsCopy
		}
	}

	// Deep copy the files map
	if r.files != nil {
		for k, v := range r.files {
			filesCopy := make([]*multipart.FileHeader, len(v))
			copy(filesCopy, v) // Note: FileHeader objects themselves aren't deep copied
			clonedReq.files[k] = filesCopy
		}
	}

	// Copy body if already read
	if r.readBody {
		clonedReq.body = make([]byte, len(r.body))
		copy(clonedReq.body, r.body)
		clonedReq.readBody = true
	}

	return clonedReq
}

func (r *Request) parseParams() {
	if r.readParams {
		return
	}

	r.readParams = true

	r.initParams()
	r.parseQueryParams()
	r.parseBodyParams()
}

func (r *Request) setPathParams(pathParams map[string]string) {
	r.initParams()

	for name, value := range pathParams {
		r.params[name] = append(r.params[name], value)
	}
}

func (r *Request) initParams() {
	if r.params == nil {
		r.params = make(map[string][]string)
	}
}

func (r *Request) parseQueryParams() {
	rawQuery := r.Raw.URL.RawQuery

	values, err := url.ParseQuery(rawQuery)
	panicIfNotNilUsingStatusCode(http.StatusBadRequest, err)
	r.copyMapToParams(values)
}

func (r *Request) parseBodyParams() {
	contentType := r.Header(ContentTypeHeader)

	if strings.Contains(contentType, ContentTypeFormUrlEncoded) {
		r.parseFormParams()
	} else if strings.Contains(contentType, ContentTypeFormData) {
		r.parseMultiPartFormParams()
	}
}

func (r *Request) recreateBodyReader(body []byte) {
	if body == nil {
		body = r.Body()
	}

	_ = r.Raw.Body.Close()
	r.Raw.Body = io.NopCloser(bytes.NewBuffer(body))
}

func (r *Request) parseFormParams() {
	body := r.Body()
	defer r.recreateBodyReader(body)

	panicIfNotNil(r.Raw.ParseForm())
	r.copyMapToParams(r.Raw.PostForm)
}

func (r *Request) parseMultiPartFormParams() {
	body := r.Body()
	defer r.recreateBodyReader(body)

	panicIfNotNil(r.Raw.ParseMultipartForm(512 * 1024))

	r.copyMapToParams(r.Raw.MultipartForm.Value)
	r.files = r.Raw.MultipartForm.File
}

func (r *Request) copyMapToParams(m map[string][]string) {
	for key, values := range m {
		if len(r.params[key]) == 0 {
			r.params[key] = values
			continue
		}

		r.params[key] = append(r.params[key], values...)
	}
}
