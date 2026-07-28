```go
// Package model_test contains table-driven tests for the ErrorMessage type.
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
// Zero-value / "no-args constructor" behaviour
// ---------------------------------------------------------------------------

func TestErrorMessage_ZeroValue(t *testing.T) {
	tests := []struct {
		name            string
		wantStatus      int
		wantMessage     string
	}{
		{
			name:        "default int is 0 and default string is empty",
			wantStatus:  0,
			wantMessage: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var em model.ErrorMessage
			assert.Equal(t, tc.wantStatus, em.Status)
			assert.Equal(t, tc.wantMessage, em.Message)
		})
	}
}

// ---------------------------------------------------------------------------
// NewErrorMessage (all-args constructor) behaviour
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
			status:      http.StatusBadRequest,
			message:     "bad request",
			wantStatus:  http.StatusBadRequest,
			wantMessage: "bad request",
		},
		{
			name:        "internal server error status",
			status:      http.StatusInternalServerError,
			message:     "something went wrong",
			wantStatus:  http.StatusInternalServerError,
			wantMessage: "something went wrong",
		},
		{
			name:        "zero status and empty message (null-equivalent)",
			status:      0,
			message:     "",
			wantStatus:  0,
			wantMessage: "",
		},
		{
			name:        "status only, empty message",
			status:      http.StatusNotFound,
			message:     "",
			wantStatus:  http.StatusNotFound,
			wantMessage: "",
		},
		{
			name:        "zero status, non-empty message",
			status:      0,
			message:     "some error",
			wantStatus:  0,
			wantMessage: "some error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			em := model.NewErrorMessage(tc.status, tc.message)
			assert.Equal(t, tc.wantStatus, em.Status)
			assert.Equal(t, tc.wantMessage, em.Message)
		})
	}
}

// ---------------------------------------------------------------------------
// Field getter semantics (Status field)
// ---------------------------------------------------------------------------

func TestErrorMessage_StatusField(t *testing.T) {
	tests := []struct {
		name       string
		em         model.ErrorMessage
		wantStatus int
	}{
		{
			name:       "status set via constructor",
			em:         model.NewErrorMessage(http.StatusOK, "ok"),
			wantStatus: http.StatusOK,
		},
		{
			name:       "status never set – returns zero value",
			em:         model.ErrorMessage{},
			wantStatus: 0,
		},
		{
			name:       "status set directly on struct literal",
			em:         model.ErrorMessage{Status: http.StatusForbidden},
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.wantStatus, tc.em.Status)
		})
	}
}

// ---------------------------------------------------------------------------
// Field getter semantics (Message field)
// ---------------------------------------------------------------------------

func TestErrorMessage_MessageField(t *testing.T) {
	tests := []struct {
		name        string
		em          model.ErrorMessage
		wantMessage string
	}{
		{
			name:        "message set via constructor",
			em:          model.NewErrorMessage(http.StatusOK, "success"),
			wantMessage: "success",
		},
		{
			name:        "message never set – returns empty string",
			em:          model.ErrorMessage{},
			wantMessage: "",
		},
		{
			name:        "message set directly on struct literal",
			em:          model.ErrorMessage{Message: "direct"},
			wantMessage: "direct",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.wantMessage, tc.em.Message)
		})
	}
}

// ---------------------------------------------------------------------------
// Setter semantics – mutating Status
// ---------------------------------------------------------------------------

func TestErrorMessage_SetStatus(t *testing.T) {
	tests := []struct {
		name       string
		initial    int
		updated    int
		wantStatus int
	}{
		{
			name:       "update from zero to valid status",
			initial:    0,
			updated:    http.StatusUnauthorized,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "overwrite existing status",
			initial:    http.StatusOK,
			updated:    http.StatusNotFound,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "reset to zero (null-equivalent)",
			initial:    http.StatusOK,
			updated:    0,
			wantStatus: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			em := model.NewErrorMessage(tc.initial, "msg")
			em.Status = tc.updated
			assert.Equal(t, tc.wantStatus, em.Status)
		})
	}
}

// ---------------------------------------------------------------------------
// Setter semantics – mutating Message
// ---------------------------------------------------------------------------

func TestErrorMessage_SetMessage(t *testing.T) {
	tests := []struct {
		name        string
		initial     string
		updated     string
		wantMessage string
	}{
		{
			name:        "update from empty to non-empty message",
			initial:     "",
			updated:     "new message",
			wantMessage: "new message",
		},
		{
			name:        "overwrite existing message",
			initial:     "old message",
			updated:     "new message",
			wantMessage: "new message",
		},
		{
			name:        "reset to empty (null-equivalent)",
			initial:     "some message",
			updated:     "",
			wantMessage: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			em := model.NewErrorMessage(http.StatusOK, tc.initial)
			em.Message = tc.updated
			assert.Equal(t, tc.wantMessage, em.Message)
		})
	}
}

// ---------------------------------------------------------------------------
// Equality semantics (Go struct comparison)
// ---------------------------------------------------------------------------

func TestErrorMessage_Equality(t *testing.T) {
	tests := []struct {
		name      string
		a         model.ErrorMessage
		b         model.ErrorMessage
		wantEqual bool
	}{
		{
			name:      "identical status and message – equal",
			a:         model.NewErrorMessage(http.StatusBadRequest, "bad"),
			b:         model.NewErrorMessage(http.StatusBadRequest, "bad"),
			wantEqual: true,
		},
		{
			name:      "different status – not equal",
			a:         model.NewErrorMessage(http.StatusBadRequest, "bad"),
			b:         model.NewErrorMessage(http.StatusInternalServerError, "bad"),
			wantEqual: false,
		},
		{
			name:      "different message – not equal",
			a:         model.NewErrorMessage(http.StatusBadRequest, "bad"),
			b:         model.NewErrorMessage(http.StatusBadRequest, "other"),
			wantEqual: false,
		},
		{
			name:      "both fields differ – not equal",
			a:         model.NewErrorMessage(http.StatusBadRequest, "bad"),
			b:         model.NewErrorMessage(http.StatusNotFound, "not found"),
			wantEqual: false,
		},
		{
			name:      "reflexive – same instance fields",
			a:         model.NewErrorMessage(http.StatusOK, "ok"),
			b:         model.NewErrorMessage(http.StatusOK, "ok"),
			wantEqual: true,
		},
		{
			name:      "zero values – equal",
			a:         model.ErrorMessage{},
			b:         model.ErrorMessage{},
			wantEqual: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.wantEqual {
				assert.Equal(t, tc.a, tc.b)
			} else {
				assert.NotEqual(t, tc.a, tc.b)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// String representation (fmt.Sprintf / fmt.Stringer fallback)
// ---------------------------------------------------------------------------

func TestErrorMessage_StringRepresentation(t *testing.T) {
	tests := []struct {
		name            string
		em              model.ErrorMessage
		wantContains    []string
	}{
		{
			name:         "default format includes status and message values",
			em:           model.NewErrorMessage(http.StatusBadRequest, "bad request"),
			wantContains: []string{"400", "bad request"},
		},
		{
			name:         "zero value representation",
			em:           model.ErrorMessage{},
			wantContains: []string{"0", ""},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repr := fmt.Sprintf("%+v", tc.em)
			for _, want := range tc.wantContains {
				if want != "" {
					assert.Contains(t, repr, want)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// JSON serialization (wire contract)
// ---------------------------------------------------------------------------

func TestErrorMessage_JSONMarshal(t *testing.T) {
	tests := []struct {
		name        string
		em          model.ErrorMessage
		wantJSON    string
	}{
		{
			name:     "marshal with status and message",
			em:       model.NewErrorMessage(http.StatusBadRequest, "bad request"),
			wantJSON: `{"status":400,"message":"bad request"}`,
		},
		{
			name:     "marshal zero value",
			em:       model.ErrorMessage{},
			wantJSON: `{"status":0,"message":""}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b, err := json.Marshal(tc.em)
			assert.NoError(t, err)
			assert.JSONEq(t, tc.wantJSON, string(b))
		})
	}
}

func TestErrorMessage_JSONUnmarshal(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantStatus  int
		wantMessage string
		wantErr     bool
	}{
		{
			name:        "valid JSON round-trip",
			input:       `{"status":400,"message":"bad request"}`,
			wantStatus:  http.StatusBadRequest,
			wantMessage: "bad request",
		},
		{
			name:        "empty JSON object defaults",
			input:       `{}`,
			wantStatus:  0,
			wantMessage: "",
		},
		{
			name:    "malformed JSON returns error",
			input:   `{bad json`,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var em model.ErrorMessage
			err := json.Unmarshal([]byte(tc.input), &em)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.wantStatus, em.Status)
			assert.Equal(t, tc.wantMessage, em.Message)
		})
	}
}

// ---------------------------------------------------------------------------
// HTTP handler integration (httptest)
// ---------------------------------------------------------------------------

// errorHandler is a minimal HTTP handler that writes an ErrorMessage as JSON.
func errorHandler(em model.ErrorMessage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(em.Status)
		_ = json.NewEncoder(w).Encode(em)
	}
}

func TestErrorMessage_HTTPHandler(t *testing.T) {
	tests := []struct {
		name           string
		em             model.ErrorMessage
		wantStatusCode int
		wantBody       model.ErrorMessage
	}{
		{
			name:           "handler returns 400 with bad request body",
			em:             model.NewErrorMessage(http.StatusBadRequest, "bad request"),
			wantStatusCode: http.StatusBadRequest,
			wantBody:       model.NewErrorMessage(http.StatusBadRequest, "bad request"),
		},
		{
			name:           "handler returns 500 with internal server error body",
			em:             model.NewErrorMessage(http.StatusInternalServerError, "internal error"),
			wantStatusCode: http.StatusInternalServerError,
			wantBody:       model.NewErrorMessage(http.StatusInternalServerError, "internal error"),
		},
		{
			name:           "handler returns 404 with not found body",
			em:             model.NewErrorMessage(http.StatusNotFound, "not found"),
			wantStatusCode: http.StatusNotFound,
			wantBody:       model.NewErrorMessage(http.StatusNotFound, "not found"),
		},
		{
			name:           "handler returns 401 with unauthorized body",
			em:             model.NewErrorMessage(http.StatusUnauthorized, "unauthorized"),
			wantStatusCode: http.StatusUnauthorized,
			wantBody:       model.NewErrorMessage(http.StatusUnauthorized, "unauthorized"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/error", nil)
			rec := httptest.NewRecorder()

			errorHandler(tc.em).ServeHTTP(rec, req)

			assert.Equal(t, tc.wantStatusCode, rec.Code)
			assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

			var got model.ErrorMessage
			err := json.NewDecoder(rec.Body).Decode(&got)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantBody.Status, got.Status)
			assert.Equal(t, tc.wantBody.Message, got.Message)
		})
	}
}

// ---------------------------------------------------------------------------
// Independence of fields – mutations to one field must not affect the other
// ---------------------------------------------------------------------------

func TestErrorMessage_FieldIndependence(t *testing.T) {
	tests := []struct {
		name            string
		initialStatus   int
		initialMessage  string
		newStatus       int
		newMessage      string
	}{
		{
			name:           "changing status does not change message",
			initialStatus:  http.StatusOK,
			initialMessage: "ok",
			newStatus:      http.StatusBadRequest,
			newMessage:     "ok", // unchanged
		},
		{
			name:           "changing message does not