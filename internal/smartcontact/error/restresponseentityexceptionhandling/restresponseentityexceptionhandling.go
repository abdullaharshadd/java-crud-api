package error

import (
	"encoding/json"
	errs "errors"
	"net/http"

	"github.com/smartContact/internal/smartcontact/model"
)

// Package error provides HTTP error-handling helpers for the SmartContact REST
// layer. It corresponds to the original Spring Java class
// com.smartContact.error.RestResponseEntityExceptionHandling.
//
// MIGRATION_NOTE: The Java source used @ControllerAdvice /
// @ExceptionHandler on a subclass of ResponseEntityExceptionHandler. That is a
// Spring MVC cross-cutting (AOP) mechanism that intercepts exceptions thrown
// anywhere in the request pipeline and converts them into ResponseEntity
// responses. Go's net/http has no equivalent framework-managed advice.
// Idiomatic Go instead handles errors explicitly at the point they occur, or
// centralizes translation in a small helper (and/or middleware) that HTTP
// handlers call. This file exposes:
//
//   - HandleUserNotFound: writes a 404 ErrorMessage JSON body for a
//     *UserNotFoundError (the direct equivalent of the Java
//     @ExceptionHandler method).
//   - WriteError: a generic dispatcher that inspects an error and writes the
//     appropriate structured response, which handlers can defer/call in place
//     of Spring's automatic advice.
//
// MIGRATION_NOTE: The extension of ResponseEntityExceptionHandler (which
// supplies default handling for Spring's built-in MVC exceptions such as
// validation and message-not-readable errors) has no direct port. Those cases
// must be handled explicitly by the individual handlers in Go; WriteError
// provides a default 500 fallback for unrecognized errors.

// HandleUserNotFound writes an HTTP 404 response whose body is an
// ErrorMessage describing the supplied UserNotFoundError. It mirrors the Java
// userNotFoundException @ExceptionHandler method.
//
// It returns an error only if the response body could not be encoded/written;
// the caller should log such failures since the HTTP status has already been
// committed.
func HandleUserNotFound(w http.ResponseWriter, err *UserNotFoundError) error {
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	return writeErrorMessage(w, http.StatusNotFound, msg)
}

// WriteError inspects err and writes an appropriate structured HTTP response.
// A *UserNotFoundError (or an error wrapping one) produces a 404; any other
// error produces a 500. It returns an error only if writing the response body
// failed.
//
// MIGRATION_NOTE: This is the explicit Go replacement for Spring's
// @ControllerAdvice dispatch. Handlers call WriteError instead of relying on
// framework interception.
func WriteError(w http.ResponseWriter, err error) error {
	var notFound *UserNotFoundError
	if errs.As(err, &notFound) {
		return HandleUserNotFound(w, notFound)
	}

	msg := ""
	if err != nil {
		msg = err.Error()
	}
	return writeErrorMessage(w, http.StatusInternalServerError, msg)
}

// writeErrorMessage encodes an ErrorMessage with the given HTTP status and
// message and writes it to w as JSON.
func writeErrorMessage(w http.ResponseWriter, status int, message string) error {
	body := model.NewErrorMessage(status, message)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(body); err != nil {
		return err
	}
	return nil
}
