```go
package service_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	smartError "github.com/smartContact/internal/smartcontact/error"
	"github.com/smartContact/internal/smartcontact/model"
	"github.com/smartContact/internal/smartcontact/repository"
	"github.com/smartContact/internal/smartcontact/service"
)

// ---------------------------------------------------------------------------
// Mock repository
// ---------------------------------------------------------------------------

type mockUserRepository struct {
	// SaveFn is called by Save.
	SaveFn func(ctx context.Context, user model.User) (*model.User, error)
	// FindAllFn is called by FindAll.
	FindAllFn func(ctx context.Context) ([]model.User, error)
	// FindByIDFn is called by FindByID.
	FindByIDFn func(ctx context.Context, id int) (*model.User, error)
	// DeleteByIDFn is called by DeleteByID.
	DeleteByIDFn func(ctx context.Context, id int) error
	// FindByNameFn is called by FindByName.
	FindByNameFn func(ctx context.Context, name string) (*model.User, error)
}

func (m *mockUserRepository) Save(ctx context.Context, user model.User) (*model.User, error) {
	if m.SaveFn != nil {
		return m.SaveFn(ctx, user)
	}
	return nil, errors.New("mock: Save not configured")
}

func (m *mockUserRepository) FindAll(ctx context.Context) ([]model.User, error) {
	if m.FindAllFn != nil {
		return m.FindAllFn(ctx)
	}
	return nil, errors.New("mock: FindAll not configured")
}

func (m *mockUserRepository) FindByID(ctx context.Context, id int) (*model.User, error) {
	if m.FindByIDFn != nil {
		return m.FindByIDFn(ctx, id)
	}
	return nil, errors.New("mock: FindByID not configured")
}

func (m *mockUserRepository) DeleteByID(ctx context.Context, id int) error {
	if m.DeleteByIDFn != nil {
		return m.DeleteByIDFn(ctx, id)
	}
	return errors.New("mock: DeleteByID not configured")
}

func (m *mockUserRepository) FindByName(ctx context.Context, name string) (*model.User, error) {
	if m.FindByNameFn != nil {
		return m.FindByNameFn(ctx, name)
	}
	return nil, errors.New("mock: FindByName not configured")
}

// Ensure the mock satisfies the interface at compile-time.
var _ repository.UserRepository = (*mockUserRepository)(nil)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// newService creates a UserService backed by the provided mock repository.
// It fails the test immediately if construction fails.
func newService(t *testing.T, repo repository.UserRepository) service.UserService {
	t.Helper()
	svc, err := service.NewUserService(repo)
	require.NoError(t, err)
	return svc
}

// validUser returns a minimal User that passes Validate().
func validUser(id int, name string) model.User {
	return model.User{
		ID:   id,
		Name: name,
	}
}

// ---------------------------------------------------------------------------
// TestNewUserService
// ---------------------------------------------------------------------------

func TestNewUserService(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		repo    repository.UserRepository
		wantErr bool
	}{
		{
			name:    "nil repository returns error",
			repo:    nil,
			wantErr: true,
		},
		{
			name:    "valid repository constructs service",
			repo:    &mockUserRepository{},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			svc, err := service.NewUserService(tc.repo)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, svc)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, svc)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestSaveUser
// ---------------------------------------------------------------------------

func TestSaveUser(t *testing.T) {
	t.Parallel()

	persistErr := errors.New("db: constraint violation")

	tests := []struct {
		name       string
		inputUser  model.User
		repoSaveFn func(ctx context.Context, user model.User) (*model.User, error)
		wantUser   *model.User
		wantErr    bool
		errContain string
	}{
		{
			name:      "given a valid user, repo persists and returns the user with generated ID",
			inputUser: validUser(0, "Alice"),
			repoSaveFn: func(_ context.Context, user model.User) (*model.User, error) {
				// Simulate DB generating ID = 1.
				saved := user
				saved.ID = 1
				return &saved, nil
			},
			wantUser: &model.User{ID: 1, Name: "Alice"},
			wantErr:  false,
		},
		{
			name:      "given a valid user, returned user reflects persisted state",
			inputUser: validUser(0, "Bob"),
			repoSaveFn: func(_ context.Context, user model.User) (*model.User, error) {
				saved := user
				saved.ID = 42
				return &saved, nil
			},
			wantUser: &model.User{ID: 42, Name: "Bob"},
			wantErr:  false,
		},
		{
			name:      "given repo persistence error, save propagates error",
			inputUser: validUser(0, "Charlie"),
			repoSaveFn: func(_ context.Context, _ model.User) (*model.User, error) {
				return nil, persistErr
			},
			wantUser:   nil,
			wantErr:    true,
			errContain: "save user",
		},
		{
			name:       "given invalid user (empty name), Validate rejects before repo call",
			inputUser:  model.User{}, // blank name – assumed invalid
			repoSaveFn: nil,          // must not be called
			wantUser:   nil,
			wantErr:    true,
			errContain: "save user",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := &mockUserRepository{SaveFn: tc.repoSaveFn}
			svc := newService(t, repo)

			got, err := svc.SaveUser(context.Background(), tc.inputUser)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.errContain != "" {
					assert.Contains(t, err.Error(), tc.errContain)
				}
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, tc.wantUser.ID, got.ID)
				assert.Equal(t, tc.wantUser.Name, got.Name)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestFetchUserList
// ---------------------------------------------------------------------------

func TestFetchUserList(t *testing.T) {
	t.Parallel()

	repoErr := errors.New("db: connection refused")

	tests := []struct {
		name         string
		repoFindAllFn func(ctx context.Context) ([]model.User, error)
		wantUsers    []model.User
		wantErr      bool
		errContain   string
	}{
		{
			name: "given users exist, returns all users",
			repoFindAllFn: func(_ context.Context) ([]model.User, error) {
				return []model.User{
					validUser(1, "Alice"),
					validUser(2, "Bob"),
				}, nil
			},
			wantUsers: []model.User{
				validUser(1, "Alice"),
				validUser(2, "Bob"),
			},
			wantErr: false,
		},
		{
			name: "given no users exist, returns empty slice",
			repoFindAllFn: func(_ context.Context) ([]model.User, error) {
				return []model.User{}, nil
			},
			wantUsers: []model.User{},
			wantErr:   false,
		},
		{
			name: "given nil returned by repo, returns nil slice without error",
			repoFindAllFn: func(_ context.Context) ([]model.User, error) {
				return nil, nil
			},
			wantUsers: nil,
			wantErr:   false,
		},
		{
			name: "given repo error, propagates error",
			repoFindAllFn: func(_ context.Context) ([]model.User, error) {
				return nil, repoErr
			},
			wantUsers:  nil,
			wantErr:    true,
			errContain: "fetch user list",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := &mockUserRepository{FindAllFn: tc.repoFindAllFn}
			svc := newService(t, repo)

			got, err := svc.FetchUserList(context.Background())

			if tc.wantErr {
				assert.Error(t, err)
				if tc.errContain != "" {
					assert.Contains(t, err.Error(), tc.errContain)
				}
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.wantUsers, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestFetchUserByID
// ---------------------------------------------------------------------------

func TestFetchUserByID(t *testing.T) {
	t.Parallel()

	notFoundErr := smartError.NewUserNotFoundErrorf("user 99 not found")

	tests := []struct {
		name           string
		id             int
		repoFindByIDFn func(ctx context.Context, id int) (*model.User, error)
		wantUser       *model.User
		wantErr        bool
		wantNotFound   bool
		errContain     string
	}{
		{
			name: "given existing id, returns matching user",
			id:   1,
			repoFindByIDFn: func(_ context.Context, id int) (*model.User, error) {
				u := validUser(id, "Alice")
				return &u, nil
			},
			wantUser: func() *model.User { u := validUser(1, "Alice"); return &u }(),
			wantErr:  false,
		},
		{
			name: "returned user ID matches requested ID",
			id:   7,
			repoFindByIDFn: func(_ context.Context, id int) (*model.User, error) {
				u := validUser(id, "Dave")
				return &u, nil
			},
			wantUser: func() *model.User { u := validUser(7, "Dave"); return &u }(),
			wantErr:  false,
		},
		{
			name: "given non-existent id, returns UserNotFound error",
			id:   99,
			repoFindByIDFn: func(_ context.Context, _ int) (*model.User, error) {
				return nil, notFoundErr
			},
			wantUser:     nil,
			wantErr:      true,
			wantNotFound: true,
			errContain:   "fetch user by id",
		},
		{
			name: "given repo error other than not-found, propagates error",
			id:   5,
			repoFindByIDFn: func(_ context.Context, _ int) (*model.User, error) {
				return nil, errors.New("db: unexpected error")
			},
			wantUser:   nil,
			wantErr:    true,
			errContain: "fetch user by id",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := &mockUserRepository{FindByIDFn: tc.repoFindByIDFn}
			svc := newService(t, repo)

			got, err := svc.FetchUserByID(context.Background(), tc.id)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.errContain != "" {
					assert.Contains(t, err.Error(), tc.errContain)
				}
				assert.Nil(t, got)
				if tc.wantNotFound {
					assert.True(t, errors.Is(err, smartError.ErrUserNotFound),
						"expected ErrUserNotFound to be in the error chain")
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, tc.id, got.ID)
				if tc.wantUser != nil {
					assert.Equal(t, tc.wantUser.Name, got.Name)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestDeleteUser
// ---------------------------------------------------------------------------

func TestDeleteUser(t *testing.T) {
	t.Parallel()

	notFoundErr := smartError.NewUserNotFoundErrorf("user 99 not found")

	tests := []struct {
		name             string
		id               int
		repoDeleteByIDFn func(ctx context.Context, id int) error
		wantErr          bool
		wantNotFound     bool
		errContain       string
	}{
		{
			name: "given existing id, deletes user without error",
			id:   1,
			repoDeleteByIDFn: func(_ context.Context, _ int) error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "given non-existent id, repo returns not-found error",
			id:   99,
			repoDeleteByIDFn: func(_ context.Context, _ int) error {
				return notFoundErr
			},
			wantErr:      true,
			wantNotFound: true,
			errContain:   "delete user",
		},
		{
			name: "given repo error, propagates error",
			id:   3,
			repoDeleteByIDFn: func(_ context.Context, _ int) error {
				return errors.New("db: delete failed")
			},
			wantErr:    true,
			errContain: "delete user",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := &mockUserRepository{DeleteByIDFn: tc.repoDeleteByIDFn}
			svc := newService(t, repo)

			err := svc.DeleteUser(context.Background(), tc.id)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.errContain != "" {
					assert.Contains(t, err.Error(), tc.errContain)
				}
				if tc.wantNotFound {
					assert.True(t, errors.Is(err, smartError.ErrUserNotFound),
						"expected ErrUserNotFound in error chain")
				}
			} else {
				assert.NoError(t, err)
			}