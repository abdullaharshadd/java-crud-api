```go
package error_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	domainerror "github.com/smartContact/internal/smartcontact/error"
	"github.com/smartContact/internal/smartcontact/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

func decodeErrorMessage(t *testing.T, body []byte) model.ErrorMessage {
	t.Helper()
	var em model.ErrorMessage
	require.NoError(t, json.Unmarshal(body, &em), "response body must be valid JSON ErrorMessage")
	return em
}

// -----------------------------------------------------------------------------
// WriteError tests
// -----------------------------------------------------------------------------

func TestWriteError(t *testing.T) {
	tests := []struct {
		name           string
		status         int
		message        string
		wantStatus     int
		wantMessage    string
		wantStatusCode int
	}{
		{
			name:        "404 with user not found message",
			status:      http.StatusNotFound,
			message:     "User not found",
			wantStatus:  http.StatusNotFound,
			wantMessage: "User not found",
		},
		{
			name:        "500 with internal error message",
			status:      http.StatusInternalServerError,
			message:     "Internal server error",
			wantStatus:  http.StatusInternalServerError,
			wantMessage: "Internal server error",
		},
		{
			name:        "400 with bad request message",
			status:      http.StatusBadRequest,
			message:     "Bad request",
			wantStatus:  http.StatusBadRequest,
			wantMessage: "Bad request",
		},
		{
			name:        "404 with empty message",
			status:      http.StatusNotFound,
			message:     "",
			wantStatus:  http.StatusNotFound,
			wantMessage: "",
		},
		{
			name:        "404 with null-like empty message",
			status:      http.StatusNotFound,
			message:     "",
			wantStatus:  http.StatusNotFound,
			wantMessage: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			err := domainerror.WriteError(w, tc.status, tc.message)

			assert.NoError(t, err)
			assert.Equal(t, tc.wantStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			em := decodeErrorMessage(t, w.Body.Bytes())
			assert.Equal(t, tc.wantStatus, em.Status)
			assert.Equal(t, tc.wantMessage, em.Message)
		})
	}
}

// -----------------------------------------------------------------------------
// WriteDomainError tests – UserNotFoundException mapping
// -----------------------------------------------------------------------------

func TestWriteDomainError_UserNotFoundException(t *testing.T) {
	tests := []struct {
		name             string
		err              error
		wantHandled      bool
		wantHTTPStatus   int
		wantErrorMessage string
		wantWriteErr     bool
	}{
		{
			name:             "UserNotFoundException with non-null message returns 404",
			err:              fmt.Errorf("%w: %s", domainerror.ErrUserNotFound, "User with id 42 not found"),
			wantHandled:      true,
			wantHTTPStatus:   http.StatusNotFound,
			wantErrorMessage: "user not found: User with id 42 not found",
		},
		{
			name:             "ErrUserNotFound sentinel itself returns 404",
			err:              domainerror.ErrUserNotFound,
			wantHandled:      true,
			wantHTTPStatus:   http.StatusNotFound,
			wantErrorMessage: domainerror.ErrUserNotFound.Error(),
		},
		{
			name:             "UserNotFoundException with empty message returns 404 with empty message body",
			err:              fmt.Errorf("%w", domainerror.ErrUserNotFound),
			wantHandled:      true,
			wantHTTPStatus:   http.StatusNotFound,
			wantErrorMessage: domainerror.ErrUserNotFound.Error(),
		},
		{
			name:        "nil error is not handled",
			err:         nil,
			wantHandled: false,
		},
		{
			name:        "unrelated error is not handled",
			err:         fmt.Errorf("some unexpected error"),
			wantHandled: false,
		},
		{
			name:        "other domain error is not handled",
			err:         fmt.Errorf("connection timeout"),
			wantHandled: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			handled, writeErr := domainerror.WriteDomainError(w, tc.err)

			assert.Equal(t, tc.wantHandled, handled)

			if !tc.wantHandled {
				// When the error is not handled, no response should have been written.
				assert.Equal(t, http.StatusOK, w.Code, "recorder default code should be 200 when nothing was written")
				assert.Empty(t, w.Body.Bytes(), "body should be empty when not handled")
				assert.NoError(t, writeErr)
				return
			}

			// Invariant: always 404 for UserNotFoundException.
			assert.Equal(t, http.StatusNotFound, w.Code,
				"HTTP status code must always be 404 for UserNotFoundException")

			// Invariant: Content-Type must be application/json.
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			// Invariant: body must be a non-null, well-formed ErrorMessage.
			require.NotEmpty(t, w.Body.Bytes(), "ErrorMessage body must not be empty")
			em := decodeErrorMessage(t, w.Body.Bytes())

			// Invariant: ErrorMessage.Status == 404.
			assert.Equal(t, http.StatusNotFound, em.Status,
				"ErrorMessage status field must always equal 404 (NOT_FOUND)")

			// Invariant: ErrorMessage.Message == exception message.
			if tc.wantErrorMessage != "" {
				assert.Equal(t, tc.wantErrorMessage, em.Message,
					"ErrorMessage message field must equal the exception message")
			}

			if !tc.wantWriteErr {
				assert.NoError(t, writeErr)
			}
		})
	}
}

// -----------------------------------------------------------------------------
// Integration-style test: simulate a controller invoking WriteDomainError
// -----------------------------------------------------------------------------

func TestWriteDomainError_IntegrationWithHTTPHandler(t *testing.T) {
	tests := []struct {
		name           string
		simulatedErr   error
		wantHTTPStatus int
		wantHandled    bool
	}{
		{
			name:           "controller throws UserNotFoundException, handler converts to 404",
			simulatedErr:   fmt.Errorf("user id=99 not found: %w", domainerror.ErrUserNotFound),
			wantHTTPStatus: http.StatusNotFound,
			wantHandled:    true,
		},
		{
			name:           "controller throws unrelated error, not handled by domain mapper",
			simulatedErr:   fmt.Errorf("database connection refused"),
			wantHTTPStatus: http.StatusInternalServerError, // caller's fallback
			wantHandled:    false,
		},
		{
			name:           "controller succeeds with no error",
			simulatedErr:   nil,
			wantHTTPStatus: http.StatusOK,
			wantHandled:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate a real HTTP handler that delegates to WriteDomainError.
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Mimic controller logic that may return a domain error.
				err := tc.simulatedErr

				handled, _ := domainerror.WriteDomainError(w, err)
				if handled {
					return
				}

				if err != nil {
					// Caller's generic fallback for unrecognised errors.
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return
				}

				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			assert.Equal(t, tc.wantHTTPStatus, w.Code)

			if tc.wantHandled {
				// Must be a valid JSON ErrorMessage with status 404.
				em := decodeErrorMessage(t, w.Body.Bytes())
				assert.Equal(t, http.StatusNotFound, em.Status)
				assert.NotNil(t, em.Message)
			}
		})
	}
}

// -----------------------------------------------------------------------------
// Invariant: handler only triggered by UserNotFoundException
// -----------------------------------------------------------------------------

func TestWriteDomainError_OnlyHandlesUserNotFoundException(t *testing.T) {
	otherErrors := []struct {
		name string
		err  error
	}{
		{"generic error", fmt.Errorf("something went wrong")},
		{"wrapped generic error", fmt.Errorf("outer: %w", fmt.Errorf("inner"))},
		{"nil", nil},
	}

	for _, tc := range otherErrors {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			handled, writeErr := domainerror.WriteDomainError(w, tc.err)

			assert.False(t, handled, "WriteDomainError must NOT handle non-UserNotFoundException errors")
			assert.NoError(t, writeErr)

			// No status code or body should have been set.
			assert.Equal(t, http.StatusOK, w.Code)
			assert.Empty(t, w.Body.Bytes())
		})
	}
}

// -----------------------------------------------------------------------------
// Invariant: response body is always a non-null ErrorMessage for 404 cases
// -----------------------------------------------------------------------------

func TestWriteDomainError_ResponseBodyIsAlwaysNonNull(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{
			name: "sentinel ErrUserNotFound produces non-null body",
			err:  domainerror.ErrUserNotFound,
		},
		{
			name: "wrapped ErrUserNotFound produces non-null body",
			err:  fmt.Errorf("wrapped: %w", domainerror.ErrUserNotFound),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			handled, err := domainerror.WriteDomainError(w, tc.err)

			assert.True(t, handled)
			assert.NoError(t, err)
			require.NotEmpty(t, w.Body.Bytes(), "body must never be null/empty for UserNotFoundException")

			var em model.ErrorMessage
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &em))
			assert.Equal(t, http.StatusNotFound, em.Status)
		})
	}
}
```