```go
package error_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	errorpkg "github.com/smartContact/internal/smartcontact/error/restresponseentityexceptionhandling"
	"github.com/smartContact/internal/smartcontact/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// decodeBody is a helper that decodes the JSON body from a ResponseRecorder
// into an ErrorMessage.
func decodeBody(t *testing.T, rr *httptest.ResponseRecorder) errorpkg.ErrorMessage {
	t.Helper()
	var em errorpkg.ErrorMessage
	err := json.NewDecoder(rr.Body).Decode(&em)
	require.NoError(t, err, "response body must be valid JSON ErrorMessage")
	return em
}

// ----------------------------------------------------------------------------
// NewErrorMessage
// ----------------------------------------------------------------------------

func TestNewErrorMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		status  int
		message string
	}{
		{
			name:    "404 with message",
			status:  http.StatusNotFound,
			message: "user not found",
		},
		{
			name:    "500 with message",
			status:  http.StatusInternalServerError,
			message: "internal error",
		},
		{
			name:    "200 with empty message",
			status:  http.StatusOK,
			message: "",
		},
		{
			name:    "400 with long message",
			status:  http.StatusBadRequest,
			message: strings.Repeat("x", 512),
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			em := errorpkg.NewErrorMessage(tc.status, tc.message)
			assert.Equal(t, tc.status, em.Status)
			assert.Equal(t, tc.message, em.Message)
		})
	}
}

// ----------------------------------------------------------------------------
// WriteError — HTTP status mapping
// ----------------------------------------------------------------------------

func TestWriteError_StatusMapping(t *testing.T) {
	t.Parallel()

	wrappedUserNotFound := fmt.Errorf("service layer: %w", errorpkg.ErrUserNotFound)
	wrappedEmptyResult := fmt.Errorf("repo layer: %w", repository.ErrEmptyResultDelete)

	tests := []struct {
		name           string
		err            error
		expectedStatus int
	}{
		// ── UserNotFoundException equivalents ──────────────────────────────
		{
			name:           "ErrUserNotFound maps to 404",
			err:            errorpkg.ErrUserNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "wrapped ErrUserNotFound maps to 404",
			err:            wrappedUserNotFound,
			expectedStatus: http.StatusNotFound,
		},
		// ── EmptyResultDataAccessException equivalents ────────────────────
		// MIGRATION_NOTE: the Java @ControllerAdvice deliberately did NOT
		// handle EmptyResultDataAccessException, so it fell through to HTTP 500.
		// We must preserve that behaviour.
		{
			name:           "ErrEmptyResultDelete maps to 500 (not 404)",
			err:            repository.ErrEmptyResultDelete,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "wrapped ErrEmptyResultDelete maps to 500",
			err:            wrappedEmptyResult,
			expectedStatus: http.StatusInternalServerError,
		},
		// ── Unhandled / generic errors ─────────────────────────────────────
		{
			name:           "generic error maps to 500",
			err:            errors.New("something went wrong"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "nil error maps to 500",
			err:            nil,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rr := httptest.NewRecorder()
			errorpkg.WriteError(rr, tc.err)

			assert.Equal(t, tc.expectedStatus, rr.Code,
				"HTTP status code must match expected mapping")
		})
	}
}

// ----------------------------------------------------------------------------
// WriteError — response headers
// ----------------------------------------------------------------------------

func TestWriteError_ContentTypeHeader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
	}{
		{"ErrUserNotFound sets JSON content-type", errorpkg.ErrUserNotFound},
		{"generic error sets JSON content-type", errors.New("boom")},
		{"nil error sets JSON content-type", nil},
		{"ErrEmptyResultDelete sets JSON content-type", repository.ErrEmptyResultDelete},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rr := httptest.NewRecorder()
			errorpkg.WriteError(rr, tc.err)

			assert.Equal(t, "application/json",
				rr.Header().Get("Content-Type"),
				"Content-Type must be application/json")
		})
	}
}

// ----------------------------------------------------------------------------
// WriteError — response body (ErrorMessage fields)
// ----------------------------------------------------------------------------

func TestWriteError_ResponseBody(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		err             error
		expectedStatus  int
		expectedMessage string
	}{
		// ── Spec: UserNotFoundException with non-null message ──────────────
		{
			name:            "ErrUserNotFound body has status 404 and correct message",
			err:             errorpkg.ErrUserNotFound,
			expectedStatus:  http.StatusNotFound,
			expectedMessage: errorpkg.ErrUserNotFound.Error(),
		},
		// ── Spec: UserNotFoundException wrapped (errors.Is chain) ──────────
		{
			name:            "wrapped ErrUserNotFound body carries wrapper message",
			err:             fmt.Errorf("lookup failed: %w", errorpkg.ErrUserNotFound),
			expectedStatus:  http.StatusNotFound,
			expectedMessage: "lookup failed: user not found",
		},
		// ── Spec: empty/null message on UserNotFoundException ─────────────
		// Go errors cannot carry a truly nil message but we can test empty string.
		{
			name:            "custom ErrUserNotFound with empty wrapping message",
			err:             fmt.Errorf(": %w", errorpkg.ErrUserNotFound),
			expectedStatus:  http.StatusNotFound,
			expectedMessage: ": user not found",
		},
		// ── MIGRATION_NOTE: EmptyResultDelete → 500 ───────────────────────
		{
			name:            "ErrEmptyResultDelete body has status 500",
			err:             repository.ErrEmptyResultDelete,
			expectedStatus:  http.StatusInternalServerError,
			expectedMessage: repository.ErrEmptyResultDelete.Error(),
		},
		// ── Generic error ─────────────────────────────────────────────────
		{
			name:            "generic error body has status 500 and correct message",
			err:             errors.New("database timeout"),
			expectedStatus:  http.StatusInternalServerError,
			expectedMessage: "database timeout",
		},
		// ── nil error ─────────────────────────────────────────────────────
		{
			name:            "nil error produces 500 with empty message",
			err:             nil,
			expectedStatus:  http.StatusInternalServerError,
			expectedMessage: "",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rr := httptest.NewRecorder()
			errorpkg.WriteError(rr, tc.err)

			em := decodeBody(t, rr)

			assert.Equal(t, tc.expectedStatus, em.Status,
				"ErrorMessage.Status must match the HTTP status code")
			assert.Equal(t, tc.expectedMessage, em.Message,
				"ErrorMessage.Message must equal the error string")
		})
	}
}

// ----------------------------------------------------------------------------
// WriteError — invariants: body is always a well-formed ErrorMessage
// ----------------------------------------------------------------------------

func TestWriteError_BodyIsAlwaysWellFormedJSON(t *testing.T) {
	t.Parallel()

	errs := []error{
		errorpkg.ErrUserNotFound,
		repository.ErrEmptyResultDelete,
		errors.New("random error"),
		nil,
		fmt.Errorf("wrapped: %w", errorpkg.ErrUserNotFound),
	}

	for _, e := range errs {
		e := e
		name := "nil"
		if e != nil {
			name = e.Error()
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			rr := httptest.NewRecorder()
			errorpkg.WriteError(rr, e)

			// body must be non-empty
			assert.NotEmpty(t, rr.Body.Bytes(), "response body must not be empty")

			// body must decode cleanly into ErrorMessage
			var em errorpkg.ErrorMessage
			decodeErr := json.NewDecoder(rr.Body).Decode(&em)
			assert.NoError(t, decodeErr, "body must be valid JSON")

			// status field must match HTTP status code
			assert.Equal(t, rr.Code, em.Status,
				"ErrorMessage.Status must equal the HTTP status code written to the response")
		})
	}
}

// ----------------------------------------------------------------------------
// WriteError — no persistent state / side effects (idempotency on recorder)
// ----------------------------------------------------------------------------

func TestWriteError_NoExternalSideEffects(t *testing.T) {
	t.Parallel()

	// Two independent calls with the same error should produce identical results.
	tests := []struct {
		name string
		err  error
	}{
		{"ErrUserNotFound is idempotent", errorpkg.ErrUserNotFound},
		{"generic error is idempotent", errors.New("boom")},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rr1 := httptest.NewRecorder()
			rr2 := httptest.NewRecorder()

			errorpkg.WriteError(rr1, tc.err)
			errorpkg.WriteError(rr2, tc.err)

			assert.Equal(t, rr1.Code, rr2.Code)
			assert.Equal(t, rr1.Header().Get("Content-Type"), rr2.Header().Get("Content-Type"))
			assert.Equal(t, rr1.Body.String(), rr2.Body.String())
		})
	}
}

// ----------------------------------------------------------------------------
// WriteError — used inside a real http.Handler (integration-style)
// ----------------------------------------------------------------------------

func TestWriteError_InsideHTTPHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		err            error
		expectedStatus int
	}{
		{
			name:           "handler returns 404 for ErrUserNotFound",
			err:            errorpkg.ErrUserNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "handler returns 500 for ErrEmptyResultDelete",
			err:            repository.ErrEmptyResultDelete,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "handler returns 500 for unknown error",
			err:            errors.New("unexpected"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Build a minimal handler that simulates a service returning an error.
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				errorpkg.WriteError(w, tc.err)
			})

			req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

			em := decodeBody(t, rr)
			assert.Equal(t, tc.expectedStatus, em.Status)
			if tc.err != nil {
				assert.Equal(t, tc.err.Error(), em.Message)
			} else {
				assert.Empty(t, em.Message)
			}
		})
	}
}

// ----------------------------------------------------------------------------
// ErrUserNotFound sentinel
// ----------------------------------------------------------------------------

func TestErrUserNotFound_Sentinel(t *testing.T) {
	t.Parallel()

	assert.True(t, errors.Is(errorpkg.ErrUserNotFound, errorpkg.ErrUserNotFound),
		"sentinel must satisfy errors.Is with itself")

	wrapped := fmt.Errorf("layer: %w", errorpkg.ErrUserNotFound)
	assert.True(t, errors.Is(wrapped, errorpkg.ErrUserNotFound),
		"wrapped sentinel must satisfy errors.Is")

	assert.False(t, errors.Is(errors.New("user not found"), errorpkg.ErrUserNotFound),
		"a different error with the same text must NOT satisfy errors.Is")
}

// ----------------------------------------------------------------------------
// ErrorMessage — JSON serialisation round-trip
// ----------------------------------------------------------------------------

func TestErrorMessage_JSONRoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		status  int
		message string
	}{
		{"404 user not found", http.StatusNotFound, "user not found"},
		{"500 internal", http.StatusInternalServerError, "internal server error"},
		{"zero values", 0, ""},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			original := errorpkg.NewErrorMessage(tc.status, tc.message)
			data, err := json.Marshal(original)
			require.NoError(t, err)

			var decoded errorpkg.ErrorMessage
			require.NoError(t, json.Unmarshal(data, &decoded))

			assert.Equal(t, original.Status, decoded.Status)
			assert.Equal(t, original.Message, decoded.Message)
		})
	}
}

// ----------------------------------------------------------------------------
// helpers — import fmt (needed for fmt.Errorf in tests above)
// ----------------------------------------------------------------------------

// The import is placed here to satisfy the compiler; fmt is used by
// fmt.Errorf calls in the table entries above.
var _ = fmt.Sprintf
```

> **Note — one import alias needed in the real file:** the package declaration is `package error` which is a reserved identifier in Go. The test file uses `errorpkg` as the import alias:
>
> ```go
> import errorpkg "github.com/smartContact/internal/smartcontact/error/restresponseentityexceptionhandling"
> ```
>
> Adjust the module path to match your actual `go.mod` if it differs.