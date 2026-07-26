```go
package service

import (
	"context"
	"testing"

	smartErrors "github.com/smartContact/internal/smartcontact/error"
	"github.com/smartContact/internal/smartcontact/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeUserDaoForTest is a hand-written test double for repository.UserDao used
// in the table-driven tests below. It satisfies the same UserDao interface as
// fakeUserDao defined in the migrated file but lives inside the _test build so
// it does not collide.
type fakeUserDaoForTest struct {
	findByNameFn func(ctx context.Context, name string) (*model.User, error)
	saveUserFn   func(ctx context.Context, u *model.User) (*model.User, error)
	findByIDFn   func(ctx context.Context, id int) (*model.User, error)
	findAllFn    func(ctx context.Context) ([]*model.User, error)
	deleteByIDFn func(ctx context.Context, id int) error
	countFn      func(ctx context.Context) (int64, error)
}

func (f *fakeUserDaoForTest) Save(ctx context.Context, u *model.User) (*model.User, error) {
	if f.saveUserFn != nil {
		return f.saveUserFn(ctx, u)
	}
	return u, nil
}

func (f *fakeUserDaoForTest) FindByID(ctx context.Context, id int) (*model.User, error) {
	if f.findByIDFn != nil {
		return f.findByIDFn(ctx, id)
	}
	return nil, smartErrors.NewUserNotFoundError()
}

func (f *fakeUserDaoForTest) FindAll(ctx context.Context) ([]*model.User, error) {
	if f.findAllFn != nil {
		return f.findAllFn(ctx)
	}
	return nil, nil
}

func (f *fakeUserDaoForTest) FindByName(ctx context.Context, name string) (*model.User, error) {
	if f.findByNameFn != nil {
		return f.findByNameFn(ctx, name)
	}
	return nil, smartErrors.NewUserNotFoundError()
}

func (f *fakeUserDaoForTest) DeleteByID(ctx context.Context, id int) error {
	if f.deleteByIDFn != nil {
		return f.deleteByIDFn(ctx, id)
	}
	return nil
}

func (f *fakeUserDaoForTest) Count(ctx context.Context) (int64, error) {
	if f.countFn != nil {
		return f.countFn(ctx)
	}
	return 0, nil
}

// ---------------------------------------------------------------------------
// TestGetUserByName_ValidName – spec: given a valid existing user name the
// service returns a User whose name equals the queried name.
// ---------------------------------------------------------------------------

func TestGetUserByName_ValidName(t *testing.T) {
	tests := []struct {
		name       string
		lookupName string
		storedUser *model.User
		wantName   string
	}{
		{
			name:       "exact name match – hemraj",
			lookupName: "hemraj",
			storedUser: model.NewUser(3, "hemraj", "hemrajmalhi1234@gmail.com", "Sr", "root", "java developer"),
			wantName:   "hemraj",
		},
		{
			name:       "exact name match – alice",
			lookupName: "alice",
			storedUser: model.NewUser(7, "alice", "alice@example.com", "about alice", "pass1", "tester"),
			wantName:   "alice",
		},
		{
			name:       "exact name match – bob",
			lookupName: "bob",
			storedUser: model.NewUser(42, "bob", "bob@example.com", "about bob", "secret", "engineer"),
			wantName:   "bob",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dao := &fakeUserDaoForTest{
				findByNameFn: func(ctx context.Context, name string) (*model.User, error) {
					assert.Equal(t, tt.lookupName, name, "DAO should receive the exact name passed to the service")
					return tt.storedUser, nil
				},
			}

			svc := NewUserService(dao)
			got, err := svc.GetUserByName(context.Background(), tt.lookupName)

			require.NoError(t, err, "should not return an error for a valid existing name")
			require.NotNil(t, got, "should return a non-nil user")

			// Global invariant: the returned user's name equals the queried name.
			assert.Equal(t, tt.wantName, got.Name,
				"returned user name must match the queried name")
		})
	}
}

// ---------------------------------------------------------------------------
// TestGetUserByName_UnknownName – spec: given a name that does not correspond
// to any user the service returns an error and no user.
// ---------------------------------------------------------------------------

func TestGetUserByName_UnknownName(t *testing.T) {
	tests := []struct {
		name       string
		lookupName string
		daoErr     error
	}{
		{
			name:       "name not in store",
			lookupName: "nobody",
			daoErr:     smartErrors.NewUserNotFoundError(),
		},
		{
			name:       "empty name not in store",
			lookupName: "",
			daoErr:     smartErrors.NewUserNotFoundError(),
		},
		{
			name:       "random unknown name",
			lookupName: "xyzzy_does_not_exist",
			daoErr:     smartErrors.NewUserNotFoundError(),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dao := &fakeUserDaoForTest{
				findByNameFn: func(ctx context.Context, name string) (*model.User, error) {
					return nil, tt.daoErr
				},
			}

			svc := NewUserService(dao)
			got, err := svc.GetUserByName(context.Background(), tt.lookupName)

			require.Error(t, err, "should return an error when no user exists for the given name")
			assert.Nil(t, got, "should return nil user when no user exists for the given name")
		})
	}
}

// ---------------------------------------------------------------------------
// TestGetUserByName_ReturnedUserFieldsIntact – global invariant: User objects
// expose accessible name, email, about, password, role, and id fields and
// retrieval by name does not modify persisted user data.
// ---------------------------------------------------------------------------

func TestGetUserByName_ReturnedUserFieldsIntact(t *testing.T) {
	tests := []struct {
		name        string
		lookupName  string
		storedUser  *model.User
		wantID      int
		wantName    string
		wantEmail   string
		wantAbout   string
		wantPassword string
		wantRole    string
	}{
		{
			name:         "all fields preserved after lookup",
			lookupName:   "hemraj",
			storedUser:   model.NewUser(3, "hemraj", "hemrajmalhi1234@gmail.com", "Sr", "root", "java developer"),
			wantID:       3,
			wantName:     "hemraj",
			wantEmail:    "hemrajmalhi1234@gmail.com",
			wantAbout:    "Sr",
			wantPassword: "root",
			wantRole:     "java developer",
		},
		{
			name:         "all fields preserved for second user",
			lookupName:   "carol",
			storedUser:   model.NewUser(99, "carol", "carol@example.com", "DevOps lead", "p@ssw0rd", "admin"),
			wantID:       99,
			wantName:     "carol",
			wantEmail:    "carol@example.com",
			wantAbout:    "DevOps lead",
			wantPassword: "p@ssw0rd",
			wantRole:     "admin",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dao := &fakeUserDaoForTest{
				findByNameFn: func(ctx context.Context, name string) (*model.User, error) {
					return tt.storedUser, nil
				},
			}

			svc := NewUserService(dao)
			got, err := svc.GetUserByName(context.Background(), tt.lookupName)

			require.NoError(t, err)
			require.NotNil(t, got)

			assert.Equal(t, tt.wantID, got.ID, "ID field must be intact")
			assert.Equal(t, tt.wantName, got.Name, "Name field must be intact")
			assert.Equal(t, tt.wantEmail, got.Email, "Email field must be intact")
			assert.Equal(t, tt.wantAbout, got.About, "About field must be intact")
			assert.Equal(t, tt.wantPassword, got.Password, "Password field must be intact")
			assert.Equal(t, tt.wantRole, got.Role, "Role field must be intact")
		})
	}
}

// ---------------------------------------------------------------------------
// TestGetUserByName_DoesNotMutateStore – global invariant: retrieval by name
// does not modify persisted user data. The Count reported by the DAO must be
// the same before and after a GetUserByName call.
// ---------------------------------------------------------------------------

func TestGetUserByName_DoesNotMutateStore(t *testing.T) {
	tests := []struct {
		name          string
		lookupName    string
		storedUser    *model.User
		initialCount  int64
	}{
		{
			name:         "count unchanged after successful lookup",
			lookupName:   "hemraj",
			storedUser:   model.NewUser(3, "hemraj", "hemrajmalhi1234@gmail.com", "Sr", "root", "java developer"),
			initialCount: 5,
		},
		{
			name:         "count unchanged after failed lookup",
			lookupName:   "ghost",
			storedUser:   nil,
			initialCount: 3,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			callCount := 0
			dao := &fakeUserDaoForTest{
				findByNameFn: func(ctx context.Context, name string) (*model.User, error) {
					callCount++
					if tt.storedUser == nil {
						return nil, smartErrors.NewUserNotFoundError()
					}
					return tt.storedUser, nil
				},
				countFn: func(ctx context.Context) (int64, error) {
					return tt.initialCount, nil
				},
			}

			svc := NewUserService(dao)

			countBefore, err := dao.Count(context.Background())
			require.NoError(t, err)

			// Perform the lookup (error is acceptable for the ghost case).
			_, _ = svc.GetUserByName(context.Background(), tt.lookupName)

			countAfter, err := dao.Count(context.Background())
			require.NoError(t, err)

			assert.Equal(t, countBefore, countAfter,
				"store count must not change after a read-only GetUserByName call")

			// Ensure FindByName was called exactly once.
			assert.Equal(t, 1, callCount,
				"FindByName on the DAO must be called exactly once per service call")
		})
	}
}

// ---------------------------------------------------------------------------
// TestGetUserByName_ContextPropagation – the context passed to the service
// must be forwarded to the DAO unchanged.
// ---------------------------------------------------------------------------

func TestGetUserByName_ContextPropagation(t *testing.T) {
	type ctxKey string
	const testKey ctxKey = "test-request-id"

	tests := []struct {
		name       string
		ctxValue   string
		lookupName string
		storedUser *model.User
	}{
		{
			name:       "context value reaches DAO",
			ctxValue:   "req-abc-123",
			lookupName: "hemraj",
			storedUser: model.NewUser(3, "hemraj", "hemrajmalhi1234@gmail.com", "Sr", "root", "java developer"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.WithValue(context.Background(), testKey, tt.ctxValue)
			var receivedCtx context.Context

			dao := &fakeUserDaoForTest{
				findByNameFn: func(c context.Context, name string) (*model.User, error) {
					receivedCtx = c
					return tt.storedUser, nil
				},
			}

			svc := NewUserService(dao)
			_, err := svc.GetUserByName(ctx, tt.lookupName)
			require.NoError(t, err)

			require.NotNil(t, receivedCtx, "DAO must receive the context")
			assert.Equal(t, tt.ctxValue, receivedCtx.Value(testKey),
				"context value must be propagated to the DAO intact")
		})
	}
}

// ---------------------------------------------------------------------------
// TestGetUserByName_UniqueByName – global invariant: a user is uniquely
// identifiable and retrievable by their name; two distinct names must yield
// two distinct users.
// ---------------------------------------------------------------------------

func TestGetUserByName_UniqueByName(t *testing.T) {
	tests := []struct {
		name    string
		users   map[string]*model.User
		lookups []string
	}{
		{
			name: "two distinct names return two distinct users",
			users: map[string]*model.User{
				"alice": model.NewUser(1, "alice", "alice@example.com", "about alice", "pw1", "dev"),
				"bob":   model.NewUser(2, "bob", "bob@example.com", "about bob", "pw2", "qa"),
			},
			lookups: []string{"alice", "bob"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dao := &fakeUserDaoForTest{
				findByNameFn: func(ctx context.Context, name string) (*model.User, error) {
					u, ok := tt.users[name]
					if !ok {
						return nil, smartErrors.NewUserNotFoundError()
					}
					return u, nil
				},
			}

			svc := NewUserService(dao)

			results := make(map[string]*model.User)
			for _, lookupName := range tt.lookups {
				got, err := svc.GetUserByName(context.Background(), lookupName)
				require.NoError(t, err, "lookup of %q should succeed", lookupName)
				require.NotNil(t, got)
				// Invariant: returned user's name equals queried name.
				assert.Equal(t, lookupName, got.Name)
				results[lookupName] = got
			}

			// All returned users must be distinct (different IDs).
			ids := make(map[int]struct{})
			for name, u := range results {
				_, duplicate := ids[u.ID]
				assert.False(t, duplicate, "user %q