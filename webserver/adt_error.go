package webserver

import (
	"net/http"
	"strconv"
)

type serverError struct {
	StatusCode int
	Message    string
	Log        any
}

func newError(log any) *serverError {
	return &serverError{Log: log}
}

func NewHTTPError(statusCode int, log any) *serverError {
	return &serverError{StatusCode: statusCode, Log: log}
}

func (this *serverError) Panic() {
	this.setDefaults()
	panic(this)
}

func (this *serverError) setDefaults() {
	if this.StatusCode == 0 {
		this.StatusCode = http.StatusInternalServerError
	}

	if this.Message == "" {
		this.Message = strconv.Itoa(this.StatusCode) + " - " + http.StatusText(this.StatusCode)
	}

	if this.Log == nil {
		this.Log = this.Message
	}
}

func panicIfNotNil(err error) {
	if err != nil {
		newError(err).Panic()
	}
}

func panicIfNotNilUsingStatusCode(statusCode int, err error) {
	if err != nil {
		NewHTTPError(statusCode, err).Panic()
	}
}
