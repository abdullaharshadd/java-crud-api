```go
// Package service_test contains table-driven tests for the UserService interface
// contract defined in userservice.go.
//
// Since UserService is an interface, we validate the contract by implementing a
// mock that records calls and returns configured responses, then drive each
// spec through that mock.  This ensures that any real implementation that
// satisfies the interface will pass the same contract tests.
package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartContact/internal/smartcontact/model"
	"github.com/smartContact/internal/smartcontact/service"
)

// ---------------------------------------------------------------------------
// Sentinel errors used in tests
// ---------------------------------------------------------------------------

// ErrUserNotFound is the sentinel that implementations must return when a user
// cannot be located.  We define it here as a stand-in; real code would import
// smartcontacterror.ErrUserNotFound.
var ErrUserNotFound = errors.New("user not found")

// ErrPersistence simulates a generic storage failure.
var ErrPersistence = errors.New("persistence error")

// ---------------------------------------------------------------------------
// mockUserService – minimal, table-friendly mock of service.UserService
// ---------------------------------------------------------------------------

// mockUserService implements service.UserService.  Each field holds the value
// that the corresponding method should return, allowing per-test configuration.
type mockUserService struct {
	// SaveUser
	saveUserResp model.UserResponse
	saveUserErr  error

	// FetchUserList
	fetchListResp []model.UserResponse
	fetchListErr  error

	// FetchUserByID
	fetchByIDResp model.UserResponse
	fetchByIDErr  error

	// DeleteUser
	deleteUserErr error

	// UpdateUser
	updateUserErr error

	// GetUserByName
	getByNameResp model.UserResponse
	getByNameErr  error

	// call recording
	lastSavedUser   model.User
	lastDeletedID   int
	lastUpdatedID   int
	lastUpdatedUser model.User
	lastQueriedName string
	lastQueriedID   int
}

// Compile-time assertion that mockUserService satisfies the interface.
var _ service.UserService = (*mockUserService)(nil)

func (m *mockUserService) SaveUser(_ context.Context, user model.User) (model.UserResponse, error) {
	m.lastSavedUser = user
	return m.saveUserResp, m.saveUserErr
}

func (m *mockUserService) FetchUserList(_ context.Context) ([]model.UserResponse, error) {
	return m.fetchListResp, m.fetchListErr
}

func (m *mockUserService) FetchUserByID(_ context.Context, id int) (model.UserResponse, error) {
	m.lastQueriedID = id
	return m.fetchByIDResp, m.fetchByIDErr
}

func (m *mockUserService) DeleteUser(_ context.Context, id int) error {
	m.lastDeletedID = id
	return m.deleteUserErr
}

func (m *mockUserService) UpdateUser(_ context.Context, id int, user model.User) error {
	m.lastUpdatedID = id
	m.lastUpdatedUser = user
	return m.updateUserErr
}

func (m *mockUserService) GetUserByName(_ context.Context, name string) (model.UserResponse, error) {
	m.lastQueriedName = name
	return m.getByNameResp, m.getByNameErr
}

// ---------------------------------------------------------------------------
// Helper constructors
// ---------------------------------------------------------------------------

func newUser(id int, name, email string) model.User {
	return model.User{
		ID:    id,
		Name:  name,
		Email: email,
	}
}

func newUserResponse(id int, name, email string) model.UserResponse {
	return model.UserResponse{
		ID:    id,
		Name:  name,
		Email: email,
	}
}

// ---------------------------------------------------------------------------
// TestSaveUser
// ---------------------------------------------------------------------------

func TestSaveUser(t *testing.T) {
	t.Parallel()

	type input struct {
		user model.User
	}
	type want struct {
		resp    model.UserResponse
		wantErr bool
		errIs   error
	}
	type mockSetup struct {
		resp model.UserResponse
		err  error
	}

	tests := []struct {
		name  string
		mock  mockSetup
		input input
		want  want
	}{
		{
			name: "valid user is persisted and returned with generated id",
			mock: mockSetup{
				resp: newUserResponse(1, "Alice", "alice@example.com"),
				err:  nil,
			},
			input: input{user: newUser(0, "Alice", "alice@example.com")},
			want: want{
				resp:    newUserResponse(1, "Alice", "alice@example.com"),
				wantErr: false,
			},
		},
		{
			name: "persistence error is propagated to the caller",
			mock: mockSetup{
				resp: model.UserResponse{},
				err:  ErrPersistence,
			},
			input: input{user: newUser(0, "Bob", "bob@example.com")},
			want: want{
				resp:    model.UserResponse{},
				wantErr: true,
				errIs:   ErrPersistence,
			},
		},
		{
			name: "constraint violation returns error",
			mock: mockSetup{
				resp: model.UserResponse{},
				err:  errors.New("unique constraint violated"),
			},
			input: input{user: newUser(0, "Alice", "alice@example.com")},
			want: want{
				resp:    model.UserResponse{},
				wantErr: true,
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svc := &mockUserService{
				saveUserResp: tc.mock.resp,
				saveUserErr:  tc.mock.err,
			}

			got, err := svc.SaveUser(context.Background(), tc.input.user)

			if tc.want.wantErr {
				require.Error(t, err)
				if tc.want.errIs != nil {
					assert.True(t, errors.Is(err, tc.want.errIs),
						"expected errors.Is(err, %v) to be true, got: %v", tc.want.errIs, err)
				}
				// returned response should be zero value on error
				assert.Equal(t, tc.want.resp, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.want.resp, got)
			}

			// Invariant: the mock recorded the exact user that was passed in.
			if !tc.want.wantErr {
				assert.Equal(t, tc.input.user, svc.lastSavedUser,
					"service must forward the user to the repository unchanged")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestFetchUserList
// ---------------------------------------------------------------------------

func TestFetchUserList(t *testing.T) {
	t.Parallel()

	type mockSetup struct {
		resp []model.UserResponse
		err  error
	}
	type want struct {
		resp    []model.UserResponse
		wantErr bool
	}

	tests := []struct {
		name string
		mock mockSetup
		want want
	}{
		{
			name: "returns all users when users exist",
			mock: mockSetup{
				resp: []model.UserResponse{
					newUserResponse(1, "Alice", "alice@example.com"),
					newUserResponse(2, "Bob", "bob@example.com"),
				},
			},
			want: want{
				resp: []model.UserResponse{
					newUserResponse(1, "Alice", "alice@example.com"),
					newUserResponse(2, "Bob", "bob@example.com"),
				},
			},
		},
		{
			name: "returns empty slice when no users exist",
			mock: mockSetup{
				resp: []model.UserResponse{},
			},
			want: want{
				resp: []model.UserResponse{},
			},
		},
		{
			name: "nil slice is treated as empty list (never nil invariant)",
			mock: mockSetup{
				resp: nil,
			},
			want: want{
				// Implementation may return nil; callers should treat it as empty.
				resp: nil,
			},
		},
		{
			name: "storage error is propagated",
			mock: mockSetup{
				resp: nil,
				err:  ErrPersistence,
			},
			want: want{
				resp:    nil,
				wantErr: true,
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svc := &mockUserService{
				fetchListResp: tc.mock.resp,
				fetchListErr:  tc.mock.err,
			}

			got, err := svc.FetchUserList(context.Background())

			if tc.want.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.want.resp, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestFetchUserByID
// ---------------------------------------------------------------------------

func TestFetchUserByID(t *testing.T) {
	t.Parallel()

	type input struct {
		id int
	}
	type mockSetup struct {
		resp model.UserResponse
		err  error
	}
	type want struct {
		resp    model.UserResponse
		wantErr bool
		errIs   error
	}

	tests := []struct {
		name  string
		mock  mockSetup
		input input
		want  want
	}{
		{
			name: "existing id returns matching user",
			mock: mockSetup{
				resp: newUserResponse(42, "Carol", "carol@example.com"),
				err:  nil,
			},
			input: input{id: 42},
			want: want{
				resp: newUserResponse(42, "Carol", "carol@example.com"),
			},
		},
		{
			name: "non-existent id returns ErrUserNotFound",
			mock: mockSetup{
				resp: model.UserResponse{},
				err:  ErrUserNotFound,
			},
			input: input{id: 999},
			want: want{
				resp:    model.UserResponse{},
				wantErr: true,
				errIs:   ErrUserNotFound,
			},
		},
		{
			name: "zero id with no matching record returns ErrUserNotFound",
			mock: mockSetup{
				resp: model.UserResponse{},
				err:  ErrUserNotFound,
			},
			input: input{id: 0},
			want: want{
				resp:    model.UserResponse{},
				wantErr: true,
				errIs:   ErrUserNotFound,
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svc := &mockUserService{
				fetchByIDResp: tc.mock.resp,
				fetchByIDErr:  tc.mock.err,
			}

			got, err := svc.FetchUserByID(context.Background(), tc.input.id)

			// Invariant: the id was forwarded correctly.
			assert.Equal(t, tc.input.id, svc.lastQueriedID)

			if tc.want.wantErr {
				require.Error(t, err)
				if tc.want.errIs != nil {
					assert.True(t, errors.Is(err, tc.want.errIs),
						"expected errors.Is(err, %v), got: %v", tc.want.errIs, err)
				}
				assert.Equal(t, tc.want.resp, got, "zero response expected on error")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.want.resp, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestDeleteUser
// ---------------------------------------------------------------------------

func TestDeleteUser(t *testing.T) {
	t.Parallel()

	type input struct {
		id int
	}
	type mockSetup struct {
		err error
	}
	type want struct {
		wantErr bool
		errIs   error
	}

	tests := []struct {
		name  string
		mock  mockSetup
		input input
		want  want
	}{
		{
			name:  "existing user is deleted without error",
			mock:  mockSetup{err: nil},
			input: input{id: 1},
			want:  want{wantErr: false},
		},
		{
			name:  "deleting non-existent user may return an error",
			mock:  mockSetup{err: ErrUserNotFound},
			input: input{id: 999},
			want:  want{wantErr: true, errIs: ErrUserNotFound},
		},
		{
			name:  "storage error during deletion is propagated",
			mock:  mockSetup{err: ErrPersistence},
			input: input{id: 5},
			want:  want{wantErr: true, errIs: ErrPersistence},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svc := &mockUserService{
				deleteUserErr: tc.mock.err,
			}

			err := svc.DeleteUser(context.Background(), tc.input.id)

			// Invariant: the id was forwarded correctly.
			assert.Equal(t, tc.input.id, svc.lastDeletedID)

			if tc.want.wantErr {
				require.Error(t, err)
				if tc.want.errIs != nil {
					assert.True(t, errors.Is(err, tc.want.errIs),
						"expected errors.Is(err, %v), got: %v", tc.want.errIs, err)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestUpdateUser
// ---------------------------------------------------------------------------

func TestUpdateUser(t *testing.T) {
	t.Parallel()

	type input struct {
		id   int
		user model.User
	}
	type mockSetup struct {
		err error
	}
	type want struct {
		wantErr bool
		errIs   error
	}

	tests := []struct {
		name  string
		mock  mockSetup
		input input
		want  want
	}{
		{
			name: "valid id and user updates successfully",
			mock: mockSetup{err: nil},
			input: input{
				id:   1,
				user: newUser(1, "Alice Updated", "alice-new@example.com"),
			},
			want: want{wantErr: false},
		},
		{
			name: "non-existent id returns ErrUserNotFound",
			mock: mockSetup{err: ErrUserNotFound},
			input: input{
				id:   999,
				user: newUser(999, "Ghost", "ghost@example.com"),
			},
			want: want{wantErr: true, errIs: ErrUserNotFound},
		},
		{
			name: "storage error during update is propagated",
			mock: mockSetup{err: ErrPersistence},
			input: input{
				id:   2,
				user: newUser(2, "Bob", "bob@example.com"),
			},
			want: want{wantErr: true, errIs: E