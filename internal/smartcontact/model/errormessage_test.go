```go
package model

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// HTTPStatus.Name()
// ---------------------------------------------------------------------------

func TestHTTPStatus_Name(t *testing.T) {
	tests := []struct {
		name       string
		status     HTTPStatus
		wantName   string
		wantFound  bool
	}{
		{"OK", HTTPStatus(http.StatusOK), "OK", true},
		{"Created", HTTPStatus(http.StatusCreated), "CREATED", true},
		{"Accepted", HTTPStatus(http.StatusAccepted), "ACCEPTED", true},
		{"NoContent", HTTPStatus(http.StatusNoContent), "NO_CONTENT", true},
		{"BadRequest", HTTPStatus(http.StatusBadRequest), "BAD_REQUEST", true},
		{"Unauthorized", HTTPStatus(http.StatusUnauthorized), "UNAUTHORIZED", true},
		{"Forbidden", HTTPStatus(http.StatusForbidden), "FORBIDDEN", true},
		{"NotFound", HTTPStatus(http.StatusNotFound), "NOT_FOUND", true},
		{"MethodNotAllowed", HTTPStatus(http.StatusMethodNotAllowed), "METHOD_NOT_ALLOWED", true},
		{"Conflict", HTTPStatus(http.StatusConflict), "CONFLICT", true},
		{"UnprocessableEntity", HTTPStatus(http.StatusUnprocessableEntity), "UNPROCESSABLE_ENTITY", true},
		{"InternalServerError", HTTPStatus(http.StatusInternalServerError), "INTERNAL_SERVER_ERROR", true},
		{"NotImplemented", HTTPStatus(http.StatusNotImplemented), "NOT_IMPLEMENTED", true},
		{"BadGateway", HTTPStatus(http.StatusBadGateway), "BAD_GATEWAY", true},
		{"ServiceUnavailable", HTTPStatus(http.StatusServiceUnavailable), "SERVICE_UNAVAILABLE", true},
		{"Unknown", HTTPStatus(999), "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotName, gotFound := tc.status.Name()
			assert.Equal(t, tc.wantName, gotName)
			assert.Equal(t, tc.wantFound, gotFound)
		})
	}
}

// ---------------------------------------------------------------------------
// HTTPStatus.Code()
// ---------------------------------------------------------------------------

func TestHTTPStatus_Code(t *testing.T) {
	tests := []struct {
		name   string
		status HTTPStatus
		want   int
	}{
		{"OK", HTTPStatus(http.StatusOK), 200},
		{"NotFound", HTTPStatus(http.StatusNotFound), 404},
		{"InternalServerError", HTTPStatus(http.StatusInternalServerError), 500},
		{"Custom", HTTPStatus(418), 418},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.status.Code())
		})
	}
}

// ---------------------------------------------------------------------------
// HTTPStatus.MarshalJSON()
// ---------------------------------------------------------------------------

func TestHTTPStatus_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		status   HTTPStatus
		wantJSON string
	}{
		{"OK", HTTPStatus(http.StatusOK), `"OK"`},
		{"NotFound", HTTPStatus(http.StatusNotFound), `"NOT_FOUND"`},
		{"BadRequest", HTTPStatus(http.StatusBadRequest), `"BAD_REQUEST"`},
		{"InternalServerError", HTTPStatus(http.StatusInternalServerError), `"INTERNAL_SERVER_ERROR"`},
		{"ServiceUnavailable", HTTPStatus(http.StatusServiceUnavailable), `"SERVICE_UNAVAILABLE"`},
		// Unknown code: fallback to http.StatusText
		{"UnknownFallback", HTTPStatus(418), `"I'm a teapot"`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.status.MarshalJSON()
			assert.NoError(t, err)
			assert.Equal(t, tc.wantJSON, string(got))
		})
	}
}

// ---------------------------------------------------------------------------
// NewErrorMessage (all-args constructor)
// ---------------------------------------------------------------------------

func TestNewErrorMessage(t *testing.T) {
	tests := []struct {
		name          string
		status        HTTPStatus
		message       string
		wantStatus    HTTPStatus
		wantMessage   string
	}{
		{
			name:        "with status and message",
			status:      HTTPStatus(http.StatusNotFound),
			message:     "resource not found",
			wantStatus:  HTTPStatus(http.StatusNotFound),
			wantMessage: "resource not found",
		},
		{
			name:        "with bad request status",
			status:      HTTPStatus(http.StatusBadRequest),
			message:     "invalid input",
			wantStatus:  HTTPStatus(http.StatusBadRequest),
			wantMessage: "invalid input",
		},
		{
			name:        "with empty message",
			status:      HTTPStatus(http.StatusInternalServerError),
			message:     "",
			wantStatus:  HTTPStatus(http.StatusInternalServerError),
			wantMessage: "",
		},
		{
			name:        "zero status (default/null equivalent)",
			status:      HTTPStatus(0),
			message:     "",
			wantStatus:  HTTPStatus(0),
			wantMessage: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			em := NewErrorMessage(tc.status, tc.message)
			assert.Equal(t, tc.wantStatus, em.Status)
			assert.Equal(t, tc.wantMessage, em.Message)
		})
	}
}

// ---------------------------------------------------------------------------
// Zero-value ErrorMessage (no-args constructor equivalent)
// ---------------------------------------------------------------------------

func TestErrorMessage_ZeroValue(t *testing.T) {
	t.Run("constructed with no arguments (zero value)", func(t *testing.T) {
		var em ErrorMessage
		// In Go, zero value for HTTPStatus (int) is 0, string is ""
		// This mirrors the Java no-args constructor where fields default to null.
		assert.Equal(t, HTTPStatus(0), em.Status)
		assert.Equal(t, "", em.Message)
	})
}

// ---------------------------------------------------------------------------
// Field get/set (direct struct field access mirrors Java getters/setters)
// ---------------------------------------------------------------------------

func TestErrorMessage_StatusField(t *testing.T) {
	tests := []struct {
		name       string
		initial    HTTPStatus
		newStatus  HTTPStatus
		wantAfter  HTTPStatus
	}{
		{
			name:      "set valid status",
			initial:   HTTPStatus(http.StatusOK),
			newStatus: HTTPStatus(http.StatusNotFound),
			wantAfter: HTTPStatus(http.StatusNotFound),
		},
		{
			name:      "set to zero (null equivalent)",
			initial:   HTTPStatus(http.StatusBadRequest),
			newStatus: HTTPStatus(0),
			wantAfter: HTTPStatus(0),
		},
		{
			name:      "status never set (zero default)",
			initial:   HTTPStatus(0),
			newStatus: HTTPStatus(0),
			wantAfter: HTTPStatus(0),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			em := NewErrorMessage(tc.initial, "")
			// Simulate setStatus
			em.Status = tc.newStatus
			// Simulate getStatus
			assert.Equal(t, tc.wantAfter, em.Status)
		})
	}
}

func TestErrorMessage_MessageField(t *testing.T) {
	tests := []struct {
		name       string
		initial    string
		newMsg     string
		wantAfter  string
	}{
		{
			name:      "set valid message",
			initial:   "old message",
			newMsg:    "new message",
			wantAfter: "new message",
		},
		{
			name:      "set to empty (null equivalent)",
			initial:   "some message",
			newMsg:    "",
			wantAfter: "",
		},
		{
			name:      "message never set (zero default)",
			initial:   "",
			newMsg:    "",
			wantAfter: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			em := NewErrorMessage(HTTPStatus(http.StatusOK), tc.initial)
			// Simulate setMessage
			em.Message = tc.newMsg
			// Simulate getMessage
			assert.Equal(t, tc.wantAfter, em.Message)
		})
	}
}

// ---------------------------------------------------------------------------
// Equality (mirrors Java equals())
// ---------------------------------------------------------------------------

func TestErrorMessage_Equality(t *testing.T) {
	statusA := HTTPStatus(http.StatusNotFound)
	statusB := HTTPStatus(http.StatusBadRequest)

	tests := []struct {
		name      string
		em1       ErrorMessage
		em2       ErrorMessage
		wantEqual bool
	}{
		{
			name:      "equal status and message",
			em1:       NewErrorMessage(statusA, "not found"),
			em2:       NewErrorMessage(statusA, "not found"),
			wantEqual: true,
		},
		{
			name:      "different status same message",
			em1:       NewErrorMessage(statusA, "error"),
			em2:       NewErrorMessage(statusB, "error"),
			wantEqual: false,
		},
		{
			name:      "same status different message",
			em1:       NewErrorMessage(statusA, "msg one"),
			em2:       NewErrorMessage(statusA, "msg two"),
			wantEqual: false,
		},
		{
			name:      "both zero value",
			em1:       ErrorMessage{},
			em2:       ErrorMessage{},
			wantEqual: true,
		},
		{
			name:      "reflexive: same instance values",
			em1:       NewErrorMessage(statusA, "hello"),
			em2:       NewErrorMessage(statusA, "hello"),
			wantEqual: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.wantEqual {
				assert.Equal(t, tc.em1, tc.em2)
			} else {
				assert.NotEqual(t, tc.em1, tc.em2)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Hash code consistency (mirrors Java hashCode())
// ---------------------------------------------------------------------------

// In Go, struct comparison is value-based; we verify that equal structs
// produce consistent behaviour (this is analogous to hashCode contract).
func TestErrorMessage_HashCodeConsistency(t *testing.T) {
	tests := []struct {
		name string
		em1  ErrorMessage
		em2  ErrorMessage
	}{
		{
			name: "equal instances have equal 'hash' (struct equality)",
			em1:  NewErrorMessage(HTTPStatus(http.StatusNotFound), "not found"),
			em2:  NewErrorMessage(HTTPStatus(http.StatusNotFound), "not found"),
		},
		{
			name: "zero values are consistently equal",
			em1:  ErrorMessage{},
			em2:  ErrorMessage{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Equal structs in Go are by definition identical in memory layout,
			// which is the Go analog of hashCode contract.
			assert.Equal(t, tc.em1, tc.em2, "equal ErrorMessages must be structurally identical")
		})
	}
}

// ---------------------------------------------------------------------------
// toString (JSON serialization as human-readable representation)
// ---------------------------------------------------------------------------

func TestErrorMessage_JSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		em       ErrorMessage
		wantJSON string
	}{
		{
			name:     "known status serializes to enum name",
			em:       NewErrorMessage(HTTPStatus(http.StatusNotFound), "resource not found"),
			wantJSON: `{"status":"NOT_FOUND","message":"resource not found"}`,
		},
		{
			name:     "bad request",
			em:       NewErrorMessage(HTTPStatus(http.StatusBadRequest), "invalid input"),
			wantJSON: `{"status":"BAD_REQUEST","message":"invalid input"}`,
		},
		{
			name:     "internal server error",
			em:       NewErrorMessage(HTTPStatus(http.StatusInternalServerError), "unexpected error"),
			wantJSON: `{"status":"INTERNAL_SERVER_ERROR","message":"unexpected error"}`,
		},
		{
			name:     "empty message",
			em:       NewErrorMessage(HTTPStatus(http.StatusOK), ""),
			wantJSON: `{"status":"OK","message":""}`,
		},
		{
			name:     "unknown status falls back to StatusText",
			em:       NewErrorMessage(HTTPStatus(418), "teapot"),
			wantJSON: `{"status":"I'm a teapot","message":"teapot"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := json.Marshal(tc.em)
			assert.NoError(t, err)
			assert.JSONEq(t, tc.wantJSON, string(got))
		})
	}
}

// ---------------------------------------------------------------------------
// toString: non-null, reflects current field values
// ---------------------------------------------------------------------------

func TestErrorMessage_StringRepresentation(t *testing.T) {
	tests := []struct {
		name        string
		em          ErrorMessage
		shouldContain []string
	}{
		{
			name:          "JSON contains status and message",
			em:            NewErrorMessage(HTTPStatus(http.StatusNotFound), "not found"),
			shouldContain: []string{"NOT_FOUND", "not found"},
		},
		{
			name:          "zero value produces non-empty JSON",
			em:            ErrorMessage{},
			shouldContain: []string{"status", "message"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := json.Marshal(tc.em)
			assert.NoError(t, err)
			assert.NotNil(t, got)
			assert.NotEmpty(t, string(got))
			for _, sub := range tc.shouldContain {
				assert.Contains(t, string(got), sub)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Round-trip JSON decode
// ---------------------------------------------------------------------------

func TestErrorMessage_JSONRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		em   ErrorMessage
	}{
		{
			name: "not found",
			em:   NewErrorMessage(HTTPStatus(http.StatusNotFound), "resource not found"),
		},
		{
			name: "bad request",
			em:   NewErrorMessage(HTTPStatus(http.StatusBadRequest), "bad input"),
		},
		{
			name: "empty message",
			em:   NewErrorMessage(HTTPStatus(http.StatusOK), ""),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := json.Marshal(tc.em)
			assert.NoError(t, err)

			// Decode into a map to verify the string value of status
			var decoded map[string]interface{}
			assert.NoError(t, json.Unmarshal(raw, &decoded))

			// The status field must be the Spring enum name string, not a number
			statusVal, ok := decoded["status"].(string)
			assert.True(t, ok, "status field must be a JSON string")
			assert.NotEmpty(t, statusVal)

			// message must match
			msg