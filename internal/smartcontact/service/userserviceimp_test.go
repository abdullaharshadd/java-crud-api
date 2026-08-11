```go
package service_test

import (
	"context"
	"errors"
	"testing"

	smarterror "internal/smartcontact/error"
	"internal/smartcontact/model"
	"internal/smartcontact/service"

	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// Mock repository
// ---------------------------------------------------------------------------

type mockUserRepository struct {
	// FindByID
	findByIDUser *model.User
	findByIDErr  error

	// FindAll
	findAllUsers []*model.User
	findAllErr   error

	// Save
	saveUser *model.User
	saveErr  error

	// DeleteByID
	deleteByIDErr error

	// FindByName
	findByNameUser *model.User
	findByNameErr  error

	// Capture calls for assertion
	savedUser     *model.User
	deletedID     int
	findByIDCalls int
}

func (m *mockUserRepository) FindByID(ctx context.Context, id int) (*model.User, error) {
	m.findByIDCalls++
	return m.findByIDUser, m.findByIDErr
}

func (m *mockUserRepository) FindAll(ctx context.Context) ([]*model.User, error) {
	return m.findAllUsers, m.findAllErr
}

func (m *mockUserRepository) Save(ctx context.Context, user *model.User) (*model.User, error) {
	m.savedUser = user
	if m.saveErr != nil {
		return nil, m.saveErr
	}
	if m.saveUser != nil {
		return m.saveUser, nil
	}
	return user, nil
}

func (m *mockUserRepository) DeleteByID(ctx context.Context, id int) error {
	m.deletedID = id
	return m.deleteByIDErr
}

func (m *mockUserRepository) FindByName(ctx context.Context, name string) (*model.User, error) {
	return m.findByNameUser, m.findByNameErr
}

// ---------------------------------------------------------------------------
// Helper – build a concrete *service.UserService (whichever constructor exists)
// ---------------------------------------------------------------------------

// newSvc creates the concrete service implementation from the migrated package.
// Adjust the constructor name if the implementation uses a different one.
func newSvc(repo *mockUserRepository) *service.UserServiceImpl {
	return service.NewUserServiceImpl(repo)
}

// ---------------------------------------------------------------------------
// FetchUserByIDStrict (standalone helper)
// ---------------------------------------------------------------------------

func TestFetchUserByIDStrict(t *testing.T) {
	tests := []struct {
		name        string
		id          int
		repoUser    *model.User
		repoErr     error
		wantUser    *model.User
		wantErrType error
		wantErrMsg  string
	}{
		{
			name:     "existing user is returned",
			id:       1,
			repoUser: &model.User{ID: 1, Name: "Alice"},
			wantUser: &model.User{ID: 1, Name: "Alice"},
		},
		{
			name:        "repo signals ErrUserNotFound – UserNotFoundError returned",
			id:          99,
			repoErr:     smarterror.ErrUserNotFound,
			wantErrType: &smarterror.UserNotFoundError{},
			wantErrMsg:  "User are not available",
		},
		{
			name:        "repo returns nil user without error – UserNotFoundError returned",
			id:          42,
			repoUser:    nil,
			repoErr:     nil,
			wantErrType: &smarterror.UserNotFoundError{},
			wantErrMsg:  "User are not available",
		},
		{
			name:    "repo returns generic error – propagated as-is",
			id:      7,
			repoErr: errors.New("db connection refused"),
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockUserRepository{
				findByIDUser: tc.repoUser,
				findByIDErr:  tc.repoErr,
			}

			got, err := service.FetchUserByIDStrict(context.Background(), repo, tc.id)

			if tc.wantUser != nil {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantUser, got)
				return
			}

			assert.Nil(t, got)
			assert.Error(t, err)

			if tc.wantErrType != nil {
				assert.ErrorAs(t, err, &tc.wantErrType)
				assert.Contains(t, err.Error(), tc.wantErrMsg)
			} else if tc.repoErr != nil && !errors.Is(tc.repoErr, smarterror.ErrUserNotFound) {
				// generic error must be propagated unchanged
				assert.Equal(t, tc.repoErr, err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// UpdateUserByID (standalone helper)
// ---------------------------------------------------------------------------

func TestUpdateUserByID(t *testing.T) {
	tests := []struct {
		name       string
		id         int
		inputUser  *model.User
		saveErr    error
		wantErr    bool
		wantUserID int
	}{
		{
			name:       "valid user – id is set and persisted",
			id:         5,
			inputUser:  &model.User{ID: 0, Name: "Bob"},
			wantErr:    false,
			wantUserID: 5,
		},
		{
			name:      "nil user – returns validation error",
			id:        5,
			inputUser: nil,
			wantErr:   true,
		},
		{
			name:      "repo.Save fails – error propagated",
			id:        3,
			inputUser: &model.User{Name: "Carol"},
			saveErr:   errors.New("constraint violation"),
			wantErr:   true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockUserRepository{saveErr: tc.saveErr}

			err := service.UpdateUserByID(context.Background(), repo, tc.id, tc.inputUser)

			if tc.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			// The helper must set user.ID before saving.
			assert.Equal(t, tc.wantUserID, repo.savedUser.ID)
		})
	}
}

// ---------------------------------------------------------------------------
// SaveUser (via concrete service)
// ---------------------------------------------------------------------------

func TestSaveUser(t *testing.T) {
	tests := []struct {
		name       string
		inputUser  *model.User
		savedUser  *model.User
		saveErr    error
		wantErr    bool
		wantResult *model.User
	}{
		{
			name:       "valid user is persisted and returned",
			inputUser:  &model.User{Name: "Dave"},
			savedUser:  &model.User{ID: 10, Name: "Dave"},
			wantErr:    false,
			wantResult: &model.User{ID: 10, Name: "Dave"},
		},
		{
			name:      "constraint violation – repo error propagated",
			inputUser: &model.User{Name: "Duplicate"},
			saveErr:   errors.New("unique constraint violated"),
			wantErr:   true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockUserRepository{saveUser: tc.savedUser, saveErr: tc.saveErr}
			svc := newSvc(repo)

			got, err := svc.SaveUser(context.Background(), tc.inputUser)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.wantResult, got)
			// Verify delegation – the repository must have been called with the input
			assert.Equal(t, tc.inputUser, repo.savedUser)
		})
	}
}

// ---------------------------------------------------------------------------
// FetchUserList (via concrete service)
// ---------------------------------------------------------------------------

func TestFetchUserList(t *testing.T) {
	tests := []struct {
		name       string
		repoUsers  []*model.User
		repoErr    error
		wantErr    bool
		wantResult []*model.User
	}{
		{
			name: "users exist – all returned",
			repoUsers: []*model.User{
				{ID: 1, Name: "Eve"},
				{ID: 2, Name: "Frank"},
			},
			wantResult: []*model.User{
				{ID: 1, Name: "Eve"},
				{ID: 2, Name: "Frank"},
			},
		},
		{
			name:       "no users exist – empty slice returned",
			repoUsers:  []*model.User{},
			wantResult: []*model.User{},
		},
		{
			name:      "repo error – propagated",
			repoErr:   errors.New("db timeout"),
			wantErr:   true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockUserRepository{findAllUsers: tc.repoUsers, findAllErr: tc.repoErr}
			svc := newSvc(repo)

			got, err := svc.FetchUserList(context.Background())

			if tc.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, got)
			assert.Equal(t, tc.wantResult, got)
		})
	}
}

// ---------------------------------------------------------------------------
// FetchUserByID (via concrete service)
// ---------------------------------------------------------------------------

func TestFetchUserByID(t *testing.T) {
	tests := []struct {
		name        string
		id          int
		repoUser    *model.User
		repoErr     error
		wantUser    *model.User
		wantErrType interface{}
		wantErrMsg  string
	}{
		{
			name:     "existing user is returned",
			id:       1,
			repoUser: &model.User{ID: 1, Name: "Grace"},
			wantUser: &model.User{ID: 1, Name: "Grace"},
		},
		{
			name:        "user not found – UserNotFoundError with correct message",
			id:          404,
			repoErr:     smarterror.ErrUserNotFound,
			wantErrType: &smarterror.UserNotFoundError{},
			wantErrMsg:  "User are not available",
		},
		{
			name:        "repo returns nil user – UserNotFoundError",
			id:          55,
			repoUser:    nil,
			repoErr:     nil,
			wantErrType: &smarterror.UserNotFoundError{},
			wantErrMsg:  "User are not available",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockUserRepository{findByIDUser: tc.repoUser, findByIDErr: tc.repoErr}
			svc := newSvc(repo)

			got, err := svc.FetchUserByID(context.Background(), tc.id)

			if tc.wantUser != nil {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantUser, got)
				return
			}

			assert.Nil(t, got)
			assert.Error(t, err)

			if tc.wantErrType != nil {
				assert.ErrorAs(t, err, tc.wantErrType)
				assert.Contains(t, err.Error(), tc.wantErrMsg)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// DeleteUser (via concrete service)
// ---------------------------------------------------------------------------

func TestDeleteUser(t *testing.T) {
	tests := []struct {
		name          string
		id            int
		deleteByIDErr error
		wantErr       bool
	}{
		{
			name:    "existing user deleted successfully",
			id:      1,
			wantErr: false,
		},
		{
			name:          "non-existing id – repo error propagated",
			id:            999,
			deleteByIDErr: errors.New("record not found"),
			wantErr:       true,
		},
		{
			name:    "repository is no-op for missing id – no error",
			id:      888,
			wantErr: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockUserRepository{deleteByIDErr: tc.deleteByIDErr}
			svc := newSvc(repo)

			err := svc.DeleteUser(context.Background(), tc.id)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Verify delegation
				assert.Equal(t, tc.id, repo.deletedID)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// UpdateUser (via concrete service)
// ---------------------------------------------------------------------------

func TestUpdateUser(t *testing.T) {
	tests := []struct {
		name       string
		id         int
		inputUser  *model.User
		saveErr    error
		wantErr    bool
		wantUserID int
	}{
		{
			name:       "valid update – id is forced onto user and persisted",
			id:         7,
			inputUser:  &model.User{ID: 0, Name: "Heidi"},
			wantErr:    false,
			wantUserID: 7,
		},
		{
			name:       "different original id – overwritten with argument id",
			id:         20,
			inputUser:  &model.User{ID: 99, Name: "Ivan"},
			wantErr:    false,
			wantUserID: 20,
		},
		{
			name:      "repo.Save fails – error propagated",
			id:        3,
			inputUser: &model.User{Name: "Judy"},
			saveErr:   errors.New("unique constraint violated"),
			wantErr:   true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockUserRepository{saveErr: tc.saveErr}
			svc := newSvc(repo)

			err := svc.UpdateUser(context.Background(), tc.id, tc.inputUser)

			if tc.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			// The service must set user.ID before delegating to repo.Save.
			assert.Equal(t, tc.wantUserID, repo.savedUser.ID)
		})
	}
}

// ---------------------------------------------------------------------------
// GetUserByName (via concrete service)
// ---------------------------------------------------------------------------

func TestGetUserByName(t *testing.T) {
	tests := []struct {
		name           string
		lookupName     string
		repoUser       *model.User
		repoErr        error
		wantUser       *model.User
		wantErr        bool
	}{
		{
			name:       "matching user returned",
			lookupName: "Alice",
			repoUser:   &model.User{ID: 1, Name: "Alice"},
			wantUser:   &model.User{ID: 1, Name: "Alice"},
		},
		{
			name:       "no match – nil returned without error",