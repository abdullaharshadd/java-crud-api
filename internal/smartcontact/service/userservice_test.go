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
	"github.com/smartContact/internal/smartcontact/service"
)

// ---------------------------------------------------------------------------
// Mock repository
// ---------------------------------------------------------------------------

// mockUserRepository is a test double for repository.UserRepository.
type mockUserRepository struct {
	// SaveFn is called by Save.
	SaveFn func(ctx context.Context, user model.User) (model.User, error)
	// FindAllFn is called by FindAll.
	FindAllFn func(ctx context.Context) ([]model.User, error)
	// FindByIDFn is called by FindByID.
	FindByIDFn func(ctx context.Context, id int) (model.User, error)
	// DeleteByIDFn is called by DeleteByID.
	DeleteByIDFn func(ctx context.Context, id int) error
	// FindByNameFn is called by FindByName.
	FindByNameFn func(ctx context.Context, name string) (model.User, error)
}

func (m *mockUserRepository) Save(ctx context.Context, user model.User) (model.User, error) {
	if m.SaveFn != nil {
		return m.SaveFn(ctx, user)
	}
	return model.User{}, errors.New("SaveFn not set")
}

func (m *mockUserRepository) FindAll(ctx context.Context) ([]model.User, error) {
	if m.FindAllFn != nil {
		return m.FindAllFn(ctx)
	}
	return nil, errors.New("FindAllFn not set")
}

func (m *mockUserRepository) FindByID(ctx context.Context, id int) (model.User, error) {
	if m.FindByIDFn != nil {
		return m.FindByIDFn(ctx, id)
	}
	return model.User{}, errors.New("FindByIDFn not set")
}

func (m *mockUserRepository) DeleteByID(ctx context.Context, id int) error {
	if m.DeleteByIDFn != nil {
		return m.DeleteByIDFn(ctx, id)
	}
	return errors.New("DeleteByIDFn not set")
}

func (m *mockUserRepository) FindByName(ctx context.Context, name string) (model.User, error) {
	if m.FindByNameFn != nil {
		return m.FindByNameFn(ctx, name)
	}
	return model.User{}, errors.New("FindByNameFn not set")
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// newService builds a userService backed by the supplied mock.
func newService(repo *mockUserRepository) service.UserService {
	return service.NewUserService(repo)
}

// validUser returns a User that passes Validate().
func validUser(id int, name string) model.User {
	return model.User{
		ID:   id,
		Name: name,
	}
}

// sentinelErr is a generic persistence error for negative-path tests.
var sentinelErr = errors.New("db error")

// ---------------------------------------------------------------------------
// SaveUser
// ---------------------------------------------------------------------------

func TestSaveUser(t *testing.T) {
	tests := []struct {
		name        string
		inputUser   model.User
		saveFn      func(ctx context.Context, u model.User) (model.User, error)
		wantUser    model.User
		wantErr     bool
		wantErrMsg  string
	}{
		{
			name:      "valid user is persisted and returned with generated id",
			inputUser: validUser(0, "Alice"),
			saveFn: func(ctx context.Context, u model.User) (model.User, error) {
				u.ID = 42
				return u, nil
			},
			wantUser: validUser(42, "Alice"),
			wantErr:  false,
		},
		{
			name:      "persistence error is propagated",
			inputUser: validUser(0, "Bob"),
			saveFn: func(ctx context.Context, u model.User) (model.User, error) {
				return model.User{}, sentinelErr
			},
			wantErr:    true,
			wantErrMsg: "save user",
		},
		{
			name: "validation failure prevents persistence",
			// An empty Name should fail Validate().
			inputUser: model.User{},
			saveFn:    nil, // must NOT be called
			wantErr:   true,
			wantErrMsg: "save user: validation failed",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			saveCalled := false
			repo := &mockUserRepository{
				SaveFn: func(ctx context.Context, u model.User) (model.User, error) {
					saveCalled = true
					if tc.saveFn != nil {
						return tc.saveFn(ctx, u)
					}
					t.Fatal("SaveFn called unexpectedly")
					return model.User{}, nil
				},
			}

			// For the validation test we do NOT want Save to be called.
			if tc.saveFn == nil {
				repo.SaveFn = nil
			}

			svc := newService(repo)
			got, err := svc.SaveUser(context.Background(), tc.inputUser)

			if tc.wantErr {
				require.Error(t, err)
				if tc.wantErrMsg != "" {
					assert.Contains(t, err.Error(), tc.wantErrMsg)
				}
				assert.Equal(t, model.User{}, got)
				// Ensure save was not called when validation failed.
				if tc.saveFn == nil {
					assert.False(t, saveCalled, "repo.Save must not be called when validation fails")
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.wantUser, got)
		})
	}
}

// ---------------------------------------------------------------------------
// FetchUserList
// ---------------------------------------------------------------------------

func TestFetchUserList(t *testing.T) {
	tests := []struct {
		name       string
		findAllFn  func(ctx context.Context) ([]model.User, error)
		wantUsers  []model.User
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "returns all users when records exist",
			findAllFn: func(ctx context.Context) ([]model.User, error) {
				return []model.User{
					validUser(1, "Alice"),
					validUser(2, "Bob"),
				}, nil
			},
			wantUsers: []model.User{
				validUser(1, "Alice"),
				validUser(2, "Bob"),
			},
		},
		{
			name: "returns empty slice when no users exist",
			findAllFn: func(ctx context.Context) ([]model.User, error) {
				return []model.User{}, nil
			},
			wantUsers: []model.User{},
		},
		{
			name: "nil slice from repo is returned as-is (list contract still holds via caller)",
			findAllFn: func(ctx context.Context) ([]model.User, error) {
				return nil, nil
			},
			wantUsers: nil,
		},
		{
			name: "propagates repository error",
			findAllFn: func(ctx context.Context) ([]model.User, error) {
				return nil, sentinelErr
			},
			wantErr:    true,
			wantErrMsg: "fetch user list",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockUserRepository{FindAllFn: tc.findAllFn}
			svc := newService(repo)
			got, err := svc.FetchUserList(context.Background())

			if tc.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErrMsg)
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.wantUsers, got)
		})
	}
}

// ---------------------------------------------------------------------------
// FetchUserByID
// ---------------------------------------------------------------------------

func TestFetchUserByID(t *testing.T) {
	tests := []struct {
		name            string
		id              int
		findByIDFn      func(ctx context.Context, id int) (model.User, error)
		wantUser        model.User
		wantErr         bool
		wantNotFound    bool
		wantErrContains string
	}{
		{
			name: "returns user when id matches",
			id:   1,
			findByIDFn: func(ctx context.Context, id int) (model.User, error) {
				assert.Equal(t, 1, id)
				return validUser(1, "Alice"), nil
			},
			wantUser: validUser(1, "Alice"),
		},
		{
			name: "returns wrapped ErrUserNotFound when no user matches",
			id:   99,
			findByIDFn: func(ctx context.Context, id int) (model.User, error) {
				return model.User{}, smartError.ErrUserNotFound
			},
			wantErr:      true,
			wantNotFound: true,
		},
		{
			name: "wraps non-not-found repository error",
			id:   1,
			findByIDFn: func(ctx context.Context, id int) (model.User, error) {
				return model.User{}, sentinelErr
			},
			wantErr:         true,
			wantNotFound:    false,
			wantErrContains: "fetch user by id 1",
		},
		{
			name: "returned user id matches requested id",
			id:   7,
			findByIDFn: func(ctx context.Context, id int) (model.User, error) {
				return validUser(7, "Carol"), nil
			},
			wantUser: validUser(7, "Carol"),
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockUserRepository{FindByIDFn: tc.findByIDFn}
			svc := newService(repo)
			got, err := svc.FetchUserByID(context.Background(), tc.id)

			if tc.wantErr {
				require.Error(t, err)
				if tc.wantNotFound {
					assert.True(t, errors.Is(err, smartError.ErrUserNotFound),
						"expected ErrUserNotFound, got: %v", err)
				}
				if tc.wantErrContains != "" {
					assert.Contains(t, err.Error(), tc.wantErrContains)
				}
				assert.Equal(t, model.User{}, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.wantUser, got)
			assert.Equal(t, tc.id, got.ID, "returned user id must match requested id")
		})
	}
}

// ---------------------------------------------------------------------------
// DeleteUser
// ---------------------------------------------------------------------------

func TestDeleteUser(t *testing.T) {
	tests := []struct {
		name            string
		id              int
		deleteByIDFn    func(ctx context.Context, id int) error
		wantErr         bool
		wantErrContains string
	}{
		{
			name: "successfully deletes existing user",
			id:   1,
			deleteByIDFn: func(ctx context.Context, id int) error {
				assert.Equal(t, 1, id)
				return nil
			},
			wantErr: false,
		},
		{
			name: "propagates error when user does not exist",
			id:   99,
			deleteByIDFn: func(ctx context.Context, id int) error {
				return sentinelErr
			},
			wantErr:         true,
			wantErrContains: "delete user 99",
		},
		{
			name: "propagates repository persistence error",
			id:   5,
			deleteByIDFn: func(ctx context.Context, id int) error {
				return fmt.Errorf("constraint violation")
			},
			wantErr:         true,
			wantErrContains: "delete user 5",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockUserRepository{DeleteByIDFn: tc.deleteByIDFn}
			svc := newService(repo)
			err := svc.DeleteUser(context.Background(), tc.id)

			if tc.wantErr {
				require.Error(t, err)
				if tc.wantErrContains != "" {
					assert.Contains(t, err.Error(), tc.wantErrContains)
				}
				return
			}

			assert.NoError(t, err)
		})
	}
}

// ---------------------------------------------------------------------------
// UpdateUser
// ---------------------------------------------------------------------------

func TestUpdateUser(t *testing.T) {
	tests := []struct {
		name            string
		id              int
		inputUser       model.User
		findByIDFn      func(ctx context.Context, id int) (model.User, error)
		saveFn          func(ctx context.Context, u model.User) (model.User, error)
		wantUser        model.User
		wantErr         bool
		wantNotFound    bool
		wantErrContains string
	}{
		{
			name:      "successfully updates existing user",
			id:        1,
			inputUser: validUser(0, "Alice Updated"),
			findByIDFn: func(ctx context.Context, id int) (model.User, error) {
				return validUser(1, "Alice"), nil
			},
			saveFn: func(ctx context.Context, u model.User) (model.User, error) {
				// The service should set user.ID = id before saving.
				assert.Equal(t, 1, u.ID)
				return u, nil
			},
			wantUser: validUser(1, "Alice Updated"),
		},
		{
			name:      "returned user id equals supplied id after update",
			id:        5,
			inputUser: validUser(0, "Dave"),
			findByIDFn: func(ctx context.Context, id int) (model.User, error) {
				return validUser(5, "Dave Old"), nil
			},
			saveFn: func(ctx context.Context, u model.User) (model.User, error) {
				return u, nil
			},
			wantUser: validUser(5, "Dave"),
		},
		{
			name:      "returns wrapped ErrUserNotFound when user does not exist",
			id:        99,
			inputUser: validUser(0, "Ghost"),
			findByIDFn: func(ctx context.Context, id int) (model.User, error) {
				return model.User{}, smartError.ErrUserNotFound
			},
			wantErr:      true,
			wantNotFound: true,
		},
		{
			name:      "propagates non-not-found error from FindByID",
			id:        3,
			inputUser: validUser(0, "Eve"),
			findByIDFn: func(ctx context.Context, id int) (model.User, error) {
				return model.User{}, sentinelErr
			},
			wantErr:         true,
			wantErrContains: "update user 3",
		},
		{
			name: "validation failure after find prevents save",
			id:   2