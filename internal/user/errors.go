package user

import "errors"

// NotFoundMessage is the verbatim message the source service used when a user
// lookup failed. It is part of the HTTP error body and must not change.
const NotFoundMessage = "User are not available"

// ErrNotFound is the sentinel for "requested user does not exist"
// (source: com.smartContact.error.UserNotFoundException). Check for it with
// errors.Is; every *NotFoundError matches it.
var ErrNotFound = errors.New(NotFoundMessage)

// ErrIncorrectResultSize is returned when a lookup that must yield at most one
// user (e.g. find by name) matched several rows. In the source this surfaced
// as Spring's IncorrectResultSizeDataAccessException, i.e. an HTTP 500.
var ErrIncorrectResultSize = errors.New("user: query did not return a unique result")

// ErrEmptyResult is returned when deleting a user id that does not exist. In
// the source this surfaced as Spring's EmptyResultDataAccessException, i.e.
// an HTTP 500 (not a 404).
var ErrEmptyResult = errors.New("user: no user entity with the given id exists")

// NotFoundError is a user-not-found error carrying an optional detail message
// and an optional underlying cause. It mirrors the constructor overloads of
// the Java UserNotFoundException; most code should simply return ErrNotFound
// or wrap it with fmt.Errorf("...: %w", ErrNotFound).
//
// MIGRATION_NOTE: the protected Java constructor controlling suppression and
// stack-trace writability has no Go equivalent (Go errors carry no stack trace
// or suppressed list) and is intentionally omitted.
type NotFoundError struct {
	// Message is the detail message; empty means "no detail message".
	Message string
	// Cause is the underlying error, if any.
	Cause error
}

// NewNotFoundError returns a NotFoundError with no detail message and no cause.
func NewNotFoundError() *NotFoundError {
	return &NotFoundError{}
}

// NewNotFoundErrorMsg returns a NotFoundError with the given detail message
// and no cause.
func NewNotFoundErrorMsg(message string) *NotFoundError {
	return &NotFoundError{Message: message}
}

// NewNotFoundErrorMsgCause returns a NotFoundError with the given detail
// message and underlying cause.
func NewNotFoundErrorMsgCause(message string, cause error) *NotFoundError {
	return &NotFoundError{Message: message, Cause: cause}
}

// NewNotFoundErrorCause returns a NotFoundError wrapping cause; its message is
// derived from the cause, as in Java's Throwable(Throwable) constructor.
func NewNotFoundErrorCause(cause error) *NotFoundError {
	return &NotFoundError{Cause: cause}
}

// Error implements the error interface. It returns the detail message if set,
// otherwise the cause's message, otherwise the default NotFoundMessage.
func (e *NotFoundError) Error() string {
	switch {
	case e.Message != "":
		return e.Message
	case e.Cause != nil:
		return e.Cause.Error()
	default:
		return NotFoundMessage
	}
}

// Unwrap returns the underlying cause, enabling errors.Is / errors.As chains.
func (e *NotFoundError) Unwrap() error {
	return e.Cause
}

// Is makes every NotFoundError match the ErrNotFound sentinel.
func (e *NotFoundError) Is(target error) bool {
	return target == ErrNotFound
}
