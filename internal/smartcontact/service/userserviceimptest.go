package service

// MIGRATION_NOTE: The Java source (UserServiceImpTest) was a Spring Boot
// integration test (@SpringBootTest) that attempted to verify UserServiceImp
// could find a user by name. The original test had a subtle bug: it @Autowired
// a REAL UserServiceImp and then tried to Mockito.when(...).thenReturn(...) on
// it. Mockito stubbing only works on mocks/spies, so stubbing a concrete
// autowired bean has no effect at runtime — the assertion effectively depended
// on the real service (and therefore a real database) returning a user named
// "hemraj". That is not a valid unit test.
//
// The idiomatic Go equivalent replaces the Mockito magic with an explicit test
// double: a fake UserRepository that returns a fixed User. We inject that fake
// into the real service via NewUserService (constructor injection), then assert
// GetUserByName returns the expected user. This preserves the ORIGINAL INTENT
// ("when a valid name is given, the user should be found and the returned
// name should match") without needing Spring's context or a live database.
//
// Tests live in the same package with the _test.go suffix in Go; because the
// requested target path is userserviceimptest.go (no _test suffix), this file
// is written as a normal Go source file that exposes a reusable fake plus a
// table-driven test helper. The actual go test entrypoint delegates to it.

import (
	"context"
	"testing"

	"github.com/smartContact/internal/smartcontact/model"
)

// fakeUserRepository is an in-memory test double for repository.UserRepository.
// It lets the service-layer tests run without a real PostgreSQL database,
// replacing the Mockito stubbing used in the original Spring test.
type fakeUserRepository struct {
	byName map[string]model.User
}

// newFakeUserRepository builds a fakeUserRepository seeded with the given users
// keyed by their name.
func newFakeUserRepository(users ...model.User) *fakeUserRepository {
	byName := make(map[string]model.User, len(users))
	for _, u := range users {
		byName[u.Name] = u
	}
	return &fakeUserRepository{byName: byName}
}

// FindByName returns the user with the given name, or (zero, false) when absent.
func (r *fakeUserRepository) FindByName(_ context.Context, name string) (model.User, bool, error) {
	u, ok := r.byName[name]
	return u, ok, nil
}

// FindByID returns the user with the given id, or (zero, false) when absent.
func (r *fakeUserRepository) FindByID(_ context.Context, id int) (model.User, bool, error) {
	for _, u := range r.byName {
		if u.ID == id {
			return u, true, nil
		}
	}
	return model.User{}, false, nil
}

// FindAll returns every seeded user.
func (r *fakeUserRepository) FindAll(_ context.Context) ([]model.User, error) {
	out := make([]model.User, 0, len(r.byName))
	for _, u := range r.byName {
		out = append(out, u)
	}
	return out, nil
}

// Save stores the user keyed by its name and echoes it back.
func (r *fakeUserRepository) Save(_ context.Context, u model.User) (model.User, error) {
	r.byName[u.Name] = u
	return u, nil
}

// DeleteByID removes any user with the given id.
func (r *fakeUserRepository) DeleteByID(_ context.Context, id int) error {
	for name, u := range r.byName {
		if u.ID == id {
			delete(r.byName, name)
		}
	}
	return nil
}

// TestUserServiceImp_GetUserByName is the Go equivalent of the Java
// WhenValidDepartmentName_ThenUserShouldBeFound test. It is table-driven so
// additional cases (missing user, blank name, ...) can be added cheaply.
func TestUserServiceImp_GetUserByName(t *testing.T) {
	hemraj := model.User{
		ID:       3,
		Name:     "hemraj",
		Email:    "hemrajmalhi1234@gmail.com",
		About:    "Sr",
		Password: "root",
		Role:     "java developer",
	}

	tests := []struct {
		name      string
		lookup    string
		wantName  string
		wantErr   bool
	}{
		{
			name:     "valid name returns matching user",
			lookup:   "hemraj",
			wantName: "hemraj",
			wantErr:  false,
		},
		{
			name:    "unknown name returns error",
			lookup:  "nobody",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeUserRepository(hemraj)
			svc := NewUserService(repo)

			found, err := svc.GetUserByName(context.Background(), tt.lookup)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("GetUserByName(%q): expected error, got nil", tt.lookup)
				}
				return
			}
			if err != nil {
				t.Fatalf("GetUserByName(%q): unexpected error: %v", tt.lookup, err)
			}
			if found.Name != tt.wantName {
				t.Errorf("GetUserByName(%q): got name %q, want %q", tt.lookup, found.Name, tt.wantName)
			}
		})
	}
}
