```go
package service_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	smarterr "github.com/smartContact/internal/smartcontact/error"
	"github.com/smartContact/internal/smartcontact/model"
	"github.com/smartContact/internal/smartcontact/service"
)

// ---------------------------------------------------------------------------
// Mock UserDao
// ---------------------------------------------------------------------------

type mockUserDao struct {
	mock.Mock
}

func (m *mockUserDao) Save(ctx context.Context, user model.User) (model.User, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(model.User), args.Error(1)
}

func (m *mockUserDao) FindAll(ctx context.Context) ([]model.User, error) {
	args := m.Called(ctx)
	if v := args.Get(0); v != nil {
		return v.([]model.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockUserDao) FindByID(ctx context.Context, id int) (model.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(model.User), args.Error(1)
}

func (m *mockUserDao) DeleteByID(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockUserDao) FindByName(ctx context.Context, name string) (model.User, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(model.User), args.Error(1)
}

// ---------------------------------------------------------------------------
// Helper
// ---------------------------------------------------------------------------

func newUserNotFoundError(id int) error {
	return &smarterr.UserNotFoundError{ID: id}
}

// ---------------------------------------------------------------------------
// SaveUser tests
// ---------------------------------------------------------------------------

func TestUserService_SaveUser(t *testing.T) {
	type input struct {
		user model.User
	}
	type daoReturn struct {
		saved model.User
		err   error
	}
	tests := []struct {
		name        string
		input       input
		daoReturn   daoReturn
		wantUser    model.User
		wantErr     bool
		wantErrWrap error
	}{
		{
			name:  "valid user is saved and returned with generated id",
			input: input{user: model.User{Name: "Alice", Email: "alice@example.com"}},
			daoReturn: daoReturn{
				saved: model.User{ID: 1, Name: "Alice", Email: "alice@example.com"},
				err:   nil,
			},
			wantUser: model.User{ID: 1, Name: "Alice", Email: "alice@example.com"},
			wantErr:  false,
		},
		{
			name:  "dao returns persistence error",
			input: input{user: model.User{Name: ""}},
			daoReturn: daoReturn{
				saved: model.User{},
				err:   errors.New("constraint violation"),
			},
			wantUser:    model.User{},
			wantErr:     true,
			wantErrWrap: errors.New("constraint violation"),
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			dao := new(mockUserDao)
			dao.On("Save", mock.Anything, tc.input.user).
				Return(tc.daoReturn.saved, tc.daoReturn.err)

			svc := service.NewUserService(dao)
			got, err := svc.SaveUser(context.Background(), tc.input.user)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Equal(t, model.User{}, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantUser, got)
				assert.NotEqual(t, model.User{}, got, "successful save must not return zero-value user")
			}
			dao.AssertExpectations(t)
		})
	}
}

// ---------------------------------------------------------------------------
// FetchUserList tests
// ---------------------------------------------------------------------------

func TestUserService_FetchUserList(t *testing.T) {
	tests := []struct {
		name      string
		daoUsers  []model.User
		daoErr    error
		wantUsers []model.User
		wantErr   bool
	}{
		{
			name: "users exist – returns full list",
			daoUsers: []model.User{
				{ID: 1, Name: "Alice"},
				{ID: 2, Name: "Bob"},
			},
			daoErr: nil,
			wantUsers: []model.User{
				{ID: 1, Name: "Alice"},
				{ID: 2, Name: "Bob"},
			},
			wantErr: false,
		},
		{
			name:      "no users exist – returns empty list",
			daoUsers:  []model.User{},
			daoErr:    nil,
			wantUsers: []model.User{},
			wantErr:   false,
		},
		{
			name:      "dao returns error",
			daoUsers:  nil,
			daoErr:    errors.New("db unavailable"),
			wantUsers: nil,
			wantErr:   true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			dao := new(mockUserDao)
			dao.On("FindAll", mock.Anything).Return(tc.daoUsers, tc.daoErr)

			svc := service.NewUserService(dao)
			got, err := svc.FetchUserList(context.Background())

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got, "list must never be nil on success")
				assert.Equal(t, tc.wantUsers, got)
			}
			dao.AssertExpectations(t)
		})
	}
}

// ---------------------------------------------------------------------------
// FetchUserByID tests
// ---------------------------------------------------------------------------

func TestUserService_FetchUserByID(t *testing.T) {
	tests := []struct {
		name     string
		id       int
		daoUser  model.User
		daoErr   error
		wantUser model.User
		wantErr  bool
		notFound bool
	}{
		{
			name:     "existing id returns user",
			id:       1,
			daoUser:  model.User{ID: 1, Name: "Alice"},
			daoErr:   nil,
			wantUser: model.User{ID: 1, Name: "Alice"},
			wantErr:  false,
		},
		{
			name:     "non-existing id returns UserNotFoundError",
			id:       99,
			daoUser:  model.User{},
			daoErr:   &smarterr.UserNotFoundError{ID: 99},
			wantUser: model.User{},
			wantErr:  true,
			notFound: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			dao := new(mockUserDao)
			dao.On("FindByID", mock.Anything, tc.id).Return(tc.daoUser, tc.daoErr)

			svc := service.NewUserService(dao)
			got, err := svc.FetchUserByID(context.Background(), tc.id)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Equal(t, model.User{}, got)
				if tc.notFound {
					var notFoundErr *smarterr.UserNotFoundError
					assert.True(t, errors.As(err, &notFoundErr),
						"error must unwrap to *smarterr.UserNotFoundError")
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantUser, got)
			}
			dao.AssertExpectations(t)
		})
	}
}

// FetchUserByID read-only invariant: dao.Save is never called.
func TestUserService_FetchUserByID_ReadOnly(t *testing.T) {
	dao := new(mockUserDao)
	dao.On("FindByID", mock.Anything, 1).Return(model.User{ID: 1, Name: "Alice"}, nil)
	// No expectation for Save/Delete/etc.

	svc := service.NewUserService(dao)
	_, _ = svc.FetchUserByID(context.Background(), 1)

	dao.AssertNotCalled(t, "Save")
	dao.AssertNotCalled(t, "DeleteByID")
	dao.AssertNotCalled(t, "FindAll")
}

// ---------------------------------------------------------------------------
// DeleteUser tests
// ---------------------------------------------------------------------------

func TestUserService_DeleteUser(t *testing.T) {
	tests := []struct {
		name    string
		id      int
		daoErr  error
		wantErr bool
	}{
		{
			name:    "existing user is deleted successfully",
			id:      1,
			daoErr:  nil,
			wantErr: false,
		},
		{
			name:    "dao propagates error when user not found",
			id:      99,
			daoErr:  &smarterr.UserNotFoundError{ID: 99},
			wantErr: true,
		},
		{
			name:    "dao propagates generic persistence error",
			id:      2,
			daoErr:  errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			dao := new(mockUserDao)
			dao.On("DeleteByID", mock.Anything, tc.id).Return(tc.daoErr)

			svc := service.NewUserService(dao)
			err := svc.DeleteUser(context.Background(), tc.id)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			dao.AssertExpectations(t)
		})
	}
}

// DeleteUser invariant: after successful delete the record is gone (dao was called once).
func TestUserService_DeleteUser_CalledOnce(t *testing.T) {
	dao := new(mockUserDao)
	dao.On("DeleteByID", mock.Anything, 5).Return(nil).Once()

	svc := service.NewUserService(dao)
	err := svc.DeleteUser(context.Background(), 5)

	assert.NoError(t, err)
	dao.AssertNumberOfCalls(t, "DeleteByID", 1)
}

// ---------------------------------------------------------------------------
// UpdateUser tests
// ---------------------------------------------------------------------------

func TestUserService_UpdateUser(t *testing.T) {
	existingUser := model.User{ID: 3, Name: "Charlie", Email: "charlie@example.com"}
	updatedValues := model.User{Name: "Charlie Updated", Email: "charlie2@example.com"}
	// Service stamps the id before saving.
	savedPayload := model.User{ID: 3, Name: "Charlie Updated", Email: "charlie2@example.com"}

	tests := []struct {
		name         string
		id           int
		user         model.User
		findReturn   model.User
		findErr      error
		savePayload  model.User
		saveReturn   model.User
		saveErr      error
		expectSave   bool
		wantErr      bool
		notFound     bool
	}{
		{
			name:        "existing user is updated successfully",
			id:          3,
			user:        updatedValues,
			findReturn:  existingUser,
			findErr:     nil,
			savePayload: savedPayload,
			saveReturn:  savedPayload,
			saveErr:     nil,
			expectSave:  true,
			wantErr:     false,
		},
		{
			name:       "non-existing id returns UserNotFoundError",
			id:         99,
			user:       updatedValues,
			findReturn: model.User{},
			findErr:    &smarterr.UserNotFoundError{ID: 99},
			expectSave: false,
			wantErr:    true,
			notFound:   true,
		},
		{
			name:        "save fails after find – propagates error",
			id:          3,
			user:        updatedValues,
			findReturn:  existingUser,
			findErr:     nil,
			savePayload: savedPayload,
			saveReturn:  model.User{},
			saveErr:     errors.New("save failed"),
			expectSave:  true,
			wantErr:     true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			dao := new(mockUserDao)
			dao.On("FindByID", mock.Anything, tc.id).Return(tc.findReturn, tc.findErr)

			if tc.expectSave {
				dao.On("Save", mock.Anything, tc.savePayload).Return(tc.saveReturn, tc.saveErr)
			}

			svc := service.NewUserService(dao)
			err := svc.UpdateUser(context.Background(), tc.id, tc.user)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.notFound {
					var notFoundErr *smarterr.UserNotFoundError
					assert.True(t, errors.As(err, &notFoundErr),
						"error must unwrap to *smarterr.UserNotFoundError")
				}
			} else {
				assert.NoError(t, err)
			}
			dao.AssertExpectations(t)
		})
	}
}

// UpdateUser invariant: id is always preserved in the persisted payload.
func TestUserService_UpdateUser_IDPreserved(t *testing.T) {
	const targetID = 7
	incoming := model.User{Name: "Dave", Email: "dave@example.com"}
	expected := model.User{ID: targetID, Name: "Dave", Email: "dave@example.com"}

	dao := new(mockUserDao)
	dao.On("FindByID", mock.Anything, targetID).Return(model.User{ID: targetID, Name: "Old"}, nil)
	// Assert that save receives payload with correct id.
	dao.On("Save", mock.Anything, expected).Return(expected, nil)

	svc := service.NewUserService(dao)
	err := svc.UpdateUser(context.Background(), targetID, incoming)

	assert.NoError(t, err)
	dao.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// GetUserByName tests
// ---------------------------------------------------------------------------

func TestUserService_GetUserByName(t *testing.T) {
	tests := []struct {
		name     string
		nameArg  string
		daoUser  model.User
		daoErr   error
		wantUser model.User
		wantErr  bool
		notFound bool
	}{
		{
			name:     "existing name returns user",
			nameArg:  "Alice",
			daoUser:  model.User{ID: 1, Name: "Alice"},
			daoErr:   nil,
			wantUser: model.User{ID: 1, Name: "Alice"},
			wantErr:  false,
		},
		{
			name:     "name matches no user returns error",
			nameArg:  "Unknown",
			daoUser:  model.User{},
			daoErr:   