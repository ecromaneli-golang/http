package webserver

import (
	"fmt"
	"net/http"
)

// serverError represents an HTTP server error with status code and message.
type serverError struct {
	statusCode int
	message    string
	log        any
}

// NewError creates a new serverError with the given log information.
//
// The log parameter contains error details that will be logged but not exposed to the client.
//
// Returns a serverError instance with default status code (500 Internal Server Error).
func NewError(log any) *serverError {
	return (&serverError{log: log}).setDefaults()
}

// NewHTTPError creates a new serverError with the specified status code and log information.
//
// The statusCode parameter is the HTTP status code to return to the client.
// The log parameter contains error details that will be logged but not exposed to the client.
//
// Returns a configured serverError instance.
func NewHTTPError(statusCode int, log any) *serverError {
	return (&serverError{statusCode: statusCode, log: log}).setDefaults()
}

// ExposeLog sets the error message to be the same as the log information.
//
// This makes the error details visible to clients, so should be used only when
// it's safe to expose internal error details.
//
// Returns the serverError instance for method chaining.
func (se *serverError) ExposeLog() *serverError {
	se.message = fmt.Sprintf("%v", se.log)
	return se
}

// Error returns a string representation of the error.
//
// The format is "[status_code] log_message".
//
// Implements the error interface.
func (se *serverError) Error() string {
	return fmt.Sprintf("[%d] %v", se.statusCode, se.log)
}

// Panic triggers a panic with the serverError.
//
// This is used to immediately stop execution and propagate the error up the call stack.
func (se *serverError) Panic() {
	panic(se)
}

func (se *serverError) setDefaults() *serverError {
	if se.statusCode == 0 {
		se.statusCode = http.StatusInternalServerError
	}

	if se.log == nil {
		se.log = se.message
	}

	return se
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
