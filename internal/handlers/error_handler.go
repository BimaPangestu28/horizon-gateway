package handlers

import (
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v2"

	apierrors "github.com/bimapangestu28/horizon/internal/errors"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
}

// CustomErrorHandler provides enhanced error handling for HTTP requests
func CustomErrorHandler(c *fiber.Ctx, err error) error {
	// Default to internal server error
	code := http.StatusInternalServerError
	message := "Internal Server Error"

	// Handle specific error types
	if fiberErr, ok := err.(*fiber.Error); ok {
		// Handle Fiber's built-in errors
		code = fiberErr.Code
		message = fiberErr.Message
	} else if errors.Is(err, apierrors.ErrRouteNotFound) {
		code = http.StatusNotFound
		message = "Route not found"
	} else if errors.Is(err, apierrors.ErrMethodNotAllowed) {
		code = http.StatusMethodNotAllowed
		message = "Method not allowed"
	}

	// Return error as JSON
	return c.Status(code).JSON(ErrorResponse{
		Error:   http.StatusText(code),
		Code:    code,
		Message: message,
	})
}
