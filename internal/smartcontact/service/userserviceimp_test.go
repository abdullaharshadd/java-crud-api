```go
package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"internal/smartcontact/apperr"
	"internal/smartcontact/model"
	"internal/smartcontact/service"
)

// mockUserRepo implements the repository interface expected by userService.
type mockUserRepo struct {
	saveFunc          func(ctx context.Context, user *model.User) (*model.User, error)
	getAllFunc         func(ctx context.Context) ([]*model.User, error)
	getByIDFunc       func(ctx context.Context, id int) (*model.User, error)
	deleteByIDFunc    func(ctx context.Context, id int) error
	updateFunc        func(ctx context.Context, user *model.User) error
	findByNameFunc    func(ctx context.Context, name string) (*model.User, error)
}

func (m *mockUserRepo) Save(ctx context.Context, user *model.User) (*model.User, error) {
	return m.saveFunc(ctx, user)
}

func (m *mockUserRepo) GetAll(ctx context.Context) ([]*model.User, error) {
	return m.getAllFunc(ctx)
}

func (m *mockUserRepo) GetByID(ctx context.Context, id int) (*model.User, error) {
	return m.getByIDFunc(ctx, id)
}

func (m *mockUserRepo) DeleteByID(ctx context.Context, id int) error {
	return m.deleteByIDFunc(ctx, id)
}

func (m *mockUserRepo) Update(ctx context.Context, user *model.User) error {
	return m.updateFunc(ctx, user)
}

func (m *mockUserRepo) FindByName(ctx context.Context, name string) (*model.User, error) {
	return m.findByNameFunc(ctx, name)
}

// ---------------------------------------------------------------------------
// SaveUser
// ---------------------------------------------------------------------------

func TestSaveUser(t *testing.T) {
	tests := []struct {
		name        string
		input       *model.User
		repoReturn  *model.User
		repoErr     error
		wantUser    *model.User
		wantErr     bool
	}{
		{
			name:  "given a valid User object returns the persisted User entity",
			input: &model.User{Name: "Alice", Email: "alice@example.com"},
			repoReturn: &model.User{ID: 1, Name: "Alice", Email: "alice@example.com"},
			repoErr:    nil,
			wantUser:   &model.User{ID: 1, Name: "Alice", Email: "alice@example.com"},
			wantErr:    false,
		},
		{
			name:  "given a User object with an existing id returns the updated User entity",
			input: &model.User{ID: 5, Name: "Bob", Email: "bob@example.com"},
			repoReturn: &model.User{ID: 5, Name: "Bob", Email: "bob@example.com"},
			repoErr:    nil,
			wantUser:   &model.User{ID: 5, Name: "Bob", Email: "bob@example.com"},
			wantErr:    false,
		},
		{
			name:    "repository returns an error propagates it",
			input:   &model.User{Name: "Err"},
			repoReturn: nil,
			repoErr: errors.New("db error"),
			wantUser: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockUserRepo{
				saveFunc: func(ctx context.Context, user *model.User) (*model.User, error) {
					return tt.repoReturn, tt.repoErr
				},
			}
			svc := service.NewUserService(repo)
			got, err := svc.SaveUser(context.Background(), tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantUser, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// GetAllUsers (fetchUserList)
// ---------------------------------------------------------------------------

func TestGetAllUsers(t *testing.T) {
	tests := []struct {
		name       string
		repoReturn []*model.User
		repoErr    error
		wantUsers  []*model.User
		wantErr    bool
	}{
		{
			name: "given users exist in the store returns all User entities",
			repoReturn: []*model.User{
				{ID: 1, Name: "Alice"},
				{ID: 2, Name: "Bob"},
			},
			repoErr:   nil,
			wantUsers: []*model.User{{ID: 1, Name: "Alice"}, {ID: 2, Name: "Bob"}},
			wantErr:   false,
		},
		{
			name:       "given no users exist returns an empty list",
			repoReturn: []*model.User{},
			repoErr:    nil,
			wantUsers:  []*model.User{},
			wantErr:    false,
		},
		{
			name:       "repository error is propagated",
			repoReturn: nil,
			repoErr:    errors.New("connection refused"),
			wantUsers:  nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockUserRepo{
				getAllFunc: func(ctx context.Context) ([]*model.User, error) {
					return tt.repoReturn, tt.repoErr
				},
			}
			svc := service.NewUserService(repo)
			got, err := svc.GetAllUsers(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantUsers, got)
				// invariant: never returns nil slice on success
				assert.NotNil(t, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// FetchUserById
// ---------------------------------------------------------------------------

func TestFetchUserById(t *testing.T) {
	tests := []struct {
		name       string
		id         int
		repoReturn *model.User
		repoErr    error
		wantUser   *model.User
		wantErrType interface{} // nil or pointer to expected error type
		wantErr    bool
	}{
		{
			name:       "given an id matching an existing User returns the User",
			id:         1,
			repoReturn: &model.User{ID: 1, Name: "Alice"},
			repoErr:    nil,
			wantUser:   &model.User{ID: 1, Name: "Alice"},
			wantErr:    false,
		},
		{
			name:       "given an id with no matching User returns UserNotFoundError",
			id:         999,
			repoReturn: nil,
			repoErr:    errors.New("not found"),
			wantUser:   nil,
			wantErr:    true,
			wantErrType: &apperr.UserNotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockUserRepo{
				getByIDFunc: func(ctx context.Context, id int) (*model.User, error) {
					return tt.repoReturn, tt.repoErr
				},
			}
			svc := service.NewUserService(repo)
			got, err := svc.FetchUserById(context.Background(), tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				if tt.wantErrType != nil {
					var notFound *apperr.UserNotFoundError
					assert.True(t, errors.As(err, &notFound), "expected UserNotFoundError")
					assert.Contains(t, notFound.Error(), "User are not available")
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantUser, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// DeleteUser
// ---------------------------------------------------------------------------

func TestDeleteUser(t *testing.T) {
	tests := []struct {
		name    string
		id      int
		repoErr error
		wantErr bool
	}{
		{
			name:    "given an id matching an existing User deletes successfully",
			id:      1,
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "given an id with no matching User propagates repository error",
			id:      999,
			repoErr: errors.New("record not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockUserRepo{
				deleteByIDFunc: func(ctx context.Context, id int) error {
					assert.Equal(t, tt.id, id)
					return tt.repoErr
				},
			}
			svc := service.NewUserService(repo)
			err := svc.DeleteUser(context.Background(), tt.id)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// UpdateUser
// ---------------------------------------------------------------------------

func TestUpdateUser(t *testing.T) {
	tests := []struct {
		name    string
		id      int
		user    *model.User
		repoErr error
		wantErr bool
		// capturedUser holds what the repo received so we can assert the id was overridden
		assertCaptured func(t *testing.T, captured *model.User)
	}{
		{
			name:    "given a valid id and User object persists with the provided id",
			id:      7,
			user:    &model.User{ID: 0, Name: "Carol", Email: "carol@example.com"},
			repoErr: nil,
			wantErr: false,
			assertCaptured: func(t *testing.T, captured *model.User) {
				assert.Equal(t, 7, captured.ID, "id must be overridden by the argument")
				assert.Equal(t, "Carol", captured.Name)
			},
		},
		{
			name:    "id argument overrides any id already set on the user object",
			id:      42,
			user:    &model.User{ID: 1, Name: "Dave"},
			repoErr: nil,
			wantErr: false,
			assertCaptured: func(t *testing.T, captured *model.User) {
				assert.Equal(t, 42, captured.ID, "id must be overridden by the argument")
			},
		},
		{
			name:    "repository error is propagated",
			id:      3,
			user:    &model.User{Name: "Eve"},
			repoErr: errors.New("update failed"),
			wantErr: true,
			assertCaptured: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var captured *model.User
			repo := &mockUserRepo{
				updateFunc: func(ctx context.Context, user *model.User) error {
					captured = user
					return tt.repoErr
				},
			}
			svc := service.NewUserService(repo)
			err := svc.UpdateUser(context.Background(), tt.id, tt.user)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.assertCaptured != nil && captured != nil {
					tt.assertCaptured(t, captured)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// GetUserNameByName (getUserNameByName)
// ---------------------------------------------------------------------------

func TestGetUserNameByName(t *testing.T) {
	tests := []struct {
		name       string
		inputName  string
		repoReturn *model.User
		repoErr    error
		wantUser   *model.User
		wantErr    bool
	}{
		{
			name:       "given a name matching an existing User returns the User",
			inputName:  "Alice",
			repoReturn: &model.User{ID: 1, Name: "Alice"},
			repoErr:    nil,
			wantUser:   &model.User{ID: 1, Name: "Alice"},
			wantErr:    false,
		},
		{
			name:       "given a name with no matching User returns nil without error",
			inputName:  "Ghost",
			repoReturn: nil,
			repoErr:    nil,
			wantUser:   nil,
			wantErr:    false,
		},
		{
			name:       "repository error is propagated",
			inputName:  "Error",
			repoReturn: nil,
			repoErr:    errors.New("db failure"),
			wantUser:   nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockUserRepo{
				findByNameFunc: func(ctx context.Context, name string) (*model.User, error) {
					assert.Equal(t, tt.inputName, name)
					return tt.repoReturn, tt.repoErr
				},
			}
			svc := service.NewUserService(repo)
			got, err := svc.GetUserNameByName(context.Background(), tt.inputName)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantUser, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Global invariants – delegation, no internal state mutation
// ---------------------------------------------------------------------------

// TestServiceDelegatesAllPersistenceToRepo verifies that the service never
// stores data internally: two consecutive GetAllUsers calls reflect whatever
// the repo returns each time, proving no internal cache / state.
func TestServiceHoldsNoInternalMutableState(t *testing.T) {
	callCount := 0
	responses := [][]*model.User{
		{{ID: 1, Name: "First call"}},
		{{ID: 1, Name: "First call"}, {ID: 2, Name: "Second call"}},
	}

	repo := &mockUserRepo{
		getAllFunc: func(ctx context.Context) ([]*model.User, error) {
			defer func() { callCount++ }()
			return responses[callCount], nil
		},
	}
	svc := service.NewUserService(repo)

	first, err := svc.GetAllUsers(context.Background())
	assert.NoError(t, err)
	assert.Len(t, first, 1)

	second, err := svc.GetAllUsers(context.Background())
	assert.NoError(t, err)
	assert.Len(t, second, 2)
}

// TestReadOperationsDoNotMutate verifies FetchUserById and GetUserNameByName
// never call Save/Update/Delete.
func TestReadOperationsDoNotCallWriteMethods(t *testing.T) {
	mutationCalled := false

	repo := &mockUserRepo{
		getByIDFunc: func(ctx context.Context, id int) (*model.User, error) {
			return &model.User{ID: id, Name: "X"}, nil
		},
		findByNameFunc: func(ctx context.Context, name string) (*model.User, error) {