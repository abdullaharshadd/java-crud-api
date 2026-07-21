// Package error contains the SmartContact service's domain errors and the
// HTTP error-handling middleware. It is the Go equivalent of the Java
// package com.smartContact.error.
//
// MIGRATION_NOTE: The source RestResponseEntityExceptionHandling.java was a
// Spring @ControllerAdvice that extended ResponseEntityExceptionHandler.
// Spring MVC used AOP-style interception to catch a UserNotFoundException
// thrown from ANY controller and convert it into a structured 404 JSON
// response, while inheriting Spring's default handlers for framework
// exceptions (validation, unsupported media type, etc).
//
// Go has no equivalent to controller-advice AOP interception. The idiomatic
// analogue is HTTP middleware that recovers/inspects errors and translates
// them into a status code and JSON body. Because Go handlers do not "throw",
// we model the two common approaches:
//
//   - A helper (WriteError) that a handler calls explicitly when it has an
//     error value, mapping the error to the correct status and body.
//   - Middleware (ExceptionHandling) that wraps a handler which returns an
//     error, centralizing the translation exactly like the @ControllerAdvice.
//
// The inherited Spring default handlers (from ResponseEntityExceptionHandler)
// have no direct Go equivalent; each router/framework surfaces its own 4xx
// responses. Only the explicitly-mapped UserNotFoundException behaviour is
// reproduced here. See requires_manual_review note.
package error

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/smartContact/internal/smartcontact/model"
)

// HandlerFunc is an HTTP handler that may return an error. Returning a
// non-nil error lets the ExceptionHandling middleware translate it into a
// structured HTTP response, mirroring how a Spring controller would let an
// exception propagate to the @ControllerAdvice.
type HandlerFunc func(http.ResponseWriter, *http.Request) error

// WriteError inspects err and writes the appropriate HTTP status and JSON
// ErrorMessage body to w. It is the explicit Go equivalent of the Spring
// @ExceptionHandler methods.
//
// A *UserNotFoundError (matched via errors.As, so wrapped errors are also
// handled) is translated into a 404 Not Found response, preserving the
// original Java behaviour that returned HttpStatus.NOT_FOUND with the
// exception message. Any other error falls through to a generic 500.
func WriteError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	var notFound *UserNotFoundError
	if errors.As(err, &notFound) {
		writeErrorMessage(w, http.StatusNotFound, notFound.Error())
		return
	}

	writeErrorMessage(w, http.StatusInternalServerError, err.Error())
}

// ExceptionHandling wraps a HandlerFunc and centralizes error-to-response
// translation, mirroring the cross-cutting @ControllerAdvice interception.
// If the wrapped handler returns an error, it is passed to WriteError.
func ExceptionHandling(h HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			WriteError(w, err)
		}
	}
}

// writeErrorMessage serializes an ErrorMessage with the given status code
// and message to the response writer as JSON.
func writeErrorMessage(w http.ResponseWriter, status int, message string) {
	errorMessage := model.NewErrorMessage(status, message)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	// MIGRATION_NOTE: If encoding fails the status/headers are already sent,
	// so we cannot change the response. We ignore the encode error here
	// rather than panic, matching library-code guidance to never panic.
	_ = json.NewEncoder(w).Encode(errorMessage)
}
