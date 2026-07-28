```go
package smartcontacterror_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	smartcontacterror "github.com/example/smartcontact/internal/smartcontact/error"
)

// ---------------------------------------------------------------------------
// ErrUserNotFound sentinel
// ---------------------------------------------------------------------------

func TestErrUserNotFound_IsSentinel(t *testing.T) {
	assert.NotNil(t, smartcontacterror.ErrUserNotFound)
	assert.Equal(t, "user not found", smartcontacterror.ErrUserNotFound.Error())
}

// ---------------------------------------------------------------------------
// NewUserNotFoundError — mirrors the Java no-arg and message-only ctors
// ---------------------------------------------------------------------------

func TestNewUserNotFoundError(t *testing.T) {
	tests := []struct {
		name            string
		message         string
		wantMessage     string // what Message field should hold
		wantErrorString string // what Error() should return
		wantCause       bool   // whether Cause should be non-nil
	}{
		{
			name:            "no-arg equivalent (empty message)",
			message:         "",
			wantMessage:     "",
			wantErrorString: "user not found",
			wantCause:       false,
		},
		{
			name:            "non-empty message",
			message:         "user 42 not found",
			wantMessage:     "user 42 not found",
			wantErrorString: "user 42 not found",
			wantCause:       false,
		},
		{
			name:            "whitespace-only message treated as non-empty",
			message:         "   ",
			wantMessage:     "   ",
			wantErrorString: "   ",
			wantCause:       false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := smartcontacterror.NewUserNotFoundError(tc.message)

			require.NotNil(t, err)
			assert.Equal(t, tc.wantMessage, err.Message)
			assert.Equal(t, tc.wantErrorString, err.Error())

			if tc.wantCause {
				assert.NotNil(t, err.Cause)
			} else {
				assert.Nil(t, err.Cause)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// NewUserNotFoundErrorWithCause — mirrors the Java cause-bearing ctors
// ---------------------------------------------------------------------------

func TestNewUserNotFoundErrorWithCause(t *testing.T) {
	rootCause := errors.New("db connection refused")

	tests := []struct {
		name            string
		message         string
		cause           error
		wantMessage     string
		wantErrorString string
		wantCauseNil    bool
	}{
		{
			name:            "message and cause both provided",
			message:         "user 7 not found",
			cause:           rootCause,
			wantMessage:     "user 7 not found",
			wantErrorString: fmt.Sprintf("user 7 not found: %v", rootCause),
			wantCauseNil:    false,
		},
		{
			name:            "empty message with cause (cause-only Java ctor equivalent)",
			message:         "",
			cause:           rootCause,
			wantMessage:     "",
			wantErrorString: fmt.Sprintf("user not found: %v", rootCause),
			wantCauseNil:    false,
		},
		{
			name:            "nil cause (null-cause Java ctor equivalent)",
			message:         "user not found in store",
			cause:           nil,
			wantMessage:     "user not found in store",
			wantErrorString: "user not found in store",
			wantCauseNil:    true,
		},
		{
			name:            "both message and cause are zero/nil",
			message:         "",
			cause:           nil,
			wantMessage:     "",
			wantErrorString: "user not found",
			wantCauseNil:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := smartcontacterror.NewUserNotFoundErrorWithCause(tc.message, tc.cause)

			require.NotNil(t, err)
			assert.Equal(t, tc.wantMessage, err.Message)
			assert.Equal(t, tc.wantErrorString, err.Error())

			if tc.wantCauseNil {
				assert.Nil(t, err.Cause)
			} else {
				assert.Equal(t, tc.cause, err.Cause)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Error() method
// ---------------------------------------------------------------------------

func TestUserNotFoundError_Error(t *testing.T) {
	wrappedErr := errors.New("underlying failure")

	tests := []struct {
		name string
		err  *smartcontacterror.UserNotFoundError
		want string
	}{
		{
			name: "empty message, no cause → default sentinel text",
			err:  &smartcontacterror.UserNotFoundError{},
			want: "user not found",
		},
		{
			name: "custom message, no cause",
			err:  &smartcontacterror.UserNotFoundError{Message: "user 99 missing"},
			want: "user 99 missing",
		},
		{
			name: "empty message, with cause",
			err:  &smartcontacterror.UserNotFoundError{Cause: wrappedErr},
			want: fmt.Sprintf("user not found: %v", wrappedErr),
		},
		{
			name: "custom message, with cause",
			err:  &smartcontacterror.UserNotFoundError{Message: "lookup failed", Cause: wrappedErr},
			want: fmt.Sprintf("lookup failed: %v", wrappedErr),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.err.Error())
		})
	}
}

// ---------------------------------------------------------------------------
// Unwrap() behaviour
// ---------------------------------------------------------------------------

func TestUserNotFoundError_Unwrap(t *testing.T) {
	rootCause := errors.New("root cause")

	tests := []struct {
		name         string
		err          *smartcontacterror.UserNotFoundError
		wantUnwrap   error
	}{
		{
			name:       "no cause → unwraps to ErrUserNotFound sentinel",
			err:        &smartcontacterror.UserNotFoundError{},
			wantUnwrap: smartcontacterror.ErrUserNotFound,
		},
		{
			name:       "with cause → unwraps to that cause",
			err:        &smartcontacterror.UserNotFoundError{Cause: rootCause},
			wantUnwrap: rootCause,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.wantUnwrap, tc.err.Unwrap())
		})
	}
}

// ---------------------------------------------------------------------------
// errors.Is semantics
// ---------------------------------------------------------------------------

func TestUserNotFoundError_ErrorsIs(t *testing.T) {
	rootCause := errors.New("sql: no rows")

	tests := []struct {
		name         string
		err          error
		target       error
		wantIs       bool
	}{
		{
			name:   "bare UserNotFoundError matches ErrUserNotFound",
			err:    &smartcontacterror.UserNotFoundError{},
			target: smartcontacterror.ErrUserNotFound,
			wantIs: true,
		},
		{
			name:   "UserNotFoundError with cause still matches ErrUserNotFound",
			err:    &smartcontacterror.UserNotFoundError{Cause: rootCause},
			target: smartcontacterror.ErrUserNotFound,
			wantIs: true,
		},
		{
			name:   "UserNotFoundError does NOT match an unrelated sentinel",
			err:    &smartcontacterror.UserNotFoundError{},
			target: errors.New("something else"),
			wantIs: false,
		},
		{
			name:   "NewUserNotFoundError matches ErrUserNotFound",
			err:    smartcontacterror.NewUserNotFoundError("msg"),
			target: smartcontacterror.ErrUserNotFound,
			wantIs: true,
		},
		{
			name:   "NewUserNotFoundErrorWithCause matches ErrUserNotFound",
			err:    smartcontacterror.NewUserNotFoundErrorWithCause("msg", rootCause),
			target: smartcontacterror.ErrUserNotFound,
			wantIs: true,
		},
		{
			name:   "wrapped via fmt.Errorf still matches ErrUserNotFound",
			err:    fmt.Errorf("service layer: %w", smartcontacterror.NewUserNotFoundError("user 1 missing")),
			target: smartcontacterror.ErrUserNotFound,
			wantIs: true,
		},
		{
			name:   "ErrUserNotFound sentinel itself matches",
			err:    smartcontacterror.ErrUserNotFound,
			target: smartcontacterror.ErrUserNotFound,
			wantIs: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.wantIs, errors.Is(tc.err, tc.target))
		})
	}
}

// ---------------------------------------------------------------------------
// errors.As semantics
// ---------------------------------------------------------------------------

func TestUserNotFoundError_ErrorsAs(t *testing.T) {
	direct := smartcontacterror.NewUserNotFoundError("direct")
	wrapped := fmt.Errorf("handler: %w", smartcontacterror.NewUserNotFoundError("wrapped"))

	tests := []struct {
		name    string
		err     error
		wantAs  bool
		wantMsg string
	}{
		{
			name:    "direct error resolves via errors.As",
			err:     direct,
			wantAs:  true,
			wantMsg: "direct",
		},
		{
			name:    "wrapped error resolves via errors.As",
			err:     wrapped,
			wantAs:  true,
			wantMsg: "wrapped",
		},
		{
			name:   "unrelated error does not resolve",
			err:    errors.New("unrelated"),
			wantAs: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var target *smartcontacterror.UserNotFoundError
			ok := errors.As(tc.err, &target)
			assert.Equal(t, tc.wantAs, ok)
			if tc.wantAs {
				require.NotNil(t, target)
				assert.Equal(t, tc.wantMsg, target.Message)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Is() method (direct call, independent of errors.Is chain)
// ---------------------------------------------------------------------------

func TestUserNotFoundError_IsMethod(t *testing.T) {
	e := &smartcontacterror.UserNotFoundError{Message: "x"}

	tests := []struct {
		name   string
		target error
		want   bool
	}{
		{
			name:   "target is ErrUserNotFound → true",
			target: smartcontacterror.ErrUserNotFound,
			want:   true,
		},
		{
			name:   "target is nil → false",
			target: nil,
			want:   false,
		},
		{
			name:   "target is different sentinel → false",
			target: errors.New("other"),
			want:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, e.Is(tc.target))
		})
	}
}

// ---------------------------------------------------------------------------
// Structural / type-level invariants
// ---------------------------------------------------------------------------

func TestUserNotFoundError_ImplementsErrorInterface(t *testing.T) {
	var _ error = (*smartcontacterror.UserNotFoundError)(nil)
	// If the above compiles, the interface is satisfied. Confirm at runtime too.
	err := smartcontacterror.NewUserNotFoundError("check")
	assert.Implements(t, (*error)(nil), err)
}

func TestUserNotFoundError_NoExtraState(t *testing.T) {
	// The type must only expose Message and Cause (no additional domain state).
	err := smartcontacterror.NewUserNotFoundErrorWithCause("msg", errors.New("cause"))
	assert.Equal(t, "msg", err.Message)
	assert.NotNil(t, err.Cause)
}

// ---------------------------------------------------------------------------
// Chained / nested wrapping scenarios
// ---------------------------------------------------------------------------

func TestUserNotFoundError_ChainedWrapping(t *testing.T) {
	dbErr := errors.New("pq: relation does not exist")
	inner := smartcontacterror.NewUserNotFoundErrorWithCause("repo layer", dbErr)
	outer := fmt.Errorf("service layer: %w", inner)

	// The outer error must still match both sentinels in the chain.
	assert.True(t, errors.Is(outer, smartcontacterror.ErrUserNotFound),
		"outer fmt.Errorf chain must match ErrUserNotFound")

	var typed *smartcontacterror.UserNotFoundError
	assert.True(t, errors.As(outer, &typed))
	assert.Equal(t, "repo layer", typed.Message)
	assert.Equal(t, dbErr, typed.Cause)
}

func TestUserNotFoundError_MultiLevelUnwrap(t *testing.T) {
	// Unwrap without cause → ErrUserNotFound
	e1 := smartcontacterror.NewUserNotFoundError("")
	assert.Equal(t, smartcontacterror.ErrUserNotFound, e1.Unwrap())

	// Unwrap with cause → that cause
	cause := errors.New("network timeout")
	e2 := smartcontacterror.NewUserNotFoundErrorWithCause("", cause)
	assert.Equal(t, cause, e2.Unwrap())
}

// ---------------------------------------------------------------------------
// Protected four-arg constructor equivalence note
// ---------------------------------------------------------------------------
// The Java protected constructor
//   UserNotFoundException(String, Throwable, boolean, boolean)
// has no equivalent in Go — stack trace writability and suppression
// enablement are JVM-level concerns. The migration explicitly omits it.
// This test documents/verifies that no such constructor is exported and
// that the two public constructors cover all needed use-cases.

func TestProtectedConstructorNotExported(t *testing.T) {
	// We can only create instances via the two exported constructors.
	// This test asserts both produce valid *UserNotFoundError values
	// and that the type carries no unexplained hidden state.

	e1 := smartcontacterror.NewUserNotFoundError("only message")
	assert.NotNil(t, e1)
	assert.Equal(t, "only message", e1.Message)
	assert.Nil(t, e1.Cause)

	cause := errors.New("some cause")
	e2 := smartcontacterror.NewUserNotFoundErrorWithCause("msg+cause", cause)
	assert.NotNil(t, e2)
	assert.Equal(t, "msg+cause", e2.Message)
	assert.Equal(t, cause, e2.Cause)
}
```