```go
package service

import (
	"context"
	"errors"
	"fmt"
	"testing"

	smartcontacterror "github.com/smartContact/internal/smartcontact/error"
	"github.com/smartContact/internal/smartcontact/model"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// Mock repository
// ---------------------------------------------------------------------------

type mockUserRepository struct {
	mergeFunc    func(ctx context.Context, user *model.User) (*model.User, error)
	findAllFunc  func(ctx context.Context) ([]*model.User, error)
	findByIDFunc func(ctx context.Context, id int) (*model.User, error)
	deleteByID   func(ctx context.Context, id int) error
	findByName   func(ctx context.Context, name string) (*model.User, error)
}

func (m *mockUserRepository) Merge(ctx context.Context, user *model.User) (*model.User, error) {
	return m.mergeFunc(ctx, user)
}

func (m *mockUserRepository) FindAll(ctx context.Context) ([]*model.User, error) {
	return m.findAllFunc(ctx)
}

func (m *mockUserRepository) FindByID(ctx context.Context, id int) (*model.User, error) {
	return m.findByIDFunc(ctx, id)
}

func (m *mockUserRepository) DeleteByID(ctx context.Context, id int) error {
	return m.deleteByID(ctx, id)
}

func (m *mockUserRepository) FindByName(ctx context.Context, name string) (*model.User, error) {
	return m.findByName(ctx, name)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newService(repo *mockUserRepository) *UserServiceImp {
	return NewUserServiceImp(repo)
}

// ---------------------------------------------------------------------------
// SaveUser tests
// ---------------------------------------------------------------------------

func TestSaveUser(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		inputUser   *model.User
		mergeReturn *model.User
		mergeErr    error
		wantUser    *model.User
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid new user is persisted and returned",
			inputUser:   &model.User{Name: "Alice"},
			mergeReturn: &model.User{ID: 1, Name: "Alice"},
			mergeErr:    nil,
			wantUser:    &model.User{ID: 1, Name: "Alice"},
			wantErr:     false,
		},
		{
			name:        "existing user (matching id) is updated and returned",
			inputUser:   &model.User{ID: 5, Name: "Bob"},
			mergeReturn: &model.User{ID: 5, Name: "Bob"},
			mergeErr:    nil,
			wantUser:    &model.User{ID: 5, Name: "Bob"},
			wantErr:     false,
		},
		{
			name:        "repository error is wrapped and returned",
			inputUser:   &model.User{Name: "Carol"},
			mergeReturn: nil,
			mergeErr:    errors.New("db unavailable"),
			wantUser:    nil,
			wantErr:     true,
			errContains: "save user",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockUserRepository{
				mergeFunc: func(_ context.Context, user *model.User) (*model.User, error) {
					return tc.mergeReturn, tc.mergeErr
				},
			}
			svc := newService(repo)

			got, err := svc.SaveUser(ctx, tc.inputUser)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantUser, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// FetchUserList tests
// ---------------------------------------------------------------------------

func TestFetchUserList(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		repoReturn  []*model.User
		repoErr     error
		wantUsers   []*model.User
		wantErr     bool
		errContains string
	}{
		{
			name: "users exist – returns all users",
			repoReturn: []*model.User{
				{ID: 1, Name: "Alice"},
				{ID: 2, Name: "Bob"},
			},
			repoErr: nil,
			wantUsers: []*model.User{
				{ID: 1, Name: "Alice"},
				{ID: 2, Name: "Bob"},
			},
			wantErr: false,
		},
		{
			name:       "no users exist – returns empty list",
			repoReturn: []*model.User{},
			repoErr:    nil,
			wantUsers:  []*model.User{},
			wantErr:    false,
		},
		{
			name:        "repository error is wrapped and returned",
			repoReturn:  nil,
			repoErr:     errors.New("connection refused"),
			wantUsers:   nil,
			wantErr:     true,
			errContains: "fetch user list",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockUserRepository{
				findAllFunc: func(_ context.Context) ([]*model.User, error) {
					return tc.repoReturn, tc.repoErr
				},
			}
			svc := newService(repo)

			got, err := svc.FetchUserList(ctx)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantUsers, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// FetchUserByID tests
// ---------------------------------------------------------------------------

func TestFetchUserByID(t *testing.T) {
	ctx := context.Background()

	notFoundErr := smartcontacterror.NewUserNotFoundError("user are not available")

	tests := []struct {
		name            string
		id              int
		repoReturn      *model.User
		repoErr         error
		wantUser        *model.User
		wantErr         bool
		wantNotFound    bool
		errMsgContains  string
	}{
		{
			name:         "existing id returns the matching user",
			id:           1,
			repoReturn:   &model.User{ID: 1, Name: "Alice"},
			repoErr:      nil,
			wantUser:     &model.User{ID: 1, Name: "Alice"},
			wantErr:      false,
			wantNotFound: false,
		},
		{
			name:           "non-existing id – repository returns ErrUserNotFound – service wraps as ErrUserNotFound",
			id:             99,
			repoReturn:     nil,
			repoErr:        notFoundErr,
			wantUser:       nil,
			wantErr:        true,
			wantNotFound:   true,
			errMsgContains: "user are not available",
		},
		{
			name:           "non-existing id – repository returns nil user and nil error",
			id:             42,
			repoReturn:     nil,
			repoErr:        nil,
			wantUser:       nil,
			wantErr:        true,
			wantNotFound:   true,
			errMsgContains: "user are not available",
		},
		{
			name:           "repository returns generic error – error is wrapped",
			id:             7,
			repoReturn:     nil,
			repoErr:        errors.New("timeout"),
			wantUser:       nil,
			wantErr:        true,
			wantNotFound:   false,
			errMsgContains: "fetch user by id 7",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockUserRepository{
				findByIDFunc: func(_ context.Context, id int) (*model.User, error) {
					return tc.repoReturn, tc.repoErr
				},
			}
			svc := newService(repo)

			got, err := svc.FetchUserByID(ctx, tc.id)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				assert.Contains(t, err.Error(), tc.errMsgContains)
				if tc.wantNotFound {
					assert.True(t, errors.Is(err, smartcontacterror.ErrUserNotFound),
						"expected ErrUserNotFound, got: %v", err)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantUser, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// DeleteUser tests
// ---------------------------------------------------------------------------

func TestDeleteUser(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		id          int
		repoErr     error
		wantErr     bool
		errContains string
	}{
		{
			name:    "existing id is deleted successfully",
			id:      1,
			repoErr: nil,
			wantErr: false,
		},
		{
			name:        "repository returns error – error is wrapped",
			id:          99,
			repoErr:     errors.New("record not found"),
			wantErr:     true,
			errContains: "delete user 99",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var capturedID int
			repo := &mockUserRepository{
				deleteByID: func(_ context.Context, id int) error {
					capturedID = id
					return tc.repoErr
				},
			}
			svc := newService(repo)

			err := svc.DeleteUser(ctx, tc.id)

			assert.Equal(t, tc.id, capturedID, "repository should be called with the supplied id")

			if tc.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// UpdateUser tests
// ---------------------------------------------------------------------------

func TestUpdateUser(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		id          int
		inputUser   *model.User
		mergeReturn *model.User
		mergeErr    error
		wantErr     bool
		errContains string
		wantID      int // expected id assigned to user before Merge
	}{
		{
			name:        "valid id and user – id is set on user and repo is called",
			id:          10,
			inputUser:   &model.User{Name: "Dave"},
			mergeReturn: &model.User{ID: 10, Name: "Dave"},
			mergeErr:    nil,
			wantErr:     false,
			wantID:      10,
		},
		{
			name:        "user id is overwritten with supplied id even if different",
			id:          20,
			inputUser:   &model.User{ID: 999, Name: "Eve"},
			mergeReturn: &model.User{ID: 20, Name: "Eve"},
			mergeErr:    nil,
			wantErr:     false,
			wantID:      20,
		},
		{
			name:        "nil user returns error without calling repository",
			id:          5,
			inputUser:   nil,
			mergeReturn: nil,
			mergeErr:    nil,
			wantErr:     true,
			errContains: "user must not be nil",
		},
		{
			name:        "repository error is wrapped",
			id:          3,
			inputUser:   &model.User{Name: "Frank"},
			mergeReturn: nil,
			mergeErr:    errors.New("db write failed"),
			wantErr:     true,
			errContains: "update user 3",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var mergedUser *model.User
			mergeCalled := false

			repo := &mockUserRepository{
				mergeFunc: func(_ context.Context, user *model.User) (*model.User, error) {
					mergeCalled = true
					mergedUser = user
					return tc.mergeReturn, tc.mergeErr
				},
			}
			svc := newService(repo)

			err := svc.UpdateUser(ctx, tc.id, tc.inputUser)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
				if tc.inputUser == nil {
					assert.False(t, mergeCalled, "Merge should not be called when user is nil")
				}
			} else {
				assert.NoError(t, err)
				assert.True(t, mergeCalled)
				assert.Equal(t, tc.wantID, mergedUser.ID,
					"user.ID should be overwritten with the supplied id")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// GetUserByName tests
// ---------------------------------------------------------------------------

func TestGetUserByName(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		queryName   string
		repoReturn  *model.User
		repoErr     error
		wantUser    *model.User
		wantErr     bool
		errContains string
	}{
		{
			name:       "name matches a user – returns matching user",
			queryName:  "Alice",
			repoReturn: &model.User{ID: 1, Name: "Alice"},
			repoErr:    nil,
			wantUser:   &model.User{ID: 1, Name: "Alice"},
			wantErr:    false,
		},
		{
			name:       "name matches no user – returns nil without error",
			queryName:  "Ghost",
			repoReturn: nil,
			repoErr:    nil,
			wantUser:   nil,
			wantErr:    false,
		},
		{
			name:        "repository error is wrapped",
			queryName:   "Error",
			repoReturn:  nil,
			repoErr:     fmt.Errorf("query failed"),
			wantUser:    nil,
			wantErr:     true,
			errContains: `get user by name "Error"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var capturedName string
			repo := &mockUserRepository{
				findByName: func(_ context.Context, name string) (*model.User, error) {
					capturedName = name
					return tc.repoReturn, tc.repoErr
				},
			}
			svc := newService(repo)

			got, err := svc.GetUserByName(ctx, tc.queryName)

			assert.Equal(t, tc.queryName, capturedName)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantUser, got)