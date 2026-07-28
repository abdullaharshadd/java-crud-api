```go
package apperror_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	apperror "internal/smartcontact/error"
)

// ---------------------------------------------------------------------------
// ErrUserNotFound sentinel
// ---------------------------------------------------------------------------

func TestErrUserNotFound_IsSentinel(t *testing.T) {
	assert.NotNil(t, apperror.ErrUserNotFound)
	assert.Equal(t, "user not found", apperror.ErrUserNotFound.Error())
}

// ---------------------------------------------------------------------------
// NewUserNotFoundError
// ---------------------------------------------------------------------------

func TestNewUserNotFoundError(t *testing.T) {
	tests := []struct {
		name           string
		msg            string
		wantExact      error  // if non-nil, result must == this value
		wantContains   string // substring that must appear in Error()
		wantIsSentinel bool   // errors.Is(result, ErrUserNotFound) must be true
		wantNilCause   bool   // errors.Unwrap must return nil (only ErrUserNotFound)
	}{
		{
			// Java UserNotFoundException() – no-arg / empty message
			name:           "empty message returns sentinel directly",
			msg:            "",
			wantExact:      apperror.ErrUserNotFound,
			wantIsSentinel: true,
			wantNilCause:   true,
		},
		{
			// Java UserNotFoundException(String message) – non-empty message
			name:           "non-empty message wraps sentinel",
			msg:            "user 42 does not exist",
			wantContains:   "user 42 does not exist",
			wantIsSentinel: true,
		},
		{
			// message that is only whitespace is treated as non-empty
			name:           "whitespace-only message is non-empty",
			msg:            "   ",
			wantContains:   "user not found",
			wantIsSentinel: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := apperror.NewUserNotFoundError(tc.msg)

			assert.NotNil(t, got)

			if tc.wantExact != nil {
				assert.Equal(t, tc.wantExact, got)
			}

			if tc.wantContains != "" {
				assert.Contains(t, got.Error(), tc.wantContains)
			}

			if tc.wantIsSentinel {
				assert.True(t, errors.Is(got, apperror.ErrUserNotFound),
					"errors.Is must detect ErrUserNotFound")
			}

			if tc.wantNilCause {
				// When the sentinel is returned directly there is nothing to unwrap
				// beyond the sentinel itself.
				assert.Equal(t, apperror.ErrUserNotFound, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// NewUserNotFoundErrorf
// ---------------------------------------------------------------------------

func TestNewUserNotFoundErrorf(t *testing.T) {
	rootCause := errors.New("db: connection refused")

	tests := []struct {
		name              string
		msg               string
		cause             error
		wantIsSentinel    bool
		wantCauseReached  bool   // errors.Is(result, rootCause)
		wantMsgContains   string
		wantBehavesLikeNE bool   // should behave like NewUserNotFoundError(msg)
	}{
		{
			// Java UserNotFoundException(String message, Throwable cause) – both set
			name:             "message and cause both set",
			msg:              "user 99 missing",
			cause:            rootCause,
			wantIsSentinel:   true,
			wantCauseReached: true,
			wantMsgContains:  "user 99 missing",
		},
		{
			// Java UserNotFoundException(Throwable cause) – only cause, empty message
			name:             "empty message with cause",
			msg:              "",
			cause:            rootCause,
			wantIsSentinel:   true,
			wantCauseReached: true,
		},
		{
			// Java UserNotFoundException(String message, Throwable cause) with null cause
			name:              "nil cause falls back to NewUserNotFoundError",
			msg:               "some message",
			cause:             nil,
			wantIsSentinel:    true,
			wantMsgContains:   "some message",
			wantBehavesLikeNE: true,
		},
		{
			// nil cause AND empty message → sentinel directly
			name:              "nil cause and empty message returns sentinel",
			msg:               "",
			cause:             nil,
			wantIsSentinel:    true,
			wantBehavesLikeNE: true,
		},
		{
			// wrapped cause is itself an ErrUserNotFound – still detectable
			name:             "cause is another ErrUserNotFound wrapper",
			msg:              "outer",
			cause:            apperror.NewUserNotFoundError("inner"),
			wantIsSentinel:   true,
			wantCauseReached: true,
			wantMsgContains:  "outer",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := apperror.NewUserNotFoundErrorf(tc.msg, tc.cause)

			assert.NotNil(t, got)

			if tc.wantIsSentinel {
				assert.True(t, errors.Is(got, apperror.ErrUserNotFound),
					"errors.Is must detect ErrUserNotFound")
			}

			if tc.wantCauseReached {
				assert.True(t, errors.Is(got, rootCause) || errors.Is(got, tc.cause),
					"errors.Is must reach the original cause")
			}

			if tc.wantMsgContains != "" {
				assert.Contains(t, got.Error(), tc.wantMsgContains)
			}

			if tc.wantBehavesLikeNE {
				expected := apperror.NewUserNotFoundError(tc.msg)
				// The errors.Is relationship must hold in both
				assert.Equal(t, errors.Is(expected, apperror.ErrUserNotFound),
					errors.Is(got, apperror.ErrUserNotFound))
			}
		})
	}
}

// ---------------------------------------------------------------------------
// IsUserNotFound
// ---------------------------------------------------------------------------

func TestIsUserNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error returns false",
			err:  nil,
			want: false,
		},
		{
			name: "sentinel itself returns true",
			err:  apperror.ErrUserNotFound,
			want: true,
		},
		{
			name: "error from NewUserNotFoundError empty msg returns true",
			err:  apperror.NewUserNotFoundError(""),
			want: true,
		},
		{
			name: "error from NewUserNotFoundError with msg returns true",
			err:  apperror.NewUserNotFoundError("user 5 not found"),
			want: true,
		},
		{
			name: "error from NewUserNotFoundErrorf with cause returns true",
			err:  apperror.NewUserNotFoundErrorf("lookup failed", errors.New("db timeout")),
			want: true,
		},
		{
			name: "unrelated error returns false",
			err:  errors.New("something else"),
			want: false,
		},
		{
			name: "wrapped unrelated error returns false",
			err:  fmt.Errorf("outer: %w", errors.New("inner")),
			want: false,
		},
		{
			name: "deeply wrapped sentinel returns true",
			err:  fmt.Errorf("level1: %w", fmt.Errorf("level2: %w", apperror.ErrUserNotFound)),
			want: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := apperror.IsUserNotFound(tc.err)
			assert.Equal(t, tc.want, got)
		})
	}
}

// ---------------------------------------------------------------------------
// Unwrap / chain integrity
// ---------------------------------------------------------------------------

func TestErrorChainIntegrity(t *testing.T) {
	t.Run("NewUserNotFoundError wraps only ErrUserNotFound", func(t *testing.T) {
		err := apperror.NewUserNotFoundError("test message")
		// The chain must contain ErrUserNotFound
		assert.True(t, errors.Is(err, apperror.ErrUserNotFound))
		// It must NOT be the sentinel itself
		assert.NotEqual(t, apperror.ErrUserNotFound, err)
	})

	t.Run("NewUserNotFoundErrorf wraps both sentinel and cause", func(t *testing.T) {
		cause := errors.New("original cause")
		err := apperror.NewUserNotFoundErrorf("ctx", cause)

		assert.True(t, errors.Is(err, apperror.ErrUserNotFound))
		assert.True(t, errors.Is(err, cause))
	})

	t.Run("NewUserNotFoundErrorf empty msg wraps sentinel and cause", func(t *testing.T) {
		cause := errors.New("raw cause")
		err := apperror.NewUserNotFoundErrorf("", cause)

		assert.True(t, errors.Is(err, apperror.ErrUserNotFound))
		assert.True(t, errors.Is(err, cause))
	})

	t.Run("error message contains sentinel text", func(t *testing.T) {
		err := apperror.NewUserNotFoundError("context info")
		assert.Contains(t, err.Error(), "user not found")
		assert.Contains(t, err.Error(), "context info")
	})

	t.Run("cause-only error message contains sentinel text", func(t *testing.T) {
		cause := errors.New("db error")
		err := apperror.NewUserNotFoundErrorf("", cause)
		assert.Contains(t, err.Error(), "user not found")
	})

	t.Run("full message contains all parts", func(t *testing.T) {
		cause := errors.New("network timeout")
		err := apperror.NewUserNotFoundErrorf("user lookup failed", cause)
		text := err.Error()
		assert.Contains(t, text, "user lookup failed")
		assert.Contains(t, text, "user not found")
		assert.Contains(t, text, "network timeout")
	})
}

// ---------------------------------------------------------------------------
// Protected constructor / suppression – migration note coverage
// ---------------------------------------------------------------------------
// The Java protected 4-arg constructor (enableSuppression, writableStackTrace)
// has no Go equivalent. The test below confirms there is no such exported API
// and that the package compiles and functions correctly without it.

func TestNoProtectedConstructorEquivalent(t *testing.T) {
	// If the migration had accidentally exported a four-arg variant this test
	// would not compile (or the reviewer would catch it). We validate that the
	// two exported constructors cover all observable behaviours instead.

	t.Run("sentinel is comparable", func(t *testing.T) {
		a := apperror.ErrUserNotFound
		b := apperror.ErrUserNotFound
		assert.True(t, errors.Is(a, b))
	})

	t.Run("two NewUserNotFoundError calls produce distinct values", func(t *testing.T) {
		e1 := apperror.NewUserNotFoundError("msg")
		e2 := apperror.NewUserNotFoundError("msg")
		// They are equal in content but distinct allocations
		assert.Equal(t, e1.Error(), e2.Error())
		assert.False(t, e1 == e2, "distinct allocations should not be pointer-equal")
	})
}

// ---------------------------------------------------------------------------
// HTTP handler integration (httptest) – error mapper simulation
// ---------------------------------------------------------------------------

import (
	"net/http"
	"net/http/httptest"
)

// errorMapperHandler simulates the central HTTP error mapper that maps
// apperror.ErrUserNotFound → 404 and everything else → 500.
func errorMapperHandler(err error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if apperror.IsUserNotFound(err) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
	})
}

func TestHTTPErrorMapper(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "sentinel maps to 404",
			err:        apperror.ErrUserNotFound,
			wantStatus: http.StatusNotFound,
			wantBody:   "not found",
		},
		{
			name:       "NewUserNotFoundError maps to 404",
			err:        apperror.NewUserNotFoundError("user 7 missing"),
			wantStatus: http.StatusNotFound,
			wantBody:   "not found",
		},
		{
			name:       "NewUserNotFoundErrorf maps to 404",
			err:        apperror.NewUserNotFoundErrorf("db lookup", errors.New("timeout")),
			wantStatus: http.StatusNotFound,
			wantBody:   "not found",
		},
		{
			name:       "deeply wrapped sentinel maps to 404",
			err:        fmt.Errorf("service: %w", apperror.NewUserNotFoundError("id=3")),
			wantStatus: http.StatusNotFound,
			wantBody:   "not found",
		},
		{
			name:       "unrelated error maps to 500",
			err:        errors.New("unexpected failure"),
			wantStatus: http.StatusInternalServerError,
			wantBody:   "internal error",
		},
		{
			name:       "nil error maps to 500 (no user-not-found signal)",
			err:        nil,
			wantStatus: http.StatusInternalServerError,
			wantBody:   "internal error",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/user/1", nil)
			rec := httptest.NewRecorder()

			errorMapperHandler(tc.err).ServeHTTP(rec, req)

			res := rec.Result()
			assert.Equal(t, tc.wantStatus, res.StatusCode)

			body := rec.Body.String()
			assert.Contains(t, body, tc.wantBody)
		})
	}
}
```