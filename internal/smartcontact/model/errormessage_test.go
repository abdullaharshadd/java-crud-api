```go
package model_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"internal/smartcontact/model"
)

// TestErrorMessage_ZeroValue validates the "no-args constructor" equivalent in
// Go: a zero-value ErrorMessage has both fields at their zero values (0 and "").
func TestErrorMessage_ZeroValue(t *testing.T) {
	t.Parallel()

	var em model.ErrorMessage

	assert.Equal(t, 0, em.Status, "zero-value Status should be 0")
	assert.Equal(t, "", em.Message, "zero-value Message should be empty string")
}

// TestNewErrorMessage covers the all-args constructor equivalent.
func TestNewErrorMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		status          int
		message         string
		expectedStatus  int
		expectedMessage string
	}{
		{
			name:            "valid status and message",
			status:          http.StatusNotFound,
			message:         "resource not found",
			expectedStatus:  http.StatusNotFound,
			expectedMessage: "resource not found",
		},
		{
			name:            "internal server error with message",
			status:          http.StatusInternalServerError,
			message:         "something went wrong",
			expectedStatus:  http.StatusInternalServerError,
			expectedMessage: "something went wrong",
		},
		{
			name:            "bad request with empty message",
			status:          http.StatusBadRequest,
			message:         "",
			expectedStatus:  http.StatusBadRequest,
			expectedMessage: "",
		},
		{
			name:            "zero status (null-equivalent) with non-empty message",
			status:          0,
			message:         "some error",
			expectedStatus:  0,
			expectedMessage: "some error",
		},
		{
			name:            "zero status and empty message (null-equivalent for both)",
			status:          0,
			message:         "",
			expectedStatus:  0,
			expectedMessage: "",
		},
		{
			name:            "status ok",
			status:          http.StatusOK,
			message:         "ok",
			expectedStatus:  http.StatusOK,
			expectedMessage: "ok",
		},
		{
			name:            "status unauthorized",
			status:          http.StatusUnauthorized,
			message:         "unauthorized access",
			expectedStatus:  http.StatusUnauthorized,
			expectedMessage: "unauthorized access",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			em := model.NewErrorMessage(tc.status, tc.message)

			assert.Equal(t, tc.expectedStatus, em.Status)
			assert.Equal(t, tc.expectedMessage, em.Message)
		})
	}
}

// TestErrorMessage_StatusField covers getStatus / setStatus equivalents via
// direct struct field access (Go's idiomatic equivalent of Lombok getters/setters).
func TestErrorMessage_StatusField(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		initialStatus  int
		updatedStatus  int
		expectInitial  int
		expectUpdated  int
	}{
		{
			name:          "set and get a valid status",
			initialStatus: http.StatusOK,
			updatedStatus: http.StatusNotFound,
			expectInitial: http.StatusOK,
			expectUpdated: http.StatusNotFound,
		},
		{
			name:          "set status to zero (null-equivalent)",
			initialStatus: http.StatusBadRequest,
			updatedStatus: 0,
			expectInitial: http.StatusBadRequest,
			expectUpdated: 0,
		},
		{
			name:          "zero initial status updated to non-zero",
			initialStatus: 0,
			updatedStatus: http.StatusForbidden,
			expectInitial: 0,
			expectUpdated: http.StatusForbidden,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			em := model.ErrorMessage{Status: tc.initialStatus}
			assert.Equal(t, tc.expectInitial, em.Status, "initial status mismatch")

			// Simulate setStatus
			em.Status = tc.updatedStatus
			assert.Equal(t, tc.expectUpdated, em.Status, "updated status mismatch")
		})
	}
}

// TestErrorMessage_MessageField covers getMessage / setMessage equivalents.
func TestErrorMessage_MessageField(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		initialMessage  string
		updatedMessage  string
		expectInitial   string
		expectUpdated   string
	}{
		{
			name:           "set and get a valid message",
			initialMessage: "initial error",
			updatedMessage: "updated error",
			expectInitial:  "initial error",
			expectUpdated:  "updated error",
		},
		{
			name:           "set message to empty string (null-equivalent)",
			initialMessage: "some error",
			updatedMessage: "",
			expectInitial:  "some error",
			expectUpdated:  "",
		},
		{
			name:           "empty initial message updated to non-empty",
			initialMessage: "",
			updatedMessage: "new message",
			expectInitial:  "",
			expectUpdated:  "new message",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			em := model.ErrorMessage{Message: tc.initialMessage}
			assert.Equal(t, tc.expectInitial, em.Message, "initial message mismatch")

			// Simulate setMessage
			em.Message = tc.updatedMessage
			assert.Equal(t, tc.expectUpdated, em.Message, "updated message mismatch")
		})
	}
}

// TestErrorMessage_Equality covers the equals / hashCode equivalents.
// In Go, struct comparison with == and reflect.DeepEqual is the idiomatic
// approach for value types.
func TestErrorMessage_Equality(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		a           model.ErrorMessage
		b           model.ErrorMessage
		expectEqual bool
	}{
		{
			name:        "equal status and message",
			a:           model.NewErrorMessage(http.StatusNotFound, "not found"),
			b:           model.NewErrorMessage(http.StatusNotFound, "not found"),
			expectEqual: true,
		},
		{
			name:        "different status same message",
			a:           model.NewErrorMessage(http.StatusNotFound, "not found"),
			b:           model.NewErrorMessage(http.StatusBadRequest, "not found"),
			expectEqual: false,
		},
		{
			name:        "same status different message",
			a:           model.NewErrorMessage(http.StatusNotFound, "not found"),
			b:           model.NewErrorMessage(http.StatusNotFound, "bad request"),
			expectEqual: false,
		},
		{
			name:        "both fields different",
			a:           model.NewErrorMessage(http.StatusNotFound, "not found"),
			b:           model.NewErrorMessage(http.StatusInternalServerError, "server error"),
			expectEqual: false,
		},
		{
			name:        "zero-value instances are equal",
			a:           model.ErrorMessage{},
			b:           model.ErrorMessage{},
			expectEqual: true,
		},
		{
			name:        "reflexive: instance equals itself",
			a:           model.NewErrorMessage(http.StatusOK, "ok"),
			b:           model.NewErrorMessage(http.StatusOK, "ok"),
			expectEqual: true,
		},
		{
			name:        "zero status equals zero status",
			a:           model.NewErrorMessage(0, ""),
			b:           model.NewErrorMessage(0, ""),
			expectEqual: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.expectEqual {
				assert.Equal(t, tc.a, tc.b, "expected instances to be equal")
				// Symmetric
				assert.Equal(t, tc.b, tc.a, "equality should be symmetric")
			} else {
				assert.NotEqual(t, tc.a, tc.b, "expected instances to be not equal")
				// Symmetric
				assert.NotEqual(t, tc.b, tc.a, "inequality should be symmetric")
			}
		})
	}
}

// TestErrorMessage_Reflexive validates reflexivity: a value equals itself.
func TestErrorMessage_Reflexive(t *testing.T) {
	t.Parallel()

	em := model.NewErrorMessage(http.StatusConflict, "conflict occurred")
	assert.Equal(t, em, em, "a struct value should equal itself")
}

// TestErrorMessage_Transitivity validates transitivity of equality.
func TestErrorMessage_Transitivity(t *testing.T) {
	t.Parallel()

	a := model.NewErrorMessage(http.StatusGone, "gone")
	b := model.NewErrorMessage(http.StatusGone, "gone")
	c := model.NewErrorMessage(http.StatusGone, "gone")

	assert.Equal(t, a, b, "a == b")
	assert.Equal(t, b, c, "b == c")
	assert.Equal(t, a, c, "a == c (transitivity)")
}

// TestErrorMessage_StringRepresentation covers the toString equivalent: a
// formatted string that reflects the current field values.
func TestErrorMessage_StringRepresentation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		em              model.ErrorMessage
		containsStatus  string
		containsMessage string
	}{
		{
			name:            "standard error message JSON representation",
			em:              model.NewErrorMessage(http.StatusNotFound, "resource not found"),
			containsStatus:  "404",
			containsMessage: "resource not found",
		},
		{
			name:            "internal server error JSON representation",
			em:              model.NewErrorMessage(http.StatusInternalServerError, "server error"),
			containsStatus:  "500",
			containsMessage: "server error",
		},
		{
			name:            "zero-value JSON representation",
			em:              model.ErrorMessage{},
			containsStatus:  "0",
			containsMessage: "",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// JSON serialization is the Go equivalent of toString for DTOs.
			b, err := json.Marshal(tc.em)
			assert.NoError(t, err)

			jsonStr := string(b)
			assert.Contains(t, jsonStr, tc.containsStatus, "JSON should contain the status code")
			if tc.containsMessage != "" {
				assert.Contains(t, jsonStr, tc.containsMessage, "JSON should contain the message")
			}
		})
	}
}

// TestErrorMessage_JSONSerialization validates round-trip JSON encoding/decoding.
func TestErrorMessage_JSONSerialization(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		em      model.ErrorMessage
		wantJSON string
	}{
		{
			name:     "not found",
			em:       model.NewErrorMessage(http.StatusNotFound, "not found"),
			wantJSON: `{"status":404,"message":"not found"}`,
		},
		{
			name:     "internal server error",
			em:       model.NewErrorMessage(http.StatusInternalServerError, "internal error"),
			wantJSON: `{"status":500,"message":"internal error"}`,
		},
		{
			name:     "zero value",
			em:       model.ErrorMessage{},
			wantJSON: `{"status":0,"message":""}`,
		},
		{
			name:     "bad request empty message",
			em:       model.NewErrorMessage(http.StatusBadRequest, ""),
			wantJSON: `{"status":400,"message":""}`,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Marshal
			b, err := json.Marshal(tc.em)
			assert.NoError(t, err)
			assert.JSONEq(t, tc.wantJSON, string(b))

			// Unmarshal round-trip
			var decoded model.ErrorMessage
			err = json.Unmarshal(b, &decoded)
			assert.NoError(t, err)
			assert.Equal(t, tc.em, decoded)
		})
	}
}

// TestErrorMessage_StatusText covers the StatusText helper method.
func TestErrorMessage_StatusText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		em             model.ErrorMessage
		expectedText   string
	}{
		{
			name:         "404 Not Found",
			em:           model.NewErrorMessage(http.StatusNotFound, "not found"),
			expectedText: "Not Found",
		},
		{
			name:         "500 Internal Server Error",
			em:           model.NewErrorMessage(http.StatusInternalServerError, "server error"),
			expectedText: "Internal Server Error",
		},
		{
			name:         "200 OK",
			em:           model.NewErrorMessage(http.StatusOK, "ok"),
			expectedText: "OK",
		},
		{
			name:         "400 Bad Request",
			em:           model.NewErrorMessage(http.StatusBadRequest, "bad request"),
			expectedText: "Bad Request",
		},
		{
			name:         "401 Unauthorized",
			em:           model.NewErrorMessage(http.StatusUnauthorized, "unauthorized"),
			expectedText: "Unauthorized",
		},
		{
			name:         "403 Forbidden",
			em:           model.NewErrorMessage(http.StatusForbidden, "forbidden"),
			expectedText: "Forbidden",
		},
		{
			name:         "unrecognized status code returns empty string",
			em:           model.NewErrorMessage(9999, "unknown"),
			expectedText: "",
		},
		{
			name:         "zero status returns empty string",
			em:           model.NewErrorMessage(0, ""),
			expectedText: "",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expectedText, tc.em.StatusText())
		})
	}
}

// TestErrorMessage_HTTPHandler validates ErrorMessage in an HTTP handler context
// using httptest, ensuring it serializes correctly as an API error response.
func TestErrorMessage_HTTPHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		em             model.ErrorMessage
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "not found response",
			em:             model.NewErrorMessage(http.StatusNotFound, "resource not found"),
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"status":404,"message":"resource not found"}`,
		},
		{
			name:           "internal server error response",
			em:             model.NewErrorMessage(http.StatusInternalServerError, "unexpected error"),
			