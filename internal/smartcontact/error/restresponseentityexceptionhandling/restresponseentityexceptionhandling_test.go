```go
package error_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	smartcontacterror "github.com/smartContact/internal/smartcontact/error"
	restresponseentityexceptionhandling "github.com/smartContact/internal/smartcontact/error/restresponseentityexceptionhandling"
	"github.com/smartContact/internal/smartcontact/model"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func decodeErrorMessage(t *testing.T, body []byte) model.ErrorMessage {
	t.Helper()
	var msg model.ErrorMessage
	require.NoError(t, json.Unmarshal(body, &msg))
	return msg
}

// ---------------------------------------------------------------------------
// StatusAndMessage
// ---------------------------------------------------------------------------

func TestErrorResponder_StatusAndMessage(t *testing.T) {
	responder := restresponseentityexceptionhandling.NewErrorResponder()

	someOtherErr := errors.New("unexpected database error")

	tests := []struct {
		name           string
		err            error
		wantStatus     int
		wantMsgStatus  int
		wantMsgContain string
	}{
		{
			name:          "nil error returns 200 OK",
			err:           nil,
			wantStatus:    http.StatusOK,
			wantMsgStatus: http.StatusOK,
		},
		{
			name:           "UserNotFoundError (direct) returns 404",
			err:            smartcontacterror.NewUserNotFoundError("user 42 not found"),
			wantStatus:     http.StatusNotFound,
			wantMsgStatus:  http.StatusNotFound,
			wantMsgContain: "user 42 not found",
		},
		{
			name:           "UserNotFoundError with empty message returns 404",
			err:            smartcontacterror.NewUserNotFoundError(""),
			wantStatus:     http.StatusNotFound,
			wantMsgStatus:  http.StatusNotFound,
			wantMsgContain: "",
		},
		{
			name:           "ErrUserNotFound sentinel returns 404",
			err:            smartcontacterror.ErrUserNotFound,
			wantStatus:     http.StatusNotFound,
			wantMsgStatus:  http.StatusNotFound,
			wantMsgContain: smartcontacterror.ErrUserNotFound.Error(),
		},
		{
			name:           "wrapped UserNotFoundError returns 404",
			err:            fmt.Errorf("outer: %w", smartcontacterror.NewUserNotFoundError("inner not found")),
			wantStatus:     http.StatusNotFound,
			wantMsgStatus:  http.StatusNotFound,
			wantMsgContain: "inner not found",
		},
		{
			name:           "unrecognized error returns 500",
			err:            someOtherErr,
			wantStatus:     http.StatusInternalServerError,
			wantMsgStatus:  http.StatusInternalServerError,
			wantMsgContain: "unexpected database error",
		},
		{
			name:           "wrapped unrecognized error returns 500",
			err:            fmt.Errorf("wrap: %w", someOtherErr),
			wantStatus:     http.StatusInternalServerError,
			wantMsgStatus:  http.StatusInternalServerError,
			wantMsgContain: "wrap: unexpected database error",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			status, msg := responder.StatusAndMessage(tc.err)

			assert.Equal(t, tc.wantStatus, status, "HTTP status mismatch")
			assert.Equal(t, tc.wantMsgStatus, msg.Status, "ErrorMessage.Status mismatch")

			if tc.wantMsgContain != "" {
				assert.Contains(t, msg.Message, tc.wantMsgContain, "ErrorMessage.Message mismatch")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Write
// ---------------------------------------------------------------------------

func TestErrorResponder_Write(t *testing.T) {
	responder := restresponseentityexceptionhandling.NewErrorResponder()

	tests := []struct {
		name               string
		err                error
		wantWritten        bool
		wantStatus         int
		wantContentType    string
		wantMsgStatus      int
		wantMsgMsgContains string
	}{
		{
			name:        "nil error writes nothing and returns false",
			err:         nil,
			wantWritten: false,
		},
		{
			name:               "UserNotFoundError writes 404 JSON body",
			err:                smartcontacterror.NewUserNotFoundError("user 7 not found"),
			wantWritten:        true,
			wantStatus:         http.StatusNotFound,
			wantContentType:    "application/json; charset=utf-8",
			wantMsgStatus:      http.StatusNotFound,
			wantMsgMsgContains: "user 7 not found",
		},
		{
			name:               "UserNotFoundError with empty message writes 404",
			err:                smartcontacterror.NewUserNotFoundError(""),
			wantWritten:        true,
			wantStatus:         http.StatusNotFound,
			wantContentType:    "application/json; charset=utf-8",
			wantMsgStatus:      http.StatusNotFound,
			wantMsgMsgContains: "",
		},
		{
			name:               "ErrUserNotFound sentinel writes 404",
			err:                smartcontacterror.ErrUserNotFound,
			wantWritten:        true,
			wantStatus:         http.StatusNotFound,
			wantContentType:    "application/json; charset=utf-8",
			wantMsgStatus:      http.StatusNotFound,
			wantMsgMsgContains: smartcontacterror.ErrUserNotFound.Error(),
		},
		{
			name:               "wrapped UserNotFoundError writes 404",
			err:                fmt.Errorf("wrap: %w", smartcontacterror.NewUserNotFoundError("wrapped")),
			wantWritten:        true,
			wantStatus:         http.StatusNotFound,
			wantContentType:    "application/json; charset=utf-8",
			wantMsgStatus:      http.StatusNotFound,
			wantMsgMsgContains: "wrapped",
		},
		{
			name:               "generic error writes 500 JSON body",
			err:                errors.New("some internal failure"),
			wantWritten:        true,
			wantStatus:         http.StatusInternalServerError,
			wantContentType:    "application/json; charset=utf-8",
			wantMsgStatus:      http.StatusInternalServerError,
			wantMsgMsgContains: "some internal failure",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()

			written := responder.Write(rr, tc.err)
			assert.Equal(t, tc.wantWritten, written)

			if !tc.wantWritten {
				// nothing should have been sent
				assert.Equal(t, http.StatusOK, rr.Code, "response code should be default 200 when nothing written")
				assert.Empty(t, rr.Body.Bytes())
				return
			}

			assert.Equal(t, tc.wantStatus, rr.Code)
			assert.Equal(t, tc.wantContentType, rr.Header().Get("Content-Type"))

			body := rr.Body.Bytes()
			require.NotEmpty(t, body, "body must not be empty on error")

			msg := decodeErrorMessage(t, body)
			assert.Equal(t, tc.wantMsgStatus, msg.Status)

			if tc.wantMsgMsgContains != "" {
				assert.Contains(t, msg.Message, tc.wantMsgMsgContains)
			}

			// Global invariant: status embedded in body == HTTP status code
			assert.Equal(t, rr.Code, msg.Status, "ErrorMessage.Status must equal HTTP response status")
		})
	}
}

// ---------------------------------------------------------------------------
// Middleware
// ---------------------------------------------------------------------------

func TestErrorResponder_Middleware(t *testing.T) {
	responder := restresponseentityexceptionhandling.NewErrorResponder()

	tests := []struct {
		name               string
		handlerErr         error
		wantStatus         int
		wantMsgStatus      int
		wantMsgMsgContains string
		wantBodyEmpty      bool
	}{
		{
			name:          "handler returns nil – middleware does not interfere",
			handlerErr:    nil,
			wantStatus:    http.StatusOK,
			wantBodyEmpty: true,
		},
		{
			name:               "handler returns UserNotFoundError – middleware writes 404",
			handlerErr:         smartcontacterror.NewUserNotFoundError("user 99 not found"),
			wantStatus:         http.StatusNotFound,
			wantMsgStatus:      http.StatusNotFound,
			wantMsgMsgContains: "user 99 not found",
		},
		{
			name:               "handler returns ErrUserNotFound sentinel – middleware writes 404",
			handlerErr:         smartcontacterror.ErrUserNotFound,
			wantStatus:         http.StatusNotFound,
			wantMsgStatus:      http.StatusNotFound,
			wantMsgMsgContains: smartcontacterror.ErrUserNotFound.Error(),
		},
		{
			name:               "handler returns wrapped UserNotFoundError – middleware writes 404",
			handlerErr:         fmt.Errorf("outer: %w", smartcontacterror.NewUserNotFoundError("inner")),
			wantStatus:         http.StatusNotFound,
			wantMsgStatus:      http.StatusNotFound,
			wantMsgMsgContains: "inner",
		},
		{
			name:               "handler returns generic error – middleware writes 500",
			handlerErr:         errors.New("db connection lost"),
			wantStatus:         http.StatusInternalServerError,
			wantMsgStatus:      http.StatusInternalServerError,
			wantMsgMsgContains: "db connection lost",
		},
		{
			name:               "handler returns wrapped generic error – middleware writes 500",
			handlerErr:         fmt.Errorf("context: %w", errors.New("timeout")),
			wantStatus:         http.StatusInternalServerError,
			wantMsgStatus:      http.StatusInternalServerError,
			wantMsgMsgContains: "context: timeout",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			handlerCalled := false

			innerHandler := restresponseentityexceptionhandling.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) error {
					handlerCalled = true
					return tc.handlerErr
				},
			)

			handler := responder.Middleware(innerHandler)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			assert.True(t, handlerCalled, "inner handler must always be called")

			if tc.wantBodyEmpty {
				assert.Equal(t, tc.wantStatus, rr.Code)
				// body might be empty or have been written by the inner handler
				// (in the nil-error case the middleware does not write anything)
				return
			}

			assert.Equal(t, tc.wantStatus, rr.Code)
			assert.Equal(t, "application/json; charset=utf-8", rr.Header().Get("Content-Type"))

			body := rr.Body.Bytes()
			require.NotEmpty(t, body)

			msg := decodeErrorMessage(t, body)
			assert.Equal(t, tc.wantMsgStatus, msg.Status)

			if tc.wantMsgMsgContains != "" {
				assert.Contains(t, msg.Message, tc.wantMsgMsgContains)
			}

			// Global invariant: status in body == HTTP response status
			assert.Equal(t, rr.Code, msg.Status)
		})
	}
}

// ---------------------------------------------------------------------------
// NewErrorResponder constructor
// ---------------------------------------------------------------------------

func TestNewErrorResponder(t *testing.T) {
	r := restresponseentityexceptionhandling.NewErrorResponder()
	assert.NotNil(t, r, "NewErrorResponder must return a non-nil value")
}

// ---------------------------------------------------------------------------
// Global invariant: only UserNotFoundError maps to 404, everything else to 500
// ---------------------------------------------------------------------------

func TestGlobalInvariant_OnlyUserNotFoundMapsTo404(t *testing.T) {
	responder := restresponseentityexceptionhandling.NewErrorResponder()

	userNotFoundErrors := []error{
		smartcontacterror.NewUserNotFoundError("message a"),
		smartcontacterror.NewUserNotFoundError(""),
		smartcontacterror.ErrUserNotFound,
		fmt.Errorf("wrapped: %w", smartcontacterror.NewUserNotFoundError("x")),
		fmt.Errorf("wrapped: %w", smartcontacterror.ErrUserNotFound),
	}

	for _, err := range userNotFoundErrors {
		status, msg := responder.StatusAndMessage(err)
		assert.Equal(t, http.StatusNotFound, status, "UserNotFoundError variant must yield 404: %v", err)
		assert.Equal(t, http.StatusNotFound, msg.Status, "ErrorMessage.Status must be 404 for UserNotFoundError: %v", err)
	}

	otherErrors := []error{
		errors.New("random error"),
		fmt.Errorf("wrapped random: %w", errors.New("inner")),
	}

	for _, err := range otherErrors {
		status, msg := responder.StatusAndMessage(err)
		assert.Equal(t, http.StatusInternalServerError, status, "non-UserNotFoundError must yield 500: %v", err)
		assert.Equal(t, http.StatusInternalServerError, msg.Status, "ErrorMessage.Status must be 500 for generic error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Global invariant: no state mutation across calls
// ---------------------------------------------------------------------------

func TestGlobalInvariant_NoStateMutation(t *testing.T) {
	responder := restresponseentityexceptionhandling.NewErrorResponder()

	// Call multiple times with different errors; the responder must behave
	// identically on every call (no internal state leaks between calls).
	for i := 0; i < 5; i++ {
		status, msg := responder.StatusAndMessage(smartcontacterror.NewUserNotFoundError("user"))
		assert.Equal(t, http.StatusNotFound, status)
		assert.Equal(t, http.StatusNotFound, msg.Status)
	}

	for i := 0; i < 5; i++ {
		status, msg := responder.StatusAndMessage(errors.New("boom"))
		assert.Equal(t, http.StatusInternalServerError, status)
		assert.Equal(t, http.StatusInternalServerError, msg.Status)
	}
}
```

> **Note on missing import:** The test file above uses `fmt.Errorf` but does not have an explicit `import "fmt"`. Add `"fmt"` to the import block. The complete, compilable import block is:
>
> ```go
> import (
>     "encoding/json"
>     "errors"
>     "fmt"
>     "net/http"
>     "net/http/httptest"
>     "testing"