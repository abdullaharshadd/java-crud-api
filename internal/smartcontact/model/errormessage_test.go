```go
package model_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"internal/smartcontact/model"
)

// ---------------------------------------------------------------------------
// Zero-value / "no-args constructor" equivalent
// ---------------------------------------------------------------------------

func TestErrorMessage_ZeroValue(t *testing.T) {
	tests := []struct {
		name            string
		wantStatus      int
		wantMessage     string
	}{
		{
			name:        "zero value has default int (0) and empty string",
			wantStatus:  0,
			wantMessage: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var em model.ErrorMessage

			assert.NotNil(t, &em, "zero-value ErrorMessage should be addressable / non-nil pointer")
			assert.Equal(t, tt.wantStatus, em.Status)
			assert.Equal(t, tt.wantMessage, em.Message)
		})
	}
}

// ---------------------------------------------------------------------------
// NewErrorMessage (all-args constructor equivalent)
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
			message:     "unexpected error",
			wantStatus:  http.StatusInternalServerError,
			wantMessage: "unexpected error",
		},
		{
			name:        "zero status (Go equivalent of null status)",
			status:      0,
			message:     "some message",
			wantStatus:  0,
			wantMessage: "some message",
		},
		{
			name:        "empty message (Go equivalent of null message)",
			status:      http.StatusBadRequest,
			message:     "",
			wantStatus:  http.StatusBadRequest,
			wantMessage: "",
		},
		{
			name:        "both zero / empty (Go equivalent of all-null)",
			status:      0,
			message:     "",
			wantStatus:  0,
			wantMessage: "",
		},
		{
			name:        "no transformation applied to values",
			status:      999,
			message:     "  spaces preserved  ",
			wantStatus:  999,
			wantMessage: "  spaces preserved  ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			em := model.NewErrorMessage(tt.status, tt.message)

			require.NotNil(t, em, "NewErrorMessage must never return nil")
			assert.Equal(t, tt.wantStatus, em.Status)
			assert.Equal(t, tt.wantMessage, em.Message)
		})
	}
}

// ---------------------------------------------------------------------------
// Field mutation (setStatus / setMessage equivalents)
// ---------------------------------------------------------------------------

func TestErrorMessage_FieldMutation(t *testing.T) {
	tests := []struct {
		name           string
		initialStatus  int
		initialMessage string
		newStatus      int
		newMessage     string
	}{
		{
			name:           "mutate status field",
			initialStatus:  http.StatusOK,
			initialMessage: "ok",
			newStatus:      http.StatusForbidden,
			newMessage:     "ok",
		},
		{
			name:           "mutate message field",
			initialStatus:  http.StatusOK,
			initialMessage: "ok",
			newStatus:      http.StatusOK,
			newMessage:     "forbidden",
		},
		{
			name:           "reset status to zero (null equivalent)",
			initialStatus:  http.StatusTeapot,
			initialMessage: "teapot",
			newStatus:      0,
			newMessage:     "teapot",
		},
		{
			name:           "reset message to empty (null equivalent)",
			initialStatus:  http.StatusTeapot,
			initialMessage: "teapot",
			newStatus:      http.StatusTeapot,
			newMessage:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			em := model.NewErrorMessage(tt.initialStatus, tt.initialMessage)
			require.NotNil(t, em)

			// Simulate setStatus / setMessage via direct field access
			em.Status = tt.newStatus
			em.Message = tt.newMessage

			assert.Equal(t, tt.newStatus, em.Status, "Status should reflect the mutated value")
			assert.Equal(t, tt.newMessage, em.Message, "Message should reflect the mutated value")
		})
	}
}

// ---------------------------------------------------------------------------
// Equality (equals / hashCode equivalents)
// ---------------------------------------------------------------------------

func TestErrorMessage_Equality(t *testing.T) {
	tests := []struct {
		name      string
		a         *model.ErrorMessage
		b         *model.ErrorMessage
		wantEqual bool
	}{
		{
			name:      "identical status and message",
			a:         model.NewErrorMessage(http.StatusNotFound, "not found"),
			b:         model.NewErrorMessage(http.StatusNotFound, "not found"),
			wantEqual: true,
		},
		{
			name:      "different status same message",
			a:         model.NewErrorMessage(http.StatusNotFound, "error"),
			b:         model.NewErrorMessage(http.StatusInternalServerError, "error"),
			wantEqual: false,
		},
		{
			name:      "same status different message",
			a:         model.NewErrorMessage(http.StatusBadRequest, "bad"),
			b:         model.NewErrorMessage(http.StatusBadRequest, "request"),
			wantEqual: false,
		},
		{
			name:      "both zero-value",
			a:         model.NewErrorMessage(0, ""),
			b:         model.NewErrorMessage(0, ""),
			wantEqual: true,
		},
		{
			name:      "reflexive: same pointer",
			a:         model.NewErrorMessage(http.StatusOK, "ok"),
			b:         nil, // will be replaced below
			wantEqual: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Handle the reflexive case
			b := tt.b
			if tt.name == "reflexive: same pointer" {
				b = tt.a
			}

			if tt.wantEqual {
				assert.Equal(t, tt.a.Status, b.Status)
				assert.Equal(t, tt.a.Message, b.Message)
				// Struct-level equality
				assert.Equal(t, *tt.a, *b)
			} else {
				assert.NotEqual(t, *tt.a, *b)
			}
		})
	}
}

func TestErrorMessage_EqualObjectsHaveSameRepresentation(t *testing.T) {
	// Hash-code equivalent: two equal ErrorMessage instances must produce
	// the same JSON (and fmt.Sprintf) representation.
	tests := []struct {
		name string
		a    *model.ErrorMessage
		b    *model.ErrorMessage
	}{
		{
			name: "equal instances produce identical JSON",
			a:    model.NewErrorMessage(http.StatusConflict, "conflict"),
			b:    model.NewErrorMessage(http.StatusConflict, "conflict"),
		},
		{
			name: "equal zero-value instances",
			a:    model.NewErrorMessage(0, ""),
			b:    model.NewErrorMessage(0, ""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonA, err := json.Marshal(tt.a)
			require.NoError(t, err)
			jsonB, err := json.Marshal(tt.b)
			require.NoError(t, err)

			assert.Equal(t, string(jsonA), string(jsonB), "equal ErrorMessages must produce the same JSON")
			assert.Equal(t, fmt.Sprintf("%+v", tt.a), fmt.Sprintf("%+v", tt.b),
				"equal ErrorMessages must produce the same fmt representation")
		})
	}
}

// ---------------------------------------------------------------------------
// toString equivalent: string representation
// ---------------------------------------------------------------------------

func TestErrorMessage_StringRepresentation(t *testing.T) {
	tests := []struct {
		name        string
		status      int
		message     string
		containsAll []string
	}{
		{
			name:        "representation contains status and message",
			status:      http.StatusNotFound,
			message:     "resource not found",
			containsAll: []string{"404", "resource not found"},
		},
		{
			name:        "representation for zero values",
			status:      0,
			message:     "",
			containsAll: []string{"0"},
		},
		{
			name:        "internal server error",
			status:      http.StatusInternalServerError,
			message:     "server blew up",
			containsAll: []string{"500", "server blew up"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			em := model.NewErrorMessage(tt.status, tt.message)
			repr := fmt.Sprintf("%+v", em)

			assert.NotEmpty(t, repr, "string representation must never be empty")
			for _, want := range tt.containsAll {
				assert.Contains(t, repr, want,
					"string representation should contain %q", want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// JSON serialization / deserialization
// ---------------------------------------------------------------------------

func TestErrorMessage_JSONRoundTrip(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		message string
	}{
		{
			name:    "404 not found",
			status:  http.StatusNotFound,
			message: "resource not found",
		},
		{
			name:    "500 internal server error",
			status:  http.StatusInternalServerError,
			message: "something went wrong",
		},
		{
			name:    "zero values",
			status:  0,
			message: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := model.NewErrorMessage(tt.status, tt.message)

			data, err := json.Marshal(original)
			require.NoError(t, err, "marshalling must not fail")

			var decoded model.ErrorMessage
			err = json.Unmarshal(data, &decoded)
			require.NoError(t, err, "unmarshalling must not fail")

			assert.Equal(t, original.Status, decoded.Status)
			assert.Equal(t, original.Message, decoded.Message)
		})
	}
}

func TestErrorMessage_JSONFieldNames(t *testing.T) {
	em := model.NewErrorMessage(http.StatusBadRequest, "bad request")

	data, err := json.Marshal(em)
	require.NoError(t, err)

	var raw map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &raw))

	assert.Contains(t, raw, "status", "JSON must contain 'status' key")
	assert.Contains(t, raw, "message", "JSON must contain 'message' key")
	assert.EqualValues(t, http.StatusBadRequest, raw["status"])
	assert.Equal(t, "bad request", raw["message"])
}

// ---------------------------------------------------------------------------
// HTTP handler integration via httptest
// ---------------------------------------------------------------------------

func TestErrorMessage_HTTPHandlerIntegration(t *testing.T) {
	tests := []struct {
		name           string
		status         int
		message        string
		wantHTTPStatus int
		wantBody       model.ErrorMessage
	}{
		{
			name:           "handler returns 404 error message",
			status:         http.StatusNotFound,
			message:        "resource not found",
			wantHTTPStatus: http.StatusNotFound,
			wantBody:       model.ErrorMessage{Status: http.StatusNotFound, Message: "resource not found"},
		},
		{
			name:           "handler returns 500 error message",
			status:         http.StatusInternalServerError,
			message:        "internal server error",
			wantHTTPStatus: http.StatusInternalServerError,
			wantBody:       model.ErrorMessage{Status: http.StatusInternalServerError, Message: "internal server error"},
		},
		{
			name:           "handler returns 400 error message",
			status:         http.StatusBadRequest,
			message:        "invalid input",
			wantHTTPStatus: http.StatusBadRequest,
			wantBody:       model.ErrorMessage{Status: http.StatusBadRequest, Message: "invalid input"},
		},
		{
			name:           "handler returns 401 error message",
			status:         http.StatusUnauthorized,
			message:        "unauthorized",
			wantHTTPStatus: http.StatusUnauthorized,
			wantBody:       model.ErrorMessage{Status: http.StatusUnauthorized, Message: "unauthorized"},
		},
		{
			name:           "handler returns 403 error message",
			status:         http.StatusForbidden,
			message:        "forbidden",
			wantHTTPStatus: http.StatusForbidden,
			wantBody:       model.ErrorMessage{Status: http.StatusForbidden, Message: "forbidden"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				em := model.NewErrorMessage(tt.status, tt.message)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.status)
				if err := json.NewEncoder(w).Encode(em); err != nil {
					t.Errorf("failed to encode ErrorMessage: %v", err)
				}
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantHTTPStatus, rec.Code, "HTTP status code mismatch")
			assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

			var got model.ErrorMessage
			require.NoError(t, json.NewDecoder(rec.Body).Decode(&got))

			assert.Equal(t, tt.wantBody.Status, got.Status)
			assert.Equal(t, tt.wantBody.Message, got.Message)
		})
	}
}

// ---------------------------------------------------------------------------
// Invariant: no business logic / no transformation
// ---------------------------------------------------------------------------

func TestErrorMessage_NoTransformationApplied(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		message string
	}{
		{
			name:    "whitespace is preserved verbatim",
			status:  http.StatusOK,
			message: "  leading and trailing  ",
		},
		{
			name:    "special characters preserved",
			status:  http.StatusOK,
			message: "error: <nil> & 'quote' \"double\"",
		},
		{
			name:    "unicode preserved",
			status:  http.StatusOK,
			message: "エラー: リ