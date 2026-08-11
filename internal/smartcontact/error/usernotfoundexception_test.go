```go
package apperr_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	apperr "github.com/example/smartcontact/internal/smartcontact/error"
)

// sentinel errors used as causes in tests
var (
	errSentinel = errors.New("underlying cause")
	errWrapped  = fmt.Errorf("wrapped: %w", errSentinel)
)

// ---------------------------------------------------------------------------
// NewUserNotFound – mirrors Java's no-arg and message-only constructors
// ---------------------------------------------------------------------------

func TestNewUserNotFound(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		message         string
		wantMessage     string
		wantCause       error
		wantErrorString string
	}{
		{
			// Java: UserNotFoundException()  – no-arg constructor
			name:            "no-arg equivalent: empty message, nil cause",
			message:         "",
			wantMessage:     "",
			wantCause:       nil,
			wantErrorString: "user not found",
		},
		{
			// Java: UserNotFoundException(String message)  – non-null message
			name:            "non-empty message, nil cause",
			message:         "user with id 42 not found",
			wantMessage:     "user with id 42 not found",
			wantCause:       nil,
			wantErrorString: "user with id 42 not found",
		},
		{
			// Java: UserNotFoundException(String message) – null/empty message
			name:            "empty string message treated as unset",
			message:         "",
			wantMessage:     "",
			wantCause:       nil,
			wantErrorString: "user not found",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			e := apperr.NewUserNotFound(tc.message)

			assert.NotNil(t, e, "constructor must return non-nil pointer")
			assert.Equal(t, tc.wantMessage, e.Message, "Message field")
			assert.Equal(t, tc.wantCause, e.Cause, "Cause field must be nil")
			assert.Nil(t, e.Unwrap(), "Unwrap must return nil when no cause given")
			assert.Equal(t, tc.wantErrorString, e.Error(), "Error() string")

			// must satisfy the error interface
			var target *apperr.UserNotFound
			assert.True(t, errors.As(e, &target), "errors.As must match *UserNotFound")
		})
	}
}

// ---------------------------------------------------------------------------
// NewUserNotFoundWithCause – mirrors Java's (String, Throwable) and
// (Throwable) constructors
// ---------------------------------------------------------------------------

func TestNewUserNotFoundWithCause(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		message         string
		cause           error
		wantMessage     string
		wantCause       error
		wantErrorString string
	}{
		{
			// Java: UserNotFoundException(String message, Throwable cause) – both non-null
			name:            "message and cause both provided",
			message:         "cannot locate user",
			cause:           errSentinel,
			wantMessage:     "cannot locate user",
			wantCause:       errSentinel,
			wantErrorString: "cannot locate user: underlying cause",
		},
		{
			// Java: UserNotFoundException(String message, Throwable cause) – null message, null cause
			name:            "empty message and nil cause",
			message:         "",
			cause:           nil,
			wantMessage:     "",
			wantCause:       nil,
			wantErrorString: "user not found",
		},
		{
			// Java: UserNotFoundException(Throwable cause) – non-null cause, message derived from cause
			// In Go we model this as empty message + cause; Error() then uses the cause.
			name:            "empty message with non-nil cause (Throwable-only ctor equivalent)",
			message:         "",
			cause:           errSentinel,
			wantMessage:     "",
			wantCause:       errSentinel,
			wantErrorString: "user not found: underlying cause",
		},
		{
			// Java: UserNotFoundException(Throwable cause) – null cause
			name:            "empty message with nil cause",
			message:         "",
			cause:           nil,
			wantMessage:     "",
			wantCause:       nil,
			wantErrorString: "user not found",
		},
		{
			// cause is itself a wrapped error
			name:            "cause is a wrapped error",
			message:         "lookup failed",
			cause:           errWrapped,
			wantMessage:     "lookup failed",
			wantCause:       errWrapped,
			wantErrorString: "lookup failed: wrapped: underlying cause",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			e := apperr.NewUserNotFoundWithCause(tc.message, tc.cause)

			assert.NotNil(t, e, "constructor must return non-nil pointer")
			assert.Equal(t, tc.wantMessage, e.Message, "Message field")
			assert.Equal(t, tc.wantCause, e.Cause, "Cause field")
			assert.Equal(t, tc.wantCause, e.Unwrap(), "Unwrap must return the same cause")
			assert.Equal(t, tc.wantErrorString, e.Error(), "Error() string")

			// must satisfy the error interface via errors.As
			var target *apperr.UserNotFound
			assert.True(t, errors.As(e, &target), "errors.As must match *UserNotFound")
		})
	}
}

// ---------------------------------------------------------------------------
// Error() method – standalone table
// ---------------------------------------------------------------------------

func TestUserNotFound_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  *apperr.UserNotFound
		want string
	}{
		{
			name: "no message, no cause",
			err:  &apperr.UserNotFound{},
			want: "user not found",
		},
		{
			name: "message only",
			err:  &apperr.UserNotFound{Message: "user 99 missing"},
			want: "user 99 missing",
		},
		{
			name: "cause only",
			err:  &apperr.UserNotFound{Cause: errSentinel},
			want: "user not found: underlying cause",
		},
		{
			name: "message and cause",
			err:  &apperr.UserNotFound{Message: "not found", Cause: errSentinel},
			want: "not found: underlying cause",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, tc.err.Error())
		})
	}
}

// ---------------------------------------------------------------------------
// Unwrap / errors.Is / errors.As chain traversal
// ---------------------------------------------------------------------------

func TestUserNotFound_ErrorChain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		err           *apperr.UserNotFound
		targetsIs     []error // errors.Is should return true for these
		targetsIsNot  []error // errors.Is should return false for these
		unwrapsTo     error
	}{
		{
			name:      "no cause: Unwrap returns nil, Is only matches itself",
			err:       apperr.NewUserNotFound("msg"),
			unwrapsTo: nil,
			targetsIsNot: []error{errSentinel},
		},
		{
			name:      "with sentinel cause: Is traverses chain",
			err:       apperr.NewUserNotFoundWithCause("msg", errSentinel),
			unwrapsTo: errSentinel,
			targetsIs: []error{errSentinel},
		},
		{
			name:      "with wrapped cause: Is finds inner sentinel",
			err:       apperr.NewUserNotFoundWithCause("msg", errWrapped),
			unwrapsTo: errWrapped,
			targetsIs: []error{errWrapped, errSentinel},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.unwrapsTo, tc.err.Unwrap(), "Unwrap")

			for _, target := range tc.targetsIs {
				assert.True(t, errors.Is(tc.err, target),
					"errors.Is(%v, %v) should be true", tc.err, target)
			}
			for _, target := range tc.targetsIsNot {
				assert.False(t, errors.Is(tc.err, target),
					"errors.Is(%v, %v) should be false", tc.err, target)
			}

			// errors.As must always find *UserNotFound itself
			var asTarget *apperr.UserNotFound
			assert.True(t, errors.As(tc.err, &asTarget))
			assert.Equal(t, tc.err, asTarget)
		})
	}
}

// ---------------------------------------------------------------------------
// HTTP middleware integration via httptest
// Demonstrates that errors.As can be used to map UserNotFound → 404.
// ---------------------------------------------------------------------------

// userLookup is the interface that an HTTP handler depends on.
// Mocked in tests to avoid real DB calls.
type userLookup interface {
	Find(id string) error
}

// mockUserLookup is a test double.
type mockUserLookup struct {
	err error
}

func (m *mockUserLookup) Find(_ string) error { return m.err }

// userHandler is a tiny HTTP handler that uses errors.As for error mapping.
func userHandler(svc userLookup) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if err := svc.Find(id); err != nil {
			var notFound *apperr.UserNotFound
			if errors.As(err, &notFound) {
				http.Error(w, notFound.Error(), http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}
}

func TestUserHandler_HTTPMiddleware(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		serviceErr error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "no error → 200",
			serviceErr: nil,
			wantStatus: http.StatusOK,
			wantBody:   "ok",
		},
		{
			name:       "UserNotFound direct → 404",
			serviceErr: apperr.NewUserNotFound("user 7 not found"),
			wantStatus: http.StatusNotFound,
			wantBody:   "user 7 not found",
		},
		{
			name:       "UserNotFound with cause → 404",
			serviceErr: apperr.NewUserNotFoundWithCause("lookup failed", errSentinel),
			wantStatus: http.StatusNotFound,
			wantBody:   "lookup failed: underlying cause",
		},
		{
			name:       "UserNotFound wrapped in fmt.Errorf → 404 via errors.As",
			serviceErr: fmt.Errorf("service layer: %w", apperr.NewUserNotFound("no such user")),
			wantStatus: http.StatusNotFound,
			wantBody:   "no such user",
		},
		{
			name:       "generic error → 500",
			serviceErr: errors.New("database connection refused"),
			wantStatus: http.StatusInternalServerError,
			wantBody:   "database connection refused",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svc := &mockUserLookup{err: tc.serviceErr}
			handler := userHandler(svc)

			req := httptest.NewRequest(http.MethodGet, "/user?id=42", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			resp := rec.Result()
			defer resp.Body.Close()

			assert.Equal(t, tc.wantStatus, resp.StatusCode, "HTTP status code")

			// http.Error appends a newline, so we check Contains
			body := rec.Body.String()
			assert.Contains(t, body, tc.wantBody, "response body")
		})
	}
}

// ---------------------------------------------------------------------------
// Protected-constructor equivalent: full-field struct literal
// Java's protected 5-arg constructor is not directly representable in Go,
// but the struct can be composed directly to achieve the same coverage.
// ---------------------------------------------------------------------------

func TestUserNotFound_DirectStructConstruction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		err             *apperr.UserNotFound
		wantMessage     string
		wantCause       error
		wantErrorString string
	}{
		{
			// Java protected ctor: (message, cause, enableSuppression=true, writableStackTrace=true)
			name:            "message and cause via direct struct literal",
			err:             &apperr.UserNotFound{Message: "protected ctor msg", Cause: errSentinel},
			wantMessage:     "protected ctor msg",
			wantCause:       errSentinel,
			wantErrorString: "protected ctor msg: underlying cause",
		},
		{
			// writableStackTrace=false equivalent: zero-value struct
			name:            "zero-value struct (all defaults)",
			err:             &apperr.UserNotFound{},
			wantMessage:     "",
			wantCause:       nil,
			wantErrorString: "user not found",
		},
		{
			// enableSuppression=false equivalent: still valid struct
			name:            "cause only via struct literal",
			err:             &apperr.UserNotFound{Cause: errSentinel},
			wantMessage:     "",
			wantCause:       errSentinel,
			wantErrorString: "user not found: underlying cause",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.wantMessage, tc.err.Message)
			assert.Equal(t, tc.wantCause, tc.err.Cause)
			assert.Equal(t, tc.wantCause, tc.err.Unwrap())
			assert.Equal(t, tc.wantErrorString, tc.err.Error())

			var target *apperr.UserNotFound
			assert.True(t, errors.As(tc.err, &target))
		})
	}
}
```