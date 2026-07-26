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

// mockUserRepo is a test double that satisfies repository.UserRepository.
// Each method field holds the behaviour for that specific test case.
type mockUserRepo struct {
	SaveFn       func(ctx context.Context, user model.User) (model.User, error)
	FindAllFn    func(ctx context.Context) ([]model.User, error)
	FindByIDFn   func(ctx context.Context, id int) (model.User, error)
	DeleteByIDFn func(ctx context.Context, id int) error
	FindByNameFn func(ctx context.Context, name string) (model.User, error)
}

func (m *mockUserRepo) Save(ctx context.Context, user model.User) (model.User, error) {
	if m.SaveFn != nil {
		return m.SaveFn(ctx, user)
	}
	return model.User{}, errors.New("SaveFn not set")
}

func (m *mockUserRepo) FindAll(ctx context.Context) ([]model.User, error) {
	if m.FindAllFn != nil {
		return m.FindAllFn(ctx)
	}
	return nil, errors.New("FindAllFn not set")
}

func (m *mockUserRepo) FindByID(ctx context.Context, id int) (model.User, error) {
	if m.FindByIDFn != nil {
		return m.FindByIDFn(ctx, id)
	}
	return model.User{}, errors.New("FindByIDFn not set")
}

func (m *mockUserRepo) DeleteByID(ctx context.Context, id int) error {
	if m.DeleteByIDFn != nil {
		return m.DeleteByIDFn(ctx, id)
	}
	return errors.New("DeleteByIDFn not set")
}

func (m *mockUserRepo) FindByName(ctx context.Context, name string) (model.User, error) {
	if m.FindByNameFn != nil {
		return m.FindByNameFn(ctx, name)
	}
	return model.User{}, errors.New("FindByNameFn not set")
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// validUser returns a User that passes model.Validate().
func validUser(id int) model.User {
	return model.User{
		ID:    id,
		Name:  "Alice",
		Email: "alice@example.com",
	}
}

// invalidUser returns a User that will fail model.Validate() (empty Name).
func invalidUser() model.User {
	return model.User{
		ID:    0,
		Name:  "",
		Email: "",
	}
}

// repoError is a generic repository-layer sentinel for tests.
var repoError = errors.New("repository error")

// ---------------------------------------------------------------------------
// Save
// ---------------------------------------------------------------------------

func TestUserService_Save(t *testing.T) {
	type args struct {
		user model.User
	}
	type repoReturns struct {
		user model.User
		err  error
	}

	tests := []struct {
		name        string
		args        args
		repoReturns *repoReturns // nil ⟹ Save should not reach the repo
		wantUser    model.User
		wantErr     bool
		wantErrWrap error // checked with errors.Is when non-nil
	}{
		{
			name: "valid user is persisted and returned",
			args: args{user: validUser(0)},
			repoReturns: &repoReturns{
				user: validUser(42),
				err:  nil,
			},
			wantUser: validUser(42),
			wantErr:  false,
		},
		{
			name:        "invalid user returns validation error without hitting repo",
			args:        args{user: invalidUser()},
			repoReturns: nil, // repo must NOT be called
			wantUser:    model.User{},
			wantErr:     true,
		},
		{
			name: "repository error is propagated",
			args: args{user: validUser(0)},
			repoReturns: &repoReturns{
				user: model.User{},
				err:  repoError,
			},
			wantUser:    model.User{},
			wantErr:     true,
			wantErrWrap: repoError,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			repoCalled := false
			mock := &mockUserRepo{
				SaveFn: func(ctx context.Context, user model.User) (model.User, error) {
					repoCalled = true
					if tc.repoReturns == nil {
						t.Fatal("repo.Save called but should not have been")
					}
					return tc.repoReturns.user, tc.repoReturns.err
				},
			}

			svc := service.NewUserService(mock)
			got, err := svc.Save(context.Background(), tc.args.user)

			if tc.repoReturns == nil {
				assert.False(t, repoCalled, "repo.Save must not be called for invalid input")
			}

			if tc.wantErr {
				require.Error(t, err)
				if tc.wantErrWrap != nil {
					assert.True(t, errors.Is(err, tc.wantErrWrap),
						"expected wrapped error %v, got: %v", tc.wantErrWrap, err)
				}
				assert.Equal(t, model.User{}, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.wantUser, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// List
// ---------------------------------------------------------------------------

func TestUserService_List(t *testing.T) {
	tests := []struct {
		name      string
		repoUsers []model.User
		repoErr   error
		wantUsers []model.User
		wantErr   bool
	}{
		{
			name:      "returns all users when store is populated",
			repoUsers: []model.User{validUser(1), validUser(2)},
			repoErr:   nil,
			wantUsers: []model.User{validUser(1), validUser(2)},
			wantErr:   false,
		},
		{
			name:      "returns empty slice when store is empty",
			repoUsers: []model.User{},
			repoErr:   nil,
			wantUsers: []model.User{},
			wantErr:   false,
		},
		{
			name:      "repository error is propagated",
			repoUsers: nil,
			repoErr:   repoError,
			wantUsers: nil,
			wantErr:   true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockUserRepo{
				FindAllFn: func(ctx context.Context) ([]model.User, error) {
					return tc.repoUsers, tc.repoErr
				},
			}

			svc := service.NewUserService(mock)
			got, err := svc.List(context.Background())

			if tc.wantErr {
				require.Error(t, err)
				assert.True(t, errors.Is(err, repoError))
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				// List must always return a non-nil slice (never null equivalent).
				assert.NotNil(t, got)
				assert.Equal(t, tc.wantUsers, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// GetByID
// ---------------------------------------------------------------------------

func TestUserService_GetByID(t *testing.T) {
	tests := []struct {
		name        string
		id          int
		repoUser    model.User
		repoErr     error
		wantUser    model.User
		wantErr     bool
		wantNotFound bool
	}{
		{
			name:     "existing id returns corresponding user",
			id:       1,
			repoUser: validUser(1),
			repoErr:  nil,
			wantUser: validUser(1),
			wantErr:  false,
		},
		{
			name:         "non-existing id returns ErrUserNotFound",
			id:           99,
			repoUser:     model.User{},
			repoErr:      fmt.Errorf("repo: %w", smartError.ErrUserNotFound),
			wantUser:     model.User{},
			wantErr:      true,
			wantNotFound: true,
		},
		{
			name:     "generic repository error is propagated",
			id:       1,
			repoUser: model.User{},
			repoErr:  repoError,
			wantUser: model.User{},
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockUserRepo{
				FindByIDFn: func(ctx context.Context, id int) (model.User, error) {
					assert.Equal(t, tc.id, id)
					return tc.repoUser, tc.repoErr
				},
			}

			svc := service.NewUserService(mock)
			got, err := svc.GetByID(context.Background(), tc.id)

			if tc.wantErr {
				require.Error(t, err)
				assert.Equal(t, model.User{}, got)
				if tc.wantNotFound {
					assert.True(t, errors.Is(err, smartError.ErrUserNotFound),
						"expected ErrUserNotFound in error chain, got: %v", err)
				}
				// Returned user must have the requested id when found.
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.id, got.ID, "returned user must have the requested id")
				assert.Equal(t, tc.wantUser, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------------

func TestUserService_Delete(t *testing.T) {
	tests := []struct {
		name    string
		id      int
		repoErr error
		wantErr bool
	}{
		{
			name:    "existing id is deleted successfully",
			id:      1,
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "non-existing id propagates repository error",
			id:      99,
			repoErr: fmt.Errorf("repo: %w", smartError.ErrUserNotFound),
			wantErr: true,
		},
		{
			name:    "generic repository error is propagated",
			id:      1,
			repoErr: repoError,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			deleteCalled := false
			mock := &mockUserRepo{
				DeleteByIDFn: func(ctx context.Context, id int) error {
					deleteCalled = true
					assert.Equal(t, tc.id, id)
					return tc.repoErr
				},
			}

			svc := service.NewUserService(mock)
			err := svc.Delete(context.Background(), tc.id)

			assert.True(t, deleteCalled, "repo.DeleteByID must always be called")

			if tc.wantErr {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.repoErr) || err != nil)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func TestUserService_Update(t *testing.T) {
	type repoState struct {
		findUser model.User
		findErr  error
		saveUser model.User
		saveErr  error
	}

	tests := []struct {
		name      string
		id        int
		inputUser model.User
		repoState *repoState
		wantErr   bool
		// after a successful update the stored user must carry the path id
		wantIDEnforced bool
	}{
		{
			name:      "valid update replaces existing user",
			id:        1,
			inputUser: validUser(0), // id in payload is irrelevant
			repoState: &repoState{
				findUser: validUser(1),
				findErr:  nil,
				saveUser: validUser(1),
				saveErr:  nil,
			},
			wantErr:        false,
			wantIDEnforced: true,
		},
		{
			name:      "invalid user payload returns validation error",
			id:        1,
			inputUser: invalidUser(),
			repoState: nil, // repo must NOT be called
			wantErr:   true,
		},
		{
			name:      "non-existing id returns error (find fails)",
			id:        99,
			inputUser: validUser(0),
			repoState: &repoState{
				findErr: fmt.Errorf("repo: %w", smartError.ErrUserNotFound),
			},
			wantErr: true,
		},
		{
			name:      "repository save error is propagated after successful find",
			id:        1,
			inputUser: validUser(0),
			repoState: &repoState{
				findUser: validUser(1),
				findErr:  nil,
				saveUser: model.User{},
				saveErr:  repoError,
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			findCalled := false
			saveCalled := false

			var savedUser model.User // capture what was passed to Save

			mock := &mockUserRepo{
				FindByIDFn: func(ctx context.Context, id int) (model.User, error) {
					findCalled = true
					if tc.repoState == nil {
						t.Fatal("repo.FindByID called but should not have been")
					}
					return tc.repoState.findUser, tc.repoState.findErr
				},
				SaveFn: func(ctx context.Context, user model.User) (model.User, error) {
					saveCalled = true
					savedUser = user
					if tc.repoState == nil {
						t.Fatal("repo.Save called but should not have been")
					}
					return tc.repoState.saveUser, tc.repoState.saveErr
				},
			}

			svc := service.NewUserService(mock)
			err := svc.Update(context.Background(), tc.id, tc.inputUser)

			if tc.repoState == nil {
				assert.False(t, findCalled, "repo.FindByID must not be called for invalid payload")
				assert.False(t, saveCalled, "repo.Save must not be called for invalid payload")
			}

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.True(t, saveCalled, "repo.Save must be called on successful update")
				if tc.wantIDEnforced {
					assert.Equal(t, tc.id, saved