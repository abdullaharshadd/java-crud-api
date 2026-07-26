```go
package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// Minimal domain type – adjust import path / field names to match the real
// model once it is available in the project.
// ---------------------------------------------------------------------------

// If the real User / UserService types are already declared in userservice.go
// (same package) we do NOT re-declare them here. The stubs below are guarded
// by a build tag so they only compile when the real declarations are absent.
// In a real project you would simply import the package; because these tests
// live IN the package we rely on the existing declarations.

// ---------------------------------------------------------------------------
// Mock UserDao / repository
// ---------------------------------------------------------------------------

// mockUserDAO implements whatever repository interface userService depends on.
// Adjust method signatures to match the real interface.
type mockUserDAO struct {
	// storage keyed by ID
	store map[int]User

	// controls what FindByName returns
	findByNameResult *User

	// force Save / FindAll / DeleteById to return an error
	saveErr   error
	findErr   error
	deleteErr error
}

func newMockUserDAO() *mockUserDAO {
	return &mockUserDAO{store: make(map[int]User)}
}

func (m *mockUserDAO) Save(u User) (User, error) {
	if m.saveErr != nil {
		return User{}, m.saveErr
	}
	if u.ID == 0 {
		// simulate auto-increment
		u.ID = len(m.store) + 1
	}
	m.store[u.ID] = u
	return u, nil
}

func (m *mockUserDAO) FindAll() ([]User, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	out := make([]User, 0, len(m.store))
	for _, v := range m.store {
		out = append(out, v)
	}
	return out, nil
}

func (m *mockUserDAO) FindById(id int) (User, error) {
	if m.findErr != nil {
		return User{}, m.findErr
	}
	u, ok := m.store[id]
	if !ok {
		return User{}, errors.New("User are not available")
	}
	return u, nil
}

func (m *mockUserDAO) DeleteById(id int) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.store, id)
	return nil
}

func (m *mockUserDAO) FindByName(name string) *User {
	return m.findByNameResult
}

// ---------------------------------------------------------------------------
// Helper: build a service wired to the mock DAO.
// Adjust NewUserService signature as needed.
// ---------------------------------------------------------------------------

func newTestService(dao *mockUserDAO) UserService {
	return NewUserService(dao)
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

// TestInterfaceGuard verifies the compile-time assertion in
// userserviceimp.go (var _ UserService = (*userService)(nil)) does not
// prevent the package from building.
func TestInterfaceGuard(t *testing.T) {
	// If the package compiles, the guard is satisfied. Nothing else needed.
	t.Log("compile-time interface guard: OK")
}

// ---------------------------------------------------------------------------
// SaveUser / CreateUser
// ---------------------------------------------------------------------------

func TestSaveUser(t *testing.T) {
	type tc struct {
		name      string
		input     User
		seedStore map[int]User // pre-populate DAO before the call
		saveErr   error
		wantErr   bool
		check     func(t *testing.T, got User)
	}

	tests := []tc{
		{
			name:  "valid new user receives generated ID",
			input: User{Name: "Alice", Email: "alice@example.com"},
			check: func(t *testing.T, got User) {
				assert.NotZero(t, got.ID, "expected auto-generated ID")
				assert.Equal(t, "Alice", got.Name)
			},
		},
		{
			name:      "user with existing ID is updated",
			seedStore: map[int]User{42: {ID: 42, Name: "Old", Email: "old@example.com"}},
			input:     User{ID: 42, Name: "New", Email: "new@example.com"},
			check: func(t *testing.T, got User) {
				assert.Equal(t, 42, got.ID)
				assert.Equal(t, "New", got.Name)
			},
		},
		{
			name:    "repository error is propagated",
			input:   User{Name: "Broken"},
			saveErr: errors.New("db connection lost"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			dao := newMockUserDAO()
			if tt.seedStore != nil {
				dao.store = tt.seedStore
			}
			dao.saveErr = tt.saveErr

			svc := newTestService(dao)
			got, err := svc.SaveUser(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// FetchUserList
// ---------------------------------------------------------------------------

func TestFetchUserList(t *testing.T) {
	type tc struct {
		name      string
		seedStore map[int]User
		findErr   error
		wantErr   bool
		wantLen   int
	}

	tests := []tc{
		{
			name: "returns all users when store is populated",
			seedStore: map[int]User{
				1: {ID: 1, Name: "A"},
				2: {ID: 2, Name: "B"},
			},
			wantLen: 2,
		},
		{
			name:    "returns empty slice when store is empty",
			wantLen: 0,
		},
		{
			name:    "repository error is propagated",
			findErr: errors.New("timeout"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			dao := newMockUserDAO()
			if tt.seedStore != nil {
				dao.store = tt.seedStore
			}
			dao.findErr = tt.findErr

			svc := newTestService(dao)
			users, err := svc.FetchUserList()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, users, "result must never be nil")
			assert.Len(t, users, tt.wantLen)
		})
	}
}

// ---------------------------------------------------------------------------
// FetchUserById
// ---------------------------------------------------------------------------

func TestFetchUserById(t *testing.T) {
	type tc struct {
		name      string
		seedStore map[int]User
		id        int
		findErr   error
		wantErr   bool
		wantUser  *User
	}

	tests := []tc{
		{
			name:      "returns user when ID exists",
			seedStore: map[int]User{1: {ID: 1, Name: "Alice"}},
			id:        1,
			wantUser:  &User{ID: 1, Name: "Alice"},
		},
		{
			name:    "returns UserNotFoundException when ID is missing",
			id:      999,
			wantErr: true,
		},
		{
			name:    "repository error is propagated",
			id:      1,
			findErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			dao := newMockUserDAO()
			if tt.seedStore != nil {
				dao.store = tt.seedStore
			}
			dao.findErr = tt.findErr

			svc := newTestService(dao)
			got, err := svc.FetchUserById(tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				// check the canonical not-found message where applicable
				if tt.findErr == nil {
					assert.Contains(t, err.Error(), "User are not available")
				}
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantUser.ID, got.ID)
			assert.Equal(t, tt.wantUser.Name, got.Name)
		})
	}
}

// ---------------------------------------------------------------------------
// DeleteUser
// ---------------------------------------------------------------------------

func TestDeleteUser(t *testing.T) {
	type tc struct {
		name      string
		seedStore map[int]User
		id        int
		deleteErr error
		wantErr   bool
		checkDAO  func(t *testing.T, dao *mockUserDAO)
	}

	tests := []tc{
		{
			name:      "removes existing user from store",
			seedStore: map[int]User{5: {ID: 5, Name: "Bob"}},
			id:        5,
			checkDAO: func(t *testing.T, dao *mockUserDAO) {
				_, ok := dao.store[5]
				assert.False(t, ok, "user should be removed")
			},
		},
		{
			name: "no error when ID does not exist (delete is idempotent by default)",
			id:   999,
		},
		{
			name:      "repository error is propagated",
			id:        1,
			deleteErr: errors.New("constraint violation"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			dao := newMockUserDAO()
			if tt.seedStore != nil {
				dao.store = tt.seedStore
			}
			dao.deleteErr = tt.deleteErr

			svc := newTestService(dao)
			err := svc.DeleteUser(tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			if tt.checkDAO != nil {
				tt.checkDAO(t, dao)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// UpdateUser
// ---------------------------------------------------------------------------

func TestUpdateUser(t *testing.T) {
	type tc struct {
		name      string
		seedStore map[int]User
		id        int
		input     User
		saveErr   error
		wantErr   bool
		// returned struct should equal input with ID set
		wantUser *User
		checkDAO func(t *testing.T, dao *mockUserDAO)
	}

	tests := []tc{
		{
			name:      "sets ID on user and persists (existing record)",
			seedStore: map[int]User{10: {ID: 10, Name: "Old"}},
			id:        10,
			input:     User{Name: "Updated", Email: "u@example.com"},
			wantUser:  &User{ID: 10, Name: "Updated", Email: "u@example.com"},
			checkDAO: func(t *testing.T, dao *mockUserDAO) {
				stored := dao.store[10]
				assert.Equal(t, 10, stored.ID)
				assert.Equal(t, "Updated", stored.Name)
			},
		},
		{
			name:  "upserts a record when ID does not exist",
			id:    77,
			input: User{Name: "New"},
			wantUser: &User{ID: 77, Name: "New"},
			checkDAO: func(t *testing.T, dao *mockUserDAO) {
				stored := dao.store[77]
				assert.Equal(t, 77, stored.ID)
			},
		},
		{
			name:    "propagates repository error",
			id:      1,
			input:   User{Name: "Broken"},
			saveErr: errors.New("write failed"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			dao := newMockUserDAO()
			if tt.seedStore != nil {
				dao.store = tt.seedStore
			}
			dao.saveErr = tt.saveErr

			svc := newTestService(dao)
			got, err := svc.UpdateUser(tt.id, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			// The persisted user must always have ID == the provided id parameter.
			assert.Equal(t, tt.id, got.ID, "persisted ID must match the supplied id parameter")
			if tt.wantUser != nil {
				assert.Equal(t, tt.wantUser.Name, got.Name)
				assert.Equal(t, tt.wantUser.Email, got.Email)
			}
			if tt.checkDAO != nil {
				tt.checkDAO(t, dao)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// GetUserByName / getUserNameByName
// ---------------------------------------------------------------------------

func TestGetUserByName(t *testing.T) {
	type tc struct {
		name             string
		searchName       string
		findByNameResult *User
		// service returns exactly what the DAO returns – no not-found check
		wantNil  bool
		wantUser *User
	}

	alice := &User{ID: 3, Name: "Alice"}

	tests := []tc{
		{
			name:             "returns matching user when found",
			searchName:       "Alice",
			findByNameResult: alice,
			wantUser:         alice,
		},
		{
			name:             "returns nil when no user matches",
			searchName:       "Ghost",
			findByNameResult: nil,
			wantNil:          true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			dao := newMockUserDAO()
			dao.findByNameResult = tt.findByNameResult

			svc := newTestService(dao)
			got := svc.GetUserByName(tt.searchName)

			if tt.wantNil {
				assert.Nil(t, got, "expected nil for no-match")
				return
			}
			assert.NotNil(t, got)
			assert.Equal(t, tt.wantUser.ID, got.ID)
			assert.Equal(t, tt.wantUser.Name, got.Name)
		})
	}
}

// ---------------------------------------------------------------------------
// Global invariants
// ---------------------------------------------------------------------------

// TestServiceHoldsNoState verifies that two independent service instances
// wired to different DAOs do not share state.
func TestServiceHoldsNoState(t *testing.T) {
	dao1 := newMockUserDAO()
	dao2 := newMockUserDAO()

	svc1 := newTestService(dao1)
	svc2 := newTestService(dao2)

	_, _ = svc1.SaveUser(User{Name: "Only-in-svc1"})

	users2, err := svc2.FetchUserList()
	assert.NoError(t, err)
	assert.Empty(t, users2, "svc2 must not see data saved through svc1")
}

// TestFetchUserByIdNeverReturnsNilUser ensures a non-error result is always a
// populated user (not zero-value / nil).
func TestFetchUserByIdNeverReturnsNilUser(t *testing.T) {
	dao := newMockUserDAO()
	dao.store[1] = User{ID: 