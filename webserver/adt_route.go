package webserver

import (
	"bytes"
	"net/http"
	"strings"
)

type routesByPattern map[string][]route

type route struct {
	dynamicHost    [][]byte
	staticPattern  string
	dynamicPattern [][]byte
	methods        []string
	handler        Handler
}

var slashSlice = []byte{'/'}
var dotSlice = []byte{'.'}

const dynamicSymbols = "{*"

func (this *routesByPattern) getRoute(method, pattern, hostPort, path string) (currentRoute *route, params map[string]string) {
	routes := (*this)[pattern]
	errorStatus := http.StatusNotFound

	for _, route := range routes {
		params, statusCode := route.matchURLAndGetParam(hostPort, path)

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

		return &route, params
	}

	NewHTTPError(errorStatus, nil).Panic()

	// Should not reach here
	return nil, nil
}

func (this *routesByPattern) Add(methods []string, pattern string, handler Handler) *route {
	route := newRoute(methods, pattern, handler)
	(*this)[route.staticPattern] = append((*this)[route.staticPattern], *route)
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

	// === DYNAMIC HOST === //

	indexOf := bytes.IndexByte(pattern, '/')

	if indexOf == -1 {
		this.dynamicHost = bytes.Split(pattern, dotSlice)
		reversePattern(this.dynamicHost)
		return
	}

	if indexOf > 0 {
		this.dynamicHost = bytes.Split(pattern[:indexOf], dotSlice)
		reversePattern(this.dynamicHost)
		pattern = pattern[indexOf:]
	}

	// === STATIC AND DYNAMIC PATH PATTERN === //

	indexOf = bytes.IndexAny(pattern, dynamicSymbols)

	if indexOf == -1 {
		this.staticPattern = string(trimSlashes(pattern, 0))
		return
	}
	dynamicPattern := pattern[indexOf:]

	staticPattern := pattern[:indexOf]
	staticPattern = staticPattern[:bytes.LastIndexByte(staticPattern, '/')+1]

	this.staticPattern = string(trimSlashes(staticPattern, 0))
	this.dynamicPattern = bytes.Split(trimSlashes(dynamicPattern, 0), slashSlice)
}

func (this *route) matchURLAndGetParam(hostPort, path string) (params map[string]string, status int) {
	params = make(map[string]string)
	pathBytes := trimSlashes([]byte(path), 0)

	// Validate dynamic host
	if len(this.dynamicHost) > 0 {
		host, _ := splitHostPort(hostPort)
		hostTokens := bytes.Split([]byte(host), dotSlice)
		reversePattern(hostTokens)
		status = matchTokens(this.dynamicHost, hostTokens, params)

		if status != 0 {
			return nil, status
		}
	}

	// The static part of the path was already validated by 'http' library
	if len(pathBytes) == len([]byte(this.staticPattern)) && len(this.dynamicPattern) == 0 {
		return nil, 0
	}

	// Split dynamic part of the path by slashes
	dynamicPath := bytes.Split(trimSlashes(pathBytes, len(this.staticPattern)), slashSlice)

	// Validate dynamic path
	return params, matchTokens(this.dynamicPattern, dynamicPath, params)
}

func matchTokens(tokensPattern, tokens [][]byte, params map[string]string) int {
	tokensLength := len(tokens)

	for index, key := range tokensPattern {

		// Handle when the path finishes before of the pattern
		if index == tokensLength {
			if isOptional(key) {
				return 0
			}
			return http.StatusNotFound
		}

		switch key[0] {

		// case '*': ignore
		case '*':
			// case '**': ignore all
			if len(key) > 1 && key[1] == '*' {
				return 0
			}

		// case '{': parse param and validate
		case '{':
			name, value, isOptional := parsePathParam(key, tokens[index])

			if len(value) != 0 {
				params[string(name)] = string(value)
			} else if !isOptional {
				return http.StatusBadRequest
			}

		// default: compare static names
		default:
			if bytes.Compare(key, tokens[index]) != 0 {
				return http.StatusNotFound
			}
		}
	}

	if len(tokensPattern) == tokensLength {
		return 0
	}

	return http.StatusNotFound
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

func splitHostPort(hostPort string) (host, port string) {
	host = hostPort

	colon := strings.LastIndexByte(host, ':')
	if colon == -1 {
		return host, ""
	}

	return hostPort[:colon], hostPort[colon+1:]
}

func reversePattern(pattern [][]byte) {
	for i, j := 0, len(pattern)-1; i < j; i, j = i+1, j-1 {
		pattern[i], pattern[j] = pattern[j], pattern[i]
	}
}
