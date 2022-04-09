package webserver

import (
	"bytes"
	"net/http"
)

type routesByPattern map[string][]route

type route struct {
	staticPart  string
	dynamicPart [][]byte
	methods     []string
	handler     Handler
}

var slashSlice = []byte{'/'}

const dynamicSymbols = "{*"

func (this *routesByPattern) getRoute(method, pattern, currentPath string) (currentRoute *route, params map[string]string, status int) {
	routes := (*this)[pattern]
	errorStatus := http.StatusNotFound

	for _, route := range routes {
		params, statusCode := route.matchPathAndGetParam(currentPath)

		if statusCode != 0 {
			if errorStatus == http.StatusNotFound {
				errorStatus = statusCode
			}
			continue
		}

		if !route.acceptsMethod(method) {
			errorStatus = http.StatusMethodNotAllowed
			continue
		}

		return &route, params, 0
	}

	return nil, nil, errorStatus
}

func (this *routesByPattern) Add(methods []string, pattern string, handler Handler) *route {
	route := newRoute(methods, pattern, handler)
	(*this)[route.staticPart] = append((*this)[route.staticPart], *route)
	return route
}

func newRoute(methods []string, pattern string, handler Handler) *route {
	route := &route{}
	route.handler = handler
	route.methods = methods

	route.extractAndSetPattern([]byte(pattern))

	return route
}

func (this *route) extractAndSetPattern(pattern []byte) {

	indexOfFirstParameter := bytes.IndexAny(pattern, dynamicSymbols)

	if indexOfFirstParameter == -1 {
		this.staticPart = string(trimSlashes(pattern, 0))
		return
	}
	dynamicPattern := pattern[indexOfFirstParameter:]

	staticPart := pattern[:indexOfFirstParameter]
	staticPart = staticPart[:bytes.LastIndexByte(staticPart, '/')+1]

	this.staticPart = string(trimSlashes(staticPart, 0))
	this.dynamicPart = bytes.Split(trimSlashes(dynamicPattern, 0), slashSlice)
}

func (this *route) matchPathAndGetParam(path string) (params map[string]string, status int) {
	pathBytes := trimSlashes([]byte(path), 0)

	// The static part of the path was already validated by 'http' library
	if len(pathBytes) == len([]byte(this.staticPart)) && len(this.dynamicPart) == 0 {
		return nil, 0
	}

	// Split dynamic part of the path by slashes
	dynamicPath := bytes.Split(trimSlashes(pathBytes, len(this.staticPart)), slashSlice)
	dynamicPathLength := len(dynamicPath)

	params = make(map[string]string)

	for index, key := range this.dynamicPart {

		// Handle when the path finishes before of the pattern
		if index == dynamicPathLength {
			if isOptional(key) {
				return params, 0
			}
			return nil, http.StatusNotFound
		}

		switch key[0] {

		// case '*': ignore
		case '*':
			// case '**': ignore all
			if len(key) > 1 && key[1] == '*' {
				return params, 0
			}

		// case '{': parse param and validate
		case '{':
			name, value, isOptional := parsePathParam(key, dynamicPath[index])

			if len(value) != 0 {
				params[string(name)] = string(value)
			} else if !isOptional {
				return nil, http.StatusBadRequest
			}

		// default: compare static names
		default:
			if bytes.Compare(key, dynamicPath[index]) != 0 {
				return nil, http.StatusNotFound
			}
		}
	}

	if len(this.dynamicPart) == len(dynamicPath) {
		return params, 0
	}

	return nil, http.StatusNotFound
}

func parsePathParam(pattern, path []byte) (name, value []byte, isOpt bool) {
	isOpt = isOptional(pattern)

	if !isOpt && len(path) == 0 {
		return nil, path, isOpt
	}

	end := len(pattern) - 1

	if isOpt {
		end--
	}

	return pattern[1:end], path, isOpt
}

func isOptional(pattern []byte) bool {
	return pattern[len(pattern)-2] == '?'
}

func trimSlashes(data []byte, begin int) []byte {
	dataLength := len(data)

	if dataLength <= begin {
		return data
	}

	end := len(data)

	if data[begin] == '/' && dataLength > 1 {
		begin++
	}

	if data[end-1] == '/' {
		end--
	}

	return data[begin:end]
}

func (this *route) acceptsMethod(method string) bool {
	if this.methods == nil {
		return true
	}

	for _, item := range this.methods {
		if item == method {
			return true
		}
	}
	return false
}
