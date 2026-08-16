```go
package apperror_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apperror "github.com/smartContact/internal/smartcontact/error/apperror"
	"github.com/smartContact/internal/smartcontact/model"
	"github.com/smartContact/internal/smartcontact/repository"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// decodeErrorMessage reads and decodes the JSON body written by WriteError.
func decodeErrorMessage(t *testing.T, body []byte) model.ErrorMessage {
	t.Helper()
	var em model.ErrorMessage
	require.NoError(t, json.Unmarshal(body, &em), "response body must be valid JSON ErrorMessage")
	return em
}

// ---------------------------------------------------------------------------
// StatusFor – table-driven
// ---------------------------------------------------------------------------

func TestStatusFor(t *testing.T) {
	pgDuplicateKey := &pgconn.PgError{Code: "23505"}
	pgOtherCode := &pgconn.PgError{Code: "42000"}

	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		// ── nil ──────────────────────────────────────────────────────────────
		{
			name:       "nil error returns 200",
			err:        nil,
			wantStatus: http.StatusOK,
		},

		// ── UserNotFound (struct type) ────────────────────────────────────────
		{
			name:       "UserNotFound pointer returns 404",
			err:        &apperror.UserNotFound{Message: "user 42 not found"},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "UserNotFound pointer with empty message returns 404",
			err:        &apperror.UserNotFound{},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "wrapped UserNotFound returns 404",
			err:        fmt.Errorf("wrapped: %w", &apperror.UserNotFound{Message: "gone"}),
			wantStatus: http.StatusNotFound,
		},

		// ── repository sentinel errors ────────────────────────────────────────
		{
			name:       "repository.ErrUserNotFound returns 404",
			err:        repository.ErrUserNotFound,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "wrapped repository.ErrUserNotFound returns 404",
			err:        fmt.Errorf("layer: %w", repository.ErrUserNotFound),
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "repository.ErrNoRowsDeleted returns 500",
			err:        repository.ErrNoRowsDeleted,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "wrapped repository.ErrNoRowsDeleted returns 500",
			err:        fmt.Errorf("layer: %w", repository.ErrNoRowsDeleted),
			wantStatus: http.StatusInternalServerError,
		},

		// ── ValidationError ───────────────────────────────────────────────────
		{
			name:       "ValidationError returns 400",
			err:        apperror.NewValidationError("name is required"),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "wrapped ValidationError returns 400",
			err:        fmt.Errorf("validation: %w", apperror.NewValidationError("bad field")),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "ValidationError with empty message returns 400",
			err:        apperror.NewValidationError(""),
			wantStatus: http.StatusBadRequest,
		},

		// ── Postgres errors ───────────────────────────────────────────────────
		{
			name:       "pgconn.PgError 23505 duplicate key returns 500",
			err:        pgDuplicateKey,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "wrapped pgconn.PgError 23505 returns 500",
			err:        fmt.Errorf("db: %w", pgDuplicateKey),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "pgconn.PgError other code returns 500",
			err:        pgOtherCode,
			wantStatus: http.StatusInternalServerError,
		},

		// ── generic / unknown errors ──────────────────────────────────────────
		{
			name:       "plain errors.New returns 500",
			err:        errors.New("something went wrong"),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "fmt.Errorf plain returns 500",
			err:        fmt.Errorf("unexpected condition"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := apperror.StatusFor(tc.err)
			assert.Equal(t, tc.wantStatus, got)
		})
	}
}

// ---------------------------------------------------------------------------
// WriteError – table-driven (via httptest.ResponseRecorder)
// ---------------------------------------------------------------------------

func TestWriteError(t *testing.T) {
	tests := []struct {
		name            string
		err             error
		wantStatus      int
		wantMsgContains string // substring expected inside ErrorMessage.Message
		wantBodyIsJSON  bool
	}{
		// ── UserNotFoundException / spec: "a controller throws UserNotFoundException with a message"
		{
			name:            "UserNotFound with message -> 404 body mirrors message",
			err:             &apperror.UserNotFound{Message: "user 99 not found"},
			wantStatus:      http.StatusNotFound,
			wantMsgContains: "user 99 not found",
			wantBodyIsJSON:  true,
		},
		// ── spec: "a controller throws UserNotFoundException with a null message"
		// In Go null maps to zero value; UserNotFound with empty Message.
		{
			name:            "UserNotFound with empty message -> 404",
			err:             &apperror.UserNotFound{},
			wantStatus:      http.StatusNotFound,
			wantBodyIsJSON:  true,
		},
		// ── spec: "exception other than UserNotFoundException is thrown" -> default handler
		{
			name:            "generic error -> 500",
			err:             errors.New("generic problem"),
			wantStatus:      http.StatusInternalServerError,
			wantMsgContains: "generic problem",
			wantBodyIsJSON:  true,
		},
		// ── ValidationError -> 400
		{
			name:            "ValidationError -> 400",
			err:             apperror.NewValidationError("email is invalid"),
			wantStatus:      http.StatusBadRequest,
			wantMsgContains: "email is invalid",
			wantBodyIsJSON:  true,
		},
		// ── repository sentinel -> 500
		{
			name:           "ErrNoRowsDeleted -> 500",
			err:            repository.ErrNoRowsDeleted,
			wantStatus:     http.StatusInternalServerError,
			wantBodyIsJSON: true,
		},
		// ── repository.ErrUserNotFound -> 404
		{
			name:           "repository.ErrUserNotFound -> 404",
			err:            repository.ErrUserNotFound,
			wantStatus:     http.StatusNotFound,
			wantBodyIsJSON: true,
		},
		// ── Postgres duplicate key -> 500
		{
			name:           "pgconn 23505 -> 500",
			err:            &pgconn.PgError{Code: "23505"},
			wantStatus:     http.StatusInternalServerError,
			wantBodyIsJSON: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()

			apperror.WriteError(rr, tc.err)

			res := rr.Result()
			defer res.Body.Close()

			// ── status code
			assert.Equal(t, tc.wantStatus, res.StatusCode)

			// ── Content-Type
			assert.Equal(t, "application/json; charset=utf-8", res.Header.Get("Content-Type"))

			// ── body is valid JSON ErrorMessage
			body := rr.Body.Bytes()
			if tc.wantBodyIsJSON {
				em := decodeErrorMessage(t, body)

				// The ErrorMessage status field must mirror the HTTP status code.
				assert.Equal(t, tc.wantStatus, em.Status,
					"ErrorMessage.Status must equal the HTTP status code")

				// The message field must not be empty (unless we explicitly allow it).
				if tc.wantMsgContains != "" {
					assert.Contains(t, em.Message, tc.wantMsgContains)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// WriteError invariants from the spec
// ---------------------------------------------------------------------------

// TestWriteError_UserNotFoundException_Invariants explicitly validates all four
// invariants stated in the behavioral spec for userNotFoundException.
func TestWriteError_UserNotFoundException_Invariants(t *testing.T) {
	tests := []struct {
		name        string
		err         *apperror.UserNotFound
		wantMessage string // exact expected message (empty string = "")
	}{
		{
			name:        "with a non-empty message",
			err:         &apperror.UserNotFound{Message: "user 1 not found"},
			wantMessage: "user 1 not found",
		},
		{
			name:        "with an empty message (null equivalent)",
			err:         &apperror.UserNotFound{},
			wantMessage: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			apperror.WriteError(rr, tc.err)

			res := rr.Result()
			defer res.Body.Close()

			// Invariant 1: HTTP response status is always 404 (NOT_FOUND).
			assert.Equal(t, http.StatusNotFound, res.StatusCode,
				"invariant: HTTP status must always be 404 for UserNotFound")

			// Invariant 2: response body is always a non-null ErrorMessage object.
			body := rr.Body.Bytes()
			assert.NotEmpty(t, body, "invariant: response body must not be empty")

			var em model.ErrorMessage
			err := json.Unmarshal(body, &em)
			require.NoError(t, err, "invariant: response body must be a valid ErrorMessage")

			// Invariant 3: ErrorMessage status field is always NOT_FOUND (404).
			assert.Equal(t, http.StatusNotFound, em.Status,
				"invariant: ErrorMessage.Status must be 404")

			// Invariant 4: ErrorMessage message field mirrors the exception's message.
			assert.Equal(t, tc.wantMessage, em.Message,
				"invariant: ErrorMessage.Message must mirror the exception's message")
		})
	}
}

// ---------------------------------------------------------------------------
// ValidationError unit tests
// ---------------------------------------------------------------------------

func TestValidationError(t *testing.T) {
	tests := []struct {
		name    string
		message string
		wantMsg string
	}{
		{
			name:    "non-empty message",
			message: "field is required",
			wantMsg: "field is required",
		},
		{
			name:    "empty message",
			message: "",
			wantMsg: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ve := apperror.NewValidationError(tc.message)
			require.NotNil(t, ve)
			assert.Equal(t, tc.wantMsg, ve.Error())
			assert.Equal(t, tc.wantMsg, ve.Message)
		})
	}
}

func TestValidationError_NilReceiver(t *testing.T) {
	var ve *apperror.ValidationError
	// nil receiver must not panic and must return the fallback string.
	assert.Equal(t, "validation error", ve.Error())
}

// ---------------------------------------------------------------------------
// Global invariants: non-UserNotFoundException errors are NOT handled as 404
// ---------------------------------------------------------------------------

func TestWriteError_OtherErrors_NotHandledAs404(t *testing.T) {
	otherErrors := []struct {
		name string
		err  error
	}{
		{"generic error", errors.New("boom")},
		{"validation error", apperror.NewValidationError("bad input")},
		{"no rows deleted", repository.ErrNoRowsDeleted},
		{"postgres error", &pgconn.PgError{Code: "23505"}},
	}

	for _, tc := range otherErrors {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			apperror.WriteError(rr, tc.err)

			assert.NotEqual(t, http.StatusNotFound, rr.Code,
				"non-UserNotFound errors must not produce 404")
		})
	}
}

// ---------------------------------------------------------------------------
// WriteError does not modify persistent state / external side-effects
// (verified by ensuring only the ResponseRecorder is mutated)
// ---------------------------------------------------------------------------

func TestWriteError_NoExternalSideEffects(t *testing.T) {
	// We simply verify that WriteError writes to the provided ResponseWriter
	// and does not panic or call any real external system. Using httptest
	// ensures isolation.
	rr := httptest.NewRecorder()
	apperror.WriteError(rr, &apperror.UserNotFound{Message: "no user"})

	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.NotEmpty(t, rr.Body.Bytes())
}

// ---------------------------------------------------------------------------
// Content-Type header is always set correctly
// ---------------------------------------------------------------------------

func TestWriteError_ContentTypeAlwaysJSON(t *testing.T) {
	errs := []error{
		nil,
		&apperror.UserNotFound{Message: "gone"},
		apperror.NewValidationError("bad"),
		errors.New("generic"),
		repository.ErrNoRowsDeleted,
		repository.ErrUserNotFound,
	}

	for _, e := range errs {
		name := "nil"
		if e != nil {
			name = e.Error()
		}
		t.Run(name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			apperror.WriteError(rr, e)
			assert.Equal(t, "application/json; charset=utf-8",
				rr.Header().Get("Content-Type"))
		})
	}
}
```