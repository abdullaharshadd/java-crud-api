```go
package apperr_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apperr "github.com/yourmodule/internal/smartcontact/error"
)

// ---------------------------------------------------------------------------
// Helper: a tiny fake cause error used across several tests.
// ---------------------------------------------------------------------------

var errCause = errors.New("underlying db error")

// ---------------------------------------------------------------------------
// NewUserNotFound – no-cause constructor
// ---------------------------------------------------------------------------

func TestNewUserNotFound(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		id          string
		wantID      string
		wantCause   error
		wantMessage string
	}{
		{
			name:        "empty id (analogous to Java no-arg constructor)",
			id:          "",
			wantID:      "",
			wantCause:   nil,
			wantMessage: "user not found",
		},
		{
			name:        "numeric id",
			id:          "42",
			wantID:      "42",
			wantCause:   nil,
			wantMessage: `user not found: id "42"`,
		},
		{
			name:        "uuid id",
			id:          "550e8400-e29b-41d4-a716-446655440000",
			wantID:      "550e8400-e29b-41d4-a716-446655440000",
			wantCause:   nil,
			wantMessage: `user not found: id "550e8400-e29b-41d4-a716-446655440000"`,
		},
		{
			name:        "email address as id",
			id:          "user@example.com",
			wantID:      "user@example.com",
			wantCause:   nil,
			wantMessage: `user not found: id "user@example.com"`,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := apperr.NewUserNotFound(tc.id)

			require.NotNil(t, err)
			assert.Equal(t, tc.wantID, err.ID, "ID field")
			assert.Equal(t, tc.wantCause, err.Cause, "Cause field")
			assert.Equal(t, tc.wantMessage, err.Error(), "Error() message")
		})
	}
}

// ---------------------------------------------------------------------------
// NewUserNotFoundWithCause – cause constructor
// ---------------------------------------------------------------------------

func TestNewUserNotFoundWithCause(t *testing.T) {
	t.Parallel()

	anotherCause := fmt.Errorf("timeout connecting to user store")

	tests := []struct {
		name        string
		id          string
		cause       error
		wantID      string
		wantCause   error
		wantMessage string
	}{
		{
			name:        "non-null id and non-null cause",
			id:          "99",
			cause:       errCause,
			wantID:      "99",
			wantCause:   errCause,
			wantMessage: `user not found: id "99": underlying db error`,
		},
		{
			name:        "empty id with non-null cause",
			id:          "",
			cause:       errCause,
			wantID:      "",
			wantCause:   errCause,
			wantMessage: "user not found: underlying db error",
		},
		{
			name:        "non-null id with nil cause",
			id:          "101",
			cause:       nil,
			wantID:      "101",
			wantCause:   nil,
			wantMessage: `user not found: id "101"`,
		},
		{
			name:        "empty id with nil cause",
			id:          "",
			cause:       nil,
			wantID:      "",
			wantCause:   nil,
			wantMessage: "user not found",
		},
		{
			name:        "wrapped cause chain",
			id:          "7",
			cause:       anotherCause,
			wantID:      "7",
			wantCause:   anotherCause,
			wantMessage: `user not found: id "7": timeout connecting to user store`,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := apperr.NewUserNotFoundWithCause(tc.id, tc.cause)

			require.NotNil(t, err)
			assert.Equal(t, tc.wantID, err.ID, "ID field")
			assert.Equal(t, tc.wantCause, err.Cause, "Cause field")
			assert.Equal(t, tc.wantMessage, err.Error(), "Error() message")
		})
	}
}

// ---------------------------------------------------------------------------
// errors.Is – sentinel matching
// ---------------------------------------------------------------------------

func TestErrorsIs_ErrUserNotFound(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		buildErr  func() error
		wantMatch bool
	}{
		{
			name:      "NewUserNotFound always matches sentinel",
			buildErr:  func() error { return apperr.NewUserNotFound("1") },
			wantMatch: true,
		},
		{
			name:      "NewUserNotFound empty id matches sentinel",
			buildErr:  func() error { return apperr.NewUserNotFound("") },
			wantMatch: true,
		},
		{
			name:      "NewUserNotFoundWithCause with non-nil cause matches sentinel",
			buildErr:  func() error { return apperr.NewUserNotFoundWithCause("2", errCause) },
			wantMatch: true,
		},
		{
			name:      "NewUserNotFoundWithCause with nil cause matches sentinel",
			buildErr:  func() error { return apperr.NewUserNotFoundWithCause("3", nil) },
			wantMatch: true,
		},
		{
			name:      "bare ErrUserNotFound matches itself",
			buildErr:  func() error { return apperr.ErrUserNotFound },
			wantMatch: true,
		},
		{
			name:      "wrapped in fmt.Errorf %w matches sentinel",
			buildErr:  func() error { return fmt.Errorf("handler: %w", apperr.NewUserNotFound("4")) },
			wantMatch: true,
		},
		{
			name:      "unrelated error does NOT match sentinel",
			buildErr:  func() error { return errors.New("something else") },
			wantMatch: false,
		},
		{
			name:      "nil error does NOT match sentinel",
			buildErr:  func() error { return nil },
			wantMatch: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.buildErr()
			got := errors.Is(err, apperr.ErrUserNotFound)
			assert.Equal(t, tc.wantMatch, got)
		})
	}
}

// ---------------------------------------------------------------------------
// errors.As – concrete type extraction
// ---------------------------------------------------------------------------

func TestErrorsAs_NotFoundError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		buildErr   func() error
		wantAs     bool
		wantID     string
		wantCause  error
	}{
		{
			name:      "direct NotFoundError unwrapped",
			buildErr:  func() error { return apperr.NewUserNotFound("abc") },
			wantAs:    true,
			wantID:    "abc",
			wantCause: nil,
		},
		{
			name:      "wrapped in fmt.Errorf still extractable",
			buildErr:  func() error { return fmt.Errorf("svc: %w", apperr.NewUserNotFound("xyz")) },
			wantAs:    true,
			wantID:    "xyz",
			wantCause: nil,
		},
		{
			name:      "with cause extractable",
			buildErr:  func() error { return apperr.NewUserNotFoundWithCause("77", errCause) },
			wantAs:    true,
			wantID:    "77",
			wantCause: errCause,
		},
		{
			name:     "unrelated error not extractable",
			buildErr: func() error { return errors.New("nope") },
			wantAs:   false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.buildErr()
			var nfe *apperr.NotFoundError
			got := errors.As(err, &nfe)

			assert.Equal(t, tc.wantAs, got)
			if tc.wantAs {
				require.NotNil(t, nfe)
				assert.Equal(t, tc.wantID, nfe.ID)
				assert.Equal(t, tc.wantCause, nfe.Cause)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Unwrap chain integrity
// ---------------------------------------------------------------------------

func TestUnwrap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		buildErr      func() *apperr.NotFoundError
		wantUnwrap    error
		wantIsSentinel bool
	}{
		{
			name:           "no cause: unwrap yields sentinel",
			buildErr:       func() *apperr.NotFoundError { return apperr.NewUserNotFound("1") },
			wantUnwrap:     apperr.ErrUserNotFound,
			wantIsSentinel: true,
		},
		{
			name:           "with cause: unwrap yields cause",
			buildErr:       func() *apperr.NotFoundError { return apperr.NewUserNotFoundWithCause("2", errCause) },
			wantUnwrap:     errCause,
			wantIsSentinel: true, // Is() still returns true for sentinel
		},
		{
			name:           "nil cause: unwrap yields sentinel",
			buildErr:       func() *apperr.NotFoundError { return apperr.NewUserNotFoundWithCause("3", nil) },
			wantUnwrap:     apperr.ErrUserNotFound,
			wantIsSentinel: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			nfe := tc.buildErr()
			unwrapped := errors.Unwrap(nfe)
			assert.Equal(t, tc.wantUnwrap, unwrapped)
			assert.Equal(t, tc.wantIsSentinel, errors.Is(nfe, apperr.ErrUserNotFound))
		})
	}
}

// ---------------------------------------------------------------------------
// Error() string formatting
// ---------------------------------------------------------------------------

func TestNotFoundError_ErrorString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		err     *apperr.NotFoundError
		wantMsg string
	}{
		{
			name:    "empty id no cause",
			err:     &apperr.NotFoundError{},
			wantMsg: "user not found",
		},
		{
			name:    "empty id with cause",
			err:     &apperr.NotFoundError{Cause: errCause},
			wantMsg: "user not found: underlying db error",
		},
		{
			name:    "non-empty id no cause",
			err:     &apperr.NotFoundError{ID: "42"},
			wantMsg: `user not found: id "42"`,
		},
		{
			name:    "non-empty id with cause",
			err:     &apperr.NotFoundError{ID: "42", Cause: errCause},
			wantMsg: `user not found: id "42": underlying db error`,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.wantMsg, tc.err.Error())
		})
	}
}

// ---------------------------------------------------------------------------
// Immutability: ID and Cause do not change after construction
// ---------------------------------------------------------------------------

func TestImmutabilityAfterConstruction(t *testing.T) {
	t.Parallel()

	t.Run("NewUserNotFound fields unchanged", func(t *testing.T) {
		t.Parallel()
		err := apperr.NewUserNotFound("immutable-id")
		// Verify the fields are as set
		assert.Equal(t, "immutable-id", err.ID)
		assert.Nil(t, err.Cause)
		// Calling Error() twice yields same result
		assert.Equal(t, err.Error(), err.Error())
	})

	t.Run("NewUserNotFoundWithCause fields unchanged", func(t *testing.T) {
		t.Parallel()
		err := apperr.NewUserNotFoundWithCause("stable-id", errCause)
		assert.Equal(t, "stable-id", err.ID)
		assert.Equal(t, errCause, err.Cause)
		assert.Equal(t, err.Error(), err.Error())
	})
}

// ---------------------------------------------------------------------------
// HTTP handler integration via httptest
// ---------------------------------------------------------------------------

// userService is the interface our HTTP handler depends on.
type userService interface {
	GetUser(id string) (string, error)
}

// mockUserService is an in-memory mock implementing userService.
type mockUserService struct {
	users map[string]string
}

func (m *mockUserService) GetUser(id string) (string, error) {
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return "", apperr.NewUserNotFound(id)
}

// userHandler is a sample HTTP handler that maps NotFoundError → 404.
func userHandler(svc userService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		user, err := svc.GetUser(id)
		if err != nil {
			if errors.Is(err, apperr.ErrUserNotFound) {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, user)
	}
}

func TestUserHandler_HTTPIntegration(t *testing.T) {
	t.Parallel()

	svc := &mockUserService{
		users: map[string]string{
			"1": "Alice",
			"2": "Bob",
		},
	}
	handler := userHandler(svc)

	tests := []struct {
		name           string
		queryID        string
		wantStatusCode int
		wantBodyContains string
	}{
		{
			name:             "existing user returns 200",
			queryID:          "1",
			wantStatusCode:   http.StatusOK,
			wantBodyContains: "Alice",
		},
		{
			name:             "existing user 2 returns 200",
			queryID:          "2",
			wantStatusCode:   http.StatusOK,
			wantBodyContains: "Bob",
		},