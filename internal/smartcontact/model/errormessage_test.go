```go
package model_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"internal/smartcontact/model"
)

// ---------------------------------------------------------------------------
// UserNotFoundMessageFormat
// ---------------------------------------------------------------------------

func TestUserNotFoundMessageFormat(t *testing.T) {
	tests := []struct {
		name     string
		id       int
		expected string
	}{
		{
			name:     "formats id 1",
			id:       1,
			expected: "User not found with id : 1",
		},
		{
			name:     "formats id 0",
			id:       0,
			expected: "User not found with id : 0",
		},
		{
			name:     "formats large id",
			id:       999999,
			expected: "User not found with id : 999999",
		},
		{
			name:     "formats negative id",
			id:       -1,
			expected: "User not found with id : -1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := fmt.Sprintf(model.UserNotFoundMessageFormat, tc.id)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// ---------------------------------------------------------------------------
// NewErrorMessage (zero-value / no-args equivalent)
// ---------------------------------------------------------------------------

func TestErrorMessage_ZeroValue(t *testing.T) {
	t.Run("zero value has zero status and empty message", func(t *testing.T) {
		var e model.ErrorMessage
		assert.Equal(t, 0, e.Status)
		assert.Equal(t, "", e.Message)
	})
}

// ---------------------------------------------------------------------------
// NewErrorMessage (all-args constructor)
// ---------------------------------------------------------------------------

func TestNewErrorMessage(t *testing.T) {
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
			message:         "not found",
			expectedStatus:  http.StatusNotFound,
			expectedMessage: "not found",
		},
		{
			name:            "status OK with message",
			status:          http.StatusOK,
			message:         "success",
			expectedStatus:  http.StatusOK,
			expectedMessage: "success",
		},
		{
			name:            "internal server error",
			status:          http.StatusInternalServerError,
			message:         "internal error",
			expectedStatus:  http.StatusInternalServerError,
			expectedMessage: "internal error",
		},
		{
			name:            "zero status (null-equivalent) and empty message (null-equivalent)",
			status:          0,
			message:         "",
			expectedStatus:  0,
			expectedMessage: "",
		},
		{
			name:            "arbitrary status code with message",
			status:          http.StatusBadRequest,
			message:         "bad request",
			expectedStatus:  http.StatusBadRequest,
			expectedMessage: "bad request",
		},
		{
			name:            "user not found format message",
			status:          http.StatusNotFound,
			message:         fmt.Sprintf(model.UserNotFoundMessageFormat, 42),
			expectedStatus:  http.StatusNotFound,
			expectedMessage: "User not found with id : 42",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := model.NewErrorMessage(tc.status, tc.message)

			assert.Equal(t, tc.expectedStatus, e.Status, "Status should match")
			assert.Equal(t, tc.expectedMessage, e.Message, "Message should match")
		})
	}
}

// ---------------------------------------------------------------------------
// Status field (getter / setter equivalent – direct field access in Go)
// ---------------------------------------------------------------------------

func TestErrorMessage_StatusField(t *testing.T) {
	tests := []struct {
		name   string
		status int
	}{
		{"status set to 404", http.StatusNotFound},
		{"status set to 200", http.StatusOK},
		{"status set to 500", http.StatusInternalServerError},
		{"status set to 0 (null-equivalent)", 0},
		{"status set to 401", http.StatusUnauthorized},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := model.ErrorMessage{}
			e.Status = tc.status
			assert.Equal(t, tc.status, e.Status)
		})
	}
}

// ---------------------------------------------------------------------------
// Message field (getter / setter equivalent – direct field access in Go)
// ---------------------------------------------------------------------------

func TestErrorMessage_MessageField(t *testing.T) {
	tests := []struct {
		name    string
		message string
	}{
		{"message set to non-empty string", "an error occurred"},
		{"message set to empty string (null-equivalent)", ""},
		{"message set to whitespace", "   "},
		{"message with special characters", "error: something went wrong!"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := model.ErrorMessage{}
			e.Message = tc.message
			assert.Equal(t, tc.message, e.Message)
		})
	}
}

// ---------------------------------------------------------------------------
// Equality (mirrors equals / hashCode specs)
// ---------------------------------------------------------------------------

func TestErrorMessage_Equality(t *testing.T) {
	tests := []struct {
		name     string
		a        model.ErrorMessage
		b        model.ErrorMessage
		areEqual bool
	}{
		{
			name:     "identical status and message are equal",
			a:        model.NewErrorMessage(http.StatusNotFound, "not found"),
			b:        model.NewErrorMessage(http.StatusNotFound, "not found"),
			areEqual: true,
		},
		{
			name:     "same instance reflected is equal to itself",
			a:        model.NewErrorMessage(http.StatusOK, "ok"),
			b:        model.NewErrorMessage(http.StatusOK, "ok"),
			areEqual: true,
		},
		{
			name:     "different status produces not equal",
			a:        model.NewErrorMessage(http.StatusNotFound, "not found"),
			b:        model.NewErrorMessage(http.StatusOK, "not found"),
			areEqual: false,
		},
		{
			name:     "different message produces not equal",
			a:        model.NewErrorMessage(http.StatusNotFound, "not found"),
			b:        model.NewErrorMessage(http.StatusNotFound, "different message"),
			areEqual: false,
		},
		{
			name:     "both status and message differ",
			a:        model.NewErrorMessage(http.StatusBadRequest, "bad request"),
			b:        model.NewErrorMessage(http.StatusInternalServerError, "server error"),
			areEqual: false,
		},
		{
			name:     "zero values are equal",
			a:        model.ErrorMessage{},
			b:        model.ErrorMessage{},
			areEqual: true,
		},
		{
			name:     "zero value vs non-zero are not equal",
			a:        model.ErrorMessage{},
			b:        model.NewErrorMessage(http.StatusNotFound, "not found"),
			areEqual: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.areEqual {
				assert.Equal(t, tc.a, tc.b, "ErrorMessage structs should be equal")
			} else {
				assert.NotEqual(t, tc.a, tc.b, "ErrorMessage structs should not be equal")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Reflexive equality
// ---------------------------------------------------------------------------

func TestErrorMessage_ReflexiveEquality(t *testing.T) {
	tests := []struct {
		name string
		e    model.ErrorMessage
	}{
		{"not found error", model.NewErrorMessage(http.StatusNotFound, "not found")},
		{"ok response", model.NewErrorMessage(http.StatusOK, "ok")},
		{"zero value", model.ErrorMessage{}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.e, tc.e, "An ErrorMessage should equal itself")
		})
	}
}

// ---------------------------------------------------------------------------
// StatusText (toString / representation)
// ---------------------------------------------------------------------------

func TestErrorMessage_StatusText(t *testing.T) {
	tests := []struct {
		name         string
		status       int
		message      string
		expectedText string
	}{
		{
			name:         "404 returns Not Found",
			status:       http.StatusNotFound,
			message:      "user not found",
			expectedText: "Not Found",
		},
		{
			name:         "200 returns OK",
			status:       http.StatusOK,
			message:      "success",
			expectedText: "OK",
		},
		{
			name:         "500 returns Internal Server Error",
			status:       http.StatusInternalServerError,
			message:      "server error",
			expectedText: "Internal Server Error",
		},
		{
			name:         "400 returns Bad Request",
			status:       http.StatusBadRequest,
			message:      "bad request",
			expectedText: "Bad Request",
		},
		{
			name:         "401 returns Unauthorized",
			status:       http.StatusUnauthorized,
			message:      "unauthorized",
			expectedText: "Unauthorized",
		},
		{
			name:         "unknown status code returns empty string",
			status:       9999,
			message:      "unknown",
			expectedText: "",
		},
		{
			name:         "zero status returns empty string",
			status:       0,
			message:      "",
			expectedText: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := model.NewErrorMessage(tc.status, tc.message)
			assert.Equal(t, tc.expectedText, e.StatusText())
		})
	}
}

// ---------------------------------------------------------------------------
// StatusText consistency (toString invariant – never nil in Go)
// ---------------------------------------------------------------------------

func TestErrorMessage_StatusText_NeverPanics(t *testing.T) {
	tests := []struct {
		name   string
		status int
	}{
		{"zero", 0},
		{"negative", -1},
		{"valid 404", http.StatusNotFound},
		{"unknown large", 99999},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := model.NewErrorMessage(tc.status, "msg")
			assert.NotPanics(t, func() {
				_ = e.StatusText()
			})
		})
	}
}

// ---------------------------------------------------------------------------
// JSON serialization (wire compatibility)
// ---------------------------------------------------------------------------

func TestErrorMessage_JSONSerialization(t *testing.T) {
	tests := []struct {
		name        string
		input       model.ErrorMessage
		expectedJSON string
	}{
		{
			name:        "404 not found",
			input:       model.NewErrorMessage(http.StatusNotFound, "not found"),
			expectedJSON: `{"status":404,"message":"not found"}`,
		},
		{
			name:        "200 ok",
			input:       model.NewErrorMessage(http.StatusOK, "ok"),
			expectedJSON: `{"status":200,"message":"ok"}`,
		},
		{
			name:        "zero value",
			input:       model.ErrorMessage{},
			expectedJSON: `{"status":0,"message":""}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.input)
			assert.NoError(t, err)
			assert.JSONEq(t, tc.expectedJSON, string(data))
		})
	}
}

func TestErrorMessage_JSONDeserialization(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		expected model.ErrorMessage
	}{
		{
			name:     "deserialize 404",
			raw:      `{"status":404,"message":"not found"}`,
			expected: model.NewErrorMessage(http.StatusNotFound, "not found"),
		},
		{
			name:     "deserialize 200",
			raw:      `{"status":200,"message":"ok"}`,
			expected: model.NewErrorMessage(http.StatusOK, "ok"),
		},
		{
			name:     "deserialize zero value",
			raw:      `{"status":0,"message":""}`,
			expected: model.ErrorMessage{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var e model.ErrorMessage
			err := json.Unmarshal([]byte(tc.raw), &e)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, e)
		})
	}
}

// ---------------------------------------------------------------------------
// HTTP handler integration tests (httptest)
// ---------------------------------------------------------------------------

func TestErrorMessage_HTTPHandler(t *testing.T) {
	tests := []struct {
		name           string
		errMsg         model.ErrorMessage
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "handler returns 404 error message",
			errMsg:         model.NewErrorMessage(http.StatusNotFound, "User not found with id : 1"),
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"status":404,"message":"User not found with id : 1"}`,
		},
		{
			name:           "handler returns 500 error message",
			errMsg:         model.NewErrorMessage(http.StatusInternalServerError, "internal error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"status":500,"message":"internal error"}`,
		},
		{
			name:           "handler returns 400 error message",
			errMsg:         model.NewErrorMessage(http.StatusBadRequest, "bad request"),
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"status":400,"message":"bad request"}`,
		},
		{
			name:           "handler returns 401 error message",
			errMsg:         model.NewErrorMessage(http.StatusUnauthorized, "unauthorized"),
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"status":401,"message":"unauthorized"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.errMsg.Status)
				data, err := json.Marshal(tc.errMsg)
				assert.NoError(t, err)
				_, _ = w.Write(data)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Code)
			assert.JSONEq(t, tc.expectedBody, rec.Body.String())
			assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		})
	}
}

func TestErrorMessage_HTTPHandler_UserNotFound(t *testing.T) {
	tests := []struct {
		name           string
		userID         int
		expectedStatus int
	}{
		{"user id 1 not found", 1, http.