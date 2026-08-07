package service

// MIGRATION_NOTE: The Java source, UserServiceImpTest, was a Spring Boot
// integration test (@SpringBootTest) that @Autowired the real UserServiceImp
// bean and then attempted to Mockito.when(...).thenReturn(...) on it. This test
// is fundamentally flawed as written: Mockito can only stub mocks/spies, not a
// concrete autowired bean, so the stub in setUp() has no effect and the test
// would either fail with a MissingMethodInvocation/UnfinishedStubbing error or
// silently exercise the real (empty) repository. The Java @BeforeAll on a
// non-static method without @TestInstance(PER_CLASS) is also invalid.
//
// The idiomatic Go replacement below preserves the *intent* of the test
// ("GetUserByName returns the user whose name matches the query") by using a
// hand-written in-memory fake of the repository.UserRepository interface,
// injecting it via the existing NewUserService constructor, and asserting the
// returned user's name with a table-driven test. This removes the flaw while
// keeping the business behaviour under test.
//
// This file lives in the same package (service) as userserviceimp.go, so it
// declares NO production types that would collide with the already-migrated
// UserService/NewUserService symbols — only test helpers and Test functions.

import (
	"context"
	"testing"

	domainerr "github.com/smartContact/internal/smartcontact/error"
	"github.com/smartContact/internal/smartcontact/model"
)

// fakeUserRepository is an in-memory implementation of
// repository.UserRepository used exclusively by the service-layer tests. It
// replaces the Mockito stub from the Java source with an explicit fake, which
// is the idiomatic Go approach for unit-testing a service against its
// repository dependency.
type fakeUserRepository struct {
	users []model.User
}

// FindByName returns the first user whose Name matches name. If no user is
// found it returns a wrapped ErrUserNotFound so callers can use errors.Is.
func (f *fakeUserRepository) FindByName(_ context.Context, name string) (model.User, error) {
	for _, u := range f.users {
		if u.Name == name {
			return u, nil
		}
	}
	return model.User{}, domainerr.WrapUserNotFound(name)
}

// FindByID returns the user with the given id, or a wrapped ErrUserNotFound.
func (f *fakeUserRepository) FindByID(_ context.Context, id int) (model.User, error) {
	for _, u := range f.users {
		if u.ID == id {
			return u, nil
		}
	}
	return model.User{}, domainerr.WrapUserNotFound("")
}

// FindAll returns all users held by the fake.
func (f *fakeUserRepository) FindAll(_ context.Context) ([]model.User, error) {
	return f.users, nil
}

// Save appends the user to the fake's backing slice and returns it.
func (f *fakeUserRepository) Save(_ context.Context, user model.User) (model.User, error) {
	f.users = append(f.users, user)
	return user, nil
}

// DeleteByID removes the user with the given id from the fake.
func (f *fakeUserRepository) DeleteByID(_ context.Context, id int) error {
	for i, u := range f.users {
		if u.ID == id {
			f.users = append(f.users[:i], f.users[i+1:]...)
			return nil
		}
	}
	return domainerr.WrapUserNotFound("")
}

// TestGetUserByName verifies that UserService.GetUserByName returns the user
// whose name matches the query. This is the Go equivalent of the Java
// WhenValidDepartmentName_ThenUserShouldBeFound test, corrected to use an
// injected fake repository instead of an ineffective Mockito stub.
func TestGetUserByName(t *testing.T) {
	seed := model.User{
		ID:       3,
		Name:     "hemraj",
		Email:    "hemrajmalhi1234@gmail.com",
		About:    "Sr",
		Password: "root",
		Role:     "java developer",
	}

	tests := []struct {
		name      string
		query     string
		wantName  string
		wantFound bool
	}{
		{
			name:      "valid name returns matching user",
			query:     "hemraj",
			wantName:  "hemraj",
			wantFound: true,
		},
		{
			name:      "unknown name returns not-found error",
			query:     "nobody",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeUserRepository{users: []model.User{seed}}
			svc := NewUserService(repo)

			found, err := svc.GetUserByName(context.Background(), tt.query)
			if tt.wantFound {
				if err != nil {
					t.Fatalf("GetUserByName(%q) returned unexpected error: %v", tt.query, err)
				}
				if found.Name != tt.wantName {
					t.Errorf("GetUserByName(%q) name = %q, want %q", tt.query, found.Name, tt.wantName)
				}
				return
			}

			if err == nil {
				t.Fatalf("GetUserByName(%q) = %+v, want error", tt.query, found)
			}
		})
	}
}
