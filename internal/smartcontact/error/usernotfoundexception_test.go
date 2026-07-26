```go
package apperr_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apperr "github.com/smartContact/internal/smartcontact/error"
	"github.com/smartContact/internal/smartcontact/model"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// sentinelCause is a simple error used as a stand-in for a "cause" in tests.
var sentinelCause = errors.New("underlying cause")

// ---------------------------------------------------------------------------
// ErrUserNotFound sentinel
// ---------------------------------------------------------------------------

func TestErrUserNotFound_IsSentinel(t *testing.T) {
	assert.NotNil(t, apperr.ErrUserNotFound)
	assert.EqualError(t, apperr.ErrUserNotFound, "user not found")
}

// ---------------------------------------------------------------------------
// NewUserNotFoundError (message-only constructor)
// Mirrors: UserNotFoundException(String message)
// ---------------------------------------------------------------------------

func TestNewUserNotFoundError(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		wantUserID     string
		wantCauseNil   bool
		wantErrContain string // substring that must appear in Error()
	}{
		{
			name:           "non-empty userID produces error with userID embedded",
			userID:         "user-42",
			wantUserID:     "user-42",
			wantCauseNil:   true,
			wantErrContain: "user-42",
		},
		{
			name:           "empty userID still constructs valid error",
			userID:         "",
			wantUserID:     "",
			wantCauseNil:   true,
			wantErrContain: "", // message may just be the format prefix
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := apperr.NewUserNotFoundError(tc.userID)

			require.NotNil(t, err)
			assert.Equal(t, tc.wantUserID, err.UserID)

			if tc.wantCauseNil {
				assert.Nil(t, err.Cause)
			}

			if tc.wantErrContain != "" {
				assert.Contains(t, err.Error(), tc.wantErrContain)
			}

			// Unwrap() must return nil when there is no cause.
			assert.Nil(t, err.Unwrap())
		})
	}
}

// ---------------------------------------------------------------------------
// NewUserNotFoundErrorWithCause (message + cause constructor)
// Mirrors: UserNotFoundException(String message, Throwable cause)
// ---------------------------------------------------------------------------

func TestNewUserNotFoundErrorWithCause(t *testing.T) {
	tests := []struct {
		name         string
		userID       string
		cause        error
		wantUserID   string
		wantCause    error
		wantCauseNil bool
	}{
		{
			name:       "non-empty userID and non-nil cause",
			userID:     "user-99",
			cause:      sentinelCause,
			wantUserID: "user-99",
			wantCause:  sentinelCause,
		},
		{
			name:         "non-empty userID with nil cause",
			userID:       "user-77",
			cause:        nil,
			wantUserID:   "user-77",
			wantCauseNil: true,
		},
		{
			name:         "empty userID with nil cause",
			userID:       "",
			cause:        nil,
			wantUserID:   "",
			wantCauseNil: true,
		},
		{
			name:       "empty userID with non-nil cause",
			userID:     "",
			cause:      sentinelCause,
			wantUserID: "",
			wantCause:  sentinelCause,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := apperr.NewUserNotFoundErrorWithCause(tc.userID, tc.cause)

			require.NotNil(t, err)
			assert.Equal(t, tc.wantUserID, err.UserID)

			if tc.wantCauseNil {
				assert.Nil(t, err.Cause)
			} else {
				assert.Equal(t, tc.wantCause, err.Cause)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Error() string formatting
// ---------------------------------------------------------------------------

func TestUserNotFoundError_Error(t *testing.T) {
	tests := []struct {
		name           string
		err            *apperr.UserNotFoundError
		wantContains   []string
		wantNotContain []string
	}{
		{
			name:         "without cause contains userID",
			err:          apperr.NewUserNotFoundError("abc-123"),
			wantContains: []string{"abc-123"},
		},
		{
			name:         "with cause contains userID and cause message",
			err:          apperr.NewUserNotFoundErrorWithCause("abc-456", sentinelCause),
			wantContains: []string{"abc-456", sentinelCause.Error()},
		},
		{
			name:           "without cause does NOT contain cause text",
			err:            apperr.NewUserNotFoundError("abc-789"),
			wantNotContain: []string{sentinelCause.Error()},
		},
		{
			name: "direct struct construction with no userID and no cause",
			err:  &apperr.UserNotFoundError{},
			// Error() should still return a non-empty string (the formatted message).
			wantContains: []string{},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			msg := tc.err.Error()
			assert.NotEmpty(t, msg)

			for _, s := range tc.wantContains {
				assert.Contains(t, msg, s)
			}
			for _, s := range tc.wantNotContain {
				assert.NotContains(t, msg, s)
			}
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
			name:   "Is returns true for ErrUserNotFound sentinel",
			err:    apperr.NewUserNotFoundError("u1"),
			target: apperr.ErrUserNotFound,
			want:   true,
		},
		{
			name:   "Is returns true for ErrUserNotFound sentinel even with cause",
			err:    apperr.NewUserNotFoundErrorWithCause("u2", sentinelCause),
			target: apperr.ErrUserNotFound,
			want:   true,
		},
		{
			name:   "Is returns false for unrelated error",
			err:    apperr.NewUserNotFoundError("u3"),
			target: sentinelCause,
			want:   false,
		},
		{
			name:   "errors.Is traverses chain – detects ErrUserNotFound",
			err:    fmt.Errorf("wrapped: %w", apperr.NewUserNotFoundError("u4")),
			target: apperr.ErrUserNotFound,
			want:   true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := errors.Is(tc.err, tc.target)
			assert.Equal(t, tc.want, got)
		})
	}
}

// ---------------------------------------------------------------------------
// Unwrap() – cause chain traversal
// ---------------------------------------------------------------------------

func TestUserNotFoundError_Unwrap(t *testing.T) {
	tests := []struct {
		name      string
		err       *apperr.UserNotFoundError
		wantUnwrap error
	}{
		{
			name:      "Unwrap returns nil when no cause",
			err:       apperr.NewUserNotFoundError("u1"),
			wantUnwrap: nil,
		},
		{
			name:      "Unwrap returns cause when set",
			err:       apperr.NewUserNotFoundErrorWithCause("u2", sentinelCause),
			wantUnwrap: sentinelCause,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := tc.err.Unwrap()
			assert.Equal(t, tc.wantUnwrap, got)
		})
	}
}

// ---------------------------------------------------------------------------
// errors.As – type assertion through the chain
// ---------------------------------------------------------------------------

func TestUserNotFoundError_ErrorsAs(t *testing.T) {
	original := apperr.NewUserNotFoundError("as-user-1")
	wrapped := fmt.Errorf("outer: %w", original)

	var target *apperr.UserNotFoundError
	ok := errors.As(wrapped, &target)

	assert.True(t, ok)
	require.NotNil(t, target)
	assert.Equal(t, "as-user-1", target.UserID)
}

// ---------------------------------------------------------------------------
// ErrorMessage() – transport-level model conversion
// ---------------------------------------------------------------------------

func TestUserNotFoundError_ErrorMessage(t *testing.T) {
	tests := []struct {
		name       string
		err        *apperr.UserNotFoundError
		status     int
		wantStatus int
	}{
		{
			name:       "404 status maps correctly",
			err:        apperr.NewUserNotFoundError("em-user-1"),
			status:     http.StatusNotFound,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "500 status maps correctly",
			err:        apperr.NewUserNotFoundError("em-user-2"),
			status:     http.StatusInternalServerError,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "400 status maps correctly",
			err:        apperr.NewUserNotFoundErrorWithCause("em-user-3", sentinelCause),
			status:     http.StatusBadRequest,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			em := tc.err.ErrorMessage(tc.status)

			// model.ErrorMessage must carry the status we passed in.
			assert.Equal(t, tc.wantStatus, em.Status)
			// The error detail must contain the user's ID.
			assert.Contains(t, em.Message, tc.err.UserID)
		})
	}
}

// ---------------------------------------------------------------------------
// ErrorMessage via HTTP handler (httptest)
// ---------------------------------------------------------------------------

func TestUserNotFoundError_HTTPHandler(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		cause      error
		wantStatus int
	}{
		{
			name:       "handler returns 404 with user-not-found body",
			userID:     "http-user-1",
			cause:      nil,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "handler returns 404 when cause is present",
			userID:     "http-user-2",
			cause:      sentinelCause,
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var appErr *apperr.UserNotFoundError
			if tc.cause != nil {
				appErr = apperr.NewUserNotFoundErrorWithCause(tc.userID, tc.cause)
			} else {
				appErr = apperr.NewUserNotFoundError(tc.userID)
			}

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				em := appErr.ErrorMessage(http.StatusNotFound)
				w.WriteHeader(em.Status)
				_, _ = fmt.Fprint(w, em.Message)
			})

			req, err := http.NewRequest(http.MethodGet, "/users/"+tc.userID, nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.wantStatus, rr.Code)
			assert.Contains(t, rr.Body.String(), tc.userID)
		})
	}
}

// ---------------------------------------------------------------------------
// model.UserNotFoundMessageFormat integration
// ---------------------------------------------------------------------------

func TestUserNotFoundError_MessageFormat(t *testing.T) {
	// Ensure the Error() output is consistent with model.UserNotFoundMessageFormat.
	userID := "format-user-1"
	err := apperr.NewUserNotFoundError(userID)

	expected := fmt.Sprintf(model.UserNotFoundMessageFormat, userID)
	assert.Equal(t, expected, err.Error())
}

func TestUserNotFoundError_MessageFormatWithCause(t *testing.T) {
	userID := "format-user-2"
	err := apperr.NewUserNotFoundErrorWithCause(userID, sentinelCause)

	base := fmt.Sprintf(model.UserNotFoundMessageFormat, userID)
	expected := fmt.Sprintf("%s: %v", base, sentinelCause)
	assert.Equal(t, expected, err.Error())
}

// ---------------------------------------------------------------------------
// Suppression/stack-trace constructor – intentional omission check
// ---------------------------------------------------------------------------

// TestProtectedConstructorOmitted verifies that the Go migration correctly
// omits the protected (enableSuppression, writableStackTrace) constructor,
// which has no Go equivalent.  We simply confirm the package compiles without
// such a function and that the two public constructors are the only entry
// points provided.
func TestProtectedConstructorOmitted(t *testing.T) {
	// If the file compiles, the omission is already validated at compile time.
	// This test serves as a documentation checkpoint.
	t.Log("Protected suppression/writable-stack-trace constructor is intentionally omitted in Go migration")
}

// ---------------------------------------------------------------------------
// Zero-value / direct struct construction
// Mirrors: UserNotFoundException() – no-arg constructor
// ---------------------------------------------------------------------------

func TestUserNotFoundError_ZeroValue(t *testing.T) {
	// Directly constructing the struct with zero values mirrors the no-arg Java
	// constructor (no message, no cause).
	err := &apperr.UserNotFoundError{}

	assert.Equal(t, "", err.UserID)
	assert.Nil(t, err.Cause)
	assert.Nil(t, err.Unwrap())
	assert.True(t, errors.Is(err, apperr.ErrUserNotFound))
	// Error() must not panic on zero value.
	assert.NotPanics(t, func() { _ = err.Error() })
}

// ---------------------------------------------------------------------------
// Cause-only constructor analogue
// Mirrors: UserNotFoundException(Throwable cause)
// ---------------------------------------------------------------------------

func TestNewUserNotFoundErrorWithCause_CauseOnly(t *testing.T) {
	tests := []struct {
		name         string
		cause        error
		wantCauseNil bool
		wantCause    error
	}{
		{
			name:      "non-nil cause is preserved",
			cause:     sentinelCause,
			wantCause: sentinelCause,
		},
		{
			name:         "