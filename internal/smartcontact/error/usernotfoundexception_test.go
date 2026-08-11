```go
package apperr_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	apperr "github.com/example/smartcontact/internal/smartcontact/error"
)

// ---------------------------------------------------------------------------
// NewUserNotFoundError – no-arg constructor
// ---------------------------------------------------------------------------

func TestNewUserNotFoundError(t *testing.T) {
	tests := []struct {
		name            string
		wantMessage     string
		wantCause       error
		wantErrorString string
	}{
		{
			name:            "instantiated with no arguments returns default message and nil cause",
			wantMessage:     apperr.DefaultUserNotFoundMessage,
			wantCause:       nil,
			wantErrorString: apperr.DefaultUserNotFoundMessage,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := apperr.NewUserNotFoundError()

			assert.NotNil(t, err, "constructor should return a non-nil pointer")
			assert.Equal(t, tc.wantMessage, err.Message, "Message field should match")
			assert.Equal(t, tc.wantCause, err.Cause, "Cause field should be nil")
			assert.Equal(t, tc.wantErrorString, err.Error(), "Error() should return expected string")

			// Ensure it satisfies the error interface
			var target *apperr.UserNotFoundError
			assert.True(t, errors.As(err, &target), "should be detectable via errors.As")
		})
	}
}

// ---------------------------------------------------------------------------
// NewUserNotFoundErrorf – message constructor
// ---------------------------------------------------------------------------

func TestNewUserNotFoundErrorf(t *testing.T) {
	tests := []struct {
		name            string
		format          string
		args            []any
		wantMessage     string
		wantCause       error
		wantErrorString string
	}{
		{
			name:            "instantiated with non-empty message stores exact message",
			format:          "user with id %d not found",
			args:            []any{42},
			wantMessage:     "user with id 42 not found",
			wantCause:       nil,
			wantErrorString: "user with id 42 not found",
		},
		{
			name:            "instantiated with plain string stores it verbatim",
			format:          "custom error message",
			args:            nil,
			wantMessage:     "custom error message",
			wantCause:       nil,
			wantErrorString: "custom error message",
		},
		{
			// Java analogue: passing a null message to the string constructor
			// causes getMessage() to return null. In Go the equivalent is an
			// empty string format producing an empty Message field; Error()
			// then falls back to DefaultUserNotFoundMessage.
			name:            "instantiated with empty format falls back to default in Error()",
			format:          "",
			args:            nil,
			wantMessage:     "",
			wantCause:       nil,
			wantErrorString: apperr.DefaultUserNotFoundMessage,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err *apperr.UserNotFoundError
			if len(tc.args) > 0 {
				err = apperr.NewUserNotFoundErrorf(tc.format, tc.args...)
			} else {
				err = apperr.NewUserNotFoundErrorf(tc.format)
			}

			assert.NotNil(t, err)
			assert.Equal(t, tc.wantMessage, err.Message, "Message field mismatch")
			assert.Equal(t, tc.wantCause, err.Cause, "Cause should be nil")
			assert.Equal(t, tc.wantErrorString, err.Error(), "Error() string mismatch")
		})
	}
}

// ---------------------------------------------------------------------------
// NewUserNotFoundErrorWithCause – message + cause constructor
// ---------------------------------------------------------------------------

func TestNewUserNotFoundErrorWithCause(t *testing.T) {
	sentinel := errors.New("underlying db error")

	tests := []struct {
		name            string
		message         string
		cause           error
		wantMessage     string
		wantCause       error
		wantErrorString string
	}{
		{
			name:            "message and cause both provided",
			message:         "user lookup failed",
			cause:           sentinel,
			wantMessage:     "user lookup failed",
			wantCause:       sentinel,
			wantErrorString: fmt.Sprintf("user lookup failed: %v", sentinel),
		},
		{
			// Java: UserNotFoundException(Throwable cause) – message derived
			// from cause.toString(). In Go: empty message falls back to
			// DefaultUserNotFoundMessage; cause is preserved.
			name:            "empty message with non-nil cause uses default message",
			message:         "",
			cause:           sentinel,
			wantMessage:     apperr.DefaultUserNotFoundMessage,
			wantCause:       sentinel,
			wantErrorString: fmt.Sprintf("%s: %v", apperr.DefaultUserNotFoundMessage, sentinel),
		},
		{
			// Java: UserNotFoundException(null, null) – both null
			name:            "empty message and nil cause",
			message:         "",
			cause:           nil,
			wantMessage:     apperr.DefaultUserNotFoundMessage,
			wantCause:       nil,
			wantErrorString: apperr.DefaultUserNotFoundMessage,
		},
		{
			name:            "non-empty message with nil cause",
			message:         "user 99 not found",
			cause:           nil,
			wantMessage:     "user 99 not found",
			wantCause:       nil,
			wantErrorString: "user 99 not found",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := apperr.NewUserNotFoundErrorWithCause(tc.message, tc.cause)

			assert.NotNil(t, err)
			assert.Equal(t, tc.wantMessage, err.Message, "Message field mismatch")
			assert.Equal(t, tc.wantCause, err.Cause, "Cause field mismatch")
			assert.Equal(t, tc.wantErrorString, err.Error(), "Error() string mismatch")
		})
	}
}

// ---------------------------------------------------------------------------
// Unwrap / errors.Is / errors.As – error chain traversal
// ---------------------------------------------------------------------------

func TestUserNotFoundError_Unwrap(t *testing.T) {
	sentinel := errors.New("root cause")

	tests := []struct {
		name          string
		err           *apperr.UserNotFoundError
		wantUnwrap    error
		isTarget      error
		wantIs        bool
		wantAsSuccess bool
	}{
		{
			name:          "Unwrap returns nil when no cause set",
			err:           apperr.NewUserNotFoundError(),
			wantUnwrap:    nil,
			isTarget:      apperr.NewUserNotFoundError(),
			wantIs:        false, // different pointer, no Is override
			wantAsSuccess: true,
		},
		{
			name:          "Unwrap returns the wrapped cause",
			err:           apperr.NewUserNotFoundErrorWithCause("msg", sentinel),
			wantUnwrap:    sentinel,
			isTarget:      sentinel,
			wantIs:        true,
			wantAsSuccess: true,
		},
		{
			name:          "errors.Is traverses chain to find sentinel",
			err:           apperr.NewUserNotFoundErrorWithCause("wrapped", sentinel),
			wantUnwrap:    sentinel,
			isTarget:      sentinel,
			wantIs:        true,
			wantAsSuccess: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.wantUnwrap, tc.err.Unwrap(), "Unwrap() result mismatch")
			assert.Equal(t, tc.wantIs, errors.Is(tc.err, tc.isTarget), "errors.Is mismatch")

			var target *apperr.UserNotFoundError
			assert.Equal(t, tc.wantAsSuccess, errors.As(tc.err, &target), "errors.As mismatch")
		})
	}
}

// ---------------------------------------------------------------------------
// Error() – string representation
// ---------------------------------------------------------------------------

func TestUserNotFoundError_Error(t *testing.T) {
	cause := errors.New("sql: no rows")

	tests := []struct {
		name    string
		err     *apperr.UserNotFoundError
		wantStr string
	}{
		{
			name:    "no message no cause returns default",
			err:     &apperr.UserNotFoundError{},
			wantStr: apperr.DefaultUserNotFoundMessage,
		},
		{
			name:    "explicit message no cause returns message",
			err:     &apperr.UserNotFoundError{Message: "specific message"},
			wantStr: "specific message",
		},
		{
			name:    "no message with cause returns default: cause",
			err:     &apperr.UserNotFoundError{Cause: cause},
			wantStr: fmt.Sprintf("%s: %v", apperr.DefaultUserNotFoundMessage, cause),
		},
		{
			name:    "message and cause returns message: cause",
			err:     &apperr.UserNotFoundError{Message: "lookup failed", Cause: cause},
			wantStr: fmt.Sprintf("lookup failed: %v", cause),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.wantStr, tc.err.Error())
		})
	}
}

// ---------------------------------------------------------------------------
// Type identity – satisfies error interface & detectable via errors.As
// ---------------------------------------------------------------------------

func TestUserNotFoundError_TypeIdentity(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{
			name: "NewUserNotFoundError satisfies error interface",
			err:  apperr.NewUserNotFoundError(),
		},
		{
			name: "NewUserNotFoundErrorf satisfies error interface",
			err:  apperr.NewUserNotFoundErrorf("user %d not found", 7),
		},
		{
			name: "NewUserNotFoundErrorWithCause satisfies error interface",
			err:  apperr.NewUserNotFoundErrorWithCause("msg", errors.New("cause")),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Must implement error
			assert.Implements(t, (*error)(nil), tc.err)

			// Must be detectable as *UserNotFoundError
			var target *apperr.UserNotFoundError
			assert.True(t, errors.As(tc.err, &target))
		})
	}
}

// ---------------------------------------------------------------------------
// HTTP handler integration – mapError converts UserNotFoundError → 404
// ---------------------------------------------------------------------------

// mapError is the Go equivalent of the Java controller advice / mapError
// function that converts a UserNotFoundError into an HTTP 404 response.
// It is defined here inline so the test remains self-contained.

import (
	"net/http"
	"net/http/httptest"
)

func mapError(err error, w http.ResponseWriter) {
	var notFound *apperr.UserNotFoundError
	if errors.As(err, &notFound) {
		http.Error(w, notFound.Error(), http.StatusNotFound)
		return
	}
	http.Error(w, "internal server error", http.StatusInternalServerError)
}

func TestMapError_UserNotFoundError_Returns404(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantCode   int
		wantBodyRe string
	}{
		{
			name:       "UserNotFoundError maps to 404",
			err:        apperr.NewUserNotFoundError(),
			wantCode:   http.StatusNotFound,
			wantBodyRe: apperr.DefaultUserNotFoundMessage,
		},
		{
			name:       "UserNotFoundError with message maps to 404 with message body",
			err:        apperr.NewUserNotFoundErrorf("user 123 not found"),
			wantCode:   http.StatusNotFound,
			wantBodyRe: "user 123 not found",
		},
		{
			name:       "wrapped UserNotFoundError still maps to 404",
			err:        fmt.Errorf("service: %w", apperr.NewUserNotFoundError()),
			wantCode:   http.StatusNotFound,
			wantBodyRe: apperr.DefaultUserNotFoundMessage,
		},
		{
			name:       "generic error maps to 500",
			err:        errors.New("something went wrong"),
			wantCode:   http.StatusInternalServerError,
			wantBodyRe: "internal server error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			mapError(tc.err, rr)

			assert.Equal(t, tc.wantCode, rr.Code)
			assert.Contains(t, rr.Body.String(), tc.wantBodyRe)
		})
	}
}

// ---------------------------------------------------------------------------
// DefaultUserNotFoundMessage constant
// ---------------------------------------------------------------------------

func TestDefaultUserNotFoundMessage(t *testing.T) {
	assert.Equal(t, "user not found", apperr.DefaultUserNotFoundMessage,
		"constant value must not change without updating callers")
}
```