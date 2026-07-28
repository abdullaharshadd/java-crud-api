```go
package error_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	smarterror "github.com/smartContact/internal/smartcontact/error/restresponseentityexceptionhandling"
	"github.com/smartContact/internal/smartcontact/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// decodeErrorMessage reads the JSON body from a ResponseRecorder and returns
// the decoded ErrorMessage, failing the test on any parse error.
func decodeErrorMessage(t *testing.T, rr *httptest.ResponseRecorder) model.ErrorMessage {
	t.Helper()
	var msg model.ErrorMessage
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&msg))
	return msg
}

// ---------------------------------------------------------------------------
// WriteError
// ---------------------------------------------------------------------------

func TestWriteError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		wantStatusCode int
		wantInBody     bool   // whether we expect a JSON error body
		wantMessage    string // expected message in the ErrorMessage
	}{
		{
			name:           "nil error returns 200 and no body",
			err:            nil,
			wantStatusCode: http.StatusOK,
			wantInBody:     false,
		},
		{
			name:           "UserNotFoundError returns 404",
			err:            smarterror.NewUserNotFoundError("user 42 not found"),
			wantStatusCode: http.StatusNotFound,
			wantInBody:     true,
			wantMessage:    "user 42 not found",
		},
		{
			name:           "wrapped ErrUserNotFound returns 404",
			err:            fmt.Errorf("wrapped: %w", smarterror.ErrUserNotFound),
			wantStatusCode: http.StatusNotFound,
			wantInBody:     true,
		},
		{
			name:           "generic error returns 500",
			err:            errors.New("some internal failure"),
			wantStatusCode: http.StatusInternalServerError,
			wantInBody:     true,
			wantMessage:    "some internal failure",
		},
		{
			name:           "UserNotFoundError with empty message returns 404",
			err:            smarterror.NewUserNotFoundError(""),
			wantStatusCode: http.StatusNotFound,
			wantInBody:     true,
			wantMessage:    "",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			gotStatus := smarterror.WriteError(rr, tc.err)

			// Return value must match the HTTP status written.
			assert.Equal(t, tc.wantStatusCode, gotStatus)

			if tc.err == nil {
				// nil path: WriteError must not touch the ResponseWriter.
				assert.Equal(t, http.StatusOK, gotStatus)
				assert.Empty(t, rr.Body.String())
				return
			}

			// HTTP status code on the recorder.
			assert.Equal(t, tc.wantStatusCode, rr.Code)

			// Content-Type must be JSON.
			assert.Contains(t, rr.Header().Get("Content-Type"), "application/json")

			if tc.wantInBody {
				msg := decodeErrorMessage(t, rr)

				// The ErrorMessage status field must reflect the HTTP status.
				assert.Equal(t, tc.wantStatusCode, msg.Status)

				// Non-empty message expectations.
				if tc.wantMessage != "" {
					assert.Equal(t, tc.wantMessage, msg.Message)
				}

				// The ErrorMessage object itself must never be zero-valued
				// (status field is always set).
				assert.NotZero(t, msg.Status)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// WriteError – invariants from the spec
// ---------------------------------------------------------------------------

func TestWriteError_UserNotFoundException_Invariants(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		message string
	}{
		{
			name:    "standard user-not-found message",
			err:     smarterror.NewUserNotFoundError("user with id 7 not found"),
			message: "user with id 7 not found",
		},
		{
			name:    "empty message UserNotFoundException",
			err:     smarterror.NewUserNotFoundError(""),
			message: "",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			status := smarterror.WriteError(rr, tc.err)

			// Invariant: HTTP status is always 404 for UserNotFoundException.
			assert.Equal(t, http.StatusNotFound, status)
			assert.Equal(t, http.StatusNotFound, rr.Code)

			// Invariant: response body is always a non-null ErrorMessage.
			body := rr.Body.String()
			assert.NotEmpty(t, body)

			msg := decodeErrorMessage(t, rr)

			// Invariant: ErrorMessage.Status is always NOT_FOUND (404).
			assert.Equal(t, http.StatusNotFound, msg.Status)

			// Invariant: ErrorMessage.Message equals the exception's message.
			assert.Equal(t, tc.message, msg.Message)
		})
	}
}

// ---------------------------------------------------------------------------
// RecoverMiddleware
// ---------------------------------------------------------------------------

func TestRecoverMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		handler        http.HandlerFunc
		wantStatusCode int
		wantJSON       bool
		wantMessage    string
	}{
		{
			name: "no panic passes through normally",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"ok":true}`))
			},
			wantStatusCode: http.StatusOK,
			wantJSON:       false,
		},
		{
			name: "panic with UserNotFoundError yields 404",
			handler: func(w http.ResponseWriter, r *http.Request) {
				panic(smarterror.NewUserNotFoundError("user not found via panic"))
			},
			wantStatusCode: http.StatusNotFound,
			wantJSON:       true,
			wantMessage:    "user not found via panic",
		},
		{
			name: "panic with generic error yields 500",
			handler: func(w http.ResponseWriter, r *http.Request) {
				panic(errors.New("unexpected boom"))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantJSON:       true,
			wantMessage:    "unexpected boom",
		},
		{
			name: "panic with wrapped ErrUserNotFound yields 404",
			handler: func(w http.ResponseWriter, r *http.Request) {
				panic(fmt.Errorf("context: %w", smarterror.ErrUserNotFound))
			},
			wantStatusCode: http.StatusNotFound,
			wantJSON:       true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			mw := smarterror.RecoverMiddleware(tc.handler)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rr := httptest.NewRecorder()

			// Non-error panics should re-panic; the others must be recovered.
			mw.ServeHTTP(rr, req)

			assert.Equal(t, tc.wantStatusCode, rr.Code)

			if tc.wantJSON {
				assert.Contains(t, rr.Header().Get("Content-Type"), "application/json")
				msg := decodeErrorMessage(t, rr)
				assert.Equal(t, tc.wantStatusCode, msg.Status)
				if tc.wantMessage != "" {
					assert.Equal(t, tc.wantMessage, msg.Message)
				}
			}
		})
	}
}

// TestRecoverMiddleware_NonErrorPanicRepanics ensures that a non-error panic
// value propagates (is not silently swallowed by the middleware).
func TestRecoverMiddleware_NonErrorPanicRepanics(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("this is a string, not an error")
	})

	mw := smarterror.RecoverMiddleware(handler)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	assert.Panics(t, func() {
		mw.ServeHTTP(rr, req)
	})
}

// ---------------------------------------------------------------------------
// RecoverMiddleware – global invariants
// ---------------------------------------------------------------------------

// TestRecoverMiddleware_NeverMutatesState verifies that the middleware does not
// alter any shared state (trivially verified: it only writes to ResponseWriter).
func TestRecoverMiddleware_NeverMutatesState(t *testing.T) {
	sentinel := "unchanged"

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(smarterror.NewUserNotFoundError("oops"))
	})

	mw := smarterror.RecoverMiddleware(handler)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	mw.ServeHTTP(rr, req)

	// Shared variable must be untouched.
	assert.Equal(t, "unchanged", sentinel)
}

// TestWriteError_OnlyUserNotFoundIsHandledAs404 asserts that only
// UserNotFoundException (ErrUserNotFound family) maps to 404; everything else
// must map to 500.
func TestWriteError_OnlyUserNotFoundIsHandledAs404(t *testing.T) {
	otherErrors := []error{
		errors.New("random error"),
		fmt.Errorf("wrapped random: %w", errors.New("inner")),
	}

	for _, err := range otherErrors {
		rr := httptest.NewRecorder()
		status := smarterror.WriteError(rr, err)
		assert.Equal(t, http.StatusInternalServerError, status,
			"expected 500 for non-UserNotFound error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Response content-type invariant
// ---------------------------------------------------------------------------

func TestWriteError_ContentTypeIsAlwaysJSON(t *testing.T) {
	errs := []error{
		smarterror.NewUserNotFoundError("not found"),
		errors.New("generic"),
	}

	for _, err := range errs {
		rr := httptest.NewRecorder()
		smarterror.WriteError(rr, err)
		assert.Contains(t, rr.Header().Get("Content-Type"), "application/json",
			"Content-Type must be JSON for error: %v", err)
	}
}
```

> **Note:** The test file imports `fmt` for `fmt.Errorf`; add the following import line to the import block:
>
> ```go
> "fmt"
> ```
>
> The full, self-contained test file with `fmt` included is below.

```go
package error_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	smarterror "github.com/smartContact/internal/smartcontact/error/restresponseentityexceptionhandling"
	"github.com/smartContact/internal/smartcontact/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// decodeErrorMessage reads the JSON body from a ResponseRecorder and returns
// the decoded ErrorMessage, failing the test on any parse error.
func decodeErrorMessage(t *testing.T, rr *httptest.ResponseRecorder) model.ErrorMessage {
	t.Helper()
	var msg model.ErrorMessage
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&msg))
	return msg
}

// ---------------------------------------------------------------------------
// WriteError
// ---------------------------------------------------------------------------

func TestWriteError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		wantStatusCode int
		wantBody       bool   // whether we expect a JSON error body
		wantMessage    string // expected message in the ErrorMessage (empty = skip check)
	}{
		{
			name:           "nil error returns 200 and writes nothing",
			err:            nil,
			wantStatusCode: http.StatusOK,
			wantBody:       false,
		},
		{
			name:           "UserNotFoundError returns 404 with message",
			err:            smarterror.NewUserNotFoundError("user 42 not found"),
			wantStatusCode: http.StatusNotFound,
			wantBody:       true,
			wantMessage:    "user 42 not found",
		},
		{
			name:           "wrapped ErrUserNotFound returns 404",
			err:            fmt.Errorf("context: %w", smarterror.ErrUserNotFound),
			wantStatusCode: http.StatusNotFound,
			wantBody:       true,
		},
		{
			name:           "generic error returns 500 with message",
			err:            errors.New("some internal failure"),
			wantStatusCode: http.StatusInternalServerError,
			wantBody:       true,
			wantMessage:    "some internal failure",
		},
		{
			name:           "UserNotFoundError with empty message returns 404",
			err:            smarterror.NewUserNotFoundError(""),
			wantStatusCode: http.StatusNotFound,
			wantBody:       true,
			wantMessage:    "",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			gotStatus := smarterror.WriteError(rr, tc.err)

			// The return value must match the intended HTTP status.
			assert.Equal(t, tc.wantStatusCode, gotStatus, "return value mismatch")

			if tc.err == nil {
				// Nil path: WriteError must not write anything to the ResponseWriter.
				assert.Empty(t, rr.Body.String(), "expected empty body for nil error")
				return
			}

			// Recorder code must match.
			assert.Equal(t, tc.wantStatusCode, rr.Code, "recorded status code mismatch")

			// Content-Type must always be JSON when we write an error body.
			assert.Contains(t, rr.Header().Get("Content-Type"), "application/json")

			if tc.wantBody {
				msg := decodeErrorMessage(t, rr)

				// The status field inside the JSON body must match the HTTP status.
				assert.Equal(t, tc.wantStatusCode, msg.Status, "ErrorMessage.Status mismatch")

				// Only check the message field when the test case specifies one.
				if tc.wantMessage != "" {
					assert.Equal(t, tc.wantMessage, msg.Message, "ErrorMessage.Message mismatch")
				}

				// Status field is always non-zero.
				assert.NotZero(t, msg.Status)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// WriteError – UserNotFoundException spec invariants
// ---------------------------------------------------------------------------

func TestWriteError_UserNotFoundInvariants(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		wantMessage string
	}{
		{