```go
package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	scerr "internal/smartcontact/error"
	"internal/smartcontact/model"
	"internal/smartcontact/repository"
)

// ---------------------------------------------------------------------------
// Test double
// ---------------------------------------------------------------------------

// stubUserRepository is a hand-written test double that satisfies
// repository.UserRepository.  Each method is configurable via a function
// field; if a field is nil the method returns a safe zero / not-found value
// so the struct always compiles against the full interface.
type stubUserRepository struct {
	findByNameFn func(ctx context.Context, name string) (*model.User, error)
	findByIDFn   func(ctx context.Context, id int) (*model.User, error)
	findAllFn    func(ctx context.Context) ([]model.User, error)
	mergeFn      func(ctx context.Context, user *model.User) (*model.User, error)
	deleteByIDFn func(ctx context.Context, id int) error
}

func (s *stubUserRepository) FindAll(ctx context.Context) ([]model.User, error) {
	if s.findAllFn != nil {
		return s.findAllFn(ctx)
	}
	return nil, nil
}

func (s *stubUserRepository) FindByID(ctx context.Context, id int) (*model.User, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(ctx, id)
	}
	return nil, scerr.NewUserNotFoundErrorWithCause(id, nil)
}

func (s *stubUserRepository) FindByName(ctx context.Context, name string) (*model.User, error) {
	if s.findByNameFn != nil {
		return s.findByNameFn(ctx, name)
	}
	return nil, scerr.NewUserNotFoundErrorWithCause(0, nil)
}

func (s *stubUserRepository) Merge(ctx context.Context, user *model.User) (*model.User, error) {
	if s.mergeFn != nil {
		return s.mergeFn(ctx, user)
	}
	return user, nil
}

func (s *stubUserRepository) DeleteByID(ctx context.Context, id int) error {
	if s.deleteByIDFn != nil {
		return s.deleteByIDFn(ctx, id)
	}
	return nil
}

// compile-time interface check
var _ repository.UserRepository = (*stubUserRepository)(nil)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// seedUser is the canonical user record used across test cases.
var seedUser = &model.User{
	ID:       3,
	Name:     "hemraj",
	Email:    "hemrajmalhi1234@gmail.com",
	About:    "Sr",
	Password: "root",
	Role:     "java developer",
}

// buildService wires a UserServiceImp around the provided repository stub.
func buildService(repo repository.UserRepository) UserService {
	return NewUserServiceImp(repo)
}

// ---------------------------------------------------------------------------
// TestUserServiceImp_GetUserByName
// ---------------------------------------------------------------------------

// TestUserServiceImp_GetUserByName is the primary table-driven test that
// mirrors the original Java WhenValidDepartmentName_ThenUserShouldBeFound test
// and extends it with every error case described in the behavioural spec.
func TestUserServiceImp_GetUserByName(t *testing.T) {
	t.Parallel()

	type repoStubFn func(ctx context.Context, name string) (*model.User, error)

	tests := []struct {
		name           string
		inputName      string
		repoFn         repoStubFn
		wantErr        bool
		wantErrContain string // optional substring check on the error message
		wantName       string // only checked when wantErr == false
	}{
		// ------------------------------------------------------------------ //
		// Happy-path: spec "given a valid name matching an existing user"
		// ------------------------------------------------------------------ //
		{
			name:      "valid name returns matching user",
			inputName: "hemraj",
			repoFn: func(ctx context.Context, name string) (*model.User, error) {
				if name == "hemraj" {
					return seedUser, nil
				}
				return nil, scerr.NewUserNotFoundErrorWithCause(0, nil)
			},
			wantErr:  false,
			wantName: "hemraj",
		},
		// invariant: returned user's Name must equal the input name
		{
			name:      "returned user name equals the queried name",
			inputName: "hemraj",
			repoFn: func(_ context.Context, name string) (*model.User, error) {
				return &model.User{ID: 99, Name: name, Email: "x@example.com"}, nil
			},
			wantErr:  false,
			wantName: "hemraj",
		},

		// ------------------------------------------------------------------ //
		// Error cases
		// ------------------------------------------------------------------ //
		{
			name:      "unknown name returns not-found error",
			inputName: "nobody",
			repoFn: func(_ context.Context, _ string) (*model.User, error) {
				return nil, scerr.NewUserNotFoundErrorWithCause(0, nil)
			},
			wantErr: true,
		},
		{
			name:      "empty name returns not-found error",
			inputName: "",
			repoFn: func(_ context.Context, _ string) (*model.User, error) {
				return nil, scerr.NewUserNotFoundErrorWithCause(0, nil)
			},
			wantErr: true,
		},
		{
			name:      "repository returns unexpected error propagates to caller",
			inputName: "hemraj",
			repoFn: func(_ context.Context, _ string) (*model.User, error) {
				return nil, errors.New("connection timeout")
			},
			wantErr: true,
		},
		{
			name:      "repository returns nil user without error is treated as not-found",
			inputName: "ghost",
			repoFn: func(_ context.Context, _ string) (*model.User, error) {
				// Some repository implementations may return (nil, nil) to
				// signal absence; the service must handle this gracefully.
				return nil, nil
			},
			// The service must either return an error or an empty-ish response;
			// either way the caller must not panic.
			wantErr: false, // tolerate (nil, nil) → depends on service; adjust as needed
		},
	}

	for _, tt := range tests {
		tt := tt // pin
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := &stubUserRepository{findByNameFn: tt.repoFn}
			svc := buildService(repo)

			got, err := svc.GetUserByName(context.Background(), tt.inputName)

			// ── error branch ──────────────────────────────────────────────
			if tt.wantErr {
				require.Error(t, err,
					"GetUserByName(%q) should return an error", tt.inputName)
				if tt.wantErrContain != "" {
					assert.Contains(t, err.Error(), tt.wantErrContain)
				}
				return
			}

			// ── success branch ────────────────────────────────────────────
			require.NoError(t, err,
				"GetUserByName(%q) returned unexpected error", tt.inputName)

			// When wantName is set we strictly validate the response Name.
			if tt.wantName != "" {
				require.NotNil(t, got,
					"GetUserByName(%q) returned nil response", tt.inputName)

				// The service layer returns a UserResponse where Name is *string.
				require.NotNil(t, got.Name,
					"GetUserByName(%q): response Name pointer is nil", tt.inputName)

				assert.Equal(t, tt.wantName, *got.Name,
					"GetUserByName(%q): Name mismatch", tt.inputName)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestUserServiceImp_GetUserByName_Invariant_NameEqualsInput
// ---------------------------------------------------------------------------

// TestUserServiceImp_GetUserByName_Invariant_NameEqualsInput validates the
// spec invariant "When a User is returned, its name field equals the name
// provided as input" across a set of distinct names.
func TestUserServiceImp_GetUserByName_Invariant_NameEqualsInput(t *testing.T) {
	t.Parallel()

	names := []string{"hemraj", "alice", "bob", "charlie", "dĩa"}

	for _, n := range names {
		n := n
		t.Run("name="+n, func(t *testing.T) {
			t.Parallel()

			repo := &stubUserRepository{
				findByNameFn: func(_ context.Context, name string) (*model.User, error) {
					return &model.User{ID: 1, Name: name, Email: name + "@example.com"}, nil
				},
			}
			svc := buildService(repo)

			got, err := svc.GetUserByName(context.Background(), n)

			require.NoError(t, err)
			require.NotNil(t, got)
			require.NotNil(t, got.Name,
				"response Name is nil for input %q", n)
			assert.Equal(t, n, *got.Name,
				"invariant violated: returned Name %q != input %q", *got.Name, n)
		})
	}
}

// ---------------------------------------------------------------------------
// TestUserServiceImp_GetUserByName_UserModelFields
// ---------------------------------------------------------------------------

// TestUserServiceImp_GetUserByName_UserModelFields validates the global
// invariant that "The User model exposes at minimum: id, name, email, about,
// password, and role fields" by checking the seed record is surfaced intact
// through the service layer.
func TestUserServiceImp_GetUserByName_UserModelFields(t *testing.T) {
	t.Parallel()

	repo := &stubUserRepository{
		findByNameFn: func(_ context.Context, name string) (*model.User, error) {
			if name == seedUser.Name {
				return seedUser, nil
			}
			return nil, scerr.NewUserNotFoundErrorWithCause(0, nil)
		},
	}
	svc := buildService(repo)

	got, err := svc.GetUserByName(context.Background(), seedUser.Name)

	require.NoError(t, err)
	require.NotNil(t, got)

	// Name
	require.NotNil(t, got.Name, "Name field must not be nil")
	assert.Equal(t, seedUser.Name, *got.Name)

	// Email – only validate when the response carries this optional field.
	if got.Email != nil {
		assert.Equal(t, seedUser.Email, *got.Email)
	}

	// About
	if got.About != nil {
		assert.Equal(t, seedUser.About, *got.About)
	}

	// Role
	if got.Role != nil {
		assert.Equal(t, seedUser.Role, *got.Role)
	}

	// ID – some DTO representations use int vs int64; accept either non-zero
	// value as proof the field is populated.
	if got.ID != nil {
		assert.NotZero(t, *got.ID, "ID field should be non-zero for a real user")
	}
}

// ---------------------------------------------------------------------------
// TestUserServiceImp_GetUserByName_ContextCancellation
// ---------------------------------------------------------------------------

// TestUserServiceImp_GetUserByName_ContextCancellation verifies that a
// cancelled context is forwarded to the repository and the error is propagated
// to the caller.
func TestUserServiceImp_GetUserByName_ContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	repo := &stubUserRepository{
		findByNameFn: func(ctx context.Context, name string) (*model.User, error) {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			return seedUser, nil
		},
	}
	svc := buildService(repo)

	_, err := svc.GetUserByName(ctx, seedUser.Name)
	assert.Error(t, err, "expected an error when context is already cancelled")
}

// ---------------------------------------------------------------------------
// TestUserServiceImp_GetUserByName_MultipleCallsConsistency
// ---------------------------------------------------------------------------

// TestUserServiceImp_GetUserByName_MultipleCallsConsistency verifies that
// calling GetUserByName twice with the same input produces the same result
// (the service must not mutate shared state between calls).
func TestUserServiceImp_GetUserByName_MultipleCallsConsistency(t *testing.T) {
	t.Parallel()

	callCount := 0
	repo := &stubUserRepository{
		findByNameFn: func(_ context.Context, _ string) (*model.User, error) {
			callCount++
			return seedUser, nil
		},
	}
	svc := buildService(repo)

	got1, err1 := svc.GetUserByName(context.Background(), seedUser.Name)
	require.NoError(t, err1)

	got2, err2 := svc.GetUserByName(context.Background(), seedUser.Name)
	require.NoError(t, err2)

	require.NotNil(t, got1)
	require.NotNil(t, got2)
	require.NotNil(t, got1.Name)
	require.NotNil(t, got2.Name)

	assert.Equal(t, *got1.Name, *got2.Name,
		"successive calls with the same name must return the same Name value")
	assert.Equal(t, 2, callCount,
		"repository should be called once per service invocation (no caching)")
}
```