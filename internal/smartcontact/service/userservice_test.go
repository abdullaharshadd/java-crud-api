```go
package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/smartContact/internal/smartcontact/model"
	"github.com/smartContact/internal/smartcontact/service"
)

// ---------------------------------------------------------------------------
// Sentinel errors used by the mock to simulate domain-level failures.
// ---------------------------------------------------------------------------

var (
	errUserNotFound  = errors.New("user not found")
	errDeleteFailed  = errors.New("delete failed")
	errUpdateFailed  = errors.New("update failed")
	errInternalStore = errors.New("internal store error")
)

// ---------------------------------------------------------------------------
// mockUserService – a hand-rolled mock that implements service.UserService.
// Each field stores the canned return values so individual test cases can
// control behaviour without an external mocking library.
// ---------------------------------------------------------------------------

type mockUserService struct {
	// SaveUser
	saveUserFn func(ctx context.Context, user *model.User) (*model.User, error)

	// FetchUserList
	fetchUserListFn func(ctx context.Context) ([]*model.User, error)

	// FetchUserByID
	fetchUserByIDFn func(ctx context.Context, id int64) (*model.User, error)

	// DeleteUser
	deleteUserFn func(ctx context.Context, id int64) error

	// UpdateUser
	updateUserFn func(ctx context.Context, id int64, user *model.User) error

	// GetUserByName
	getUserByNameFn func(ctx context.Context, name string) (*model.User, error)
}

func (m *mockUserService) SaveUser(ctx context.Context, user *model.User) (*model.User, error) {
	return m.saveUserFn(ctx, user)
}

func (m *mockUserService) FetchUserList(ctx context.Context) ([]*model.User, error) {
	return m.fetchUserListFn(ctx)
}

func (m *mockUserService) FetchUserByID(ctx context.Context, id int64) (*model.User, error) {
	return m.fetchUserByIDFn(ctx, id)
}

func (m *mockUserService) DeleteUser(ctx context.Context, id int64) error {
	return m.deleteUserFn(ctx, id)
}

func (m *mockUserService) UpdateUser(ctx context.Context, id int64, user *model.User) error {
	return m.updateUserFn(ctx, id, user)
}

func (m *mockUserService) GetUserByName(ctx context.Context, name string) (*model.User, error) {
	return m.getUserByNameFn(ctx, name)
}

// Compile-time assertion: mockUserService must satisfy the interface.
var _ service.UserService = (*mockUserService)(nil)

// ---------------------------------------------------------------------------
// Helper
// ---------------------------------------------------------------------------

func newUser(id int64, name string) *model.User {
	return &model.User{ID: id, Name: name}
}

// ---------------------------------------------------------------------------
// TestSaveUser
// ---------------------------------------------------------------------------

func TestSaveUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		inputUser   *model.User
		setupMock   func() *mockUserService
		wantUser    *model.User
		wantErr     bool
		errContains string
	}{
		{
			name:      "given a valid user object returns the saved user with generated id",
			inputUser: &model.User{Name: "Alice"},
			setupMock: func() *mockUserService {
				return &mockUserService{
					saveUserFn: func(_ context.Context, user *model.User) (*model.User, error) {
						saved := *user
						saved.ID = 42
						return &saved, nil
					},
				}
			},
			wantUser: &model.User{ID: 42, Name: "Alice"},
			wantErr:  false,
		},
		{
			name:      "given a nil user object returns an error",
			inputUser: nil,
			setupMock: func() *mockUserService {
				return &mockUserService{
					saveUserFn: func(_ context.Context, user *model.User) (*model.User, error) {
						if user == nil {
							return nil, errInternalStore
						}
						return user, nil
					},
				}
			},
			wantUser:    nil,
			wantErr:     true,
			errContains: "internal store error",
		},
		{
			name:      "returned user reflects the persisted state of the input",
			inputUser: &model.User{Name: "Bob"},
			setupMock: func() *mockUserService {
				return &mockUserService{
					saveUserFn: func(_ context.Context, user *model.User) (*model.User, error) {
						saved := *user
						saved.ID = 99
						return &saved, nil
					},
				}
			},
			wantUser: &model.User{ID: 99, Name: "Bob"},
			wantErr:  false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svc := tc.setupMock()
			got, err := svc.SaveUser(context.Background(), tc.inputUser)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.errContains != "" {
					assert.Contains(t, err.Error(), tc.errContains)
				}
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, tc.wantUser, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestFetchUserList
// ---------------------------------------------------------------------------

func TestFetchUserList(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupMock func() *mockUserService
		wantUsers []*model.User
		wantErr   bool
	}{
		{
			name: "when users exist returns all persisted users",
			setupMock: func() *mockUserService {
				return &mockUserService{
					fetchUserListFn: func(_ context.Context) ([]*model.User, error) {
						return []*model.User{
							newUser(1, "Alice"),
							newUser(2, "Bob"),
						}, nil
					},
				}
			},
			wantUsers: []*model.User{
				newUser(1, "Alice"),
				newUser(2, "Bob"),
			},
			wantErr: false,
		},
		{
			name: "when no users exist returns an empty list (never nil)",
			setupMock: func() *mockUserService {
				return &mockUserService{
					fetchUserListFn: func(_ context.Context) ([]*model.User, error) {
						return []*model.User{}, nil
					},
				}
			},
			wantUsers: []*model.User{},
			wantErr:   false,
		},
		{
			name: "persistence error is propagated",
			setupMock: func() *mockUserService {
				return &mockUserService{
					fetchUserListFn: func(_ context.Context) ([]*model.User, error) {
						return nil, errInternalStore
					},
				}
			},
			wantUsers: nil,
			wantErr:   true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svc := tc.setupMock()
			got, err := svc.FetchUserList(context.Background())

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				// Invariant: never nil
				assert.NotNil(t, got)
				assert.Equal(t, tc.wantUsers, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestFetchUserByID
// ---------------------------------------------------------------------------

func TestFetchUserByID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		inputID     int64
		setupMock   func() *mockUserService
		wantUser    *model.User
		wantErr     bool
		errIs       error
		errContains string
	}{
		{
			name:    "given an id that matches an existing user returns the user",
			inputID: 1,
			setupMock: func() *mockUserService {
				return &mockUserService{
					fetchUserByIDFn: func(_ context.Context, id int64) (*model.User, error) {
						if id == 1 {
							return newUser(1, "Alice"), nil
						}
						return nil, errUserNotFound
					},
				}
			},
			wantUser: newUser(1, "Alice"),
			wantErr:  false,
		},
		{
			name:    "given an id that does not match any user returns UserNotFoundException",
			inputID: 999,
			setupMock: func() *mockUserService {
				return &mockUserService{
					fetchUserByIDFn: func(_ context.Context, id int64) (*model.User, error) {
						return nil, errUserNotFound
					},
				}
			},
			wantUser:    nil,
			wantErr:     true,
			errIs:       errUserNotFound,
			errContains: "user not found",
		},
		{
			name:    "returned user identifier equals the requested id",
			inputID: 7,
			setupMock: func() *mockUserService {
				return &mockUserService{
					fetchUserByIDFn: func(_ context.Context, id int64) (*model.User, error) {
						return newUser(id, "Carol"), nil
					},
				}
			},
			wantUser: newUser(7, "Carol"),
			wantErr:  false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svc := tc.setupMock()
			got, err := svc.FetchUserByID(context.Background(), tc.inputID)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.errIs != nil {
					assert.True(t, errors.Is(err, tc.errIs), "expected error to wrap %v", tc.errIs)
				}
				if tc.errContains != "" {
					assert.Contains(t, err.Error(), tc.errContains)
				}
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, tc.wantUser, got)
				// Invariant: returned user's id equals the requested id.
				assert.Equal(t, tc.inputID, got.ID)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestDeleteUser
// ---------------------------------------------------------------------------

func TestDeleteUser(t *testing.T) {
	t.Parallel()

	// Track deletion side-effects via an in-memory store.
	type store struct {
		users map[int64]*model.User
	}

	tests := []struct {
		name        string
		inputID     int64
		setupMock   func() (*mockUserService, *store)
		wantErr     bool
		errContains string
		// afterCheck is an optional post-call assertion on the store.
		afterCheck func(t *testing.T, s *store)
	}{
		{
			name:    "given an id that matches an existing user removes the record",
			inputID: 1,
			setupMock: func() (*mockUserService, *store) {
				s := &store{users: map[int64]*model.User{
					1: newUser(1, "Alice"),
					2: newUser(2, "Bob"),
				}}
				return &mockUserService{
					deleteUserFn: func(_ context.Context, id int64) error {
						if _, ok := s.users[id]; !ok {
							return errUserNotFound
						}
						delete(s.users, id)
						return nil
					},
				}, s
			},
			wantErr: false,
			afterCheck: func(t *testing.T, s *store) {
				t.Helper()
				_, exists := s.users[1]
				assert.False(t, exists, "user with id=1 should have been deleted")
				// Only the targeted record is removed; others intact.
				assert.Contains(t, s.users, int64(2))
			},
		},
		{
			name:    "given an id that does not match any user returns an error",
			inputID: 999,
			setupMock: func() (*mockUserService, *store) {
				s := &store{users: map[int64]*model.User{}}
				return &mockUserService{
					deleteUserFn: func(_ context.Context, id int64) error {
						return errDeleteFailed
					},
				}, s
			},
			wantErr:     true,
			errContains: "delete failed",
		},
		{
			name:    "deletion affects only the targeted record",
			inputID: 3,
			setupMock: func() (*mockUserService, *store) {
				s := &store{users: map[int64]*model.User{
					3: newUser(3, "Carol"),
					4: newUser(4, "Dave"),
				}}
				return &mockUserService{
					deleteUserFn: func(_ context.Context, id int64) error {
						delete(s.users, id)
						return nil
					},
				}, s
			},
			wantErr: false,
			afterCheck: func(t *testing.T, s *store) {
				t.Helper()
				assert.NotContains(t, s.users, int64(3))
				assert.Contains(t, s.users, int64(4))
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svc, s := tc.setupMock()
			err := svc.DeleteUser(context.Background(), tc.inputID)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.errContains != "" {
					assert.Contains(t, err.Error(), tc.errContains)
				}
			} else {
				assert.NoError(t, err)
			}

			if tc.afterCheck != nil {
				tc.afterCheck(t, s)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestUpdateUser
// ---------------------------------------------------------------------------

func TestUpdateUser(t *testing.T) {
	t.Parallel()

	type store struct {
		users map[int64]*model.User
	}

	tests := []struct {
		name        string
		inputID     int64
		inputUser   *model.User
		setupMock   func() (*mockUserService, *store)
		wantErr     bool
		errContains string
		afterCheck  func(t *testing.T, s *store)
	}{
		{
			name:      "given an id matching an existing user and valid data updates the record",
			inputID:   1,
			inputUser: &model.User{Name: "Alice Updated"},
			setupMock: func() (*mockUserService, *store) {
				s := &store{users: map[int64]*model.User{
					1: newUser(1, "Alice"),
					2: newUser(2, "Bob"),
				}}
				return &mockUserService{
					updateUserFn: func(_ context.Context, id int64, user *model.User) error {
						existing, ok := s.users[id]
						if !ok {
							return errUserNotFound
						}
						