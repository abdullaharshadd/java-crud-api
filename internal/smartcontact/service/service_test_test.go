```go
package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	smartError "github.com/smartContact/internal/smartcontact/error"
	"github.com/smartContact/internal/smartcontact/model"
	"github.com/smartContact/internal/smartcontact/repository"
)

// ---------------------------------------------------------------------------
// fakeQuerier – hand-written test double implementing repository.Querier.
// ---------------------------------------------------------------------------

type fakeQuerier struct {
	findByNameFn func(ctx context.Context, name string) (model.User, error)
	findByIDFn   func(ctx context.Context, id int64) (model.User, error)
	findAllFn    func(ctx context.Context) ([]model.User, error)
	saveFn       func(ctx context.Context, u model.User) (model.User, error)
	deleteByIDFn func(ctx context.Context, id int64) error
}

func (f *fakeQuerier) FindByName(ctx context.Context, name string) (model.User, error) {
	if f.findByNameFn != nil {
		return f.findByNameFn(ctx, name)
	}
	return model.User{}, smartError.NewUserNotFoundErrorf("user %q not found", name)
}

func (f *fakeQuerier) FindByID(ctx context.Context, id int64) (model.User, error) {
	if f.findByIDFn != nil {
		return f.findByIDFn(ctx, id)
	}
	return model.User{}, smartError.NewUserNotFoundErrorf("user %d not found", id)
}

func (f *fakeQuerier) FindAll(ctx context.Context) ([]model.User, error) {
	if f.findAllFn != nil {
		return f.findAllFn(ctx)
	}
	return nil, nil
}

func (f *fakeQuerier) Save(ctx context.Context, u model.User) (model.User, error) {
	if f.saveFn != nil {
		return f.saveFn(ctx, u)
	}
	return u, nil
}

func (f *fakeQuerier) DeleteByID(ctx context.Context, id int64) error {
	if f.deleteByIDFn != nil {
		return f.deleteByIDFn(ctx, id)
	}
	return nil
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func newService(q *fakeQuerier) UserService {
	return NewUserService(repository.NewUserRepository(q))
}

// seededUser returns the canonical test user used across multiple test cases.
func seededUser() model.User {
	return model.NewUser(
		3,
		"hemraj",
		"hemrajmalhi1234@gmail.com",
		"Sr",
		"root",
		"java developer",
	)
}

// ---------------------------------------------------------------------------
// TestGetUserByName_WhenValidName_ThenUserShouldBeFound
// Direct migration of the Java WhenValidDepartmentName_ThenUserShouldBeFound.
// ---------------------------------------------------------------------------

func TestGetUserByName_WhenValidName_ThenUserShouldBeFound(t *testing.T) {
	t.Parallel()

	const validName = "hemraj"
	user := seededUser()

	tests := []struct {
		name        string
		lookupName  string
		querier     *fakeQuerier
		wantName    string
		wantErr     bool
		wantNotFnd  bool
		wantWrapped error
	}{
		{
			name:       "valid name returns matching user",
			lookupName: validName,
			querier: &fakeQuerier{
				findByNameFn: func(_ context.Context, name string) (model.User, error) {
					if name == validName {
						return user, nil
					}
					return model.User{}, smartError.NewUserNotFoundErrorf("user %q not found", name)
				},
			},
			wantName: validName,
		},
		{
			name:       "unknown name returns not-found error",
			lookupName: "nobody",
			querier: &fakeQuerier{
				findByNameFn: func(_ context.Context, name string) (model.User, error) {
					return model.User{}, smartError.NewUserNotFoundErrorf("user %q not found", name)
				},
			},
			wantErr:    true,
			wantNotFnd: true,
		},
		{
			name:       "empty name returns not-found error",
			lookupName: "",
			querier: &fakeQuerier{
				findByNameFn: func(_ context.Context, name string) (model.User, error) {
					return model.User{}, smartError.NewUserNotFoundErrorf("user %q not found", name)
				},
			},
			wantErr:    true,
			wantNotFnd: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := newService(tt.querier)
			got, err := svc.GetUserByName(context.Background(), tt.lookupName)

			if tt.wantErr {
				require.Error(t, err, "expected an error but got nil")
				if tt.wantNotFnd {
					assert.True(t, smartError.IsUserNotFound(err),
						"expected user-not-found error, got: %v", err)
				}
				if tt.wantWrapped != nil {
					assert.True(t, errors.Is(err, tt.wantWrapped),
						"expected error to wrap %v, got: %v", tt.wantWrapped, err)
				}
				return
			}

			require.NoError(t, err, "unexpected error for name %q", tt.lookupName)

			// Invariant: returned User's name equals the requested name.
			assert.Equal(t, tt.wantName, got.Name,
				"User.Name mismatch: want %q, got %q", tt.wantName, got.Name)
		})
	}
}

// ---------------------------------------------------------------------------
// TestGetUserByName_ReturnedUserFields
// Global invariant: a retrieved User exposes accessible fields including
// name, email, about, password, role, and id.
// ---------------------------------------------------------------------------

func TestGetUserByName_ReturnedUserFields(t *testing.T) {
	t.Parallel()

	user := seededUser()

	tests := []struct {
		name       string
		lookupName string
		seeded     model.User
		wantID     int64
		wantName   string
		wantEmail  string
		wantAbout  string
		wantPass   string
		wantRole   string
	}{
		{
			name:       "all user fields are accessible after lookup",
			lookupName: "hemraj",
			seeded:     user,
			wantID:     3,
			wantName:   "hemraj",
			wantEmail:  "hemrajmalhi1234@gmail.com",
			wantAbout:  "Sr",
			wantPass:   "root",
			wantRole:   "java developer",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := newService(&fakeQuerier{
				findByNameFn: func(_ context.Context, name string) (model.User, error) {
					if name == tt.lookupName {
						return tt.seeded, nil
					}
					return model.User{}, smartError.NewUserNotFoundErrorf("user %q not found", name)
				},
			})

			got, err := svc.GetUserByName(context.Background(), tt.lookupName)
			require.NoError(t, err)

			assert.Equal(t, tt.wantID, got.ID, "User.ID")
			assert.Equal(t, tt.wantName, got.Name, "User.Name")
			assert.Equal(t, tt.wantEmail, got.Email, "User.Email")
			assert.Equal(t, tt.wantAbout, got.About, "User.About")
			assert.Equal(t, tt.wantPass, got.Password, "User.Password")
			assert.Equal(t, tt.wantRole, got.Role, "User.Role")
		})
	}
}

// ---------------------------------------------------------------------------
// TestGetUserByName_Consistency
// Global invariant: querying with the same name twice yields the same result.
// ---------------------------------------------------------------------------

func TestGetUserByName_Consistency(t *testing.T) {
	t.Parallel()

	user := seededUser()
	const lookupName = "hemraj"

	tests := []struct {
		name       string
		lookupName string
		seeded     model.User
	}{
		{
			name:       "same name queried twice returns equal users",
			lookupName: lookupName,
			seeded:     user,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := newService(&fakeQuerier{
				findByNameFn: func(_ context.Context, name string) (model.User, error) {
					if name == tt.lookupName {
						return tt.seeded, nil
					}
					return model.User{}, smartError.NewUserNotFoundErrorf("user %q not found", name)
				},
			})

			ctx := context.Background()
			first, err := svc.GetUserByName(ctx, tt.lookupName)
			require.NoError(t, err, "first call")

			second, err := svc.GetUserByName(ctx, tt.lookupName)
			require.NoError(t, err, "second call")

			assert.Equal(t, first, second,
				"consistency invariant violated: two lookups of %q returned different results", tt.lookupName)
		})
	}
}

// ---------------------------------------------------------------------------
// TestGetUserByName_PropagatesRepositoryError
// Infrastructure errors must not be swallowed or misclassified as not-found.
// ---------------------------------------------------------------------------

func TestGetUserByName_PropagatesRepositoryError(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("connection reset by peer")

	tests := []struct {
		name        string
		lookupName  string
		repoErr     error
		wantNotFnd  bool
		wantWrapped error
	}{
		{
			name:        "infrastructure error is surfaced unchanged",
			lookupName:  "hemraj",
			repoErr:     sentinel,
			wantNotFnd:  false,
			wantWrapped: sentinel,
		},
		{
			name:       "timeout error is propagated",
			lookupName: "hemraj",
			repoErr:    context.DeadlineExceeded,
			wantNotFnd: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := newService(&fakeQuerier{
				findByNameFn: func(_ context.Context, _ string) (model.User, error) {
					return model.User{}, tt.repoErr
				},
			})

			_, err := svc.GetUserByName(context.Background(), tt.lookupName)

			require.Error(t, err, "expected error, got nil")
			assert.False(t, smartError.IsUserNotFound(err),
				"infrastructure error must not be classified as not-found: %v", err)

			if tt.wantWrapped != nil {
				assert.True(t, errors.Is(err, tt.wantWrapped),
					"expected error to wrap %v, got: %v", tt.wantWrapped, err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestGetUserByName_NotFoundDistinctFromInfraError
// A not-found error must be distinguishable from a generic infrastructure error.
// ---------------------------------------------------------------------------

func TestGetUserByName_NotFoundDistinctFromInfraError(t *testing.T) {
	t.Parallel()

	infraErr := errors.New("dial tcp: connection refused")

	tests := []struct {
		name       string
		querier    *fakeQuerier
		lookupName string
		wantNotFnd bool
	}{
		{
			name:       "user-not-found error is identified as not-found",
			lookupName: "ghost",
			querier: &fakeQuerier{
				findByNameFn: func(_ context.Context, name string) (model.User, error) {
					return model.User{}, smartError.NewUserNotFoundErrorf("user %q not found", name)
				},
			},
			wantNotFnd: true,
		},
		{
			name:       "infrastructure error is NOT identified as not-found",
			lookupName: "hemraj",
			querier: &fakeQuerier{
				findByNameFn: func(_ context.Context, _ string) (model.User, error) {
					return model.User{}, infraErr
				},
			},
			wantNotFnd: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := newService(tt.querier)
			_, err := svc.GetUserByName(context.Background(), tt.lookupName)

			require.Error(t, err)
			assert.Equal(t, tt.wantNotFnd, smartError.IsUserNotFound(err),
				"IsUserNotFound classification mismatch for error: %v", err)
		})
	}
}
```