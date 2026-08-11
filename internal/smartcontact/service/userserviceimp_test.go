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
	"github.com/smartContact/internal/smartcontact/service"
)

// ---------------------------------------------------------------------------
// Fake repository
// ---------------------------------------------------------------------------

type fakeUserRepository struct {
	saveFunc     func(ctx context.Context, user *model.User) (*model.User, error)
	findAllFunc  func(ctx context.Context) ([]*model.User, error)
	findByIDFunc func(ctx context.Context, id int) (*model.User, bool, error)
	deleteByID   func(ctx context.Context, id int) error
	findByName   func(ctx context.Context, name string) (*model.User, error)
}

func (f *fakeUserRepository) Save(ctx context.Context, user *model.User) (*model.User, error) {
	if f.saveFunc != nil {
		return f.saveFunc(ctx, user)
	}
	return nil, errors.New("Save not configured")
}

func (f *fakeUserRepository) FindAll(ctx context.Context) ([]*model.User, error) {
	if f.findAllFunc != nil {
		return f.findAllFunc(ctx)
	}
	return nil, errors.New("FindAll not configured")
}

func (f *fakeUserRepository) FindByID(ctx context.Context, id int) (*model.User, bool, error) {
	if f.findByIDFunc != nil {
		return f.findByIDFunc(ctx, id)
	}
	return nil, false, errors.New("FindByID not configured")
}

func (f *fakeUserRepository) DeleteByID(ctx context.Context, id int) error {
	if f.deleteByID != nil {
		return f.deleteByID(ctx, id)
	}
	return errors.New("DeleteByID not configured")
}

func (f *fakeUserRepository) FindByName(ctx context.Context, name string) (*model.User, error) {
	if f.findByName != nil {
		return f.findByName(ctx, name)
	}
	return nil, errors.New("FindByName not configured")
}

// ---------------------------------------------------------------------------
// Helper
// ---------------------------------------------------------------------------

func newService(repo *fakeUserRepository) service.UserService {
	return service.NewUserService(repo)
}

// ---------------------------------------------------------------------------
// SaveUser tests
// ---------------------------------------------------------------------------

func TestSaveUser(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		input       *model.User
		repoReturn  *model.User
		repoErr     error
		wantUser    *model.User
		wantErr     bool
		wantErrWrap string
	}{
		{
			name:       "given a valid User object returns saved entity",
			input:      &model.User{Name: "Alice"},
			repoReturn: &model.User{ID: 1, Name: "Alice"},
			wantUser:   &model.User{ID: 1, Name: "Alice"},
		},
		{
			name:       "given a User with an existing id returns updated entity",
			input:      &model.User{ID: 5, Name: "Bob"},
			repoReturn: &model.User{ID: 5, Name: "Bob"},
			wantUser:   &model.User{ID: 5, Name: "Bob"},
		},
		{
			name:        "repository error is wrapped and propagated",
			input:       &model.User{Name: "Charlie"},
			repoErr:     errors.New("db connection failed"),
			wantErr:     true,
			wantErrWrap: "save user",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeUserRepository{
				saveFunc: func(_ context.Context, user *model.User) (*model.User, error) {
					if tc.repoErr != nil {
						return nil, tc.repoErr
					}
					return tc.repoReturn, nil
				},
			}
			svc := newService(repo)

			got, err := svc.SaveUser(ctx, tc.input)

			if tc.wantErr {
				require.Error(t, err)
				if tc.wantErrWrap != "" {
					assert.Contains(t, err.Error(), tc.wantErrWrap)
				}
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.wantUser, got)
		})
	}
}

// Invariant: SaveUser passes the user to the repository without transformation.
func TestSaveUser_NoTransformation(t *testing.T) {
	ctx := context.Background()

	input := &model.User{ID: 0, Name: "No-Transform"}
	var captured *model.User

	repo := &fakeUserRepository{
		saveFunc: func(_ context.Context, u *model.User) (*model.User, error) {
			captured = u
			return u, nil
		},
	}
	svc := newService(repo)

	_, err := svc.SaveUser(ctx, input)
	require.NoError(t, err)
	assert.Equal(t, input, captured, "SaveUser must pass the user to the repository unmodified")
}

// ---------------------------------------------------------------------------
// FetchUserList tests
// ---------------------------------------------------------------------------

func TestFetchUserList(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		repoReturn []*model.User
		repoErr    error
		wantUsers  []*model.User
		wantErr    bool
	}{
		{
			name: "given users exist returns all entities",
			repoReturn: []*model.User{
				{ID: 1, Name: "Alice"},
				{ID: 2, Name: "Bob"},
			},
			wantUsers: []*model.User{
				{ID: 1, Name: "Alice"},
				{ID: 2, Name: "Bob"},
			},
		},
		{
			name:       "given no users exist returns empty list",
			repoReturn: []*model.User{},
			wantUsers:  []*model.User{},
		},
		{
			name:    "repository error is wrapped and propagated",
			repoErr: errors.New("timeout"),
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeUserRepository{
				findAllFunc: func(_ context.Context) ([]*model.User, error) {
					return tc.repoReturn, tc.repoErr
				},
			}
			svc := newService(repo)

			got, err := svc.FetchUserList(ctx)

			if tc.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "fetch user list")
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.wantUsers, got)
		})
	}
}

// Invariant: FetchUserList returns exactly what the repository provides.
func TestFetchUserList_ReturnsExactRepositoryResult(t *testing.T) {
	ctx := context.Background()
	expected := []*model.User{{ID: 99, Name: "Exact"}}

	repo := &fakeUserRepository{
		findAllFunc: func(_ context.Context) ([]*model.User, error) {
			return expected, nil
		},
	}
	svc := newService(repo)

	got, err := svc.FetchUserList(ctx)
	require.NoError(t, err)
	assert.Equal(t, expected, got)
}

// ---------------------------------------------------------------------------
// FetchUserByID tests
// ---------------------------------------------------------------------------

func TestFetchUserByID(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		id           int
		repoUser     *model.User
		repoFound    bool
		repoErr      error
		wantUser     *model.User
		wantErr      bool
		wantNotFound bool
		wantMsg      string
	}{
		{
			name:      "given matching id returns user entity",
			id:        1,
			repoUser:  &model.User{ID: 1, Name: "Alice"},
			repoFound: true,
			wantUser:  &model.User{ID: 1, Name: "Alice"},
		},
		{
			name:         "given non-matching id returns UserNotFound error",
			id:           42,
			repoFound:    false,
			wantErr:      true,
			wantNotFound: true,
			wantMsg:      "User are not available",
		},
		{
			name:    "repository error is wrapped and propagated",
			id:      7,
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeUserRepository{
				findByIDFunc: func(_ context.Context, id int) (*model.User, bool, error) {
					return tc.repoUser, tc.repoFound, tc.repoErr
				},
			}
			svc := newService(repo)

			got, err := svc.FetchUserByID(ctx, tc.id)

			if tc.wantErr {
				require.Error(t, err)

				if tc.wantNotFound {
					var notFound *apperr.UserNotFound
					assert.True(t, errors.As(err, &notFound),
						"expected *apperr.UserNotFound, got %T: %v", err, err)
					if tc.wantMsg != "" {
						assert.Contains(t, err.Error(), tc.wantMsg)
					}
				}
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got, "FetchUserByID must never return nil without an error")
			assert.Equal(t, tc.wantUser, got)
		})
	}
}

// Invariant: FetchUserByID always returns UserNotFound when not present.
func TestFetchUserByID_AlwaysUserNotFoundWhenMissing(t *testing.T) {
	ctx := context.Background()
	ids := []int{0, 1, 100, 999}

	for _, id := range ids {
		t.Run(fmt.Sprintf("id=%d", id), func(t *testing.T) {
			repo := &fakeUserRepository{
				findByIDFunc: func(_ context.Context, _ int) (*model.User, bool, error) {
					return nil, false, nil
				},
			}
			svc := newService(repo)

			_, err := svc.FetchUserByID(ctx, id)
			require.Error(t, err)

			var notFound *apperr.UserNotFound
			assert.True(t, errors.As(err, &notFound))
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
		wantErrWrap string
		capturedID  int
	}{
		{
			name: "given existing id deletes user successfully",
			id:   3,
		},
		{
			name:        "repository error on non-existent id is propagated",
			id:          999,
			repoErr:     errors.New("not found"),
			wantErr:     true,
			wantErrWrap: "delete user",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var deletedID int
			repo := &fakeUserRepository{
				deleteByID: func(_ context.Context, id int) error {
					deletedID = id
					return tc.repoErr
				},
			}
			svc := newService(repo)

			err := svc.DeleteUser(ctx, tc.id)

			if tc.wantErr {
				require.Error(t, err)
				if tc.wantErrWrap != "" {
					assert.Contains(t, err.Error(), tc.wantErrWrap)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.id, deletedID, "must delegate with correct id")
		})
	}
}

// Invariant: DeleteUser always delegates to the repository with the exact id.
func TestDeleteUser_DelegatesToRepositoryWithCorrectID(t *testing.T) {
	ctx := context.Background()
	wantID := 77

	var got int
	repo := &fakeUserRepository{
		deleteByID: func(_ context.Context, id int) error {
			got = id
			return nil
		},
	}
	svc := newService(repo)

	err := svc.DeleteUser(ctx, wantID)
	require.NoError(t, err)
	assert.Equal(t, wantID, got)
}

// ---------------------------------------------------------------------------
// UpdateUser tests
// ---------------------------------------------------------------------------

func TestUpdateUser(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		id           int
		user         *model.User
		repoErr      error
		wantErr      bool
		wantErrWrap  string
		wantSavedID  int
		wantSavedPtr bool
	}{
		{
			name:        "given id and user sets id on user and persists",
			id:          10,
			user:        &model.User{Name: "Updated"},
			wantSavedID: 10,
		},
		{
			name:        "given id overwrites existing id on user object",
			id:          20,
			user:        &model.User{ID: 999, Name: "Overwrite"},
			wantSavedID: 20,
		},
		{
			name:        "nil user returns error immediately",
			id:          1,
			user:        nil,
			wantErr:     true,
			wantErrWrap: "update user",
		},
		{
			name:        "repository error is wrapped and propagated",
			id:          5,
			user:        &model.User{Name: "Fail"},
			repoErr:     errors.New("db write error"),
			wantErr:     true,
			wantErrWrap: "update user",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var savedUser *model.User
			repo := &fakeUserRepository{
				saveFunc: func(_ context.Context, u *model.User) (*model.User, error) {
					savedUser = u
					return u, tc.repoErr
				},
			}
			svc := newService(repo)

			err := svc.UpdateUser(ctx, tc.id, tc.user)

			if tc.wantErr {
				require.Error(t, err)
				if tc.wantErrWrap != "" {
					assert.Contains(t, err.Error(), tc.wantErrWrap)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, savedUser)
			assert.Equal(t, tc.wantSavedID, savedUser.ID,
				"id on saved user must equal the id parameter")
			// Verify the in-place mutation as well
			assert.Equal(