package utils

import "fmt"

// HTTPError represents an HTTP error with status code
type HTTPError struct {
	StatusCode int
	Message    string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}

// NewHTTPError creates a new HTTP error
func NewHTTPError(statusCode int, message string) *HTTPError {
	return &HTTPError{
		StatusCode: statusCode,
		Message:    message,
	}
}

// Common errors

// ErrNotFound represents a 404 error
var ErrNotFound = NewHTTPError(404, "Not Found")

// ErrForbidden represents a 403 error
var ErrForbidden = NewHTTPError(403, "Forbidden")

// ErrServiceUnavailable represents a 503 error
var ErrServiceUnavailable = NewHTTPError(503, "Service Unavailable")

// ErrRequestInProgress represents a request in progress error
var ErrRequestInProgress = NewHTTPError(503, "Request in progress, please retry")

// ErrUnauthorized represents a 401 error
var ErrUnauthorized = NewHTTPError(401, "Unauthorized")

// ErrBadRequest represents a 400 error
var ErrBadRequest = NewHTTPError(400, "Bad Request")
