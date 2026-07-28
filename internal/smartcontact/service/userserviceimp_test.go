```go
package service

import (
	"context"
	"errors"
	"fmt"
	"testing"

	apperror "github.com/smartcontact/internal/smartcontact/error"
	"github.com/smartcontact/internal/smartcontact/model"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// Mock repository
// ---------------------------------------------------------------------------

// mockUserRepo is a test double that implements whatever repository interface
// *userService depends on.  We need to support FindByName (from this file) as
// well as the other CRUD methods expected by the rest of the service.
type mockUserRepo struct {
	// FindByName
	findByNameFn func(ctx context.Context, name string) (*model.User, error)
	// FindByID
	findByIDFn func(ctx context.Context, id int) (*model.User, error)
	// Save
	saveFn func(ctx context.Context, user *model.User) (*model.User, error)
	// FindAll
	findAllFn func(ctx context.Context) ([]*model.User, error)
	// Delete
	deleteFn func(ctx context.Context, id int) error
}

func (m *mockUserRepo) FindByName(ctx context.Context, name string) (*model.User, error) {
	if m.findByNameFn != nil {
		return m.findByNameFn(ctx, name)
	}
	return nil, fmt.Errorf("FindByName not configured in mock")
}

func (m *mockUserRepo) FindByID(ctx context.Context, id int) (*model.User, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, fmt.Errorf("FindByID not configured in mock")
}

func (m *mockUserRepo) Save(ctx context.Context, user *model.User) (*model.User, error) {
	if m.saveFn != nil {
		return m.saveFn(ctx, user)
	}
	return nil, fmt.Errorf("Save not configured in mock")
}

func (m *mockUserRepo) FindAll(ctx context.Context) ([]*model.User, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx)
	}
	return nil, fmt.Errorf("FindAll not configured in mock")
}

func (m *mockUserRepo) Delete(ctx context.Context, id int) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return fmt.Errorf("Delete not configured in mock")
}

// ---------------------------------------------------------------------------
// Helper – build a *userService backed by the mock repo.
// We call NewUserService (declared in userservice.go) so we always go through
// the real constructor and we never duplicate the struct definition.
// ---------------------------------------------------------------------------
func newServiceWithMock(repo *mockUserRepo) UserService {
	return NewUserService(repo)
}

// ---------------------------------------------------------------------------
// Tests for GetUserByName  (the method added in this file)
// ---------------------------------------------------------------------------

func TestGetUserByName(t *testing.T) {
	ctx := context.Background()

	existingUser := &model.User{ID: 1, Name: "alice"}

	tests := []struct {
		name        string
		inputName   string
		repoUser    *model.User
		repoErr     error
		wantUser    *model.User
		wantErr     bool
		wantNotFound bool
	}{
		{
			name:      "found – repo returns a user",
			inputName: "alice",
			repoUser:  existingUser,
			repoErr:   nil,
			wantUser:  existingUser,
			wantErr:   false,
		},
		{
			name:         "not found – repo returns nil user and no error",
			inputName:    "ghost",
			repoUser:     nil,
			repoErr:      nil,
			wantUser:     nil,
			wantErr:      true,
			wantNotFound: true,
		},
		{
			name:         "not found – repo returns UserNotFound sentinel",
			inputName:    "ghost",
			repoUser:     nil,
			repoErr:      apperror.NewUserNotFoundErrorf("user with name %q not found", "ghost"),
			wantUser:     nil,
			wantErr:      true,
			wantNotFound: true,
		},
		{
			name:      "generic repo error is wrapped and propagated",
			inputName: "bob",
			repoUser:  nil,
			repoErr:   errors.New("connection refused"),
			wantUser:  nil,
			wantErr:   true,
			wantNotFound: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockUserRepo{
				findByNameFn: func(_ context.Context, _ string) (*model.User, error) {
					return tc.repoUser, tc.repoErr
				},
			}

			svc := newServiceWithMock(repo)

			got, err := svc.GetUserByName(ctx, tc.inputName)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				if tc.wantNotFound {
					assert.True(t, apperror.IsUserNotFound(err),
						"expected a UserNotFound error, got: %v", err)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantUser, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Tests for SaveUser
// ---------------------------------------------------------------------------

func TestSaveUser(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		inputUser   *model.User
		repoReturns *model.User
		repoErr     error
		wantUser    *model.User
		wantErr     bool
	}{
		{
			name:        "valid user – persisted and returned",
			inputUser:   &model.User{Name: "carol"},
			repoReturns: &model.User{ID: 42, Name: "carol"},
			wantUser:    &model.User{ID: 42, Name: "carol"},
		},
		{
			name:        "upsert – existing id updated",
			inputUser:   &model.User{ID: 7, Name: "dave"},
			repoReturns: &model.User{ID: 7, Name: "dave"},
			wantUser:    &model.User{ID: 7, Name: "dave"},
		},
		{
			name:      "repo error propagated",
			inputUser: &model.User{Name: "eve"},
			repoErr:   errors.New("db error"),
			wantErr:   true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var captured *model.User
			repo := &mockUserRepo{
				saveFn: func(_ context.Context, u *model.User) (*model.User, error) {
					captured = u
					return tc.repoReturns, tc.repoErr
				},
			}

			svc := newServiceWithMock(repo)
			got, err := svc.SaveUser(ctx, tc.inputUser)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantUser, got)
				// No transformation before persisting
				assert.Equal(t, tc.inputUser, captured)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Tests for FetchUserList
// ---------------------------------------------------------------------------

func TestFetchUserList(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		repoReturns []*model.User
		repoErr     error
		wantLen     int
		wantErr     bool
	}{
		{
			name: "users exist – all returned",
			repoReturns: []*model.User{
				{ID: 1, Name: "alice"},
				{ID: 2, Name: "bob"},
			},
			wantLen: 2,
		},
		{
			name:        "no users – empty slice returned",
			repoReturns: []*model.User{},
			wantLen:     0,
		},
		{
			name:    "repo error propagated",
			repoErr: errors.New("db failure"),
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockUserRepo{
				findAllFn: func(_ context.Context) ([]*model.User, error) {
					return tc.repoReturns, tc.repoErr
				},
			}

			svc := newServiceWithMock(repo)
			got, err := svc.FetchUserList(ctx)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Len(t, got, tc.wantLen)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Tests for FetchUserById / GetUserByID
// ---------------------------------------------------------------------------

func TestFetchUserById(t *testing.T) {
	ctx := context.Background()

	existingUser := &model.User{ID: 5, Name: "frank"}

	tests := []struct {
		name         string
		inputID      int
		repoUser     *model.User
		repoErr      error
		wantUser     *model.User
		wantErr      bool
		wantNotFound bool
	}{
		{
			name:     "id matches existing user – returned",
			inputID:  5,
			repoUser: existingUser,
			wantUser: existingUser,
		},
		{
			name:         "id not found – UserNotFound error raised",
			inputID:      999,
			repoErr:      apperror.NewUserNotFoundErrorf("User are not available"),
			wantErr:      true,
			wantNotFound: true,
		},
		{
			name:         "repo returns nil user without error – treated as not found",
			inputID:      123,
			repoUser:     nil,
			wantErr:      true,
			wantNotFound: true,
		},
		{
			name:    "generic repo error propagated",
			inputID: 1,
			repoErr: errors.New("network timeout"),
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockUserRepo{
				findByIDFn: func(_ context.Context, id int) (*model.User, error) {
					return tc.repoUser, tc.repoErr
				},
			}

			svc := newServiceWithMock(repo)
			got, err := svc.GetUserByID(ctx, tc.inputID)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				if tc.wantNotFound {
					assert.True(t, apperror.IsUserNotFound(err),
						"expected UserNotFound error, got: %v", err)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantUser, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Tests for DeleteUser
// ---------------------------------------------------------------------------

func TestDeleteUser(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		inputID int
		repoErr error
		wantErr bool
	}{
		{
			name:    "existing user deleted successfully",
			inputID: 3,
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "non-existing user – repo raises error",
			inputID: 999,
			repoErr: errors.New("no entity with id 999"),
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var deletedID int
			repo := &mockUserRepo{
				deleteFn: func(_ context.Context, id int) error {
					deletedID = id
					return tc.repoErr
				},
			}

			svc := newServiceWithMock(repo)
			err := svc.DeleteUser(ctx, tc.inputID)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.inputID, deletedID)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Tests for UpdateUser
// ---------------------------------------------------------------------------

func TestUpdateUser(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		inputID   int
		inputUser *model.User
		repoErr   error
		wantErr   bool
	}{
		{
			name:      "id is set on user before persisting",
			inputID:   10,
			inputUser: &model.User{ID: 0, Name: "grace"},
			wantErr:   false,
		},
		{
			name:      "existing id overwritten",
			inputID:   20,
			inputUser: &model.User{ID: 99, Name: "heidi"},
			wantErr:   false,
		},
		{
			name:      "repo error propagated",
			inputID:   5,
			inputUser: &model.User{Name: "ivan"},
			repoErr:   errors.New("write failure"),
			wantErr:   true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var saved *model.User
			repo := &mockUserRepo{
				saveFn: func(_ context.Context, u *model.User) (*model.User, error) {
					saved = u
					return u, tc.repoErr
				},
			}

			svc := newServiceWithMock(repo)
			err := svc.UpdateUser(ctx, tc.inputID, tc.inputUser)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// The persisted user must carry the supplied id.
				assert.Equal(t, tc.inputID, saved.ID,
					"expected saved user id to equal inputID")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Invariant: GetUserByName never returns both a non-nil user AND an error.
// ---------------------------------------------------------------------------

func TestGetUserByName_NeverReturnsUserAndError(t *testing.T) {
	ctx := context.Background()

	scenarios := []struct {
		label   string
		user    *model.User
		repoErr error
	}{
		{"user found", &model.User{ID: 1, Name: "judy"}, nil},
		{"not found sentinel", nil, apperror.NewUserNotFoundErrorf("not found")},
		{"nil user no error", nil, nil},
		{"generic error", nil, errors.New("oops")},
	}

	for _, sc := range scenarios {
		sc := sc
		t.Run(sc.label, func(t *testing.T) {
			repo := &mockUserRepo{
				findByNameFn: func(_ context.Context, _ string) (*model.User,