```go
package smarterror_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	smarterror "github.com/example/smartcontact/internal/smartcontact/error"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// ErrUserNotFound sentinel
// ---------------------------------------------------------------------------

func TestErrUserNotFound_IsSentinel(t *testing.T) {
	assert.NotNil(t, smarterror.ErrUserNotFound)
	assert.Equal(t, "user not found", smarterror.ErrUserNotFound.Error())
}

// ---------------------------------------------------------------------------
// NewUserNotFoundError – mirrors Java UserNotFoundException(String message)
// ---------------------------------------------------------------------------

func TestNewUserNotFoundError(t *testing.T) {
	tests := []struct {
		name        string
		message     string
		wantErrText string
		wantCause   error
	}{
		{
			name:        "non-empty message",
			message:     "user 42 not found",
			wantErrText: "user 42 not found",
			wantCause:   nil,
		},
		{
			name:        "empty message falls back to sentinel text",
			message:     "",
			wantErrText: "user not found",
			wantCause:   nil,
		},
		{
			name:        "whitespace message is preserved",
			message:     "   ",
			wantErrText: "   ",
			wantCause:   nil,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := smarterror.NewUserNotFoundError(tc.message)

			require.NotNil(t, err)
			assert.Equal(t, tc.message, err.Message)
			assert.Equal(t, tc.wantErrText, err.Error())
			assert.Nil(t, err.Cause, "Cause should be nil when not provided")
			assert.Equal(t, tc.wantCause, errors.Unwrap(err))
		})
	}
}

// ---------------------------------------------------------------------------
// NewUserNotFoundErrorWithCause – mirrors Java
//   UserNotFoundException(String message, Throwable cause)
// ---------------------------------------------------------------------------

func TestNewUserNotFoundErrorWithCause(t *testing.T) {
	rootCause := errors.New("db connection lost")

	tests := []struct {
		name        string
		message     string
		cause       error
		wantErrText string
		wantCause   error
	}{
		{
			name:        "message and non-nil cause",
			message:     "user 99 not found",
			cause:       rootCause,
			wantErrText: "user 99 not found",
			wantCause:   rootCause,
		},
		{
			name:        "empty message with non-nil cause",
			message:     "",
			cause:       rootCause,
			wantErrText: "user not found",
			wantCause:   rootCause,
		},
		{
			name:        "non-empty message with nil cause",
			message:     "user not found in store",
			cause:       nil,
			wantErrText: "user not found in store",
			wantCause:   nil,
		},
		{
			name:        "empty message with nil cause",
			message:     "",
			cause:       nil,
			wantErrText: "user not found",
			wantCause:   nil,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := smarterror.NewUserNotFoundErrorWithCause(tc.message, tc.cause)

			require.NotNil(t, err)
			assert.Equal(t, tc.message, err.Message)
			assert.Equal(t, tc.wantErrText, err.Error())
			assert.Equal(t, tc.wantCause, err.Cause)
			assert.Equal(t, tc.wantCause, errors.Unwrap(err))
		})
	}
}

// ---------------------------------------------------------------------------
// Direct struct construction – mirrors Java no-arg constructor
// ---------------------------------------------------------------------------

func TestUserNotFoundError_ZeroValue(t *testing.T) {
	// Corresponds to the Java no-arg constructor: both message and cause are nil/zero.
	err := &smarterror.UserNotFoundError{}

	assert.Empty(t, err.Message, "Message should be empty (analogous to null)")
	assert.Nil(t, err.Cause, "Cause should be nil (analogous to null)")
	assert.Equal(t, "user not found", err.Error(), "falls back to sentinel text")
	assert.Nil(t, errors.Unwrap(err))
}

// ---------------------------------------------------------------------------
// Cause-only construction – mirrors Java UserNotFoundException(Throwable cause)
// ---------------------------------------------------------------------------

func TestUserNotFoundError_CauseOnly(t *testing.T) {
	rootCause := errors.New("record not found")

	tests := []struct {
		name      string
		cause     error
		wantCause error
	}{
		{
			name:      "non-nil cause",
			cause:     rootCause,
			wantCause: rootCause,
		},
		{
			name:      "nil cause",
			cause:     nil,
			wantCause: nil,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// In Go there is no cause-only constructor; we model it by leaving
			// Message empty and supplying only Cause.
			err := &smarterror.UserNotFoundError{Cause: tc.cause}

			assert.Equal(t, tc.wantCause, err.Cause)
			assert.Equal(t, tc.wantCause, errors.Unwrap(err))

			if tc.cause != nil {
				// The Java cause-only constructor sets message to cause.toString().
				// In Go we don't auto-derive the message, so we verify the caller
				// can replicate that by setting Message manually.
				derived := &smarterror.UserNotFoundError{
					Message: tc.cause.Error(),
					Cause:   tc.cause,
				}
				assert.Equal(t, tc.cause.Error(), derived.Error())
				assert.Equal(t, tc.cause, derived.Cause)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Error() method
// ---------------------------------------------------------------------------

func TestUserNotFoundError_Error(t *testing.T) {
	tests := []struct {
		name    string
		err     *smarterror.UserNotFoundError
		wantMsg string
	}{
		{
			name:    "custom message returned verbatim",
			err:     &smarterror.UserNotFoundError{Message: "could not locate user"},
			wantMsg: "could not locate user",
		},
		{
			name:    "empty message returns sentinel text",
			err:     &smarterror.UserNotFoundError{},
			wantMsg: "user not found",
		},
		{
			name:    "message with cause still returns message",
			err:     &smarterror.UserNotFoundError{Message: "boom", Cause: errors.New("root")},
			wantMsg: "boom",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.wantMsg, tc.err.Error())
		})
	}
}

// ---------------------------------------------------------------------------
// Unwrap() – enables errors.Is / errors.As chain traversal
// ---------------------------------------------------------------------------

func TestUserNotFoundError_Unwrap(t *testing.T) {
	root := errors.New("root cause")

	tests := []struct {
		name      string
		err       *smarterror.UserNotFoundError
		wantCause error
	}{
		{
			name:      "returns wrapped cause when set",
			err:       &smarterror.UserNotFoundError{Cause: root},
			wantCause: root,
		},
		{
			name:      "returns nil when no cause",
			err:       &smarterror.UserNotFoundError{},
			wantCause: nil,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.wantCause, tc.err.Unwrap())
			assert.Equal(t, tc.wantCause, errors.Unwrap(tc.err))
		})
	}
}

// ---------------------------------------------------------------------------
// Is() – sentinel matching
// ---------------------------------------------------------------------------

func TestUserNotFoundError_Is(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		target error
		want   bool
	}{
		{
			name:   "matches ErrUserNotFound sentinel – no message",
			err:    &smarterror.UserNotFoundError{},
			target: smarterror.ErrUserNotFound,
			want:   true,
		},
		{
			name:   "matches ErrUserNotFound sentinel – with message",
			err:    smarterror.NewUserNotFoundError("user 1 not found"),
			target: smarterror.ErrUserNotFound,
			want:   true,
		},
		{
			name:   "matches ErrUserNotFound sentinel – with cause",
			err:    smarterror.NewUserNotFoundErrorWithCause("msg", errors.New("root")),
			target: smarterror.ErrUserNotFound,
			want:   true,
		},
		{
			name:   "does not match an arbitrary error",
			err:    &smarterror.UserNotFoundError{},
			target: errors.New("something else"),
			want:   false,
		},
		{
			name:   "errors.Is traverses chain to sentinel",
			err:    fmt.Errorf("wrapper: %w", smarterror.NewUserNotFoundError("inner")),
			target: smarterror.ErrUserNotFound,
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

// ---------------------------------------------------------------------------
// errors.As – type assertion through the chain
// ---------------------------------------------------------------------------

func TestUserNotFoundError_As(t *testing.T) {
	original := smarterror.NewUserNotFoundError("user 7 not found")
	wrapped := fmt.Errorf("service layer: %w", original)

	var target *smarterror.UserNotFoundError
	found := errors.As(wrapped, &target)

	assert.True(t, found)
	require.NotNil(t, target)
	assert.Equal(t, "user 7 not found", target.Message)
}

// ---------------------------------------------------------------------------
// Cause chain – nested errors.Is
// ---------------------------------------------------------------------------

func TestUserNotFoundError_CauseChain(t *testing.T) {
	dbErr := errors.New("sql: no rows")
	userErr := smarterror.NewUserNotFoundErrorWithCause("user not found", dbErr)
	wrapped := fmt.Errorf("handler: %w", userErr)

	assert.True(t, errors.Is(wrapped, smarterror.ErrUserNotFound))
	assert.True(t, errors.Is(wrapped, dbErr))
}

// ---------------------------------------------------------------------------
// HTTP handler integration – maps UserNotFoundError → 404
// ---------------------------------------------------------------------------

// userService is a minimal interface to allow dependency injection / mocking.
type userService interface {
	GetUser(id string) error
}

// mockUserService lets individual tests control the returned error.
type mockUserService struct {
	err error
}

func (m *mockUserService) GetUser(_ string) error { return m.err }

// userHandler is a thin HTTP handler that uses the service and converts the
// domain error to an HTTP status code, mirroring what Spring's
// @ExceptionHandler would do for UserNotFoundException → 404.
func userHandler(svc userService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if err := svc.GetUser(id); err != nil {
			if errors.Is(err, smarterror.ErrUserNotFound) {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}
}

func TestUserHandler_HTTP(t *testing.T) {
	tests := []struct {
		name       string
		serviceErr error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "user found – 200",
			serviceErr: nil,
			wantStatus: http.StatusOK,
			wantBody:   "ok",
		},
		{
			name:       "UserNotFoundError no message – 404 with sentinel text",
			serviceErr: &smarterror.UserNotFoundError{},
			wantStatus: http.StatusNotFound,
			wantBody:   "user not found\n",
		},
		{
			name:       "UserNotFoundError with message – 404 with custom message",
			serviceErr: smarterror.NewUserNotFoundError("user 42 not found"),
			wantStatus: http.StatusNotFound,
			wantBody:   "user 42 not found\n",
		},
		{
			name:       "UserNotFoundError with cause – 404",
			serviceErr: smarterror.NewUserNotFoundErrorWithCause("user gone", errors.New("db err")),
			wantStatus: http.StatusNotFound,
			wantBody:   "user gone\n",
		},
		{
			name:       "wrapped UserNotFoundError – 404 (errors.Is traverses chain)",
			serviceErr: fmt.Errorf("repo: %w", smarterror.NewUserNotFoundError("wrapped user not found")),
			wantStatus: http.StatusNotFound,
			wantBody:   "repo: wrapped user not found\n",
		},
		{
			name:       "unrelated error – 500",
			serviceErr: errors.New("unexpected db failure"),
			wantStatus: http.StatusInternalServerError,
			wantBody:   "internal server error\n",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			svc := &mockUserService{err: tc.serviceErr}
			handler := userHandler(svc)

			req := httptest.NewRequest(http.MethodGet, "/user?id=42", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tc.wantStatus, rec.Code)
			assert.Equal(t, tc.wantBody, rec.Body.String())
		})
	}
}

// ---------------------------------------------------------------------------
// Implements error interface (compile-time check)
// ---------------------------------------------------------------------------

func TestUserNotFoundError_ImplementsError(t *testing.T) {
	var _ error = (*smarterror.UserNotFoundError)(nil)
	// If this compiles, the interface is satisfied.
	assert.True(t, true)
}

// ---------------------------------------------------------------------------
// Message field preserved exactly (Java invariant: getMessage() exact match)
// ---------------------------------------------------------------------------

func TestUserNotFoundError_MessagePreservedExactly(t *testing.T) {
	tests := []struct {
		name    string
		message string
	}{
		{"normal string", "user 1 not found"},
		{"unicode", "utilisateur introuvable: こんにちは"},
		{"empty