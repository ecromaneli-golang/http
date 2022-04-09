package webserver

import (
	"bytes"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
)

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

func (this *Request) AllHeaders() http.Header {
	return this.Raw.Header
}

func (this *Request) Headers(name string) []string {
	this.parseParams()
	return this.AllHeaders()[name]
}

func (this *Request) Header(name string) string {
	this.parseParams()

	header := this.Headers(name)

	if len(header) == 0 {
		return ""
	}

	return header[0]
}

func (this *Request) AllParams() map[string][]string {
	this.parseParams()
	return this.params
}

func (this *Request) Params(paramName string) []string {
	this.parseParams()
	return this.params[paramName]
}

func (this *Request) Param(paramName string) string {
	this.parseParams()

	param := this.params[paramName]

	if len(param) == 0 {
		return ""
	}

	return param[0]
}

func (this *Request) AllFiles() map[string][]*multipart.FileHeader {
	this.parseParams()
	return this.files
}

func (this *Request) Files(paramName string) []*multipart.FileHeader {
	this.parseParams()
	return this.files[paramName]
}

func (this *Request) File(paramName string) *multipart.FileHeader {
	this.parseParams()

	files := this.files[paramName]

	if len(files) == 0 {
		return nil
	}

	return files[0]
}

func (this *Request) UIntParam(paramName string) uint {
	return uint(this.IntParam(paramName))
}

func (this *Request) IntParam(paramName string) int {
	strParam := this.Param(paramName)

	if len(strParam) == 0 {
		return 0
	}

	param, err := strconv.Atoi(strParam)

	panicIfNotNil(err)

	return param
}

func (this *Request) Float64Param(paramName string) float64 {
	strParam := this.Param(paramName)

	if len(strParam) == 0 {
		return 0
	}

	param, err := strconv.ParseFloat(strParam, 64)
	panicIfNotNil(err)

	return param
}

func (this *Request) Float32Param(paramName string) float32 {
	strParam := this.Param(paramName)

	if len(strParam) == 0 {
		return 0
	}

	param, err := strconv.ParseFloat(strParam, 32)
	panicIfNotNil(err)

	return float32(param)
}

func (this *Request) Body() []byte {
	if !this.readBody {
		this.readBody = true

		body, err := ioutil.ReadAll(this.Raw.Body)
		panicIfNotNil(err)

		this.Raw.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		this.body = body
	}

	return this.body
}

func (this *Request) IsDone() bool {
	if this.isDone {
		return true
	}

	select {
	case <-this.Raw.Context().Done():
		this.isDone = true
		return true
	default:
		return false
	}
}

func (this *Request) parseParams() {
	if this.readParams {
		return
	}

	this.readParams = true

	this.initParams()
	this.parseQueryParams()
	this.parseBodyParams()
}

func (this *Request) setPathParams(pathParams map[string]string) {
	this.initParams()

	for name, value := range pathParams {
		this.params[name] = append(this.params[name], value)
	}
}

func (this *Request) initParams() {
	if this.params == nil {
		this.params = make(map[string][]string)
	}
}

func (this *Request) parseQueryParams() {
	rawQuery := this.Raw.URL.RawQuery

	values, err := url.ParseQuery(rawQuery)
	panicIfNotNilUsingStatusCode(http.StatusBadRequest, err)
	this.copyMapToParams(values)
}

func (this *Request) parseBodyParams() {
	switch this.Header(ContentTypeHeader) {

	case ContentTypeFormUrlEncoded:
		this.parseFormParams()

	case ContentTypeFormData:
		this.parseMultiPartFormParams()
	}
}

func (this *Request) parseFormParams() {
	panicIfNotNil(this.Raw.ParseForm())
	this.copyMapToParams(this.Raw.PostForm)
}

func (this *Request) parseMultiPartFormParams() {
	panicIfNotNil(this.Raw.ParseMultipartForm(512 * 1024))

	this.copyMapToParams(this.Raw.MultipartForm.Value)
	this.files = this.Raw.MultipartForm.File
}

func (this *Request) copyMapToParams(m map[string][]string) {
	for key, values := range m {
		if len(this.params[key]) == 0 {
			this.params[key] = values
			continue
		}

		for _, value := range values {
			this.params[key] = append(this.params[key], value)
		}
	}
}
