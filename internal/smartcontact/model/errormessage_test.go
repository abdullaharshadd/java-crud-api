```go
package model

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestErrorMessage_ZeroValue validates that a zero-value ErrorMessage has
// empty strings for both fields (Go's equivalent of Java's null for strings).
func TestErrorMessage_ZeroValue(t *testing.T) {
	tests := []struct {
		name            string
		instance        ErrorMessage
		expectedStatus  string
		expectedMessage string
	}{
		{
			name:            "instantiated with no arguments (zero value)",
			instance:        ErrorMessage{},
			expectedStatus:  "",
			expectedMessage: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedStatus, tt.instance.Status)
			assert.Equal(t, tt.expectedMessage, tt.instance.Message)
		})
	}
}

// TestNewErrorMessage validates the all-args constructor.
func TestNewErrorMessage(t *testing.T) {
	tests := []struct {
		name            string
		status          string
		message         string
		expectedStatus  string
		expectedMessage string
	}{
		{
			name:            "instantiated with valid status and message",
			status:          "404",
			message:         "Not Found",
			expectedStatus:  "404",
			expectedMessage: "Not Found",
		},
		{
			name:            "instantiated with empty status and message",
			status:          "",
			message:         "",
			expectedStatus:  "",
			expectedMessage: "",
		},
		{
			name:            "instantiated with http status text and message",
			status:          http.StatusText(http.StatusInternalServerError),
			message:         "an unexpected error occurred",
			expectedStatus:  "Internal Server Error",
			expectedMessage: "an unexpected error occurred",
		},
		{
			name:            "instantiated with numeric status string",
			status:          "500",
			message:         "Internal Server Error",
			expectedStatus:  "500",
			expectedMessage: "Internal Server Error",
		},
		{
			name:            "instantiated with only status set",
			status:          "200",
			message:         "",
			expectedStatus:  "200",
			expectedMessage: "",
		},
		{
			name:            "instantiated with only message set",
			status:          "",
			message:         "some error",
			expectedStatus:  "",
			expectedMessage: "some error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			em := NewErrorMessage(tt.status, tt.message)
			assert.Equal(t, tt.expectedStatus, em.Status)
			assert.Equal(t, tt.expectedMessage, em.Message)
		})
	}
}

// TestErrorMessage_StatusField validates getting and setting the Status field.
func TestErrorMessage_StatusField(t *testing.T) {
	tests := []struct {
		name           string
		initialStatus  string
		updatedStatus  string
		expectedBefore string
		expectedAfter  string
	}{
		{
			name:           "status is empty by default",
			initialStatus:  "",
			updatedStatus:  "400",
			expectedBefore: "",
			expectedAfter:  "400",
		},
		{
			name:           "status set then updated",
			initialStatus:  "200",
			updatedStatus:  "404",
			expectedBefore: "200",
			expectedAfter:  "404",
		},
		{
			name:           "status set then cleared",
			initialStatus:  "403",
			updatedStatus:  "",
			expectedBefore: "403",
			expectedAfter:  "",
		},
		{
			name:           "status set with http status text",
			initialStatus:  http.StatusText(http.StatusOK),
			updatedStatus:  http.StatusText(http.StatusNotFound),
			expectedBefore: "OK",
			expectedAfter:  "Not Found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			em := ErrorMessage{Status: tt.initialStatus}
			assert.Equal(t, tt.expectedBefore, em.Status)

			em.Status = tt.updatedStatus
			assert.Equal(t, tt.expectedAfter, em.Status)
		})
	}
}

// TestErrorMessage_MessageField validates getting and setting the Message field.
func TestErrorMessage_MessageField(t *testing.T) {
	tests := []struct {
		name            string
		initialMessage  string
		updatedMessage  string
		expectedBefore  string
		expectedAfter   string
	}{
		{
			name:           "message is empty by default",
			initialMessage: "",
			updatedMessage: "record not found",
			expectedBefore: "",
			expectedAfter:  "record not found",
		},
		{
			name:           "message set then updated",
			initialMessage: "bad request",
			updatedMessage: "unauthorized access",
			expectedBefore: "bad request",
			expectedAfter:  "unauthorized access",
		},
		{
			name:           "message set then cleared",
			initialMessage: "some error",
			updatedMessage: "",
			expectedBefore: "some error",
			expectedAfter:  "",
		},
		{
			name:           "message with special characters",
			initialMessage: "error: invalid JSON",
			updatedMessage: "error: <special> & chars",
			expectedBefore: "error: invalid JSON",
			expectedAfter:  "error: <special> & chars",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			em := ErrorMessage{Message: tt.initialMessage}
			assert.Equal(t, tt.expectedBefore, em.Message)

			em.Message = tt.updatedMessage
			assert.Equal(t, tt.expectedAfter, em.Message)
		})
	}
}

// TestErrorMessage_Equality validates value equality semantics.
func TestErrorMessage_Equality(t *testing.T) {
	tests := []struct {
		name     string
		a        ErrorMessage
		b        ErrorMessage
		expected bool
	}{
		{
			name:     "two instances with identical status and message are equal",
			a:        NewErrorMessage("404", "Not Found"),
			b:        NewErrorMessage("404", "Not Found"),
			expected: true,
		},
		{
			name:     "two instances with differing status are not equal",
			a:        NewErrorMessage("404", "Not Found"),
			b:        NewErrorMessage("500", "Not Found"),
			expected: false,
		},
		{
			name:     "two instances with differing message are not equal",
			a:        NewErrorMessage("404", "Not Found"),
			b:        NewErrorMessage("404", "Internal Server Error"),
			expected: false,
		},
		{
			name:     "two instances with both fields differing are not equal",
			a:        NewErrorMessage("404", "Not Found"),
			b:        NewErrorMessage("500", "Internal Server Error"),
			expected: false,
		},
		{
			name:     "two zero-value instances are equal",
			a:        ErrorMessage{},
			b:        ErrorMessage{},
			expected: true,
		},
		{
			name:     "instance compared to itself is equal (reflexive)",
			a:        NewErrorMessage("200", "OK"),
			b:        NewErrorMessage("200", "OK"),
			expected: true,
		},
		{
			name:     "instance with empty status vs non-empty",
			a:        NewErrorMessage("", "Not Found"),
			b:        NewErrorMessage("404", "Not Found"),
			expected: false,
		},
		{
			name:     "instance with empty message vs non-empty",
			a:        NewErrorMessage("404", ""),
			b:        NewErrorMessage("404", "Not Found"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expected {
				assert.Equal(t, tt.a, tt.b)
			} else {
				assert.NotEqual(t, tt.a, tt.b)
			}
		})
	}
}

// TestErrorMessage_Equality_Symmetry validates that equality is symmetric.
func TestErrorMessage_Equality_Symmetry(t *testing.T) {
	a := NewErrorMessage("404", "Not Found")
	b := NewErrorMessage("404", "Not Found")

	assert.Equal(t, a, b, "a == b should hold")
	assert.Equal(t, b, a, "b == a should hold (symmetry)")
}

// TestErrorMessage_Equality_Transitivity validates that equality is transitive.
func TestErrorMessage_Equality_Transitivity(t *testing.T) {
	a := NewErrorMessage("404", "Not Found")
	b := NewErrorMessage("404", "Not Found")
	c := NewErrorMessage("404", "Not Found")

	assert.Equal(t, a, b)
	assert.Equal(t, b, c)
	assert.Equal(t, a, c, "transitivity: a==b && b==c => a==c")
}

// TestErrorMessage_ToString validates human-readable string representation.
func TestErrorMessage_ToString(t *testing.T) {
	tests := []struct {
		name            string
		em              ErrorMessage
		containsStatus  string
		containsMessage string
	}{
		{
			name:            "toString contains status and message values",
			em:              NewErrorMessage("404", "Not Found"),
			containsStatus:  "404",
			containsMessage: "Not Found",
		},
		{
			name:            "toString with empty fields still produces a string",
			em:              ErrorMessage{},
			containsStatus:  "",
			containsMessage: "",
		},
		{
			name:            "toString with http status text",
			em:              NewErrorMessage("Internal Server Error", "an unexpected error occurred"),
			containsStatus:  "Internal Server Error",
			containsMessage: "an unexpected error occurred",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := fmt.Sprintf("%v", tt.em)
			assert.NotEmpty(t, s)

			if tt.containsStatus != "" {
				assert.Contains(t, s, tt.containsStatus)
			}
			if tt.containsMessage != "" {
				assert.Contains(t, s, tt.containsMessage)
			}
		})
	}
}

// TestErrorMessage_JSONSerialization validates that ErrorMessage serializes
// correctly to/from JSON (used by writeError HTTP handler pattern).
func TestErrorMessage_JSONSerialization(t *testing.T) {
	tests := []struct {
		name        string
		em          ErrorMessage
		expectedJSON string
	}{
		{
			name:         "serializes with both fields set",
			em:           NewErrorMessage("404", "Not Found"),
			expectedJSON: `{"status":"404","message":"Not Found"}`,
		},
		{
			name:         "serializes with empty fields",
			em:           ErrorMessage{},
			expectedJSON: `{"status":"","message":""}`,
		},
		{
			name:         "serializes with status text",
			em:           NewErrorMessage("Internal Server Error", "unexpected error"),
			expectedJSON: `{"status":"Internal Server Error","message":"unexpected error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.em)
			assert.NoError(t, err)
			assert.JSONEq(t, tt.expectedJSON, string(data))
		})
	}
}

// TestErrorMessage_JSONDeserialization validates that ErrorMessage deserializes
// correctly from JSON.
func TestErrorMessage_JSONDeserialization(t *testing.T) {
	tests := []struct {
		name            string
		jsonInput       string
		expectedStatus  string
		expectedMessage string
		expectError     bool
	}{
		{
			name:            "deserializes with both fields",
			jsonInput:       `{"status":"404","message":"Not Found"}`,
			expectedStatus:  "404",
			expectedMessage: "Not Found",
			expectError:     false,
		},
		{
			name:            "deserializes with empty fields",
			jsonInput:       `{"status":"","message":""}`,
			expectedStatus:  "",
			expectedMessage: "",
			expectError:     false,
		},
		{
			name:            "deserializes with missing fields defaults to empty",
			jsonInput:       `{}`,
			expectedStatus:  "",
			expectedMessage: "",
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var em ErrorMessage
			err := json.Unmarshal([]byte(tt.jsonInput), &em)
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, em.Status)
			assert.Equal(t, tt.expectedMessage, em.Message)
		})
	}
}

// TestErrorMessage_HTTPHandler validates ErrorMessage usage in an HTTP handler
// context using httptest, simulating the writeError pattern.
func TestErrorMessage_HTTPHandler(t *testing.T) {
	tests := []struct {
		name               string
		status             string
		message            string
		httpStatusCode     int
		expectedStatusJSON string
		expectedMsgJSON    string
	}{
		{
			name:               "handler writes 404 error message",
			status:             "404",
			message:            "resource not found",
			httpStatusCode:     http.StatusNotFound,
			expectedStatusJSON: "404",
			expectedMsgJSON:    "resource not found",
		},
		{
			name:               "handler writes 500 error message",
			status:             "500",
			message:            "internal server error",
			httpStatusCode:     http.StatusInternalServerError,
			expectedStatusJSON: "500",
			expectedMsgJSON:    "internal server error",
		},
		{
			name:               "handler writes 400 error message",
			status:             http.StatusText(http.StatusBadRequest),
			message:            "invalid request body",
			httpStatusCode:     http.StatusBadRequest,
			expectedStatusJSON: "Bad Request",
			expectedMsgJSON:    "invalid request body",
		},
		{
			name:               "handler writes 401 error message",
			status:             "401",
			message:            "unauthorized",
			httpStatusCode:     http.StatusUnauthorized,
			expectedStatusJSON: "401",
			expectedMsgJSON:    "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				em := NewErrorMessage(tt.status, tt.message)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.httpStatusCode)
				if err := json.NewEncoder(w).Encode(em); err != nil {
					t.Fatalf("failed to encode error message: %v", err)
				}
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.httpStatusCode, rr.Code)
			assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

			var decoded ErrorMessage
			err := json.NewDecoder(rr.Body).Decode(&decoded)
			assert.NoError(