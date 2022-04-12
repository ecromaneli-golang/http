package webserver

import (
	"fmt"
	"net/http"
)

type serverError struct {
	StatusCode int
	Message    string
	Log        any
}

func NewError(log any) *serverError {
	return (&serverError{Log: log}).setDefaults()
}

func NewHTTPError(statusCode int, log any) *serverError {
	return (&serverError{StatusCode: statusCode, Log: log}).setDefaults()
}

func (this *serverError) Error() string {
	return fmt.Sprintf("[%d] %v", this.StatusCode, this.Log)
}

func (this *serverError) Panic() {
	panic(this)
}

func (this *serverError) setDefaults() *serverError {
	if this.StatusCode == 0 {
		this.StatusCode = http.StatusInternalServerError
	}

	if this.Message == "" {
		this.Message = http.StatusText(this.StatusCode)
	}

	if this.Log == nil {
		this.Log = this.Message
	}

	return this
}

func panicIfNotNil(err error) {
	if err != nil {
		NewError(err).Panic()
	}
}

func panicIfNotNilUsingStatusCode(statusCode int, err error) {
	if err != nil {
		NewHTTPError(statusCode, err).Panic()
	}
}
