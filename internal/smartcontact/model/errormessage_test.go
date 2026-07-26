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
// Struct zero-value (no-args constructor equivalent)
// ---------------------------------------------------------------------------

func TestErrorMessage_ZeroValue(t *testing.T) {
	tests := []struct {
		name            string
		expectedStatus  int
		expectedMessage string
	}{
		{
			name:            "zero value has zero status and empty message",
			expectedStatus:  0,
			expectedMessage: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var em model.ErrorMessage
			assert.Equal(t, tc.expectedStatus, em.Status)
			assert.Equal(t, tc.expectedMessage, em.Message)
		})
	}
}

// ---------------------------------------------------------------------------
// NewErrorMessage (all-args constructor equivalent)
// ---------------------------------------------------------------------------

func TestNewErrorMessage(t *testing.T) {
	tests := []struct {
		name            string
		inputStatus     int
		inputMessage    string
		expectedStatus  int
		expectedMessage string
	}{
		{
			name:            "constructs with status and message",
			inputStatus:     http.StatusNotFound,
			inputMessage:    "resource not found",
			expectedStatus:  http.StatusNotFound,
			expectedMessage: "resource not found",
		},
		{
			name:            "constructs with internal server error status",
			inputStatus:     http.StatusInternalServerError,
			inputMessage:    "unexpected error",
			expectedStatus:  http.StatusInternalServerError,
			expectedMessage: "unexpected error",
		},
		{
			name:            "constructs with zero status (null-equivalent) and empty message",
			inputStatus:     0,
			inputMessage:    "",
			expectedStatus:  0,
			expectedMessage: "",
		},
		{
			name:            "constructs with zero status and non-empty message",
			inputStatus:     0,
			inputMessage:    "some message",
			expectedStatus:  0,
			expectedMessage: "some message",
		},
		{
			name:            "constructs with valid status and empty message",
			inputStatus:     http.StatusBadRequest,
			inputMessage:    "",
			expectedStatus:  http.StatusBadRequest,
			expectedMessage: "",
		},
		{
			name:            "constructs with conflict status",
			inputStatus:     http.StatusConflict,
			inputMessage:    "conflict occurred",
			expectedStatus:  http.StatusConflict,
			expectedMessage: "conflict occurred",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			em := model.NewErrorMessage(tc.inputStatus, tc.inputMessage)
			assert.Equal(t, tc.expectedStatus, em.Status)
			assert.Equal(t, tc.expectedMessage, em.Message)
		})
	}
}

// ---------------------------------------------------------------------------
// Field access (getStatus / setStatus / getMessage / setMessage equivalents)
// Go exported fields are accessed directly; mutation is done via assignment.
// ---------------------------------------------------------------------------

func TestErrorMessage_StatusField(t *testing.T) {
	tests := []struct {
		name           string
		initialStatus  int
		setStatus      int
		expectedStatus int
	}{
		{
			name:           "status returns value set at construction",
			initialStatus:  http.StatusNotFound,
			setStatus:      http.StatusNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "status zero (null-equivalent) before any set",
			initialStatus:  0,
			setStatus:      0,
			expectedStatus: 0,
		},
		{
			name:           "status can be changed after construction",
			initialStatus:  http.StatusOK,
			setStatus:      http.StatusForbidden,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "status can be reset to zero after being set",
			initialStatus:  http.StatusBadRequest,
			setStatus:      0,
			expectedStatus: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			em := model.NewErrorMessage(tc.initialStatus, "msg")
			em.Status = tc.setStatus
			assert.Equal(t, tc.expectedStatus, em.Status)
		})
	}
}

func TestErrorMessage_MessageField(t *testing.T) {
	tests := []struct {
		name            string
		initialMessage  string
		setMessage      string
		expectedMessage string
	}{
		{
			name:            "message returns value set at construction",
			initialMessage:  "initial message",
			setMessage:      "initial message",
			expectedMessage: "initial message",
		},
		{
			name:            "message empty string (null-equivalent) before any set",
			initialMessage:  "",
			setMessage:      "",
			expectedMessage: "",
		},
		{
			name:            "message can be changed after construction",
			initialMessage:  "old message",
			setMessage:      "new message",
			expectedMessage: "new message",
		},
		{
			name:            "message can be reset to empty after being set",
			initialMessage:  "some message",
			setMessage:      "",
			expectedMessage: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			em := model.NewErrorMessage(http.StatusOK, tc.initialMessage)
			em.Message = tc.setMessage
			assert.Equal(t, tc.expectedMessage, em.Message)
		})
	}
}

// ---------------------------------------------------------------------------
// Equality (equals equivalent) — Go struct value equality via ==
// ---------------------------------------------------------------------------

func TestErrorMessage_Equality(t *testing.T) {
	tests := []struct {
		name     string
		a        model.ErrorMessage
		b        model.ErrorMessage
		expected bool
	}{
		{
			name:     "equal when both status and message match",
			a:        model.NewErrorMessage(http.StatusNotFound, "not found"),
			b:        model.NewErrorMessage(http.StatusNotFound, "not found"),
			expected: true,
		},
		{
			name:     "not equal when status differs",
			a:        model.NewErrorMessage(http.StatusNotFound, "error"),
			b:        model.NewErrorMessage(http.StatusBadRequest, "error"),
			expected: false,
		},
		{
			name:     "not equal when message differs",
			a:        model.NewErrorMessage(http.StatusNotFound, "error a"),
			b:        model.NewErrorMessage(http.StatusNotFound, "error b"),
			expected: false,
		},
		{
			name:     "not equal when both fields differ",
			a:        model.NewErrorMessage(http.StatusNotFound, "not found"),
			b:        model.NewErrorMessage(http.StatusInternalServerError, "server error"),
			expected: false,
		},
		{
			name:     "equal zero values",
			a:        model.ErrorMessage{},
			b:        model.ErrorMessage{},
			expected: true,
		},
		{
			name:     "reflexive: instance equals itself",
			a:        model.NewErrorMessage(http.StatusOK, "ok"),
			b:        model.NewErrorMessage(http.StatusOK, "ok"),
			expected: true,
		},
		{
			name:     "equal with empty message",
			a:        model.NewErrorMessage(http.StatusBadRequest, ""),
			b:        model.NewErrorMessage(http.StatusBadRequest, ""),
			expected: true,
		},
		{
			name:     "equal with zero status",
			a:        model.NewErrorMessage(0, "msg"),
			b:        model.NewErrorMessage(0, "msg"),
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.a == tc.b
			assert.Equal(t, tc.expected, result)
		})
	}
}

// ---------------------------------------------------------------------------
// Reflexive, symmetric and transitive equality invariants
// ---------------------------------------------------------------------------

func TestErrorMessage_EqualityInvariants(t *testing.T) {
	t.Run("reflexive equality", func(t *testing.T) {
		em := model.NewErrorMessage(http.StatusNotFound, "not found")
		assert.Equal(t, em, em)
	})

	t.Run("symmetric equality", func(t *testing.T) {
		a := model.NewErrorMessage(http.StatusNotFound, "not found")
		b := model.NewErrorMessage(http.StatusNotFound, "not found")
		assert.Equal(t, a == b, b == a)
	})

	t.Run("transitive equality", func(t *testing.T) {
		a := model.NewErrorMessage(http.StatusNotFound, "not found")
		b := model.NewErrorMessage(http.StatusNotFound, "not found")
		c := model.NewErrorMessage(http.StatusNotFound, "not found")
		assert.True(t, a == b)
		assert.True(t, b == c)
		assert.True(t, a == c)
	})
}

// ---------------------------------------------------------------------------
// String representation (toString equivalent) via fmt.Sprintf
// ---------------------------------------------------------------------------

func TestErrorMessage_StringRepresentation(t *testing.T) {
	tests := []struct {
		name            string
		em              model.ErrorMessage
		containsStatus  string
		containsMessage string
	}{
		{
			name:            "string contains status and message for typical error",
			em:              model.NewErrorMessage(http.StatusNotFound, "not found"),
			containsStatus:  "404",
			containsMessage: "not found",
		},
		{
			name:            "string contains status and message for server error",
			em:              model.NewErrorMessage(http.StatusInternalServerError, "internal error"),
			containsStatus:  "500",
			containsMessage: "internal error",
		},
		{
			name:            "string contains zero values",
			em:              model.ErrorMessage{},
			containsStatus:  "0",
			containsMessage: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			str := fmt.Sprintf("%+v", tc.em)
			assert.Contains(t, str, tc.containsStatus)
			if tc.containsMessage != "" {
				assert.Contains(t, str, tc.containsMessage)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// JSON serialization / deserialization
// ---------------------------------------------------------------------------

func TestErrorMessage_JSONSerialization(t *testing.T) {
	tests := []struct {
		name         string
		em           model.ErrorMessage
		expectedJSON string
	}{
		{
			name:         "serializes status and message as JSON",
			em:           model.NewErrorMessage(http.StatusNotFound, "not found"),
			expectedJSON: `{"status":404,"message":"not found"}`,
		},
		{
			name:         "serializes zero values as JSON",
			em:           model.ErrorMessage{},
			expectedJSON: `{"status":0,"message":""}`,
		},
		{
			name:         "serializes internal server error",
			em:           model.NewErrorMessage(http.StatusInternalServerError, "server error"),
			expectedJSON: `{"status":500,"message":"server error"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.em)
			assert.NoError(t, err)
			assert.JSONEq(t, tc.expectedJSON, string(data))
		})
	}
}

func TestErrorMessage_JSONDeserialization(t *testing.T) {
	tests := []struct {
		name            string
		jsonInput       string
		expectedStatus  int
		expectedMessage string
	}{
		{
			name:            "deserializes status and message from JSON",
			jsonInput:       `{"status":404,"message":"not found"}`,
			expectedStatus:  http.StatusNotFound,
			expectedMessage: "not found",
		},
		{
			name:            "deserializes zero values from JSON",
			jsonInput:       `{"status":0,"message":""}`,
			expectedStatus:  0,
			expectedMessage: "",
		},
		{
			name:            "deserializes partial JSON with only status",
			jsonInput:       `{"status":500}`,
			expectedStatus:  http.StatusInternalServerError,
			expectedMessage: "",
		},
		{
			name:            "deserializes partial JSON with only message",
			jsonInput:       `{"message":"bad request"}`,
			expectedStatus:  0,
			expectedMessage: "bad request",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var em model.ErrorMessage
			err := json.Unmarshal([]byte(tc.jsonInput), &em)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, em.Status)
			assert.Equal(t, tc.expectedMessage, em.Message)
		})
	}
}

// ---------------------------------------------------------------------------
// HTTP handler integration tests using httptest
// ---------------------------------------------------------------------------

func TestErrorMessage_HTTPHandler(t *testing.T) {
	tests := []struct {
		name               string
		em                 model.ErrorMessage
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name:               "handler writes 404 ErrorMessage",
			em:                 model.NewErrorMessage(http.StatusNotFound, "resource not found"),
			expectedStatusCode: http.StatusNotFound,
			expectedBody:       `{"status":404,"message":"resource not found"}`,
		},
		{
			name:               "handler writes 500 ErrorMessage",
			em:                 model.NewErrorMessage(http.StatusInternalServerError, "internal server error"),
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       `{"status":500,"message":"internal server error"}`,
		},
		{
			name:               "handler writes 400 ErrorMessage",
			em:                 model.NewErrorMessage(http.StatusBadRequest, "bad request"),
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `{"status":400,"message":"bad request"}`,
		},
		{
			name:               "handler writes ErrorMessage with empty message",
			em:                 model.NewErrorMessage(http.StatusForbidden, ""),
			expectedStatusCode: http.StatusForbidden,
			expectedBody:       `{"status":403,"message":""}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.em.Status)
				data, err := json.Marshal(tc.em)
				assert.NoError(t, err)
				_, _ = w.Write(data)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatusCode, rec.Code)
			assert.JSONEq(t, tc.expectedBody, rec.Body.String())

			var decoded model.ErrorMessage
			err := json.Unmarshal(rec.Body.Bytes(), &decoded)
			assert.NoError(t, err)
			assert.Equal(t, tc.em, decoded)
		})
	}
}

// ---------------------------------------------------------------------------
// No business logic / no transformation invari