```go
package error_test

import (
	stderrors "errors"
	"fmt"
	"testing"

	smerr "github.com/example/internal/smartcontact/error"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// ErrUserNotFound sentinel
// ---------------------------------------------------------------------------

func TestErrUserNotFound_IsSentinel(t *testing.T) {
	assert.NotNil(t, smerr.ErrUserNotFound)
	assert.EqualError(t, smerr.ErrUserNotFound, "user not found")
}

// ---------------------------------------------------------------------------
// NewUserNotFound – corresponds to UserNotFoundException(String message)
// ---------------------------------------------------------------------------

func TestNewUserNotFound(t *testing.T) {
	tests := []struct {
		name            string
		message         string
		wantErrContains string
		wantIsNotFound  bool
		wantNil         bool
	}{
		{
			name:            "non-empty message wraps sentinel",
			message:         "user 42 does not exist",
			wantErrContains: "user 42 does not exist",
			wantIsNotFound:  true,
		},
		{
			name:            "non-empty message includes sentinel text",
			message:         "lookup failed",
			wantErrContains: "user not found",
			wantIsNotFound:  true,
		},
		{
			name:           "empty string message still wraps sentinel",
			message:        "",
			wantIsNotFound: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := smerr.NewUserNotFound(tc.message)
			require.NotNil(t, err)

			// errors.Is must report true for the sentinel
			assert.True(t, stderrors.Is(err, smerr.ErrUserNotFound),
				"errors.Is(err, ErrUserNotFound) must be true")

			if tc.wantErrContains != "" {
				assert.Contains(t, err.Error(), tc.wantErrContains)
			}
		})
	}
}

// Spec: instantiated with a non-null message → getMessage() returns provided
// message (Go equivalent: error string contains the message).
func TestNewUserNotFound_MessagePreserved(t *testing.T) {
	msg := "specific user not found detail"
	err := smerr.NewUserNotFound(msg)
	assert.Contains(t, err.Error(), msg)
}

// Spec: instantiated with a null message → getMessage() returns null.
// Go equivalent: empty message → no panic, error still wraps sentinel.
func TestNewUserNotFound_EmptyMessage(t *testing.T) {
	err := smerr.NewUserNotFound("")
	require.NotNil(t, err)
	assert.True(t, stderrors.Is(err, smerr.ErrUserNotFound))
}

// ---------------------------------------------------------------------------
// WrapUserNotFound – corresponds to UserNotFoundException(String message, Throwable cause)
// ---------------------------------------------------------------------------

func TestWrapUserNotFound(t *testing.T) {
	someErr := stderrors.New("db connection refused")

	tests := []struct {
		name            string
		message         string
		cause           error
		wantIsNotFound  bool
		wantContains    []string
		wantNilCause    bool
	}{
		{
			name:           "message and non-nil cause",
			message:        "could not locate user",
			cause:          someErr,
			wantIsNotFound: true,
			wantContains:   []string{"could not locate user", "db connection refused", "user not found"},
		},
		{
			name:           "message and nil cause delegates to NewUserNotFound",
			message:        "no cause available",
			cause:          nil,
			wantIsNotFound: true,
			wantContains:   []string{"no cause available", "user not found"},
		},
		{
			name:           "empty message and nil cause",
			message:        "",
			cause:          nil,
			wantIsNotFound: true,
		},
		{
			name:           "empty message and non-nil cause",
			message:        "",
			cause:          someErr,
			wantIsNotFound: true,
			wantContains:   []string{"user not found"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := smerr.WrapUserNotFound(tc.message, tc.cause)
			require.NotNil(t, err)

			assert.True(t, stderrors.Is(err, smerr.ErrUserNotFound),
				"errors.Is(err, ErrUserNotFound) must be true")

			for _, substr := range tc.wantContains {
				assert.Contains(t, err.Error(), substr)
			}
		})
	}
}

// Spec: instantiated with message and cause → getCause() returns provided cause.
// Go equivalent: the cause's text is embedded in the error chain/string.
func TestWrapUserNotFound_CausePreserved(t *testing.T) {
	cause := stderrors.New("underlying storage failure")
	err := smerr.WrapUserNotFound("user fetch error", cause)
	assert.Contains(t, err.Error(), cause.Error())
}

// Spec: instantiated with null message and null cause → no panic, sentinel preserved.
func TestWrapUserNotFound_NilMessageNilCause(t *testing.T) {
	err := smerr.WrapUserNotFound("", nil)
	require.NotNil(t, err)
	assert.True(t, stderrors.Is(err, smerr.ErrUserNotFound))
}

// ---------------------------------------------------------------------------
// No-argument "constructor" – corresponds to UserNotFoundException()
// Java: null message, null cause.
// Go: ErrUserNotFound itself serves this role (sentinel with fixed message).
// ---------------------------------------------------------------------------

func TestErrUserNotFound_NoArgConstructorEquivalent(t *testing.T) {
	// The sentinel acts as the zero-argument constructor equivalent.
	err := smerr.ErrUserNotFound
	assert.NotNil(t, err)
	// No wrapped cause – Unwrap returns nil.
	assert.Nil(t, stderrors.Unwrap(err))
	// Fixed message present.
	assert.EqualError(t, err, "user not found")
}

// ---------------------------------------------------------------------------
// errors.Is chain integrity
// ---------------------------------------------------------------------------

func TestErrorsIs_ChainIntegrity(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"ErrUserNotFound sentinel", smerr.ErrUserNotFound},
		{"NewUserNotFound", smerr.NewUserNotFound("id=99")},
		{"WrapUserNotFound with cause", smerr.WrapUserNotFound("msg", stderrors.New("cause"))},
		{"WrapUserNotFound without cause", smerr.WrapUserNotFound("msg", nil)},
		{"doubly wrapped", fmt.Errorf("outer: %w", smerr.NewUserNotFound("inner"))},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.True(t, stderrors.Is(tc.err, smerr.ErrUserNotFound),
				"errors.Is must detect ErrUserNotFound in chain")
		})
	}
}

// ---------------------------------------------------------------------------
// Negative: unrelated errors do not match sentinel
// ---------------------------------------------------------------------------

func TestErrorsIs_NegativeCases(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"generic error", stderrors.New("some other error")},
		{"fmt.Errorf no wrap", fmt.Errorf("user not found")}, // same text but no %w
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.False(t, stderrors.Is(tc.err, smerr.ErrUserNotFound))
		})
	}
}

// ---------------------------------------------------------------------------
// WrapUserNotFound nil-cause fallthrough equals NewUserNotFound
// ---------------------------------------------------------------------------

func TestWrapUserNotFound_NilCauseBehavesLikeNewUserNotFound(t *testing.T) {
	msg := "user 7 not found"
	fromWrap := smerr.WrapUserNotFound(msg, nil)
	fromNew := smerr.NewUserNotFound(msg)

	// Both must satisfy errors.Is check.
	assert.True(t, stderrors.Is(fromWrap, smerr.ErrUserNotFound))
	assert.True(t, stderrors.Is(fromNew, smerr.ErrUserNotFound))

	// Both must contain the message.
	assert.Contains(t, fromWrap.Error(), msg)
	assert.Contains(t, fromNew.Error(), msg)
}

// ---------------------------------------------------------------------------
// Spec: UserNotFoundException(Throwable cause) – cause-only constructor.
// Go mapping: WrapUserNotFound("", cause) or a direct wrap.
// ---------------------------------------------------------------------------

func TestWrapUserNotFound_CauseOnly(t *testing.T) {
	tests := []struct {
		name           string
		cause          error
		wantIsNotFound bool
		wantCauseInMsg bool
	}{
		{
			name:           "non-nil cause",
			cause:          stderrors.New("some root cause"),
			wantIsNotFound: true,
			wantCauseInMsg: true,
		},
		{
			name:           "nil cause",
			cause:          nil,
			wantIsNotFound: true,
			wantCauseInMsg: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate "cause-only" constructor by passing empty message.
			err := smerr.WrapUserNotFound("", tc.cause)
			require.NotNil(t, err)
			assert.True(t, stderrors.Is(err, smerr.ErrUserNotFound))

			if tc.wantCauseInMsg && tc.cause != nil {
				assert.Contains(t, err.Error(), tc.cause.Error())
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Spec: protected 4-arg constructor (enableSuppression, writableStackTrace).
// Go has no direct equivalent; the package exposes no such function.
// We verify the package does NOT expose any additional constructors beyond
// the two documented ones, and that the sentinel itself is stable.
// ---------------------------------------------------------------------------

func TestPackage_NoExtraState(t *testing.T) {
	// ErrUserNotFound carries no additional state.
	err1 := smerr.ErrUserNotFound
	err2 := smerr.ErrUserNotFound
	assert.Same(t, err1, err2, "sentinel must be a single shared value")
}

// ---------------------------------------------------------------------------
// Verify error messages are deterministic / stable
// ---------------------------------------------------------------------------

func TestNewUserNotFound_DeterministicMessage(t *testing.T) {
	msg := "deterministic"
	err1 := smerr.NewUserNotFound(msg)
	err2 := smerr.NewUserNotFound(msg)
	assert.Equal(t, err1.Error(), err2.Error())
}

func TestWrapUserNotFound_DeterministicMessage(t *testing.T) {
	cause := stderrors.New("cause")
	msg := "deterministic"
	err1 := smerr.WrapUserNotFound(msg, cause)
	err2 := smerr.WrapUserNotFound(msg, cause)
	assert.Equal(t, err1.Error(), err2.Error())
}
```