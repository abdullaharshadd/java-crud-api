```go
package service

import (
	"context"
	"errors"
	"testing"

	domainerr "github.com/smartContact/internal/smartcontact/error"
	"github.com/smartContact/internal/smartcontact/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// seedUser is the canonical test fixture mirroring the Java @BeforeAll seed.
var seedUser = model.User{
	ID:       3,
	Name:     "hemraj",
	Email:    "hemrajmalhi1234@gmail.com",
	About:    "Sr",
	Password: "root",
	Role:     "java developer",
}

// newPopulatedRepo returns a fakeUserRepository pre-loaded with seedUser so
// every test case starts from the same baseline.
func newPopulatedRepo() *fakeUserRepository {
	return &fakeUserRepository{users: []model.User{seedUser}}
}

// ---------------------------------------------------------------------------
// TestGetUserByName – table-driven tests for UserService.GetUserByName
// ---------------------------------------------------------------------------

func TestGetUserByName_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		query       string
		wantName    string
		wantFound   bool
		wantErrType error // optional: check error identity via errors.Is
	}{
		// ---------------------------------------------------------------
		// Happy-path: spec "given a valid existing user name ('hemraj')"
		// ---------------------------------------------------------------
		{
			name:      "valid existing name returns matching user",
			query:     "hemraj",
			wantName:  "hemraj",
			wantFound: true,
		},
		// ---------------------------------------------------------------
		// Error-path: spec "given a name that does not correspond to any user"
		// ---------------------------------------------------------------
		{
			name:        "unknown name returns not-found error",
			query:       "nobody",
			wantFound:   false,
			wantErrType: domainerr.ErrUserNotFound,
		},
		{
			name:        "empty string returns not-found error",
			query:       "",
			wantFound:   false,
			wantErrType: domainerr.ErrUserNotFound,
		},
		{
			name:        "name with different casing returns not-found error",
			query:       "Hemraj",
			wantFound:   false,
			wantErrType: domainerr.ErrUserNotFound,
		},
		{
			name:        "partial name returns not-found error",
			query:       "hem",
			wantFound:   false,
			wantErrType: domainerr.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := newPopulatedRepo()
			svc := NewUserService(repo)

			got, err := svc.GetUserByName(context.Background(), tt.query)

			if tt.wantFound {
				// Spec invariant: "returns a User object whose name equals the
				// queried name"
				require.NoError(t, err, "expected no error for a valid user name")
				assert.Equal(t, tt.wantName, got.Name,
					"returned user's name must equal the queried name")

				// Global invariant: identity fields must round-trip correctly.
				assert.Equal(t, seedUser.ID, got.ID,
					"returned user ID must match stored ID")
				assert.Equal(t, seedUser.Email, got.Email,
					"returned user email must match stored email")
				return
			}

			// Error cases -------------------------------------------------------
			require.Error(t, err,
				"expected an error when querying for a non-existent user")
			assert.Equal(t, model.User{}, got,
				"zero-value User should be returned on error")

			if tt.wantErrType != nil {
				assert.True(t, errors.Is(err, tt.wantErrType),
					"error must wrap %v, got: %v", tt.wantErrType, err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestGetUserByName_Determinism – invariant: lookup is deterministic
// ---------------------------------------------------------------------------

func TestGetUserByName_Determinism(t *testing.T) {
	t.Parallel()

	repo := newPopulatedRepo()
	svc := NewUserService(repo)

	const runs = 5
	for i := 0; i < runs; i++ {
		got, err := svc.GetUserByName(context.Background(), "hemraj")
		require.NoError(t, err, "run %d: unexpected error", i)
		assert.Equal(t, "hemraj", got.Name, "run %d: name mismatch", i)
		assert.Equal(t, seedUser.ID, got.ID, "run %d: ID mismatch", i)
	}
}

// ---------------------------------------------------------------------------
// TestGetUserByName_EmptyRepository
// ---------------------------------------------------------------------------

func TestGetUserByName_EmptyRepository(t *testing.T) {
	t.Parallel()

	repo := &fakeUserRepository{users: []model.User{}} // no seed
	svc := NewUserService(repo)

	got, err := svc.GetUserByName(context.Background(), "hemraj")

	require.Error(t, err, "empty repo must return an error")
	assert.Equal(t, model.User{}, got, "zero-value User expected on error")
	assert.True(t, errors.Is(err, domainerr.ErrUserNotFound),
		"error must be ErrUserNotFound, got: %v", err)
}

// ---------------------------------------------------------------------------
// TestGetUserByName_MultipleUsers – lookup stays accurate with many users
// ---------------------------------------------------------------------------

func TestGetUserByName_MultipleUsers(t *testing.T) {
	t.Parallel()

	extra := []model.User{
		{ID: 1, Name: "alice", Email: "alice@example.com"},
		{ID: 2, Name: "bob", Email: "bob@example.com"},
		seedUser,
		{ID: 4, Name: "dave", Email: "dave@example.com"},
	}
	repo := &fakeUserRepository{users: extra}
	svc := NewUserService(repo)

	tests := []struct {
		query    string
		wantName string
		wantID   int
	}{
		{"alice", "alice", 1},
		{"bob", "bob", 2},
		{"hemraj", "hemraj", 3},
		{"dave", "dave", 4},
	}

	for _, tt := range tests {
		tt := tt
		t.Run("lookup_"+tt.query, func(t *testing.T) {
			t.Parallel()

			got, err := svc.GetUserByName(context.Background(), tt.query)
			require.NoError(t, err)
			assert.Equal(t, tt.wantName, got.Name)
			assert.Equal(t, tt.wantID, got.ID)
		})
	}
}

// ---------------------------------------------------------------------------
// TestGetUserByName_IdentityConsistency – global invariant
// ---------------------------------------------------------------------------

func TestGetUserByName_IdentityConsistency(t *testing.T) {
	t.Parallel()

	repo := newPopulatedRepo()
	svc := NewUserService(repo)

	got, err := svc.GetUserByName(context.Background(), seedUser.Name)
	require.NoError(t, err)

	// All identity fields must be consistent between storage and retrieval.
	assert.Equal(t, seedUser.Name, got.Name, "name must be consistent")
	assert.Equal(t, seedUser.Email, got.Email, "email must be consistent")
	assert.Equal(t, seedUser.ID, got.ID, "id must be consistent")
	assert.Equal(t, seedUser.About, got.About, "about must be consistent")
	assert.Equal(t, seedUser.Role, got.Role, "role must be consistent")
}
```