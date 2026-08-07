```go
package service_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	smartcontacterror "github.com/smartContact/internal/smartcontact/error"
	"github.com/smartContact/internal/smartcontact/model"
	"github.com/smartContact/internal/smartcontact/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ---------------------------------------------------------------------------
// Mock repository
// ---------------------------------------------------------------------------

type mockUserRepository struct {
	mock.Mock
}

func (m *mockUserRepository) Save(ctx context.Context, user *model.User) (*model.User, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *mockUserRepository) FindAll(ctx context.Context) ([]*model.User, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.User), args.Error(1)
}

func (m *mockUserRepository) FindByID(ctx context.Context, id int) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *mockUserRepository) DeleteByID(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockUserRepository) FindByName(ctx context.Context, name string) (*model.User, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

// ---------------------------------------------------------------------------
// Helper
// ---------------------------------------------------------------------------

func newService(repo *mockUserRepository) service.UserService {
	return service.NewUserService(repo)
}

// ---------------------------------------------------------------------------
// SaveUser
// ---------------------------------------------------------------------------

func TestSaveUser(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		inputUser   *model.User
		repoReturn  *model.User
		repoErr     error
		wantUser    *model.User
		wantErr     bool
		errContains string
	}{
		{
			name:      "given a valid user object returns the persisted user",
			inputUser: &model.User{Name: "Alice", Email: "alice@example.com"},
			repoReturn: &model.User{ID: 1, Name: "Alice", Email: "alice@example.com"},
			repoErr:    nil,
			wantUser:  &model.User{ID: 1, Name: "Alice", Email: "alice@example.com"},
			wantErr:   false,
		},
		{
			name:      "given a user with an existing id returns the updated user",
			inputUser: &model.User{ID: 5, Name: "Bob", Email: "bob@example.com"},
			repoReturn: &model.User{ID: 5, Name: "Bob", Email: "bob@example.com"},
			repoErr:    nil,
			wantUser:  &model.User{ID: 5, Name: "Bob", Email: "bob@example.com"},
			wantErr:   false,
		},
		{
			name:        "given a repository error wraps and returns the error",
			inputUser:   &model.User{Name: "Charlie"},
			repoReturn:  nil,
			repoErr:     errors.New("db connection failed"),
			wantUser:    nil,
			wantErr:     true,
			errContains: "save user",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := new(mockUserRepository)
			repo.On("Save", ctx, tc.inputUser).Return(tc.repoReturn, tc.repoErr)

			svc := newService(repo)
			got, err := svc.SaveUser(ctx, tc.inputUser)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.errContains != "" {
					assert.Contains(t, err.Error(), tc.errContains)
				}
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantUser, got)
			}
			repo.AssertExpectations(t)
		})
	}
}

// ---------------------------------------------------------------------------
// FetchUserList
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
			name: "given the repository has users returns a list of all users",
			repoReturn: []*model.User{
				{ID: 1, Name: "Alice"},
				{ID: 2, Name: "Bob"},
			},
			repoErr:   nil,
			wantUsers: []*model.User{
				{ID: 1, Name: "Alice"},
				{ID: 2, Name: "Bob"},
			},
			wantErr: false,
		},
		{
			name:        "given the repository has no users returns an empty list",
			repoReturn:  []*model.User{},
			repoErr:     nil,
			wantUsers:   []*model.User{},
			wantErr:     false,
		},
		{
			name:        "given a repository error wraps and returns the error",
			repoReturn:  nil,
			repoErr:     errors.New("db failure"),
			wantUsers:   nil,
			wantErr:     true,
			errContains: "fetch user list",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := new(mockUserRepository)
			repo.On("FindAll", ctx).Return(tc.repoReturn, tc.repoErr)

			svc := newService(repo)
			got, err := svc.FetchUserList(ctx)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.errContains != "" {
					assert.Contains(t, err.Error(), tc.errContains)
				}
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantUsers, got)
			}
			repo.AssertExpectations(t)
		})
	}
}

// ---------------------------------------------------------------------------
// FetchUserByID
// ---------------------------------------------------------------------------

func TestFetchUserByID(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name            string
		id              int
		repoReturn      *model.User
		repoErr         error
		wantUser        *model.User
		wantErr         bool
		wantNotFound    bool
		errContains     string
	}{
		{
			name:         "given an id that exists returns the matching user",
			id:           1,
			repoReturn:   &model.User{ID: 1, Name: "Alice"},
			repoErr:      nil,
			wantUser:     &model.User{ID: 1, Name: "Alice"},
			wantErr:      false,
			wantNotFound: false,
		},
		{
			name:         "given an id that does not exist repo returns ErrUserNotFound",
			id:           999,
			repoReturn:   nil,
			repoErr:      smartcontacterror.ErrUserNotFound,
			wantUser:     nil,
			wantErr:      true,
			wantNotFound: true,
			errContains:  "User are not available",
		},
		{
			name:         "given an id that does not exist repo returns nil user",
			id:           888,
			repoReturn:   nil,
			repoErr:      nil,
			wantUser:     nil,
			wantErr:      true,
			wantNotFound: true,
			errContains:  "User are not available",
		},
		{
			name:         "given a generic repository error wraps and returns the error",
			id:           2,
			repoReturn:   nil,
			repoErr:      errors.New("connection timeout"),
			wantUser:     nil,
			wantErr:      true,
			wantNotFound: false,
			errContains:  fmt.Sprintf("fetch user by id %d", 2),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := new(mockUserRepository)
			repo.On("FindByID", ctx, tc.id).Return(tc.repoReturn, tc.repoErr)

			svc := newService(repo)
			got, err := svc.FetchUserByID(ctx, tc.id)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.errContains != "" {
					assert.Contains(t, err.Error(), tc.errContains)
				}
				assert.Nil(t, got)
				if tc.wantNotFound {
					assert.True(t, errors.Is(err, smartcontacterror.ErrUserNotFound),
						"expected ErrUserNotFound but got: %v", err)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantUser, got)
			}
			repo.AssertExpectations(t)
		})
	}
}

// ---------------------------------------------------------------------------
// DeleteUser
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
			name:    "given an id that exists deletes successfully",
			id:      1,
			repoErr: nil,
			wantErr: false,
		},
		{
			name:        "given an id that does not exist propagates repository error",
			id:          999,
			repoErr:     errors.New("record not found"),
			wantErr:     true,
			errContains: fmt.Sprintf("delete user %d", 999),
		},
		{
			name:        "given a generic repository error wraps and returns the error",
			id:          2,
			repoErr:     errors.New("db unavailable"),
			wantErr:     true,
			errContains: fmt.Sprintf("delete user %d", 2),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := new(mockUserRepository)
			repo.On("DeleteByID", ctx, tc.id).Return(tc.repoErr)

			svc := newService(repo)
			err := svc.DeleteUser(ctx, tc.id)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.errContains != "" {
					assert.Contains(t, err.Error(), tc.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
			repo.AssertExpectations(t)
		})
	}
}

// ---------------------------------------------------------------------------
// UpdateUser
// ---------------------------------------------------------------------------

func TestUpdateUser(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		id          int
		inputUser   *model.User
		repoReturn  *model.User
		repoErr     error
		wantUser    *model.User
		wantErr     bool
		errContains string
		// wantIDSet verifies that the user passed to Save has id set to tc.id
		wantIDSet bool
	}{
		{
			name:      "given a valid id and user object persists and returns updated user",
			id:        10,
			inputUser: &model.User{Name: "Alice", Email: "alice@example.com"},
			repoReturn: &model.User{ID: 10, Name: "Alice", Email: "alice@example.com"},
			repoErr:   nil,
			wantUser:  &model.User{ID: 10, Name: "Alice", Email: "alice@example.com"},
			wantErr:   false,
			wantIDSet: true,
		},
		{
			name:      "given a user whose original id differs it is overwritten by the provided id",
			id:        7,
			inputUser: &model.User{ID: 99, Name: "Bob"},
			repoReturn: &model.User{ID: 7, Name: "Bob"},
			repoErr:   nil,
			wantUser:  &model.User{ID: 7, Name: "Bob"},
			wantErr:   false,
			wantIDSet: true,
		},
		{
			name:        "given a nil user returns an error",
			id:          1,
			inputUser:   nil,
			repoReturn:  nil,
			repoErr:     nil,
			wantUser:    nil,
			wantErr:     true,
			errContains: "user must not be nil",
		},
		{
			name:        "given a repository error wraps and returns the error",
			id:          3,
			inputUser:   &model.User{Name: "Charlie"},
			repoReturn:  nil,
			repoErr:     errors.New("constraint violation"),
			wantUser:    nil,
			wantErr:     true,
			errContains: fmt.Sprintf("update user %d", 3),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := new(mockUserRepository)

			// Only set up Save expectation when inputUser is non-nil (nil user errors early)
			if tc.inputUser != nil {
				// After UpdateUser sets ID, the user passed to Save has the expected ID.
				expectedUser := *tc.inputUser
				expectedUser.ID = tc.id
				repo.On("Save", ctx, &expectedUser).Return(tc.repoReturn, tc.repoErr)
			}

			svc := newService(repo)
			got, err := svc.UpdateUser(ctx, tc.id, tc.inputUser)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.errContains != "" {
					assert.Contains(t, err.Error(), tc.errContains)
				}
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantUser, got)
				// Invariant: id was set on the user
				if tc.wantIDSet && tc.inputUser != nil {
					assert.Equal(t, tc.id, tc.inputUser.ID)
				}
			}
			repo.AssertExpectations(t)
		})
	}
}

// ---------------------------------------------------------------------------
// GetUserByName
// ---------------------------------------------------------------------------

func TestGetUserByName(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		inputName    string
		repoReturn   *model.User
		repoErr      error
		wantUser     *model.User
		wantErr      bool
		wantNotFound bool
		errContains  string
	}{
		{
			name:         "given a name that matches