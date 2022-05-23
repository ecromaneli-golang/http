package webserver

import (
	"fmt"
	"net/http"
)

type serverError struct {
	statusCode int
	message    string
	log        any
}

func NewError(log any) *serverError {
	return (&serverError{log: log}).setDefaults()
}

func NewHTTPError(statusCode int, log any) *serverError {
	return (&serverError{statusCode: statusCode, log: log}).setDefaults()
}

func (this *serverError) ExposeLog() *serverError {
	this.message = fmt.Sprintf("%v", this.log)
	return this
}

func (this *serverError) Error() string {
	return fmt.Sprintf("[%d] %v", this.statusCode, this.log)
}

func (this *serverError) Panic() {
	panic(this)
}

func (this *serverError) setDefaults() *serverError {
	if this.statusCode == 0 {
		this.statusCode = http.StatusInternalServerError
	}

	if this.message == "" {
		this.message = http.StatusText(this.statusCode)
	}

	if this.log == nil {
		this.log = this.message
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
