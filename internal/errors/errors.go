package errors

import (
	"errors"
)

var (
	// ErrRouteNotFound is returned when no matching route is found
	ErrRouteNotFound = errors.New("route not found")

	// ErrMethodNotAllowed is returned when the route is found but the method is not allowed
	ErrMethodNotAllowed = errors.New("method not allowed")

	// ErrProxyTimeout is returned when a proxy request times out
	ErrProxyTimeout = errors.New("proxy request timed out")
)
