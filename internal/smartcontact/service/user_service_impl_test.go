package service

// MIGRATION_NOTE: The Java source (UserServiceImpTest) was a @SpringBootTest
// integration test that was itself partly broken: it @Autowired a REAL
// UserServiceImp bean and then tried to Mockito.when(...).thenReturn(...) on
// that non-mock instance. Mockito.when only works on mocks/spies, so the
// stubbing had no effect and the test relied on undefined behavior.
//
// The idiomatic Go migration preserves the *intent* — "GetUserByName('hemraj')
// returns a user whose name is 'hemraj'" — using a hand-written mock UserDao
// injected into the real UserService via NewUserService. This is a proper unit
// test with no Spring context and no reflective mocking framework required.
//
// The Go file lives at internal/smartcontact/service/user_service_impl_test.go
// per the migration plan; the requested target filename
// (userserviceimptest.go) is not a valid Go test filename because `go test`
// only compiles files ending in _test.go. This content is a _test.go file.

import (
	"context"
	"testing"

	"github.com/smartContact/internal/smartcontact/model"
	"github.com/smartContact/internal/smartcontact/repository"
)

// mockUserDao is a hand-written test double implementing the same surface the
// UserService depends on. It replaces Mockito's reflective stubbing.
type mockUserDao struct {
	findByNameFn func(ctx context.Context, name string) (*model.User, error)
}

func (m *mockUserDao) FindByName(ctx context.Context, name string) (*model.User, error) {
	if m.findByNameFn != nil {
		return m.findByNameFn(ctx, name)
	}
	return nil, repository.ErrUserNotFound
}

func (m *mockUserDao) FindByID(ctx context.Context, id int) (*model.User, error) {
	return nil, repository.ErrUserNotFound
}

func (m *mockUserDao) FindAll(ctx context.Context) ([]*model.User, error) {
	return nil, nil
}

func (m *mockUserDao) Save(ctx context.Context, u *model.User) (*model.User, error) {
	return u, nil
}

func (m *mockUserDao) DeleteByID(ctx context.Context, id int) error {
	return nil
}

func (m *mockUserDao) ExistsByID(ctx context.Context, id int) (bool, error) {
	return false, nil
}

func (m *mockUserDao) Count(ctx context.Context) (int64, error) {
	return 0, nil
}

// TestGetUserByName mirrors the source test
// WhenValidDepartmentName_ThenUserShouldBeFound: for a valid name, the service
// returns the matching user.
//
// MIGRATION_NOTE: This assumes NewUserService accepts an interface satisfied by
// *mockUserDao (the migrated repository.UserDao surface). If NewUserService
// requires the concrete *repository.UserDao struct rather than an interface,
// introduce a small interface in the service package and depend on that — this
// is the one spot flagged for manual review.
func TestGetUserByName(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		stored    *model.User
		wantName  string
		wantError bool
	}{
		{
			name:  "valid name returns matching user",
			input: "hemraj",
			stored: &model.User{
				ID:       3,
				Name:     "hemraj",
				Email:    "hemrajmalhi1234@gmail.com",
				About:    "Sr",
				Password: "root",
				Role:     "java developer",
			},
			wantName:  "hemraj",
			wantError: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			dao := &mockUserDao{
				findByNameFn: func(ctx context.Context, name string) (*model.User, error) {
					if tc.stored != nil && tc.stored.Name == name {
						return tc.stored, nil
					}
					return nil, repository.ErrUserNotFound
				},
			}

			svc := NewUserService(dao)

			found, err := svc.GetUserByName(context.Background(), tc.input)
			if tc.wantError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if found == nil {
				t.Fatalf("expected a user, got nil")
			}
			if found.Name != tc.wantName {
				t.Errorf("GetUserByName(%q) name = %q, want %q", tc.input, found.Name, tc.wantName)
			}
		})
	}
}
