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
var emptySlice = make([]byte, 0)
var emptyMatrix = make([][]byte, 0)

const dynamicSymbols = "{*"

func (this *routesByPattern) getRoute(method, pattern, hostPort, path string) (currentRoute *route, params map[string]string) {
	routes := (*this)[pattern]
	errorStatus := http.StatusNotFound

	for _, route := range routes {
		params, status := route.matchURLAndGetParam(hostPort, path)

		if !status {
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
		this.staticPattern = string(trimSlashes(pattern))
		return
	}
	dynamicPattern := pattern[indexOf:]

	staticPattern := pattern[:indexOf]
	staticPattern = staticPattern[:bytes.LastIndexByte(staticPattern, '/')+1]

	this.staticPattern = string(trimSlashes(staticPattern))
	this.dynamicPattern = bytes.Split(trimSlashes(dynamicPattern), slashSlice)
}

func (this *route) matchURLAndGetParam(hostPort, path string) (params map[string]string, status bool) {
	params = make(map[string]string)

	// Validate dynamic host
	if len(this.dynamicHost) > 0 {
		host, _ := splitHostPort(hostPort)
		hostTokens := bytes.Split([]byte(host), dotSlice)
		reversePattern(hostTokens)

		if !matchTokens(this.dynamicHost, hostTokens, params) {
			return nil, false
		}
	}

	// The static part of the path was already validated by 'http' library
	if len(path) == len(this.staticPattern) && len(this.dynamicPattern) == 0 {
		return params, true
	}

	// Split dynamic part of the path by slashes
	pathBytes := trimSlashes(trimSlashes([]byte(path))[len(this.staticPattern):])

	var dynamicPath [][]byte
	if len(pathBytes) > 0 {
		dynamicPath = bytes.Split(pathBytes, slashSlice)
	} else {
		dynamicPath = emptyMatrix
	}

	// Validate dynamic path
	return params, matchTokens(this.dynamicPattern, dynamicPath, params)
}

func matchTokens(tokensPattern, tokens [][]byte, params map[string]string) bool {
	tokensLength := len(tokens)

	for index, key := range tokensPattern {
		var hasToken bool = index < tokensLength
		var tokenValue []byte

		if hasToken {
			tokenValue = tokens[index]
		}

		switch key[0] {

		// case '*': ignore
		case '*':
			// case '**': ignore all
			if len(key) > 1 && key[1] == '*' {
				return true
			}

		// case '{': parse param and validate
		case '{':
			name, isOptional := parsePathParam(key, tokenValue)

			if !hasToken {
				return isOptional
			}

			params[string(name)] = string(tokenValue)

		// default: compare static names
		default:
			if bytes.Compare(key, tokenValue) != 0 {
				return false
			}
		}
	}

	return len(tokensPattern) >= tokensLength
}

func parsePathParam(pattern, path []byte) (name []byte, isOpt bool) {
	isOpt = isOptional(pattern)

	if !isOpt && len(path) == 0 {
		return nil, isOpt
	}

	end := len(pattern) - 1

	if isOpt {
		end--
	}

	return pattern[1:end], isOpt
}

func isOptional(pattern []byte) bool {
	tokenIndex := len(pattern) - 2

	if tokenIndex < 2 {
		return false
	}

	return pattern[tokenIndex] == '?'
}

func trimSlashes(data []byte) []byte {
	begin, end := 0, len(data)

	if end == 0 {
		return data
	}

	if data[begin] == '/' {
		begin++
	}

	if end > 1 && data[end-1] == '/' {
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
