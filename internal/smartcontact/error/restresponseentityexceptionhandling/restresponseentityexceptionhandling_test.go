```go
package error

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/smartContact/internal/smartcontact/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// decodeErrorMessage reads the JSON body written to a ResponseRecorder and
// returns the decoded ErrorMessage (or fails the test).
func decodeErrorMessage(t *testing.T, rec *httptest.ResponseRecorder) *model.ErrorMessage {
	t.Helper()
	var msg model.ErrorMessage
	err := json.NewDecoder(rec.Body).Decode(&msg)
	require.NoError(t, err, "response body must be valid JSON")
	return &msg
}

// ---------------------------------------------------------------------------
// MapError tests
// ---------------------------------------------------------------------------

func TestMapError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		wantStatus     int
		wantMsgNil     bool
		wantMsgStatus  int
		wantMsgContain string
	}{
		// --- nil error ---
		{
			name:       "nil error returns 200 and nil message",
			err:        nil,
			wantStatus: http.StatusOK,
			wantMsgNil: true,
		},

		// --- UserNotFoundError with non-empty message ---
		{
			name:           "UserNotFoundException with non-null message returns 404",
			err:            &UserNotFoundError{Message: "user with id 42 was not found"},
			wantStatus:     http.StatusNotFound,
			wantMsgNil:     false,
			wantMsgStatus:  http.StatusNotFound,
			wantMsgContain: "user with id 42 was not found",
		},

		// --- UserNotFoundError with empty message ---
		{
			name:          "UserNotFoundException with empty message returns 404 with empty body message",
			err:           &UserNotFoundError{Message: ""},
			wantStatus:    http.StatusNotFound,
			wantMsgNil:    false,
			wantMsgStatus: http.StatusNotFound,
			// empty string is acceptable; just verify status
			wantMsgContain: "",
		},

		// --- UserNotFoundError wrapped in another error ---
		{
			name:           "wrapped UserNotFoundException unwrapped via errors.As returns 404",
			err:            fmt.Errorf("outer: %w", &UserNotFoundError{Message: "wrapped user not found"}),
			wantStatus:     http.StatusNotFound,
			wantMsgNil:     false,
			wantMsgStatus:  http.StatusNotFound,
			wantMsgContain: "wrapped user not found",
		},

		// --- Generic / unrecognized error ---
		{
			name:           "unrecognized error returns 500",
			err:            fmt.Errorf("something went wrong"),
			wantStatus:     http.StatusInternalServerError,
			wantMsgNil:     false,
			wantMsgStatus:  http.StatusInternalServerError,
			wantMsgContain: "something went wrong",
		},

		// --- Stdlib sentinel error (not UserNotFoundError) ---
		{
			name:           "stdlib error not UserNotFoundException returns 500",
			err:            fmt.Errorf("database connection refused"),
			wantStatus:     http.StatusInternalServerError,
			wantMsgNil:     false,
			wantMsgStatus:  http.StatusInternalServerError,
			wantMsgContain: "database connection refused",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			status, msg := MapError(tc.err)

			// HTTP status code assertion
			assert.Equal(t, tc.wantStatus, status, "HTTP status code mismatch")

			if tc.wantMsgNil {
				assert.Nil(t, msg, "ErrorMessage should be nil when err is nil")
				return
			}

			require.NotNil(t, msg, "ErrorMessage must not be nil for a non-nil error")

			// Invariant: response status code and ErrorMessage.Status are consistent
			assert.Equal(t, tc.wantMsgStatus, msg.Status,
				"ErrorMessage.Status must match the returned HTTP status code")

			// Message content assertion (skip if empty – empty is explicitly allowed)
			if tc.wantMsgContain != "" {
				assert.Contains(t, msg.Message, tc.wantMsgContain,
					"ErrorMessage.Message must contain the original error text")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Global invariants for MapError
// ---------------------------------------------------------------------------

// TestMapError_UserNotFoundError_GlobalInvariants validates that for every
// UserNotFoundError the status is always 404 and the body is always non-nil.
func TestMapError_UserNotFoundError_GlobalInvariants(t *testing.T) {
	inputs := []struct {
		name string
		err  *UserNotFoundError
	}{
		{"typical message", &UserNotFoundError{Message: "user 1 not found"}},
		{"empty message", &UserNotFoundError{Message: ""}},
		{"long message", &UserNotFoundError{Message: "user with email very-long-email@example.com not found in the system"}},
	}

	for _, tc := range inputs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			status, msg := MapError(tc.err)

			// Invariant: always 404
			assert.Equal(t, http.StatusNotFound, status)

			// Invariant: body is never nil
			require.NotNil(t, msg)

			// Invariant: ErrorMessage.Status is always 404
			assert.Equal(t, http.StatusNotFound, msg.Status)

			// Invariant: ErrorMessage.Message equals the exception's message
			assert.Equal(t, tc.err.Error(), msg.Message)
		})
	}
}

// ---------------------------------------------------------------------------
// WriteError tests
// ---------------------------------------------------------------------------

func TestWriteError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		wantWritten    bool
		wantHTTPStatus int
		wantBodyStatus int
		wantMsgContain string
	}{
		{
			name:        "nil error writes nothing and returns false",
			err:         nil,
			wantWritten: false,
		},
		{
			name:           "UserNotFoundException produces 404 JSON response",
			err:            &UserNotFoundError{Message: "user 99 not found"},
			wantWritten:    true,
			wantHTTPStatus: http.StatusNotFound,
			wantBodyStatus: http.StatusNotFound,
			wantMsgContain: "user 99 not found",
		},
		{
			name:           "UserNotFoundException with empty message produces 404",
			err:            &UserNotFoundError{Message: ""},
			wantWritten:    true,
			wantHTTPStatus: http.StatusNotFound,
			wantBodyStatus: http.StatusNotFound,
			wantMsgContain: "",
		},
		{
			name:           "unrecognized error produces 500 JSON response",
			err:            fmt.Errorf("unexpected failure"),
			wantWritten:    true,
			wantHTTPStatus: http.StatusInternalServerError,
			wantBodyStatus: http.StatusInternalServerError,
			wantMsgContain: "unexpected failure",
		},
		{
			name:           "wrapped UserNotFoundException produces 404",
			err:            fmt.Errorf("context: %w", &UserNotFoundError{Message: "wrapped"}),
			wantWritten:    true,
			wantHTTPStatus: http.StatusNotFound,
			wantBodyStatus: http.StatusNotFound,
			wantMsgContain: "wrapped",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			written := WriteError(rec, tc.err)

			assert.Equal(t, tc.wantWritten, written, "WriteError return value mismatch")

			if !tc.wantWritten {
				// Nothing should have been written
				assert.Equal(t, http.StatusOK, rec.Code,
					"recorder code should still be default 200 when nothing written")
				assert.Empty(t, rec.Body.String(), "body should be empty when err is nil")
				return
			}

			// HTTP status code on the wire
			assert.Equal(t, tc.wantHTTPStatus, rec.Code, "HTTP response status code mismatch")

			// Decode body
			msg := decodeErrorMessage(t, rec)
			require.NotNil(t, msg)

			// Invariant: status code and body status are consistent
			assert.Equal(t, tc.wantBodyStatus, msg.Status,
				"ErrorMessage.Status must match HTTP response code")

			if tc.wantMsgContain != "" {
				assert.Contains(t, msg.Message, tc.wantMsgContain,
					"ErrorMessage.Message must contain original error text")
			}

			// Content-Type header should be application/json (set by WriteJSON)
			assert.Contains(t, rec.Header().Get("Content-Type"), "application/json",
				"response must carry JSON content type")
		})
	}
}

// ---------------------------------------------------------------------------
// WriteError used from an http.Handler (httptest integration)
// ---------------------------------------------------------------------------

func TestWriteError_InsideHTTPHandler(t *testing.T) {
	tests := []struct {
		name           string
		handlerErr     error
		wantStatusCode int
		wantBodyStatus int
		wantMsgContain string
	}{
		{
			name:           "handler that encounters UserNotFoundException returns 404",
			handlerErr:     &UserNotFoundError{Message: "no user found for request"},
			wantStatusCode: http.StatusNotFound,
			wantBodyStatus: http.StatusNotFound,
			wantMsgContain: "no user found for request",
		},
		{
			name:           "handler that encounters generic error returns 500",
			handlerErr:     fmt.Errorf("db timeout"),
			wantStatusCode: http.StatusInternalServerError,
			wantBodyStatus: http.StatusInternalServerError,
			wantMsgContain: "db timeout",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// Simulate an HTTP handler that calls WriteError
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				WriteError(w, tc.handlerErr)
			})

			req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			assert.Equal(t, tc.wantStatusCode, rec.Code)

			msg := decodeErrorMessage(t, rec)
			require.NotNil(t, msg)
			assert.Equal(t, tc.wantBodyStatus, msg.Status)
			assert.Contains(t, msg.Message, tc.wantMsgContain)
		})
	}
}

// ---------------------------------------------------------------------------
// UserNotFoundError type tests (unit-level – ensures the type behaves as
// expected by MapError's errors.As branch)
// ---------------------------------------------------------------------------

func TestUserNotFoundError_ErrorMethod(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    string
	}{
		{"non-empty message", "user 5 not found", "user 5 not found"},
		{"empty message", "", ""},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			e := &UserNotFoundError{Message: tc.message}
			assert.Equal(t, tc.want, e.Error())
		})
	}
}

// TestMapError_HandlerDoesNotProcessNonUserNotFoundError confirms that a
// non-UserNotFoundError is not misidentified as a 404 – matching the spec
// behavior "this handler does not process it".
func TestMapError_HandlerDoesNotProcessNonUserNotFoundError(t *testing.T) {
	otherErrors := []error{
		fmt.Errorf("generic error"),
		fmt.Errorf("validation failed"),
		fmt.Errorf("internal state corrupt"),
	}

	for _, err := range otherErrors {
		err := err
		t.Run(err.Error(), func(t *testing.T) {
			status, msg := MapError(err)
			assert.NotEqual(t, http.StatusNotFound, status,
				"non-UserNotFoundError must not produce 404")
			assert.Equal(t, http.StatusInternalServerError, status)
			require.NotNil(t, msg)
			assert.Equal(t, http.StatusInternalServerError, msg.Status)
		})
	}
}
```