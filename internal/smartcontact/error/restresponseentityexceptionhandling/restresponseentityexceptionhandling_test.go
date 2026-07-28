```go
package error

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/smartContact/internal/smartcontact/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// userNotFoundError is a local sentinel that satisfies IsUserNotFound so we can
// test without importing the real domain error package.
type userNotFoundError struct{ msg string }

func (e *userNotFoundError) Error() string        { return e.msg }
func (e *userNotFoundError) IsUserNotFound() bool { return true }

// genericError is an ordinary error unrelated to user-not-found.
type genericError struct{ msg string }

func (e *genericError) Error() string { return e.msg }

// decodeBody is a test helper that decodes the response body into an
// ErrorMessage, failing the test immediately on any issue.
func decodeBody(t *testing.T, rr *httptest.ResponseRecorder) model.ErrorMessage {
	t.Helper()
	var em model.ErrorMessage
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&em),
		"response body must be a valid ErrorMessage JSON object")
	return em
}

// ---------------------------------------------------------------------------
// WriteError
// ---------------------------------------------------------------------------

func TestWriteError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		wantStatus     int
		wantMsgContain string
	}{
		{
			name:           "UserNotFoundException maps to 404",
			err:            &userNotFoundError{msg: "user 42 not found"},
			wantStatus:     http.StatusNotFound,
			wantMsgContain: "user 42 not found",
		},
		{
			name:           "UserNotFoundException with empty message still returns 404",
			err:            &userNotFoundError{msg: ""},
			wantStatus:     http.StatusNotFound,
			wantMsgContain: "",
		},
		{
			name:           "generic error maps to 500",
			err:            &genericError{msg: "something went wrong"},
			wantStatus:     http.StatusInternalServerError,
			wantMsgContain: "something went wrong",
		},
		{
			name:           "standard errors.New maps to 500",
			err:            errors.New("unexpected failure"),
			wantStatus:     http.StatusInternalServerError,
			wantMsgContain: "unexpected failure",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()

			writeErr := WriteError(rr, tc.err)

			// WriteError itself should not return an error for an in-memory recorder.
			assert.NoError(t, writeErr, "WriteError must not return an encoding error for a healthy ResponseRecorder")

			// HTTP status
			assert.Equal(t, tc.wantStatus, rr.Code, "HTTP status code mismatch")

			// Content-Type header
			assert.Equal(t, "application/json", rr.Header().Get("Content-Type"),
				"Content-Type must be application/json")

			// Decode body
			em := decodeBody(t, rr)

			// ErrorMessage.Status must mirror the HTTP status.
			assert.Equal(t, tc.wantStatus, em.Status(),
				"ErrorMessage status field must equal HTTP status code")

			// ErrorMessage.Message must mirror the error string.
			assert.Equal(t, tc.wantMsgContain, em.Message(),
				"ErrorMessage message field must mirror the error message")

			// Body must be a non-null / non-zero ErrorMessage.
			assert.NotZero(t, em, "response body must be a non-zero ErrorMessage")
		})
	}
}

// TestWriteError_ResponseBodyIsNonNull checks the invariant that the body is
// always present regardless of error type.
func TestWriteError_ResponseBodyIsNonNull(t *testing.T) {
	errs := []error{
		&userNotFoundError{msg: "some user"},
		&genericError{msg: "boom"},
		errors.New("plain"),
	}
	for _, e := range errs {
		rr := httptest.NewRecorder()
		require.NoError(t, WriteError(rr, e))
		assert.Positive(t, rr.Body.Len(), "response body must not be empty")
	}
}

// TestWriteError_StatusFieldMatchesHTTPStatus validates the invariant that the
// ErrorMessage status field always matches the HTTP response status code.
func TestWriteError_StatusFieldMatchesHTTPStatus(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"user not found", &userNotFoundError{msg: "gone"}},
		{"internal server error", errors.New("kaboom")},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			require.NoError(t, WriteError(rr, tc.err))
			em := decodeBody(t, rr)
			assert.Equal(t, rr.Code, em.Status(),
				"ErrorMessage.Status must equal HTTP response code")
		})
	}
}

// ---------------------------------------------------------------------------
// ErrorMapper middleware
// ---------------------------------------------------------------------------

func TestErrorMapper_NormalHandler(t *testing.T) {
	// A handler that writes 200 and a simple body should pass through unchanged.
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	handler := ErrorMapper(inner)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "ok", rr.Body.String())
}

func TestErrorMapper_PanicRecovery(t *testing.T) {
	tests := []struct {
		name      string
		panicWith interface{}
	}{
		{"panic with string", "something panicked"},
		{"panic with error", errors.New("fatal error")},
		{"panic with nil value", "nil-ish panic"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic(tc.panicWith)
			})

			handler := ErrorMapper(inner)
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/panic", nil)

			// ErrorMapper must not re-panic; if it does, the test will fail.
			assert.NotPanics(t, func() {
				handler.ServeHTTP(rr, req)
			})

			// Status must be 500.
			assert.Equal(t, http.StatusInternalServerError, rr.Code)

			// Content-Type must be application/json.
			assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

			// Body must be a valid ErrorMessage.
			em := decodeBody(t, rr)
			assert.Equal(t, http.StatusInternalServerError, em.Status())
			assert.Equal(t, http.StatusText(http.StatusInternalServerError), em.Message())
		})
	}
}

func TestErrorMapper_HandlerWritesErrorViaWriteError(t *testing.T) {
	// Simulate the expected usage pattern: a handler calls WriteError for a
	// UserNotFoundException; ErrorMapper just passes through.
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = WriteError(w, &userNotFoundError{msg: "user 7 not found"})
	})

	handler := ErrorMapper(inner)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/users/7", nil)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	em := decodeBody(t, rr)
	assert.Equal(t, http.StatusNotFound, em.Status())
	assert.Equal(t, "user 7 not found", em.Message())
}

// TestErrorMapper_DoesNotAffectNonUserNotFoundErrors ensures the middleware
// simply passes through responses written by WriteError for non-404 errors.
func TestErrorMapper_DoesNotAffectNonUserNotFoundErrors(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = WriteError(w, errors.New("database offline"))
	})

	handler := ErrorMapper(inner)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/resource", nil)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	em := decodeBody(t, rr)
	assert.Equal(t, http.StatusInternalServerError, em.Status())
	assert.Equal(t, "database offline", em.Message())
}

// ---------------------------------------------------------------------------
// Invariant: only UserNotFoundException produces 404; everything else → 500
// ---------------------------------------------------------------------------

func TestWriteError_OnlyUserNotFoundProduces404(t *testing.T) {
	nonUserNotFoundErrors := []error{
		errors.New("generic"),
		&genericError{msg: "other domain error"},
	}
	for _, e := range nonUserNotFoundErrors {
		rr := httptest.NewRecorder()
		require.NoError(t, WriteError(rr, e))
		assert.Equal(t, http.StatusInternalServerError, rr.Code,
			"non-UserNotFoundException errors must produce 500, not 404")
	}
}

func TestWriteError_UserNotFoundAlwaysProduces404(t *testing.T) {
	messages := []string{"not found", "", "user 0 missing", "user abc missing"}
	for _, msg := range messages {
		msg := msg
		t.Run("msg="+msg, func(t *testing.T) {
			rr := httptest.NewRecorder()
			require.NoError(t, WriteError(rr, &userNotFoundError{msg: msg}))
			assert.Equal(t, http.StatusNotFound, rr.Code,
				"UserNotFoundException must always produce 404")
			em := decodeBody(t, rr)
			assert.Equal(t, http.StatusNotFound, em.Status())
			assert.Equal(t, msg, em.Message())
		})
	}
}

// ---------------------------------------------------------------------------
// No external side-effects: repeated calls are idempotent on independent recorders
// ---------------------------------------------------------------------------

func TestWriteError_NoExternalSideEffects(t *testing.T) {
	err := &userNotFoundError{msg: "user 1"}

	rr1 := httptest.NewRecorder()
	rr2 := httptest.NewRecorder()

	require.NoError(t, WriteError(rr1, err))
	require.NoError(t, WriteError(rr2, err))

	assert.Equal(t, rr1.Code, rr2.Code)
	assert.Equal(t, rr1.Body.String(), rr2.Body.String())
}
```