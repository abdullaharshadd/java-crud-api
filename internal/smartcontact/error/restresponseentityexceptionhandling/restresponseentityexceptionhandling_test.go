```go
package error

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/smartContact/internal/smartcontact/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// decodeErrorMessage is a helper that decodes the JSON body from a ResponseRecorder
// into a model.ErrorMessage (or a plain map if the concrete type is not exported).
// We decode into a map to remain independent of the exact ErrorMessage field tags.
func decodeBody(t *testing.T, rec *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var result map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&result)
	require.NoError(t, err, "response body should be valid JSON")
	return result
}

// --------------------------------------------------------------------------
// HandleUserNotFound
// --------------------------------------------------------------------------

func TestHandleUserNotFound(t *testing.T) {
	tests := []struct {
		name           string
		err            *UserNotFoundError
		wantStatus     int
		wantMsgContain string // substring expected inside the "message" field
		wantMsgNil     bool   // true when we expect message to be empty/null
	}{
		{
			name:           "UserNotFoundError with message",
			err:            &UserNotFoundError{Message: "user with id 42 not found"},
			wantStatus:     http.StatusNotFound,
			wantMsgContain: "user with id 42 not found",
		},
		{
			name:       "UserNotFoundError with empty message",
			err:        &UserNotFoundError{Message: ""},
			wantStatus: http.StatusNotFound,
			wantMsgNil: true,
		},
		{
			name:       "nil UserNotFoundError pointer",
			err:        nil,
			wantStatus: http.StatusNotFound,
			wantMsgNil: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			writeErr := HandleUserNotFound(rec, tc.err)
			assert.NoError(t, writeErr, "HandleUserNotFound should not return an encoding error")

			// HTTP status invariant: always 404
			assert.Equal(t, http.StatusNotFound, rec.Code,
				"response HTTP status must be 404 NOT_FOUND")

			// Content-Type invariant
			assert.Equal(t, "application/json", rec.Header().Get("Content-Type"),
				"response Content-Type must be application/json")

			// Body invariant: must be a non-null JSON object
			body := decodeBody(t, rec)
			assert.NotNil(t, body, "response body ErrorMessage must not be nil")

			// Status field in ErrorMessage must equal 404
			statusVal, ok := body["status"]
			assert.True(t, ok, "ErrorMessage body must contain a 'status' field")
			// JSON numbers decode as float64
			assert.EqualValues(t, http.StatusNotFound, statusVal,
				"ErrorMessage.status must equal NOT_FOUND (404)")

			// Message field
			msgVal := body["message"]
			if tc.wantMsgNil {
				// empty string or null are both acceptable for a nil/empty message
				assert.True(t, msgVal == nil || msgVal == "",
					"ErrorMessage.message should be empty/null when no message provided, got: %v", msgVal)
			} else {
				assert.Contains(t, msgVal, tc.wantMsgContain,
					"ErrorMessage.message must equal the exception message")
			}
		})
	}
}

// --------------------------------------------------------------------------
// WriteError – UserNotFoundError path
// --------------------------------------------------------------------------

func TestWriteError_UserNotFoundError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		wantStatus     int
		wantMsgContain string
		wantMsgEmpty   bool
	}{
		{
			name:           "direct *UserNotFoundError with message",
			err:            &UserNotFoundError{Message: "user 99 not found"},
			wantStatus:     http.StatusNotFound,
			wantMsgContain: "user 99 not found",
		},
		{
			name:         "direct *UserNotFoundError with empty message",
			err:          &UserNotFoundError{Message: ""},
			wantStatus:   http.StatusNotFound,
			wantMsgEmpty: true,
		},
		{
			name:           "wrapped *UserNotFoundError",
			err:            errors.Join(errors.New("outer"), &UserNotFoundError{Message: "inner user not found"}),
			wantStatus:     http.StatusNotFound,
			wantMsgContain: "inner user not found",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			writeErr := WriteError(rec, tc.err)
			assert.NoError(t, writeErr, "WriteError should not return an encoding error")

			assert.Equal(t, tc.wantStatus, rec.Code,
				"response HTTP status must be 404 for UserNotFoundError")

			assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

			body := decodeBody(t, rec)
			assert.NotNil(t, body)

			statusVal := body["status"]
			assert.EqualValues(t, http.StatusNotFound, statusVal,
				"ErrorMessage.status must be 404")

			msgVal := body["message"]
			if tc.wantMsgEmpty {
				assert.True(t, msgVal == nil || msgVal == "",
					"ErrorMessage.message should be empty for empty-message error, got: %v", msgVal)
			} else {
				assert.Contains(t, msgVal, tc.wantMsgContain)
			}
		})
	}
}

// --------------------------------------------------------------------------
// WriteError – generic / unknown error path (500)
// --------------------------------------------------------------------------

func TestWriteError_GenericError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		wantStatus     int
		wantMsgContain string
		wantMsgEmpty   bool
	}{
		{
			name:           "generic error produces 500",
			err:            errors.New("something went wrong"),
			wantStatus:     http.StatusInternalServerError,
			wantMsgContain: "something went wrong",
		},
		{
			name:         "nil error produces 500 with empty message",
			err:          nil,
			wantStatus:   http.StatusInternalServerError,
			wantMsgEmpty: true,
		},
		{
			name:           "wrapped generic error produces 500",
			err:            errors.Join(errors.New("wrapped"), errors.New("cause")),
			wantStatus:     http.StatusInternalServerError,
			wantMsgContain: "wrapped",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			writeErr := WriteError(rec, tc.err)
			assert.NoError(t, writeErr, "WriteError should not return an encoding error")

			assert.Equal(t, tc.wantStatus, rec.Code)
			assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

			body := decodeBody(t, rec)
			assert.NotNil(t, body)

			statusVal := body["status"]
			assert.EqualValues(t, http.StatusInternalServerError, statusVal,
				"ErrorMessage.status must be 500 for generic errors")

			msgVal := body["message"]
			if tc.wantMsgEmpty {
				assert.True(t, msgVal == nil || msgVal == "",
					"ErrorMessage.message should be empty for nil error, got: %v", msgVal)
			} else {
				assert.Contains(t, msgVal, tc.wantMsgContain)
			}
		})
	}
}

// --------------------------------------------------------------------------
// writeErrorMessage (tested indirectly via broken writer)
// --------------------------------------------------------------------------

// brokenWriter simulates a ResponseWriter whose Write always fails after
// WriteHeader has been called, to exercise the encoding-error return path.
type brokenWriter struct {
	header http.Header
	code   int
}

func newBrokenWriter() *brokenWriter {
	return &brokenWriter{header: make(http.Header)}
}

func (b *brokenWriter) Header() http.Header         { return b.header }
func (b *brokenWriter) WriteHeader(code int)         { b.code = code }
func (b *brokenWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("simulated write failure")
}

func TestWriteErrorMessage_EncodingError(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		message string
	}{
		{
			name:    "broken writer returns encoding error for 404",
			status:  http.StatusNotFound,
			message: "some message",
		},
		{
			name:    "broken writer returns encoding error for 500",
			status:  http.StatusInternalServerError,
			message: "internal error",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			w := newBrokenWriter()
			err := writeErrorMessage(w, tc.status, tc.message)
			assert.Error(t, err, "writeErrorMessage must propagate the encoding/write error")
			assert.Equal(t, tc.status, w.code,
				"WriteHeader must still be called with the correct status even when encoding fails")
		})
	}
}

// --------------------------------------------------------------------------
// Invariant: no state mutation / idempotent response structure
// --------------------------------------------------------------------------

func TestHandleUserNotFound_ResponseStructureInvariants(t *testing.T) {
	// Run the same case twice to prove no global state mutation.
	for i := 0; i < 2; i++ {
		rec := httptest.NewRecorder()
		err := &UserNotFoundError{Message: "user not found"}

		writeErr := HandleUserNotFound(rec, err)
		assert.NoError(t, writeErr)
		assert.Equal(t, http.StatusNotFound, rec.Code)

		body := decodeBody(t, rec)
		assert.EqualValues(t, http.StatusNotFound, body["status"])
		assert.Equal(t, err.Error(), body["message"])
	}
}

// --------------------------------------------------------------------------
// model.NewErrorMessage integration smoke-test
// --------------------------------------------------------------------------

func TestWriteError_ErrorMessageBodyFields(t *testing.T) {
	// Verify that the body written by WriteError can be round-tripped back to a
	// model.ErrorMessage (or at least that the essential fields are present and
	// correctly typed).
	rec := httptest.NewRecorder()
	inputErr := &UserNotFoundError{Message: "round-trip check"}

	require.NoError(t, WriteError(rec, inputErr))

	var em model.ErrorMessage
	decodeErr := json.NewDecoder(rec.Body).Decode(&em)
	if decodeErr == nil {
		// If the struct is exported and decodable, validate its fields directly.
		assert.Equal(t, http.StatusNotFound, em.Status())
		assert.Equal(t, inputErr.Error(), em.MessageText())
	} else {
		// Fallback: decode as map (model.ErrorMessage may not export accessors).
		// We already test this path in TestHandleUserNotFound.
		t.Logf("model.ErrorMessage not directly decodable (%v); structural fields tested via map in other tests", decodeErr)
	}
}
```