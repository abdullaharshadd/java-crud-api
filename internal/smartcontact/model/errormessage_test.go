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

// TestErrorMessage_ZeroValue validates the "no-args constructor" spec:
// a zero-value ErrorMessage has status == 0 and message == "".
func TestErrorMessage_ZeroValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		em              model.ErrorMessage
		expectedStatus  int
		expectedMessage string
	}{
		{
			name:            "instantiated with no arguments – all fields default to zero values",
			em:              model.ErrorMessage{},
			expectedStatus:  0,
			expectedMessage: "",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expectedStatus, tc.em.Status)
			assert.Equal(t, tc.expectedMessage, tc.em.Message)
		})
	}
}

// TestNewErrorMessage validates the all-args constructor.
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
			status:          http.StatusBadRequest,
			message:         "bad request",
			expectedStatus:  http.StatusBadRequest,
			expectedMessage: "bad request",
		},
		{
			name:            "internal server error status",
			status:          http.StatusInternalServerError,
			message:         "internal server error",
			expectedStatus:  http.StatusInternalServerError,
			expectedMessage: "internal server error",
		},
		{
			name:            "zero status and empty message (null-like values)",
			status:          0,
			message:         "",
			expectedStatus:  0,
			expectedMessage: "",
		},
		{
			name:            "not found status with descriptive message",
			status:          http.StatusNotFound,
			message:         "resource not found",
			expectedStatus:  http.StatusNotFound,
			expectedMessage: "resource not found",
		},
		{
			name:            "status only, no message",
			status:          http.StatusUnauthorized,
			message:         "",
			expectedStatus:  http.StatusUnauthorized,
			expectedMessage: "",
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

// TestErrorMessage_GetStatus validates getter-like access to the Status field.
func TestErrorMessage_GetStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		em             model.ErrorMessage
		expectedStatus int
	}{
		{
			name:           "status was set via constructor",
			em:             model.NewErrorMessage(http.StatusOK, "ok"),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "status was never set – returns zero value",
			em:             model.ErrorMessage{},
			expectedStatus: 0,
		},
		{
			name:           "status set directly on struct field",
			em:             model.ErrorMessage{Status: http.StatusForbidden},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expectedStatus, tc.em.Status)
		})
	}
}

// TestErrorMessage_SetStatus validates setter-like mutation of the Status field.
func TestErrorMessage_SetStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		initial        int
		newStatus      int
		expectedStatus int
	}{
		{
			name:           "set to a valid HTTP status",
			initial:        http.StatusOK,
			newStatus:      http.StatusBadRequest,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "set to zero (null-like)",
			initial:        http.StatusBadRequest,
			newStatus:      0,
			expectedStatus: 0,
		},
		{
			name:           "set to same value",
			initial:        http.StatusNotFound,
			newStatus:      http.StatusNotFound,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			em := model.ErrorMessage{Status: tc.initial}
			em.Status = tc.newStatus
			assert.Equal(t, tc.expectedStatus, em.Status)
		})
	}
}

// TestErrorMessage_GetMessage validates getter-like access to the Message field.
func TestErrorMessage_GetMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		em              model.ErrorMessage
		expectedMessage string
	}{
		{
			name:            "message was set via constructor",
			em:              model.NewErrorMessage(http.StatusBadRequest, "some error"),
			expectedMessage: "some error",
		},
		{
			name:            "message was never set – returns empty string",
			em:              model.ErrorMessage{},
			expectedMessage: "",
		},
		{
			name:            "message set directly on struct field",
			em:              model.ErrorMessage{Message: "direct assignment"},
			expectedMessage: "direct assignment",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expectedMessage, tc.em.Message)
		})
	}
}

// TestErrorMessage_SetMessage validates setter-like mutation of the Message field.
func TestErrorMessage_SetMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		initial         string
		newMessage      string
		expectedMessage string
	}{
		{
			name:            "set to a non-empty string",
			initial:         "old message",
			newMessage:      "new message",
			expectedMessage: "new message",
		},
		{
			name:            "set to empty string (null-like)",
			initial:         "some message",
			newMessage:      "",
			expectedMessage: "",
		},
		{
			name:            "set to same value",
			initial:         "unchanged",
			newMessage:      "unchanged",
			expectedMessage: "unchanged",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			em := model.ErrorMessage{Message: tc.initial}
			em.Message = tc.newMessage
			assert.Equal(t, tc.expectedMessage, em.Message)
		})
	}
}

// TestErrorMessage_Equality validates struct value equality (Go equality semantics).
func TestErrorMessage_Equality(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		a        model.ErrorMessage
		b        model.ErrorMessage
		areEqual bool
	}{
		{
			name:     "identical status and message – equal",
			a:        model.NewErrorMessage(http.StatusBadRequest, "bad request"),
			b:        model.NewErrorMessage(http.StatusBadRequest, "bad request"),
			areEqual: true,
		},
		{
			name:     "different status – not equal",
			a:        model.NewErrorMessage(http.StatusBadRequest, "bad request"),
			b:        model.NewErrorMessage(http.StatusNotFound, "bad request"),
			areEqual: false,
		},
		{
			name:     "different message – not equal",
			a:        model.NewErrorMessage(http.StatusBadRequest, "bad request"),
			b:        model.NewErrorMessage(http.StatusBadRequest, "not found"),
			areEqual: false,
		},
		{
			name:     "both zero values – equal",
			a:        model.ErrorMessage{},
			b:        model.ErrorMessage{},
			areEqual: true,
		},
		{
			name:     "one zero value, one non-zero – not equal",
			a:        model.ErrorMessage{},
			b:        model.NewErrorMessage(http.StatusOK, "ok"),
			areEqual: false,
		},
		{
			name:     "reflexive – same instance value compared to itself",
			a:        model.NewErrorMessage(http.StatusTeapot, "i am a teapot"),
			b:        model.NewErrorMessage(http.StatusTeapot, "i am a teapot"),
			areEqual: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if tc.areEqual {
				assert.Equal(t, tc.a, tc.b)
			} else {
				assert.NotEqual(t, tc.a, tc.b)
			}
		})
	}
}

// TestErrorMessage_Symmetry validates symmetric equality: if a == b then b == a.
func TestErrorMessage_Symmetry(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a    model.ErrorMessage
		b    model.ErrorMessage
	}{
		{
			name: "symmetric equality for matching instances",
			a:    model.NewErrorMessage(http.StatusOK, "ok"),
			b:    model.NewErrorMessage(http.StatusOK, "ok"),
		},
		{
			name: "symmetric equality for zero values",
			a:    model.ErrorMessage{},
			b:    model.ErrorMessage{},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.a, tc.b)
			assert.Equal(t, tc.b, tc.a)
		})
	}
}

// TestErrorMessage_Transitivity validates transitive equality: if a==b and b==c then a==c.
func TestErrorMessage_Transitivity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a    model.ErrorMessage
		b    model.ErrorMessage
		c    model.ErrorMessage
	}{
		{
			name: "transitive equality",
			a:    model.NewErrorMessage(http.StatusCreated, "created"),
			b:    model.NewErrorMessage(http.StatusCreated, "created"),
			c:    model.NewErrorMessage(http.StatusCreated, "created"),
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.a, tc.b)
			assert.Equal(t, tc.b, tc.c)
			assert.Equal(t, tc.a, tc.c)
		})
	}
}

// TestErrorMessage_NoExternalSideEffects validates that creating or mutating an
// ErrorMessage has no external side effects (no panics, no global state changes).
func TestErrorMessage_NoExternalSideEffects(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		fn   func()
	}{
		{
			name: "NewErrorMessage does not panic",
			fn: func() {
				_ = model.NewErrorMessage(http.StatusBadRequest, "bad")
			},
		},
		{
			name: "mutating Status field does not panic",
			fn: func() {
				em := model.NewErrorMessage(http.StatusOK, "ok")
				em.Status = http.StatusNotFound
				_ = em
			},
		},
		{
			name: "mutating Message field does not panic",
			fn: func() {
				em := model.NewErrorMessage(http.StatusOK, "ok")
				em.Message = "new message"
				_ = em
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.NotPanics(t, tc.fn)
		})
	}
}

// TestErrorMessage_JSONSerialization validates that the struct serialises to and
// from JSON correctly, exercising the json struct tags.
func TestErrorMessage_JSONSerialization(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		em          model.ErrorMessage
		expectedJSON string
	}{
		{
			name:        "standard error message",
			em:          model.NewErrorMessage(http.StatusBadRequest, "bad request"),
			expectedJSON: `{"status":400,"message":"bad request"}`,
		},
		{
			name:        "zero value",
			em:          model.ErrorMessage{},
			expectedJSON: `{"status":0,"message":""}`,
		},
		{
			name:        "internal server error",
			em:          model.NewErrorMessage(http.StatusInternalServerError, "internal server error"),
			expectedJSON: `{"status":500,"message":"internal server error"}`,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Marshal
			data, err := json.Marshal(tc.em)
			assert.NoError(t, err)
			assert.JSONEq(t, tc.expectedJSON, string(data))

			// Unmarshal round-trip
			var decoded model.ErrorMessage
			err = json.Unmarshal(data, &decoded)
			assert.NoError(t, err)
			assert.Equal(t, tc.em, decoded)
		})
	}
}

// TestErrorMessage_HTTPHandler exercises ErrorMessage in a realistic HTTP handler
// context using httptest, validating JSON serialisation over the wire.
func TestErrorMessage_HTTPHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		em             model.ErrorMessage
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "bad request response",
			em:             model.NewErrorMessage(http.StatusBadRequest, "invalid input"),
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"status":400,"message":"invalid input"}`,
		},
		{
			name:           "not found response",
			em:             model.NewErrorMessage(http.StatusNotFound, "resource not found"),
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"status":404,"message":"resource not found"}`,
		},
		{
			name:           "internal server error response",
			em:             model.NewErrorMessage(http.StatusInternalServerError, "unexpected error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"status":500,"message":"unexpected error"}`,
		},
		{
			name:           "unauthorised response",
			em:             model.NewErrorMessage(http.StatusUnauthorized, "not authorised"),
			expectedStatus: http.StatusUnauthorized,
			expected