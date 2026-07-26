// Package apierror provides structured error handling for API responses.
//
// MIGRATION_NOTE: The original Java ErrorMessage was a Lombok @Data DTO pairing
// a Spring HttpStatus with a message. In idiomatic Go we replace this DTO with:
//   - An ErrorMessage struct that serializes cleanly to JSON for API responses.
//   - Sentinel errors (errors.Is-comparable) for common failure modes.
//   - Helper functions (WriteError / WriteValidationError) that map errors to
//     HTTP status codes and write consistent JSON payloads.
//
// The package path is internal/smartcontact/model per the migration target, but
// the logical grouping is API error handling. Move to internal/smartcontact/apierror
// if a dedicated error package is preferred during manual review.
package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// Sentinel errors represent well-known failure conditions that the API layer
// can map to specific HTTP status codes. Wrap them with fmt.Errorf("...: %w", err)
// to add context while remaining comparable via errors.Is.
var (
	// ErrDuplicateEmail indicates an attempt to create a resource with an
	// email address that already exists.
	ErrDuplicateEmail = errors.New("duplicate email")

	// ErrBadRequest indicates the request was malformed or failed validation.
	ErrBadRequest = errors.New("bad request")

	// ErrNotFound indicates the requested resource does not exist.
	ErrNotFound = errors.New("not found")
)

// ErrorMessage is the JSON payload returned to clients when a request fails.
// It pairs an HTTP status code with a human-readable message.
type ErrorMessage struct {
	// Status is the HTTP status code associated with the error.
	Status int `json:"status"`

	// Message is a human-readable description of the error.
	Message string `json:"message"`
}

// NewErrorMessage constructs an ErrorMessage from an HTTP status code and message.
func NewErrorMessage(status int, message string) ErrorMessage {
	return ErrorMessage{
		Status:  status,
		Message: message,
	}
}

// String implements fmt.Stringer, mirroring the Lombok-generated toString.
func (e ErrorMessage) String() string {
	return fmt.Sprintf("ErrorMessage(status=%d, message=%q)", e.Status, e.Message)
}

// statusForError maps a sentinel error to its corresponding HTTP status code.
// Unrecognized errors default to 500 Internal Server Error.
func statusForError(err error) int {
	switch {
	case errors.Is(err, ErrDuplicateEmail):
		return http.StatusConflict
	case errors.Is(err, ErrBadRequest):
		return http.StatusBadRequest
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

// WriteError writes a JSON ErrorMessage to w, selecting the HTTP status code
// based on the wrapped sentinel error. It returns any encoding error so callers
// can log it; it never panics.
func WriteError(w http.ResponseWriter, err error) error {
	if err == nil {
		err = errors.New("internal server error")
	}

	status := statusForError(err)
	payload := NewErrorMessage(status, err.Error())

	return writeJSON(w, status, payload)
}

// WriteValidationError writes a 400 Bad Request JSON ErrorMessage. It is a
// convenience wrapper for validation failures where the message is already
// known and no sentinel error is involved.
func WriteValidationError(w http.ResponseWriter, message string) error {
	if message == "" {
		message = ErrBadRequest.Error()
	}

	payload := NewErrorMessage(http.StatusBadRequest, message)
	return writeJSON(w, http.StatusBadRequest, payload)
}

// writeJSON encodes payload as JSON and writes it with the given status code.
func writeJSON(w http.ResponseWriter, status int, payload ErrorMessage) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		return fmt.Errorf("encoding error response: %w", err)
	}
	return nil
}
