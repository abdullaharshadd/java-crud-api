```go
package service

import (
	"context"
	"testing"

	"github.com/smartContact/internal/smartcontact/model"
	"github.com/smartContact/internal/smartcontact/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockUserDao is a hand-written test double implementing the same surface the
// UserService depends on. It replaces Mockito's reflective stubbing.
type mockUserDao struct {
	findByNameFn func(ctx context.Context, name string) (*model.User, error)
	findByIDFn   func(ctx context.Context, id int) (*model.User, error)
	findAllFn    func(ctx context.Context) ([]*model.User, error)
	saveFn       func(ctx context.Context, u *model.User) (*model.User, error)
	deleteByIDFn func(ctx context.Context, id int) error
	existsByIDFn func(ctx context.Context, id int) (bool, error)
	countFn      func(ctx context.Context) (int64, error)
}

func (m *mockUserDao) FindByName(ctx context.Context, name string) (*model.User, error) {
	if m.findByNameFn != nil {
		return m.findByNameFn(ctx, name)
	}
	return nil, repository.ErrUserNotFound
}

func (m *mockUserDao) FindByID(ctx context.Context, id int) (*model.User, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, repository.ErrUserNotFound
}

func (m *mockUserDao) FindAll(ctx context.Context) ([]*model.User, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx)
	}
	return nil, nil
}

func (m *mockUserDao) Save(ctx context.Context, u *model.User) (*model.User, error) {
	if m.saveFn != nil {
		return m.saveFn(ctx, u)
	}
	return u, nil
}

func (m *mockUserDao) DeleteByID(ctx context.Context, id int) error {
	if m.deleteByIDFn != nil {
		return m.deleteByIDFn(ctx, id)
	}
	return nil
}

func (m *mockUserDao) ExistsByID(ctx context.Context, id int) (bool, error) {
	if m.existsByIDFn != nil {
		return m.existsByIDFn(ctx, id)
	}
	return false, nil
}

func (m *mockUserDao) Count(ctx context.Context) (int64, error) {
	if m.countFn != nil {
		return m.countFn(ctx)
	}
	return 0, nil
}

// TestGetUserByName covers the behavioral spec: getUserNameByName
// It validates that looking up a user by a name that exists returns a User
// whose Name field equals the queried name.
func TestGetUserByName(t *testing.T) {
	tests := []struct {
		name        string
		queryName   string
		storedUser  *model.User
		setupDAO    func(stored *model.User) *mockUserDao
		wantName    string
		wantNonNil  bool
		wantError   bool
		errContains string
	}{
		{
			name:      "given a valid name 'hemraj' that corresponds to an existing user, returns matching user",
			queryName: "hemraj",
			storedUser: &model.User{
				ID:       3,
				Name:     "hemraj",
				Email:    "hemrajmalhi1234@gmail.com",
				About:    "Sr",
				Password: "root",
				Role:     "java developer",
			},
			setupDAO: func(stored *model.User) *mockUserDao {
				return &mockUserDao{
					findByNameFn: func(ctx context.Context, name string) (*model.User, error) {
						if stored != nil && stored.Name == name {
							return stored, nil
						}
						return nil, repository.ErrUserNotFound
					},
				}
			},
			wantName:   "hemraj",
			wantNonNil: true,
			wantError:  false,
		},
		{
			name:       "given a name that does not correspond to any existing user, returns error",
			queryName:  "nonexistent",
			storedUser: nil,
			setupDAO: func(stored *model.User) *mockUserDao {
				return &mockUserDao{
					findByNameFn: func(ctx context.Context, name string) (*model.User, error) {
						return nil, repository.ErrUserNotFound
					},
				}
			},
			wantNonNil: false,
			wantError:  true,
		},
		{
			name:      "returned user name must exactly match queried name (invariant check)",
			queryName: "hemraj",
			storedUser: &model.User{
				ID:   3,
				Name: "hemraj",
			},
			setupDAO: func(stored *model.User) *mockUserDao {
				return &mockUserDao{
					findByNameFn: func(ctx context.Context, name string) (*model.User, error) {
						if stored != nil && stored.Name == name {
							return stored, nil
						}
						return nil, repository.ErrUserNotFound
					},
				}
			},
			wantName:   "hemraj",
			wantNonNil: true,
			wantError:  false,
		},
		{
			name:      "querying with correct name returns user with all fields intact",
			queryName: "hemraj",
			storedUser: &model.User{
				ID:       3,
				Name:     "hemraj",
				Email:    "hemrajmalhi1234@gmail.com",
				About:    "Sr",
				Password: "root",
				Role:     "java developer",
			},
			setupDAO: func(stored *model.User) *mockUserDao {
				return &mockUserDao{
					findByNameFn: func(ctx context.Context, name string) (*model.User, error) {
						if stored != nil && stored.Name == name {
							return stored, nil
						}
						return nil, repository.ErrUserNotFound
					},
				}
			},
			wantName:   "hemraj",
			wantNonNil: true,
			wantError:  false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			dao := tc.setupDAO(tc.storedUser)
			svc := NewUserService(dao)

			found, err := svc.GetUserByName(context.Background(), tc.queryName)

			if tc.wantError {
				assert.Error(t, err, "expected an error to be returned")
				assert.Nil(t, found, "expected returned user to be nil on error")
				return
			}

			require.NoError(t, err, "unexpected error calling GetUserByName")

			if tc.wantNonNil {
				require.NotNil(t, found, "expected a non-nil user to be returned")
				// Global invariant: returned user's name must match queried name
				assert.Equal(t, tc.wantName, found.Name,
					"GetUserByName(%q): returned user name must equal the queried name", tc.queryName)
			}

			// Additional field assertions for full-user test cases
			if tc.storedUser != nil && found != nil {
				assert.Equal(t, tc.storedUser.ID, found.ID,
					"returned user ID should match stored user ID")
				assert.Equal(t, tc.storedUser.Email, found.Email,
					"returned user Email should match stored user Email")
			}
		})
	}
}

// TestGetUserByName_NameInvariant explicitly validates the global invariant:
// "Looking up a user by a name that exists must return a User whose name field
// equals the queried name."
func TestGetUserByName_NameInvariant(t *testing.T) {
	const targetName = "hemraj"

	storedUser := &model.User{
		ID:       3,
		Name:     targetName,
		Email:    "hemrajmalhi1234@gmail.com",
		About:    "Sr",
		Password: "root",
		Role:     "java developer",
	}

	dao := &mockUserDao{
		findByNameFn: func(ctx context.Context, name string) (*model.User, error) {
			if name == storedUser.Name {
				return storedUser, nil
			}
			return nil, repository.ErrUserNotFound
		},
	}

	svc := NewUserService(dao)

	found, err := svc.GetUserByName(context.Background(), targetName)

	require.NoError(t, err)
	require.NotNil(t, found, "invariant violated: user must be non-nil for a valid existing name")
	assert.Equal(t, targetName, found.Name,
		"invariant violated: returned user's name must match the queried name")
}

// TestGetUserByName_ServiceResolvesFromDataStore validates the global invariant:
// "The service must be able to resolve a user identified by name from the
// underlying data store."
func TestGetUserByName_ServiceResolvesFromDataStore(t *testing.T) {
	tests := []struct {
		name       string
		queryName  string
		dataStore  map[string]*model.User
		wantFound  bool
		wantError  bool
	}{
		{
			name:      "service resolves existing user 'hemraj' from data store",
			queryName: "hemraj",
			dataStore: map[string]*model.User{
				"hemraj": {
					ID:    3,
					Name:  "hemraj",
					Email: "hemrajmalhi1234@gmail.com",
				},
			},
			wantFound: true,
			wantError: false,
		},
		{
			name:      "service returns error when name not in data store",
			queryName: "unknown_user",
			dataStore: map[string]*model.User{
				"hemraj": {
					ID:   3,
					Name: "hemraj",
				},
			},
			wantFound: false,
			wantError: true,
		},
		{
			name:      "service returns error when data store is empty",
			queryName: "hemraj",
			dataStore: map[string]*model.User{},
			wantFound: false,
			wantError: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			dao := &mockUserDao{
				findByNameFn: func(ctx context.Context, name string) (*model.User, error) {
					if u, ok := tc.dataStore[name]; ok {
						return u, nil
					}
					return nil, repository.ErrUserNotFound
				},
			}

			svc := NewUserService(dao)

			found, err := svc.GetUserByName(context.Background(), tc.queryName)

			if tc.wantError {
				assert.Error(t, err)
				assert.Nil(t, found)
				return
			}

			require.NoError(t, err)
			if tc.wantFound {
				require.NotNil(t, found)
				// Global invariant: name must match
				assert.Equal(t, tc.queryName, found.Name,
					"service must resolve a user whose name matches the queried name")
			}
		})
	}
}
```