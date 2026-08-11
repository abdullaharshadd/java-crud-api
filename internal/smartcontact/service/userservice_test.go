```go
package service_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apperr "github.com/smartContact/internal/smartcontact/error"
	"github.com/smartContact/internal/smartcontact/model"
	"github.com/smartContact/internal/smartcontact/repository"
	"github.com/smartContact/internal/smartcontact/service"
)

// ---------------------------------------------------------------------------
// Helpers – pointer-from-literal shorthands
// ---------------------------------------------------------------------------

func strPtr(s string) *string { return &s }

// ---------------------------------------------------------------------------
// mockUserRepository – in-memory stub that satisfies repository.UserRepository
// ---------------------------------------------------------------------------

type mockUserRepository struct {
	// Injected behaviour per test case
	saveFunc       func(ctx context.Context, user *model.User) (*model.User, error)
	findAllFunc    func(ctx context.Context) ([]*model.User, error)
	findByIDFunc   func(ctx context.Context, id int) (*model.User, error)
	deleteFunc     func(ctx context.Context, id int) error
	findByNameFunc func(ctx context.Context, name string) (*model.User, error)
}

func (m *mockUserRepository) Save(ctx context.Context, user *model.User) (*model.User, error) {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, user)
	}
	return nil, errors.New("Save: not configured")
}

func (m *mockUserRepository) FindAll(ctx context.Context) ([]*model.User, error) {
	if m.findAllFunc != nil {
		return m.findAllFunc(ctx)
	}
	return nil, errors.New("FindAll: not configured")
}

func (m *mockUserRepository) FindByID(ctx context.Context, id int) (*model.User, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return nil, errors.New("FindByID: not configured")
}

func (m *mockUserRepository) Delete(ctx context.Context, id int) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return errors.New("Delete: not configured")
}

func (m *mockUserRepository) FindByName(ctx context.Context, name string) (*model.User, error) {
	if m.findByNameFunc != nil {
		return m.findByNameFunc(ctx, name)
	}
	return nil, errors.New("FindByName: not configured")
}

// ---------------------------------------------------------------------------
// Compile-time assertion that the mock satisfies the interface.
// ---------------------------------------------------------------------------
var _ repository.UserRepository = (*mockUserRepository)(nil)

// ---------------------------------------------------------------------------
// SaveUser
// ---------------------------------------------------------------------------

func TestUserService_SaveUser(t *testing.T) {
	ctx := context.Background()

	validUser := &model.User{
		Name:     strPtr("Alice"),
		Email:    strPtr("alice@example.com"),
		Password: strPtr("secret"),
		Role:     strPtr("user"),
		About:    strPtr("about alice"),
	}

	savedUser := &model.User{
		ID:       1,
		Name:     strPtr("Alice"),
		Email:    strPtr("alice@example.com"),
		Password: strPtr("secret"),
		Role:     strPtr("user"),
		About:    strPtr("about alice"),
	}

	tests := []struct {
		name      string
		inputUser *model.User
		repoSave  func(ctx context.Context, user *model.User) (*model.User, error)
		wantUser  *model.User
		wantErr   bool
		errCheck  func(t *testing.T, err error)
	}{
		{
			name:      "valid user is persisted and returned with generated id",
			inputUser: validUser,
			repoSave: func(_ context.Context, _ *model.User) (*model.User, error) {
				return savedUser, nil
			},
			wantUser: savedUser,
			wantErr:  false,
		},
		{
			name:      "nil user returns error without calling repository",
			inputUser: nil,
			repoSave: func(_ context.Context, _ *model.User) (*model.User, error) {
				t.Fatal("Save should not be called for nil user")
				return nil, nil
			},
			wantUser: nil,
			wantErr:  true,
			errCheck: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "nil")
			},
		},
		{
			name: "repository error is propagated",
			inputUser: &model.User{
				Name:     strPtr("Bob"),
				Email:    strPtr("bob@example.com"),
				Password: strPtr("pass"),
				Role:     strPtr("admin"),
				About:    strPtr("about bob"),
			},
			repoSave: func(_ context.Context, _ *model.User) (*model.User, error) {
				return nil, errors.New("db: connection refused")
			},
			wantUser: nil,
			wantErr:  true,
			errCheck: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "save user")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockUserRepository{saveFunc: tt.repoSave}
			svc := service.NewUserService(repo)

			got, err := svc.SaveUser(ctx, tt.inputUser)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errCheck != nil {
					tt.errCheck(t, err)
				}
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantUser, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// GetAllUsers (mirrors Java fetchUserList)
// ---------------------------------------------------------------------------

func TestUserService_GetAllUsers(t *testing.T) {
	ctx := context.Background()

	alice := &model.User{ID: 1, Name: strPtr("Alice"), Email: strPtr("a@x.com"), Password: strPtr("p"), Role: strPtr("user"), About: strPtr("")}
	bob := &model.User{ID: 2, Name: strPtr("Bob"), Email: strPtr("b@x.com"), Password: strPtr("p"), Role: strPtr("admin"), About: strPtr("")}

	tests := []struct {
		name        string
		repoFindAll func(ctx context.Context) ([]*model.User, error)
		wantUsers   []*model.User
		wantErr     bool
	}{
		{
			name: "returns all users when records exist",
			repoFindAll: func(_ context.Context) ([]*model.User, error) {
				return []*model.User{alice, bob}, nil
			},
			wantUsers: []*model.User{alice, bob},
			wantErr:   false,
		},
		{
			name: "returns empty slice when no users exist",
			repoFindAll: func(_ context.Context) ([]*model.User, error) {
				return []*model.User{}, nil
			},
			wantUsers: []*model.User{},
			wantErr:   false,
		},
		{
			name: "repository error is wrapped and propagated",
			repoFindAll: func(_ context.Context) ([]*model.User, error) {
				return nil, errors.New("db: timeout")
			},
			wantUsers: nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockUserRepository{findAllFunc: tt.repoFindAll}
			svc := service.NewUserService(repo)

			got, err := svc.GetAllUsers(ctx)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				// Never nil – even for empty list per spec invariant
				require.NotNil(t, got)
				assert.Equal(t, tt.wantUsers, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// FetchUserById
// ---------------------------------------------------------------------------

func TestUserService_FetchUserById(t *testing.T) {
	ctx := context.Background()

	alice := &model.User{ID: 42, Name: strPtr("Alice"), Email: strPtr("a@x.com"), Password: strPtr("p"), Role: strPtr("user"), About: strPtr("")}

	tests := []struct {
		name         string
		inputID      int
		repoFindByID func(ctx context.Context, id int) (*model.User, error)
		wantUser     *model.User
		wantErr      bool
		errCheck     func(t *testing.T, err error)
	}{
		{
			name:    "returns user when id matches",
			inputID: 42,
			repoFindByID: func(_ context.Context, id int) (*model.User, error) {
				assert.Equal(t, 42, id)
				return alice, nil
			},
			wantUser: alice,
			wantErr:  false,
		},
		{
			name:    "returns UserNotFoundError when id does not match",
			inputID: 99,
			repoFindByID: func(_ context.Context, _ int) (*model.User, error) {
				return nil, repository.ErrUserNotFound
			},
			wantUser: nil,
			wantErr:  true,
			errCheck: func(t *testing.T, err error) {
				var notFound *apperr.UserNotFoundError
				assert.True(t, errors.As(err, &notFound), "expected *apperr.UserNotFoundError, got %T: %v", err, err)
			},
		},
		{
			name:    "non-not-found repository error is wrapped",
			inputID: 1,
			repoFindByID: func(_ context.Context, _ int) (*model.User, error) {
				return nil, errors.New("db: network error")
			},
			wantUser: nil,
			wantErr:  true,
			errCheck: func(t *testing.T, err error) {
				var notFound *apperr.UserNotFoundError
				assert.False(t, errors.As(err, &notFound))
				assert.Contains(t, err.Error(), "fetch user by id")
			},
		},
		{
			name:    "returned user id matches requested id",
			inputID: 42,
			repoFindByID: func(_ context.Context, _ int) (*model.User, error) {
				return alice, nil
			},
			wantUser: alice,
			wantErr:  false,
			errCheck: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockUserRepository{findByIDFunc: tt.repoFindByID}
			svc := service.NewUserService(repo)

			got, err := svc.FetchUserById(ctx, tt.inputID)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errCheck != nil {
					tt.errCheck(t, err)
				}
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, tt.wantUser.ID, got.ID)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// DeleteUser
// ---------------------------------------------------------------------------

func TestUserService_DeleteUser(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		inputID    int
		repoDelete func(ctx context.Context, id int) (*model.User, error) // unused sig; see below
		deleteFunc func(ctx context.Context, id int) error
		wantErr    bool
		errCheck   func(t *testing.T, err error)
	}{
		{
			name:    "successfully deletes existing user",
			inputID: 10,
			deleteFunc: func(_ context.Context, id int) error {
				assert.Equal(t, 10, id)
				return nil
			},
			wantErr: false,
		},
		{
			name:    "returns UserNotFoundError when user does not exist",
			inputID: 99,
			deleteFunc: func(_ context.Context, _ int) error {
				return repository.ErrUserNotFound
			},
			wantErr: true,
			errCheck: func(t *testing.T, err error) {
				var notFound *apperr.UserNotFoundError
				assert.True(t, errors.As(err, &notFound))
			},
		},
		{
			name:    "generic repository error is wrapped",
			inputID: 5,
			deleteFunc: func(_ context.Context, _ int) error {
				return errors.New("db: disk full")
			},
			wantErr: true,
			errCheck: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "delete user")
				var notFound *apperr.UserNotFoundError
				assert.False(t, errors.As(err, &notFound))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockUserRepository{deleteFunc: tt.deleteFunc}
			svc := service.NewUserService(repo)

			err := svc.DeleteUser(ctx, tt.inputID)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errCheck != nil {
					tt.errCheck(t, err)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// UpdateUser
// ---------------------------------------------------------------------------

func TestUserService_UpdateUser(t *testing.T) {
	ctx := context.Background()

	existingUser := &model.User{
		ID:       7,
		Name:     strPtr("OldName"),
		Email:    strPtr("old@example.com"),
		Password: strPtr("oldpass"),
		Role:     strPtr("user"),
		About:    strPtr("old about"),
	}

	tests := []struct {
		name         string
		inputID      int
		inputUser    *model.User
		repoFindByID func(ctx context.Context, id int) (*model.User, error)
		repoSave     func(ctx context.Context, user *model.User) (*model.User, error)
		wantErr      bool
		errCheck     func(t *testing.T, err error)
		// afterSave lets us inspect what was actually passed to Save
		capturedSave func(t *testing.T, user *model.User)
	}{
		{
			name:    "updates all mutable fields when user is found",
			inputID: 7,
			inputUser: &model.User{
				Name:     strPtr("NewName"),
				Email:    strPtr("new@example.com"),
				Password: strPtr("newpass"),
				Role:     strPtr("admin"),
				About:    strPtr("new about"),
			},
			repoFindByID: func(_ context.Context, _ int) (*model.User, error) {
				// Return a copy so we do not mutate existingUser across cases
				cp := *existingUser
				cpName := *existingUser.Name
				cpEmail := *existingUser.Email
				cpPass := *existingUser.Password
				cpRole := *existingUser.Role
				cpAbout := *existingUser.About