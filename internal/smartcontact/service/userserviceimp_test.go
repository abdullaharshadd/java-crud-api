```go
package service

import (
	"context"
	"errors"
	"fmt"
	"testing"

	smartcontacterror "internal/smartcontact/error"
	"internal/smartcontact/model"

	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// Mock repository
// ---------------------------------------------------------------------------

type mockUserRepository struct {
	// SaveFunc is called when Save is invoked.
	SaveFunc func(ctx context.Context, user *model.User) (*model.User, error)
	// FindAllFunc is called when FindAll is invoked.
	FindAllFunc func(ctx context.Context) ([]*model.User, error)
	// FindByIDFunc is called when FindByID is invoked.
	FindByIDFunc func(ctx context.Context, id int) (*model.User, error)
	// DeleteByIDFunc is called when DeleteByID is invoked.
	DeleteByIDFunc func(ctx context.Context, id int) error
	// FindByNameFunc is called when FindByName is invoked.
	FindByNameFunc func(ctx context.Context, name string) (*model.User, error)
}

func (m *mockUserRepository) Save(ctx context.Context, user *model.User) (*model.User, error) {
	return m.SaveFunc(ctx, user)
}

func (m *mockUserRepository) FindAll(ctx context.Context) ([]*model.User, error) {
	return m.FindAllFunc(ctx)
}

func (m *mockUserRepository) FindByID(ctx context.Context, id int) (*model.User, error) {
	return m.FindByIDFunc(ctx, id)
}

func (m *mockUserRepository) DeleteByID(ctx context.Context, id int) error {
	return m.DeleteByIDFunc(ctx, id)
}

func (m *mockUserRepository) FindByName(ctx context.Context, name string) (*model.User, error) {
	return m.FindByNameFunc(ctx, name)
}

// ---------------------------------------------------------------------------
// Helper: build a userService with a given mock repo.
// ---------------------------------------------------------------------------

func newTestService(repo *mockUserRepository) *userService {
	return &userService{repo: repo}
}

// ---------------------------------------------------------------------------
// Save tests
// ---------------------------------------------------------------------------

func TestSave(t *testing.T) {
	validUser := &model.User{ID: 0, Name: "Alice"}
	savedUser := &model.User{ID: 1, Name: "Alice"}
	repoErr := errors.New("db error")

	tests := []struct {
		name        string
		inputUser   *model.User
		repoSave    func(ctx context.Context, u *model.User) (*model.User, error)
		wantUser    *model.User
		wantErr     bool
		errContains string
	}{
		{
			name:      "valid user is persisted and returned",
			inputUser: validUser,
			repoSave: func(_ context.Context, u *model.User) (*model.User, error) {
				return savedUser, nil
			},
			wantUser: savedUser,
			wantErr:  false,
		},
		{
			name:      "user with existing id is updated and returned",
			inputUser: &model.User{ID: 5, Name: "Bob"},
			repoSave: func(_ context.Context, u *model.User) (*model.User, error) {
				return u, nil // echo back what was sent
			},
			wantUser: &model.User{ID: 5, Name: "Bob"},
			wantErr:  false,
		},
		{
			name:        "nil user returns error without calling repo",
			inputUser:   nil,
			repoSave:    nil, // must not be called
			wantUser:    nil,
			wantErr:     true,
			errContains: "user must not be nil",
		},
		{
			name:      "repository error is propagated wrapped",
			inputUser: validUser,
			repoSave: func(_ context.Context, u *model.User) (*model.User, error) {
				return nil, repoErr
			},
			wantUser:    nil,
			wantErr:     true,
			errContains: "save user",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockUserRepository{}
			if tc.repoSave != nil {
				repo.SaveFunc = tc.repoSave
			}
			svc := newTestService(repo)

			got, err := svc.Save(context.Background(), tc.inputUser)

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
		})
	}
}

// ---------------------------------------------------------------------------
// List tests
// ---------------------------------------------------------------------------

func TestList(t *testing.T) {
	users := []*model.User{
		{ID: 1, Name: "Alice"},
		{ID: 2, Name: "Bob"},
	}
	repoErr := errors.New("db error")

	tests := []struct {
		name        string
		repoFindAll func(ctx context.Context) ([]*model.User, error)
		wantUsers   []*model.User
		wantErr     bool
		errContains string
	}{
		{
			name: "users exist — all returned",
			repoFindAll: func(_ context.Context) ([]*model.User, error) {
				return users, nil
			},
			wantUsers: users,
			wantErr:   false,
		},
		{
			name: "no users exist — empty slice returned",
			repoFindAll: func(_ context.Context) ([]*model.User, error) {
				return []*model.User{}, nil
			},
			wantUsers: []*model.User{},
			wantErr:   false,
		},
		{
			name: "repository error is propagated wrapped",
			repoFindAll: func(_ context.Context) ([]*model.User, error) {
				return nil, repoErr
			},
			wantUsers:   nil,
			wantErr:     true,
			errContains: "list users",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockUserRepository{FindAllFunc: tc.repoFindAll}
			svc := newTestService(repo)

			got, err := svc.List(context.Background())

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
		})
	}
}

// ---------------------------------------------------------------------------
// GetByID tests
// ---------------------------------------------------------------------------

func TestGetByID(t *testing.T) {
	foundUser := &model.User{ID: 42, Name: "Alice"}

	tests := []struct {
		name             string
		id               int
		repoFindByID     func(ctx context.Context, id int) (*model.User, error)
		wantUser         *model.User
		wantErr          bool
		wantNotFound     bool   // errors.Is(err, ErrUserNotFound)
		errMsgContains   string
	}{
		{
			name: "existing id returns user",
			id:   42,
			repoFindByID: func(_ context.Context, id int) (*model.User, error) {
				return foundUser, nil
			},
			wantUser: foundUser,
			wantErr:  false,
		},
		{
			name: "repo returns ErrUserNotFound — service returns not-found error",
			id:   99,
			repoFindByID: func(_ context.Context, id int) (*model.User, error) {
				return nil, smartcontacterror.ErrUserNotFound
			},
			wantUser:       nil,
			wantErr:        true,
			wantNotFound:   true,
			errMsgContains: "User are not available",
		},
		{
			name: "repo returns nil user without error — service returns not-found error",
			id:   88,
			repoFindByID: func(_ context.Context, id int) (*model.User, error) {
				return nil, nil
			},
			wantUser:       nil,
			wantErr:        true,
			wantNotFound:   true,
			errMsgContains: "User are not available",
		},
		{
			name: "repo returns generic error — propagated wrapped",
			id:   7,
			repoFindByID: func(_ context.Context, id int) (*model.User, error) {
				return nil, fmt.Errorf("connection refused")
			},
			wantUser:       nil,
			wantErr:        true,
			wantNotFound:   false,
			errMsgContains: "get user by id 7",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockUserRepository{FindByIDFunc: tc.repoFindByID}
			svc := newTestService(repo)

			got, err := svc.GetByID(context.Background(), tc.id)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.errMsgContains != "" {
					assert.Contains(t, err.Error(), tc.errMsgContains)
				}
				if tc.wantNotFound {
					assert.True(t, errors.Is(err, smartcontacterror.ErrUserNotFound),
						"expected ErrUserNotFound in error chain")
				}
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantUser, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Delete tests
// ---------------------------------------------------------------------------

func TestDelete(t *testing.T) {
	errEmptyResult := smartcontacterror.ErrEmptyResultDelete // repository sentinel
	genericErr := errors.New("db connection lost")

	tests := []struct {
		name           string
		id             int
		repoDeleteByID func(ctx context.Context, id int) error
		wantErr        bool
		errContains    string
		wrapsEmptyResult bool
	}{
		{
			name: "existing id — deleted successfully",
			id:   1,
			repoDeleteByID: func(_ context.Context, id int) error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "non-existing id — repo ErrEmptyResultDelete propagates wrapped",
			id:   999,
			repoDeleteByID: func(_ context.Context, id int) error {
				return errEmptyResult
			},
			wantErr:          true,
			errContains:      "delete user 999",
			wrapsEmptyResult: true,
		},
		{
			name: "generic repo error is propagated wrapped",
			id:   2,
			repoDeleteByID: func(_ context.Context, id int) error {
				return genericErr
			},
			wantErr:     true,
			errContains: "delete user 2",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockUserRepository{DeleteByIDFunc: tc.repoDeleteByID}
			svc := newTestService(repo)

			err := svc.Delete(context.Background(), tc.id)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.errContains != "" {
					assert.Contains(t, err.Error(), tc.errContains)
				}
				if tc.wrapsEmptyResult {
					assert.True(t, errors.Is(err, smartcontacterror.ErrEmptyResultDelete),
						"expected ErrEmptyResultDelete in error chain")
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Update tests
// ---------------------------------------------------------------------------

func TestUpdate(t *testing.T) {
	repoErr := errors.New("db error")

	tests := []struct {
		name        string
		id          int
		inputUser   *model.User
		repoSave    func(ctx context.Context, u *model.User) (*model.User, error)
		wantUser    *model.User
		wantErr     bool
		errContains string
		// used to verify that the id is set on the user before save
		wantSavedID int
	}{
		{
			name:      "valid id and user — id is forced onto user before save",
			id:        10,
			inputUser: &model.User{ID: 0, Name: "Charlie"},
			repoSave: func(_ context.Context, u *model.User) (*model.User, error) {
				// the service must have set u.ID = 10 before calling Save
				return u, nil
			},
			wantUser:    &model.User{ID: 10, Name: "Charlie"},
			wantErr:     false,
			wantSavedID: 10,
		},
		{
			name:        "nil user returns error without calling repo",
			id:          5,
			inputUser:   nil,
			repoSave:    nil,
			wantUser:    nil,
			wantErr:     true,
			errContains: "user must not be nil",
		},
		{
			name:      "repository error is propagated wrapped",
			id:        3,
			inputUser: &model.User{Name: "Dave"},
			repoSave: func(_ context.Context, u *model.User) (*model.User, error) {
				return nil, repoErr
			},
			wantUser:    nil,
			wantErr:     true,
			errContains: "update user 3",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var savedUser *model.User
			repo := &mockUserRepository{}
			if tc.repoSave != nil {
				repo.SaveFunc = func(ctx context.Context, u *model.User) (*model.User, error) {
					savedUser = u
					return tc.repoSave(ctx, u)
				}
			}
			svc := newTestService(repo)

			got, err := svc.Update(context.Background(), tc.id, tc.inputUser)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.errContains != "" {
					assert.Contains(t, err.Error(), tc.errContains)
				}
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantUser, got)
				// Verify the service set the id before calling repo.Save
				if tc.wantSavedID != 0 {
					assert.Equal(t, tc.wantSavedID, savedUser.ID,
						"id must be forced onto the user entity before persistence")
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// GetByName tests
// ---------------------------------------------------------------------------

func TestGetByName(t *testing.T) {
	foundUser := &model.User{ID: 7, Name: "Eve"}

	tests := []struct {
		name             string
		inputName        string
		repoFindByName   func(ctx context.Context, name string) (*model.User, error)
		wantUser         *model.User
		wantErr          bool
		wantNotFound     bool
		errMsgContains   string