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

// ---------------------------------------------------------------------------
// TestErrorMessage_ZeroValue – "no-args constructor" equivalent
// ---------------------------------------------------------------------------

func TestErrorMessage_ZeroValue(t *testing.T) {
	tests := []struct {
		name            string
		wantStatus      int
		wantMessage     string
	}{
		{
			name:        "instantiated with no arguments gives zero values",
			wantStatus:  0,
			wantMessage: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var em ErrorMessage
			assert.Equal(t, tc.wantStatus, em.Status)
			assert.Equal(t, tc.wantMessage, em.Message)
		})
	}
}

// ---------------------------------------------------------------------------
// TestNewErrorMessage – "all-args constructor" equivalent
// ---------------------------------------------------------------------------

func TestNewErrorMessage(t *testing.T) {
	tests := []struct {
		name        string
		status      int
		message     string
		wantStatus  int
		wantMessage string
	}{
		{
			name:        "valid status and message",
			status:      http.StatusNotFound,
			message:     "resource not found",
			wantStatus:  http.StatusNotFound,
			wantMessage: "resource not found",
		},
		{
			name:        "internal server error status",
			status:      http.StatusInternalServerError,
			message:     "internal error",
			wantStatus:  http.StatusInternalServerError,
			wantMessage: "internal error",
		},
		{
			name:        "zero status and empty message (null-equivalent)",
			status:      0,
			message:     "",
			wantStatus:  0,
			wantMessage: "",
		},
		{
			name:        "status ok with message",
			status:      http.StatusOK,
			message:     "ok",
			wantStatus:  http.StatusOK,
			wantMessage: "ok",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			em := NewErrorMessage(tc.status, tc.message)
			assert.NotNil(t, em)
			assert.Equal(t, tc.wantStatus, em.Status)
			assert.Equal(t, tc.wantMessage, em.Message)
		})
	}
}

// ---------------------------------------------------------------------------
// TestErrorMessage_GetStatus – getStatus equivalent
// ---------------------------------------------------------------------------

func TestErrorMessage_GetStatus(t *testing.T) {
	tests := []struct {
		name       string
		setup      func() *ErrorMessage
		wantStatus int
	}{
		{
			name: "status was previously set via constructor",
			setup: func() *ErrorMessage {
				return NewErrorMessage(http.StatusBadRequest, "bad request")
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "status was never set (zero value)",
			setup: func() *ErrorMessage {
				em := &ErrorMessage{}
				return em
			},
			wantStatus: 0,
		},
		{
			name: "status set to 404",
			setup: func() *ErrorMessage {
				return NewErrorMessage(http.StatusNotFound, "not found")
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			em := tc.setup()
			// Reading Status must not modify the object.
			originalMessage := em.Message
			got := em.Status
			assert.Equal(t, tc.wantStatus, got)
			assert.Equal(t, originalMessage, em.Message, "getMessage must not change after getStatus")
		})
	}
}

// ---------------------------------------------------------------------------
// TestErrorMessage_SetStatus – setStatus equivalent
// ---------------------------------------------------------------------------

func TestErrorMessage_SetStatus(t *testing.T) {
	tests := []struct {
		name           string
		initial        int
		setTo          int
		wantStatus     int
		initialMessage string
		wantMessage    string
	}{
		{
			name:           "set valid HTTP status",
			initial:        0,
			setTo:          http.StatusForbidden,
			wantStatus:     http.StatusForbidden,
			initialMessage: "forbidden",
			wantMessage:    "forbidden",
		},
		{
			name:           "overwrite existing status",
			initial:        http.StatusOK,
			setTo:          http.StatusConflict,
			wantStatus:     http.StatusConflict,
			initialMessage: "conflict",
			wantMessage:    "conflict",
		},
		{
			name:           "set to zero (null-equivalent)",
			initial:        http.StatusNotFound,
			setTo:          0,
			wantStatus:     0,
			initialMessage: "msg",
			wantMessage:    "msg",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			em := NewErrorMessage(tc.initial, tc.initialMessage)
			em.Status = tc.setTo

			assert.Equal(t, tc.wantStatus, em.Status)
			// Only status must change; message must stay the same.
			assert.Equal(t, tc.wantMessage, em.Message)
		})
	}
}

// ---------------------------------------------------------------------------
// TestErrorMessage_GetMessage – getMessage equivalent
// ---------------------------------------------------------------------------

func TestErrorMessage_GetMessage(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *ErrorMessage
		wantMessage string
	}{
		{
			name: "message was previously set via constructor",
			setup: func() *ErrorMessage {
				return NewErrorMessage(http.StatusUnauthorized, "unauthorized")
			},
			wantMessage: "unauthorized",
		},
		{
			name: "message was never set (zero value)",
			setup: func() *ErrorMessage {
				return &ErrorMessage{}
			},
			wantMessage: "",
		},
		{
			name: "message set to empty string",
			setup: func() *ErrorMessage {
				return NewErrorMessage(http.StatusBadRequest, "")
			},
			wantMessage: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			em := tc.setup()
			// Reading Message must not modify the object.
			originalStatus := em.Status
			got := em.Message
			assert.Equal(t, tc.wantMessage, got)
			assert.Equal(t, originalStatus, em.Status, "getStatus must not change after getMessage")
		})
	}
}

// ---------------------------------------------------------------------------
// TestErrorMessage_SetMessage – setMessage equivalent
// ---------------------------------------------------------------------------

func TestErrorMessage_SetMessage(t *testing.T) {
	tests := []struct {
		name           string
		initialStatus  int
		initialMessage string
		setMessageTo   string
		wantMessage    string
		wantStatus     int
	}{
		{
			name:           "set a string value",
			initialStatus:  http.StatusOK,
			initialMessage: "old",
			setMessageTo:   "new message",
			wantMessage:    "new message",
			wantStatus:     http.StatusOK,
		},
		{
			name:           "set to empty string (null-equivalent)",
			initialStatus:  http.StatusNotFound,
			initialMessage: "something",
			setMessageTo:   "",
			wantMessage:    "",
			wantStatus:     http.StatusNotFound,
		},
		{
			name:           "overwrite message",
			initialStatus:  http.StatusBadGateway,
			initialMessage: "first",
			setMessageTo:   "second",
			wantMessage:    "second",
			wantStatus:     http.StatusBadGateway,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			em := NewErrorMessage(tc.initialStatus, tc.initialMessage)
			em.Message = tc.setMessageTo

			assert.Equal(t, tc.wantMessage, em.Message)
			// Only message must change; status must stay the same.
			assert.Equal(t, tc.wantStatus, em.Status)
		})
	}
}

// ---------------------------------------------------------------------------
// TestErrorMessage_Equals – equals equivalent
// ---------------------------------------------------------------------------

func TestErrorMessage_Equals(t *testing.T) {
	tests := []struct {
		name    string
		a       *ErrorMessage
		b       *ErrorMessage
		wantEq  bool
	}{
		{
			name:   "equal status and message",
			a:      NewErrorMessage(http.StatusNotFound, "not found"),
			b:      NewErrorMessage(http.StatusNotFound, "not found"),
			wantEq: true,
		},
		{
			name:   "different status same message",
			a:      NewErrorMessage(http.StatusNotFound, "error"),
			b:      NewErrorMessage(http.StatusBadRequest, "error"),
			wantEq: false,
		},
		{
			name:   "same status different message",
			a:      NewErrorMessage(http.StatusOK, "all good"),
			b:      NewErrorMessage(http.StatusOK, "different"),
			wantEq: false,
		},
		{
			name:   "both zero values",
			a:      &ErrorMessage{},
			b:      &ErrorMessage{},
			wantEq: true,
		},
		{
			name:   "one zero one non-zero",
			a:      &ErrorMessage{},
			b:      NewErrorMessage(http.StatusInternalServerError, "error"),
			wantEq: false,
		},
		{
			name:   "reflexive: object equals itself",
			a:      NewErrorMessage(http.StatusCreated, "created"),
			b:      NewErrorMessage(http.StatusCreated, "created"),
			wantEq: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			eq := *tc.a == *tc.b
			assert.Equal(t, tc.wantEq, eq)

			// Symmetry: b == a must return the same result.
			eqSym := *tc.b == *tc.a
			assert.Equal(t, tc.wantEq, eqSym)
		})
	}
}

// ---------------------------------------------------------------------------
// TestErrorMessage_HashCode – hashCode equivalent (via map usage)
// ---------------------------------------------------------------------------

func TestErrorMessage_HashCode(t *testing.T) {
	tests := []struct {
		name string
		a    ErrorMessage
		b    ErrorMessage
	}{
		{
			name: "equal objects produce equal hash codes",
			a:    ErrorMessage{Status: http.StatusNotFound, Message: "not found"},
			b:    ErrorMessage{Status: http.StatusNotFound, Message: "not found"},
		},
		{
			name: "equal zero-value objects produce equal hash codes",
			a:    ErrorMessage{},
			b:    ErrorMessage{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// In Go, structs that are comparable can be used as map keys.
			// If a == b then map[a] and map[b] refer to the same slot.
			m := map[ErrorMessage]int{}
			m[tc.a] = 42
			assert.Equal(t, 42, m[tc.b], "equal ErrorMessage keys must hash to the same bucket")
		})
	}
}

// ---------------------------------------------------------------------------
// TestErrorMessage_ToString – toString equivalent
// ---------------------------------------------------------------------------

func TestErrorMessage_ToString(t *testing.T) {
	tests := []struct {
		name            string
		em              *ErrorMessage
		wantContains    []string
	}{
		{
			name:         "contains status and message values",
			em:           NewErrorMessage(http.StatusNotFound, "resource not found"),
			wantContains: []string{"404", "resource not found"},
		},
		{
			name:         "zero values are represented",
			em:           &ErrorMessage{},
			wantContains: []string{"0", ""},
		},
		{
			name:         "internal server error is represented",
			em:           NewErrorMessage(http.StatusInternalServerError, "internal error"),
			wantContains: []string{"500", "internal error"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			str := fmt.Sprintf("%+v", tc.em)
			for _, want := range tc.wantContains {
				assert.Contains(t, str, want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestErrorMessage_StatusText – StatusText helper
// ---------------------------------------------------------------------------

func TestErrorMessage_StatusText(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		wantText string
	}{
		{
			name:     "404 returns Not Found",
			status:   http.StatusNotFound,
			wantText: "Not Found",
		},
		{
			name:     "200 returns OK",
			status:   http.StatusOK,
			wantText: "OK",
		},
		{
			name:     "500 returns Internal Server Error",
			status:   http.StatusInternalServerError,
			wantText: "Internal Server Error",
		},
		{
			name:     "unknown status returns empty string",
			status:   9999,
			wantText: "",
		},
		{
			name:     "zero status returns empty string",
			status:   0,
			wantText: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			em := NewErrorMessage(tc.status, "some message")
			assert.Equal(t, tc.wantText, em.StatusText())
		})
	}
}

// ---------------------------------------------------------------------------
// TestErrorMessage_JSONSerialization – validate JSON round-trip
// ---------------------------------------------------------------------------

func TestErrorMessage_JSONSerialization(t *testing.T) {
	tests := []struct {
		name    string
		em      *ErrorMessage
		wantJSON string
	}{
		{
			name:     "404 not found",
			em:       NewErrorMessage(http.StatusNotFound, "not found"),
			wantJSON: `{"status":404,"message":"not found"}`,
		},
		{
			name:     "zero value",
			em:       &ErrorMessage{},
			wantJSON: `{"status":0,"message":""}`,
		},
		{
			name:     "500 internal server error",
			em:       NewErrorMessage(http.StatusInternalServerError, "server error"),
			wantJSON: `{"status":500,"message":"server error"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Marshal
			data, err := json.Marshal(tc.em)
			assert.NoError(t, err)
			assert.JSONEq(t, tc.wantJSON, string(data))

			// Unmarshal round-trip
			var got ErrorMessage
			err = json.Unmarshal(data, &got)
			assert.NoError(t, err)
			assert.Equal(t, *tc.em, got)
		})
	}
}

// ---------------------------------------------------------------------------
// TestErrorMessage_HTTPHandler – httptest integration
// ---------------------------------------------------------------------------

func TestErrorMessage_HTTPHandler(t *testing.T) {
	tests := []struct {
		name           string
		status         int
		message        string
		wantHTTPStatus int
		wantBodyStatus int
		wantMessage    string
	}{
		{
			name