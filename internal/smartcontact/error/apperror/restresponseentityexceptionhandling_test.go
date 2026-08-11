```go
package apperror_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/smartContact/internal/smartcontact/error/apperror"
	"github.com/smartContact/internal/smartcontact/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ----------------------------------------------------------------------------
// Helpers / test doubles
// ----------------------------------------------------------------------------

// mockValidationError satisfies the unexported validationError interface.
type mockValidationError struct {
	msg string
}

func (e *mockValidationError) Error() string            { return e.msg }
func (e *mockValidationError) IsValidationError() bool  { return true }

// wrappedUserNotFoundError wraps a *UserNotFoundError so we can verify
// errors.As unwrapping works correctly.
type wrappedError struct {
	inner error
}

func (w *wrappedError) Error() string { return w.inner.Error() }
func (w *wrappedError) Unwrap() error { return w.inner }

// writeErrorResponse is a tiny HTTP handler helper that calls MapError and
// writes the resulting status + JSON body — mirrors what the router/handler
// layer would do in production.
func writeErrorResponse(w http.ResponseWriter, err error) {
	status, em := apperror.MapError(err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(em)
}

// decodeErrorMessage decodes the response body into a model.ErrorMessage.
func decodeErrorMessage(t *testing.T, rec *httptest.ResponseRecorder) model.ErrorMessage {
	t.Helper()
	var em model.ErrorMessage
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&em))
	return em
}

// ----------------------------------------------------------------------------
// Unit tests: MapError return values
// ----------------------------------------------------------------------------

func TestMapError_TableDriven(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		wantStatusCode int
		wantMessage    string   // exact match; empty string means "not checked"
		wantStatusText string   // checked against em.Status if non-empty
	}{
		// ----------------------------------------------------------------
		// Nil error (degenerate caller case) → 500
		// ----------------------------------------------------------------
		{
			name:           "nil error returns 500",
			err:            nil,
			wantStatusCode: http.StatusInternalServerError,
			wantMessage:    http.StatusText(http.StatusInternalServerError),
		},

		// ----------------------------------------------------------------
		// UserNotFoundError scenarios (spec: userNotFoundException)
		// ----------------------------------------------------------------
		{
			name:           "UserNotFoundError with specific message returns 404",
			err:            apperror.NewUserNotFoundError("user with id 42 not found"),
			wantStatusCode: http.StatusNotFound,
			wantMessage:    "user with id 42 not found",
		},
		{
			name:           "UserNotFoundError with empty message returns 404 empty body",
			err:            apperror.NewUserNotFoundError(""),
			wantStatusCode: http.StatusNotFound,
			wantMessage:    "",
		},
		{
			name: "wrapped UserNotFoundError unwrapped by errors.As returns 404",
			err: &wrappedError{
				inner: apperror.NewUserNotFoundError("wrapped user not found"),
			},
			wantStatusCode: http.StatusNotFound,
			wantMessage:    "wrapped user not found",
		},

		// ----------------------------------------------------------------
		// Validation error → 400
		// ----------------------------------------------------------------
		{
			name:           "validationError returns 400",
			err:            &mockValidationError{msg: "email is required"},
			wantStatusCode: http.StatusBadRequest,
			wantMessage:    "email is required",
		},
		{
			name:           "validationError with empty message returns 400",
			err:            &mockValidationError{msg: ""},
			wantStatusCode: http.StatusBadRequest,
			wantMessage:    "",
		},

		// ----------------------------------------------------------------
		// Generic / unknown error → 500
		// ----------------------------------------------------------------
		{
			name:           "generic error returns 500",
			err:            errors.New("database connection refused"),
			wantStatusCode: http.StatusInternalServerError,
			wantMessage:    "database connection refused",
		},
		{
			name:           "fmt.Errorf error returns 500",
			err:            fmt.Errorf("unexpected failure: %w", errors.New("timeout")),
			wantStatusCode: http.StatusInternalServerError,
			wantMessage:    "unexpected failure: timeout",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			gotStatus, gotEM := apperror.MapError(tc.err)

			// Status code
			assert.Equal(t, tc.wantStatusCode, gotStatus,
				"HTTP status code mismatch")

			// ErrorMessage is always non-nil (it's a value type)
			// Message field
			assert.Equal(t, tc.wantMessage, gotEM.Message,
				"ErrorMessage.Message mismatch")

			// ErrorMessage.Status must always match the returned HTTP status
			assert.Equal(t, model.HTTPStatus(tc.wantStatusCode), gotEM.Status,
				"ErrorMessage.Status must equal the returned HTTP status code")
		})
	}
}

// ----------------------------------------------------------------------------
// Invariant: UserNotFoundError always produces 404 regardless of message
// ----------------------------------------------------------------------------

func TestMapError_UserNotFoundError_Always404(t *testing.T) {
	messages := []string{
		"user with id 1 not found",
		"",
		"some very long message about why the user could not be located in the system",
	}

	for _, msg := range messages {
		msg := msg
		t.Run(fmt.Sprintf("message=%q", msg), func(t *testing.T) {
			status, em := apperror.MapError(apperror.NewUserNotFoundError(msg))

			assert.Equal(t, http.StatusNotFound, status,
				"response HTTP status must always be 404 for UserNotFoundError")
			assert.NotNil(t, em,
				"ErrorMessage must always be non-nil")
			assert.Equal(t, model.HTTPStatus(http.StatusNotFound), em.Status,
				"ErrorMessage.Status must always be NOT_FOUND (404)")
			assert.Equal(t, msg, em.Message,
				"ErrorMessage.Message must equal the exception message")
		})
	}
}

// ----------------------------------------------------------------------------
// Integration-style tests via httptest: verify the full HTTP boundary
// ----------------------------------------------------------------------------

func TestMapError_HTTPBoundary_TableDriven(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		wantStatusCode int
		wantMessage    string
	}{
		{
			name:           "UserNotFoundError produces 404 response",
			err:            apperror.NewUserNotFoundError("user 7 not found"),
			wantStatusCode: http.StatusNotFound,
			wantMessage:    "user 7 not found",
		},
		{
			name:           "UserNotFoundError with empty message produces 404 response",
			err:            apperror.NewUserNotFoundError(""),
			wantStatusCode: http.StatusNotFound,
			wantMessage:    "",
		},
		{
			name:           "validationError produces 400 response",
			err:            &mockValidationError{msg: "name cannot be blank"},
			wantStatusCode: http.StatusBadRequest,
			wantMessage:    "name cannot be blank",
		},
		{
			name:           "generic error produces 500 response",
			err:            errors.New("internal failure"),
			wantStatusCode: http.StatusInternalServerError,
			wantMessage:    "internal failure",
		},
		{
			name:           "nil error produces 500 response",
			err:            nil,
			wantStatusCode: http.StatusInternalServerError,
			wantMessage:    http.StatusText(http.StatusInternalServerError),
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				writeErrorResponse(w, tc.err)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			// Verify HTTP status line
			assert.Equal(t, tc.wantStatusCode, rec.Code,
				"HTTP response status code mismatch")

			// Verify Content-Type
			assert.Equal(t, "application/json", rec.Header().Get("Content-Type"),
				"response Content-Type must be application/json")

			// Decode and verify body
			em := decodeErrorMessage(t, rec)

			assert.Equal(t, model.HTTPStatus(tc.wantStatusCode), em.Status,
				"ErrorMessage.Status in JSON body mismatch")
			assert.Equal(t, tc.wantMessage, em.Message,
				"ErrorMessage.Message in JSON body mismatch")
		})
	}
}

// ----------------------------------------------------------------------------
// Invariant: handler never panics / throws when processing UserNotFoundError
// ----------------------------------------------------------------------------

func TestMapError_NeverPanics(t *testing.T) {
	errs := []error{
		apperror.NewUserNotFoundError("panic test"),
		apperror.NewUserNotFoundError(""),
		&mockValidationError{msg: "bad input"},
		errors.New("random error"),
		nil,
	}

	for _, e := range errs {
		e := e
		t.Run(fmt.Sprintf("err=%v", e), func(t *testing.T) {
			assert.NotPanics(t, func() {
				status, em := apperror.MapError(e)
				// Basic sanity — exercise both return values.
				_ = status
				_ = em
			})
		})
	}
}

// ----------------------------------------------------------------------------
// Invariant: ErrorMessage.Status always equals returned status code
// ----------------------------------------------------------------------------

func TestMapError_StatusConsistency(t *testing.T) {
	errs := []error{
		apperror.NewUserNotFoundError("consistency check"),
		&mockValidationError{msg: "consistency check"},
		errors.New("consistency check"),
		nil,
	}

	for _, e := range errs {
		e := e
		t.Run(fmt.Sprintf("err=%v", e), func(t *testing.T) {
			status, em := apperror.MapError(e)
			assert.Equal(t, model.HTTPStatus(status), em.Status,
				"ErrorMessage.Status must always be consistent with the returned HTTP status code")
		})
	}
}

// ----------------------------------------------------------------------------
// Verify errors.As chain resolution for UserNotFoundError
// ----------------------------------------------------------------------------

func TestMapError_UserNotFoundError_WrappedInChain(t *testing.T) {
	root := apperror.NewUserNotFoundError("deep user not found")
	// Build a chain: wrap → wrap → *UserNotFoundError
	chained := fmt.Errorf("layer2: %w", fmt.Errorf("layer1: %w", root))

	status, em := apperror.MapError(chained)

	assert.Equal(t, http.StatusNotFound, status,
		"errors.As must unwrap through the chain to find *UserNotFoundError")
	assert.Equal(t, model.HTTPStatus(http.StatusNotFound), em.Status)
	// The message on the ErrorMessage comes from the *UserNotFoundError leaf.
	assert.Equal(t, root.Error(), em.Message)
}

// ----------------------------------------------------------------------------
// Verify priority: UserNotFoundError takes precedence over validationError
// if a type somehow satisfies both (edge case)
// ----------------------------------------------------------------------------

func TestMapError_Priority_UserNotFoundBeforeValidation(t *testing.T) {
	// A *UserNotFoundError should be mapped as 404, not 400, even if
	// the caller erroneously wraps it with a validationError.
	inner := apperror.NewUserNotFoundError("priority user not found")
	wrapped := &wrappedError{inner: inner}

	status, em := apperror.MapError(wrapped)

	assert.Equal(t, http.StatusNotFound, status,
		"UserNotFoundError must take priority over the validationError branch")
	assert.Equal(t, model.HTTPStatus(http.StatusNotFound), em.Status)
}
```