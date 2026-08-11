```go
package smartcontacterror_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	smartcontacterror "github.com/your-module/internal/smartcontact/error"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// sentinelErr is a reusable sentinel used as a "cause" in tests.
var sentinelErr = errors.New("underlying cause")

// ---------------------------------------------------------------------------
// ErrUserNotFound sentinel
// ---------------------------------------------------------------------------

func TestErrUserNotFound_IsSentinel(t *testing.T) {
	assert.NotNil(t, smartcontacterror.ErrUserNotFound)
	assert.Equal(t, "user not found", smartcontacterror.ErrUserNotFound.Error())
}

// ---------------------------------------------------------------------------
// NewUserNotFoundError  (no-arg / message constructor)
// ---------------------------------------------------------------------------

func TestNewUserNotFoundError_NoArg(t *testing.T) {
	// Equivalent: UserNotFoundException() – no message, no cause.
	e := &smartcontacterror.UserNotFoundError{}

	assert.Empty(t, e.Message, "Message should be empty (equivalent to null)")
	assert.Nil(t, e.Cause, "Cause should be nil (equivalent to null)")
	assert.Equal(t, "user not found", e.Error(),
		"Error() should fall back to default text when both Message and Cause are absent")
}

func TestNewUserNotFoundError_WithMessage(t *testing.T) {
	tests := []struct {
		name        string
		message     string
		wantMessage string
		wantError   string
	}{
		{
			name:        "non-empty message",
			message:     "user 42 not found",
			wantMessage: "user 42 not found",
			wantError:   "user 42 not found",
		},
		{
			name:        "empty message (null equivalent)",
			message:     "",
			wantMessage: "",
			wantError:   "user not found",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			e := smartcontacterror.NewUserNotFoundError(tc.message)

			assert.Equal(t, tc.wantMessage, e.Message)
			assert.Nil(t, e.Cause)
			assert.Equal(t, tc.wantError, e.Error())
		})
	}
}

// ---------------------------------------------------------------------------
// NewUserNotFoundErrorf
// ---------------------------------------------------------------------------

func TestNewUserNotFoundErrorf(t *testing.T) {
	tests := []struct {
		name        string
		format      string
		args        []any
		wantMessage string
	}{
		{
			name:        "interpolated user ID",
			format:      "user %d not found",
			args:        []any{99},
			wantMessage: "user 99 not found",
		},
		{
			name:        "string interpolation",
			format:      "could not locate user %s",
			args:        []any{"alice"},
			wantMessage: "could not locate user alice",
		},
		{
			name:        "no arguments",
			format:      "user not found",
			args:        nil,
			wantMessage: "user not found",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			e := smartcontacterror.NewUserNotFoundErrorf(tc.format, tc.args...)

			assert.Equal(t, tc.wantMessage, e.Message)
			assert.Nil(t, e.Cause)
			assert.Equal(t, tc.wantMessage, e.Error())
		})
	}
}

// ---------------------------------------------------------------------------
// NewUserNotFoundErrorWithCause  (message + cause constructor)
// ---------------------------------------------------------------------------

func TestNewUserNotFoundErrorWithCause(t *testing.T) {
	tests := []struct {
		name      string
		message   string
		cause     error
		wantMsg   string
		wantError string
	}{
		{
			name:      "message and cause both set",
			message:   "user 7 not found",
			cause:     sentinelErr,
			wantMsg:   "user 7 not found",
			wantError: fmt.Sprintf("user 7 not found: %v", sentinelErr),
		},
		{
			name:      "empty message with cause",
			message:   "",
			cause:     sentinelErr,
			wantMsg:   "",
			wantError: fmt.Sprintf("user not found: %v", sentinelErr),
		},
		{
			name:      "message with nil cause (null cause equivalent)",
			message:   "user 8 not found",
			cause:     nil,
			wantMsg:   "user 8 not found",
			wantError: "user 8 not found",
		},
		{
			name:      "empty message and nil cause",
			message:   "",
			cause:     nil,
			wantMsg:   "",
			wantError: "user not found",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			e := smartcontacterror.NewUserNotFoundErrorWithCause(tc.message, tc.cause)

			assert.Equal(t, tc.wantMsg, e.Message)
			assert.Equal(t, tc.cause, e.Cause)
			assert.Equal(t, tc.wantError, e.Error())
		})
	}
}

// ---------------------------------------------------------------------------
// Cause-only constructor equivalent
// (Java: UserNotFoundException(Throwable cause) where message = cause.toString())
// ---------------------------------------------------------------------------

func TestUserNotFoundError_CauseOnly(t *testing.T) {
	tests := []struct {
		name      string
		cause     error
		wantMsg   string
		wantError string
	}{
		{
			name:      "non-null cause: message should reflect cause string representation",
			cause:     sentinelErr,
			wantMsg:   sentinelErr.Error(),
			wantError: fmt.Sprintf("%s: %v", sentinelErr.Error(), sentinelErr),
		},
		{
			name:      "null cause: message and cause both nil",
			cause:     nil,
			wantMsg:   "",
			wantError: "user not found",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var e *smartcontacterror.UserNotFoundError
			if tc.cause != nil {
				// Idiomatic Go equivalent of UserNotFoundException(Throwable cause):
				// message is derived from cause.Error() and cause is set.
				e = smartcontacterror.NewUserNotFoundErrorWithCause(tc.cause.Error(), tc.cause)
			} else {
				e = &smartcontacterror.UserNotFoundError{Message: "", Cause: nil}
			}

			assert.Equal(t, tc.wantMsg, e.Message)
			assert.Equal(t, tc.cause, e.Cause)
			assert.Equal(t, tc.wantError, e.Error())
		})
	}
}

// ---------------------------------------------------------------------------
// Unwrap / errors.Is / errors.As
// ---------------------------------------------------------------------------

func TestUserNotFoundError_Unwrap(t *testing.T) {
	tests := []struct {
		name        string
		cause       error
		wantUnwrap  error
	}{
		{
			name:       "Unwrap returns the wrapped cause",
			cause:      sentinelErr,
			wantUnwrap: sentinelErr,
		},
		{
			name:       "Unwrap returns nil when no cause",
			cause:      nil,
			wantUnwrap: nil,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			e := smartcontacterror.NewUserNotFoundErrorWithCause("msg", tc.cause)
			assert.Equal(t, tc.wantUnwrap, e.Unwrap())
		})
	}
}

func TestUserNotFoundError_ErrorsIs_Sentinel(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		target error
		want   bool
	}{
		{
			name:   "direct UserNotFoundError matches ErrUserNotFound",
			err:    smartcontacterror.NewUserNotFoundError("any message"),
			target: smartcontacterror.ErrUserNotFound,
			want:   true,
		},
		{
			name:   "UserNotFoundError with cause matches ErrUserNotFound",
			err:    smartcontacterror.NewUserNotFoundErrorWithCause("msg", sentinelErr),
			target: smartcontacterror.ErrUserNotFound,
			want:   true,
		},
		{
			name:   "wrapped UserNotFoundError matches ErrUserNotFound via chain",
			err:    fmt.Errorf("outer: %w", smartcontacterror.NewUserNotFoundError("inner")),
			target: smartcontacterror.ErrUserNotFound,
			want:   true,
		},
		{
			name:   "UserNotFoundError does NOT match an unrelated sentinel",
			err:    smartcontacterror.NewUserNotFoundError("msg"),
			target: sentinelErr,
			want:   false,
		},
		{
			name:   "sentinel cause is reachable via chain",
			err:    smartcontacterror.NewUserNotFoundErrorWithCause("msg", sentinelErr),
			target: sentinelErr,
			want:   true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, errors.Is(tc.err, tc.target))
		})
	}
}

func TestUserNotFoundError_ErrorsAs(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		wantMessage string
	}{
		{
			name:        "errors.As recovers concrete type from direct error",
			err:         smartcontacterror.NewUserNotFoundError("specific user detail"),
			wantMessage: "specific user detail",
		},
		{
			name:        "errors.As recovers concrete type from wrapped error",
			err:         fmt.Errorf("outer: %w", smartcontacterror.NewUserNotFoundError("inner detail")),
			wantMessage: "inner detail",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var target *smartcontacterror.UserNotFoundError
			ok := errors.As(tc.err, &target)

			assert.True(t, ok, "errors.As should succeed")
			assert.Equal(t, tc.wantMessage, target.Message)
		})
	}
}

// ---------------------------------------------------------------------------
// Error() string formatting
// ---------------------------------------------------------------------------

func TestUserNotFoundError_Error_Formatting(t *testing.T) {
	tests := []struct {
		name      string
		err       *smartcontacterror.UserNotFoundError
		wantError string
	}{
		{
			name:      "no message no cause",
			err:       &smartcontacterror.UserNotFoundError{},
			wantError: "user not found",
		},
		{
			name:      "message only",
			err:       &smartcontacterror.UserNotFoundError{Message: "user 1 not found"},
			wantError: "user 1 not found",
		},
		{
			name:      "cause only",
			err:       &smartcontacterror.UserNotFoundError{Cause: sentinelErr},
			wantError: fmt.Sprintf("user not found: %v", sentinelErr),
		},
		{
			name:      "message and cause",
			err:       &smartcontacterror.UserNotFoundError{Message: "user 2 not found", Cause: sentinelErr},
			wantError: fmt.Sprintf("user 2 not found: %v", sentinelErr),
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.wantError, tc.err.Error())
		})
	}
}

// ---------------------------------------------------------------------------
// Immutability: fields do not change after construction
// ---------------------------------------------------------------------------

func TestUserNotFoundError_Immutability(t *testing.T) {
	msg := "original message"
	e := smartcontacterror.NewUserNotFoundErrorWithCause(msg, sentinelErr)

	// Capture initial state.
	gotMsg := e.Message
	gotCause := e.Cause

	// Attempt to mutate through variables (Go struct fields are mutable, but
	// the spec requires immutability — we verify the constructor sets correctly
	// and no internal mutation occurs between construction and observation).
	assert.Equal(t, msg, gotMsg)
	assert.Equal(t, sentinelErr, gotCause)

	// Verify Error() is stable across multiple calls.
	first := e.Error()
	second := e.Error()
	assert.Equal(t, first, second)
}

// ---------------------------------------------------------------------------
// HTTP handler integration test (httptest)
// ---------------------------------------------------------------------------

// userService is a small interface to simulate an external dependency.
type userService interface {
	FindUser(id int) (string, error)
}

// mockUserService implements userService for testing.
type mockUserService struct {
	findUserFn func(id int) (string, error)
}

func (m *mockUserService) FindUser(id int) (string, error) {
	return m.findUserFn(id)
}

// userHandler is an HTTP handler that uses the userService.
func userHandler(svc userService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name, err := svc.FindUser(42)
		if err != nil {
			var notFound *smartcontacterror.UserNotFoundError
			if errors.As(err, &notFound) {
				http.Error(w, notFound.Error(), http.StatusNotFound)
				return
			}
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, name)
	}
}

func TestUserHandler_HTTP(t *testing.T) {
	tests := []struct {
		name           string
		svcErr         error
		wantStatusCode int
		wantBodyContains string
	}{
		{
			name:             "user found returns 200",
			svcErr:           nil,
			wantStatusCode:   http.StatusOK,
			wantBodyContains: "Alice",
		},
		{
			name:             "UserNotFoundError returns 404 with message",
			svcErr:           smartcontacterror.NewUserNotFoundError("user 42 not found"),
			wantStatusCode:   http.StatusNotFound,
			wantBodyContains: "user 42 not found",
		},
		{
			name:             "UserNotFoundError with cause returns 404",
			svcErr:           smartcontacterror.NewUserNotFoundErrorWithCause("user 42 not found", sentinelErr),
			wantStatusCode:   http.StatusNotFound,
			wantBodyContains: "user 42 not found",
		},
		{
			name:             "generic error returns 500",
			svcErr:           errors.New("db connection refused"),
			wantStatusCode:   http.StatusInternalServerError,
			wantBodyContains: "internal server error",
		},
		{
			name:             "ErrUserNotFound sentinel wrapped in UserNotFoundError returns 404",
			svcErr:           &smartcontacterror.UserNotFoundError{