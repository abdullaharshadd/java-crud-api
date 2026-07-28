```go
package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apperror "github.com/smartcontact/internal/smartcontact/error"
	"github.com/smartcontact/internal/smartcontact/model"
	"github.com/smartcontact/internal/smartcontact/repository"
	"github.com/smartcontact/internal/smartcontact/service"
)

// ---------------------------------------------------------------------------
// Fake / stub repository
// ---------------------------------------------------------------------------

// fakeQuerier is a test double for repository.Querier. Each field holds an
// optional override function so individual test cases can inject any behaviour
// without writing a separate type per case.
type fakeQuerier struct {
	saveFn       func(ctx context.Context, user model.User) (model.User, error)
	findAllFn    func(ctx context.Context) ([]model.User, error)
	findByIDFn   func(ctx context.Context, id int) (model.User, error)
	deleteByIDFn func(ctx context.Context, id int) error
	findByNameFn func(ctx context.Context, name string) (model.User, error)
}

func (f *fakeQuerier) Save(ctx context.Context, user model.User) (model.User, error) {
	if f.saveFn != nil {
		return f.saveFn(ctx, user)
	}
	return model.User{}, errors.New("Save: not implemented")
}

func (f *fakeQuerier) FindAll(ctx context.Context) ([]model.User, error) {
	if f.findAllFn != nil {
		return f.findAllFn(ctx)
	}
	return nil, errors.New("FindAll: not implemented")
}

func (f *fakeQuerier) FindByID(ctx context.Context, id int) (model.User, error) {
	if f.findByIDFn != nil {
		return f.findByIDFn(ctx, id)
	}
	return model.User{}, errors.New("FindByID: not implemented")
}

func (f *fakeQuerier) DeleteByID(ctx context.Context, id int) error {
	if f.deleteByIDFn != nil {
		return f.deleteByIDFn(ctx, id)
	}
	return errors.New("DeleteByID: not implemented")
}

func (f *fakeQuerier) FindByName(ctx context.Context, name string) (model.User, error) {
	if f.findByNameFn != nil {
		return f.findByNameFn(ctx, name)
	}
	return model.User{}, errors.New("FindByName: not implemented")
}

// Ensure fakeQuerier satisfies the interface at compile time.
var _ repository.Querier = (*fakeQuerier)(nil)

// ---------------------------------------------------------------------------
// Helper builders
// ---------------------------------------------------------------------------

func newService(repo repository.Querier) service.UserService {
	return service.NewUserService(repo)
}

func userNotFoundErr() error {
	return apperror.ErrUserNotFound
}

// validUser returns a user that should pass model.Validate().
func validUser(id int) model.User {
	return model.User{
		ID:    id,
		Name:  "Alice",
		Email: "alice@example.com",
	}
}

// ---------------------------------------------------------------------------
// SaveUser
// ---------------------------------------------------------------------------

func TestSaveUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		inputUser   model.User
		repoSaveFn  func(ctx context.Context, u model.User) (model.User, error)
		wantUser    model.User
		wantErr     bool
		wantErrWrap string
	}{
		{
			name:      "valid user is persisted and returned with generated id",
			inputUser: model.User{Name: "Alice", Email: "alice@example.com"},
			repoSaveFn: func(_ context.Context, u model.User) (model.User, error) {
				u.ID = 42
				return u, nil
			},
			wantUser: model.User{ID: 42, Name: "Alice", Email: "alice@example.com"},
			wantErr:  false,
		},
		{
			name:      "user with existing id is updated/overwritten and returned",
			inputUser: model.User{ID: 7, Name: "Bob", Email: "bob@example.com"},
			repoSaveFn: func(_ context.Context, u model.User) (model.User, error) {
				return u, nil
			},
			wantUser: model.User{ID: 7, Name: "Bob", Email: "bob@example.com"},
			wantErr:  false,
		},
		{
			name:        "validation failure returns wrapped error and no user",
			inputUser:   model.User{}, // empty — should fail Validate
			wantErr:     true,
			wantErrWrap: "save user",
		},
		{
			name:      "repository error is propagated",
			inputUser: model.User{Name: "Carol", Email: "carol@example.com"},
			repoSaveFn: func(_ context.Context, _ model.User) (model.User, error) {
				return model.User{}, errors.New("db unavailable")
			},
			wantErr:     true,
			wantErrWrap: "save user",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := &fakeQuerier{saveFn: tc.repoSaveFn}
			svc := newService(repo)

			got, err := svc.SaveUser(context.Background(), tc.inputUser)

			if tc.wantErr {
				require.Error(t, err)
				if tc.wantErrWrap != "" {
					assert.Contains(t, err.Error(), tc.wantErrWrap)
				}
				assert.Equal(t, model.User{}, got)
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
	t.Parallel()

	tests := []struct {
		name        string
		repoFn      func(ctx context.Context) ([]model.User, error)
		wantUsers   []model.User
		wantErr     bool
		wantErrWrap string
	}{
		{
			name: "returns all users when records exist",
			repoFn: func(_ context.Context) ([]model.User, error) {
				return []model.User{
					{ID: 1, Name: "Alice", Email: "alice@example.com"},
					{ID: 2, Name: "Bob", Email: "bob@example.com"},
				}, nil
			},
			wantUsers: []model.User{
				{ID: 1, Name: "Alice", Email: "alice@example.com"},
				{ID: 2, Name: "Bob", Email: "bob@example.com"},
			},
		},
		{
			name: "returns empty (non-nil) slice when no users exist",
			repoFn: func(_ context.Context) ([]model.User, error) {
				return []model.User{}, nil
			},
			wantUsers: []model.User{},
		},
		{
			name: "repository error is propagated",
			repoFn: func(_ context.Context) ([]model.User, error) {
				return nil, errors.New("connection reset")
			},
			wantErr:     true,
			wantErrWrap: "fetch user list",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := &fakeQuerier{findAllFn: tc.repoFn}
			svc := newService(repo)

			got, err := svc.FetchUserList(context.Background())

			if tc.wantErr {
				require.Error(t, err)
				if tc.wantErrWrap != "" {
					assert.Contains(t, err.Error(), tc.wantErrWrap)
				}
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.wantUsers, got)
		})
	}
}

// ---------------------------------------------------------------------------
// GetUserByID
// ---------------------------------------------------------------------------

func TestGetUserByID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		requestID      int
		repoFn         func(ctx context.Context, id int) (model.User, error)
		wantUser       model.User
		wantErr        bool
		wantNotFound   bool
		wantErrContain string
	}{
		{
			name:      "returns user when id matches",
			requestID: 1,
			repoFn: func(_ context.Context, id int) (model.User, error) {
				return model.User{ID: id, Name: "Alice", Email: "alice@example.com"}, nil
			},
			wantUser: model.User{ID: 1, Name: "Alice", Email: "alice@example.com"},
		},
		{
			name:      "returns ErrUserNotFound when no matching user",
			requestID: 99,
			repoFn: func(_ context.Context, _ int) (model.User, error) {
				return model.User{}, apperror.ErrUserNotFound
			},
			wantErr:      true,
			wantNotFound: true,
		},
		{
			name:      "propagates non-not-found repository error",
			requestID: 5,
			repoFn: func(_ context.Context, _ int) (model.User, error) {
				return model.User{}, errors.New("timeout")
			},
			wantErr:        true,
			wantNotFound:   false,
			wantErrContain: "get user by id 5",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := &fakeQuerier{findByIDFn: tc.repoFn}
			svc := newService(repo)

			got, err := svc.GetUserByID(context.Background(), tc.requestID)

			if tc.wantErr {
				require.Error(t, err)
				if tc.wantNotFound {
					assert.True(t, apperror.IsUserNotFound(err),
						"expected a UserNotFound error but got: %v", err)
				}
				if tc.wantErrContain != "" {
					assert.Contains(t, err.Error(), tc.wantErrContain)
				}
				assert.Equal(t, model.User{}, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.wantUser, got)
			assert.Equal(t, tc.requestID, got.ID,
				"returned user ID must match requested ID")
		})
	}
}

// ---------------------------------------------------------------------------
// DeleteUser
// ---------------------------------------------------------------------------

func TestDeleteUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		requestID      int
		findByIDFn     func(ctx context.Context, id int) (model.User, error)
		deleteByIDFn   func(ctx context.Context, id int) error
		wantErr        bool
		wantNotFound   bool
		wantErrContain string
	}{
		{
			name:      "deletes existing user successfully",
			requestID: 1,
			findByIDFn: func(_ context.Context, id int) (model.User, error) {
				return model.User{ID: id}, nil
			},
			deleteByIDFn: func(_ context.Context, _ int) error {
				return nil
			},
			wantErr: false,
		},
		{
			name:      "returns ErrUserNotFound when no matching user",
			requestID: 404,
			findByIDFn: func(_ context.Context, _ int) (model.User, error) {
				return model.User{}, apperror.ErrUserNotFound
			},
			wantErr:      true,
			wantNotFound: true,
		},
		{
			name:      "propagates FindByID non-not-found error",
			requestID: 5,
			findByIDFn: func(_ context.Context, _ int) (model.User, error) {
				return model.User{}, errors.New("db error")
			},
			wantErr:        true,
			wantNotFound:   false,
			wantErrContain: "delete user 5",
		},
		{
			name:      "propagates DeleteByID error",
			requestID: 2,
			findByIDFn: func(_ context.Context, id int) (model.User, error) {
				return model.User{ID: id}, nil
			},
			deleteByIDFn: func(_ context.Context, _ int) error {
				return errors.New("delete failed")
			},
			wantErr:        true,
			wantNotFound:   false,
			wantErrContain: "delete user 2",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			deleted := false
			deleteByIDFn := tc.deleteByIDFn
			if deleteByIDFn == nil {
				deleteByIDFn = func(_ context.Context, _ int) error {
					deleted = true
					return nil
				}
			} else {
				origFn := deleteByIDFn
				deleteByIDFn = func(ctx context.Context, id int) error {
					deleted = true
					return origFn(ctx, id)
				}
			}

			repo := &fakeQuerier{
				findByIDFn:   tc.findByIDFn,
				deleteByIDFn: deleteByIDFn,
			}
			svc := newService(repo)

			err := svc.DeleteUser(context.Background(), tc.requestID)

			if tc.wantErr {
				require.Error(t, err)
				if tc.wantNotFound {
					assert.True(t, apperror.IsUserNotFound(err),
						"expected UserNotFound but got: %v", err)
				}
				if tc.wantErrContain != "" {
					assert.Contains(t, err.Error(), tc.wantErrContain)
				}
				return
			}

			require.NoError(t, err)
			assert.True(t, deleted, "DeleteByID should have been called")
		})
	}
}

// ---------------------------------------------------------------------------
// UpdateUser
// ---------------------------------------------------------------------------

func TestUpdateUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		requestID      int
		inputUser      model.User
		findByIDFn     func(ctx context.Context, id int) (model.User, error)
		saveFn         func(ctx context.Context, u model.User) (model.User, error)
		wantUser       model.User
		wantErr        bool
		wantNotFound   bool
		wantErrContain string
	}{
		{
			name:      "updates existing user and returns updated record",
			requestID: 3,
			inputUser: model.User{Name: "Updated", Email: "updated@example.com"},
			findByIDFn: func(_ context.Context, id int) (model.User, error) {
				return model.User{ID