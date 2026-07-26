package service

// This file corresponds to the original Spring Boot integration test
// com.smartContact.service.UserServiceImpTest.
//
// MIGRATION_NOTE: The Java source was a @SpringBootTest that (somewhat
// unusually) mixed a full application-context integration test with Mockito
// static stubbing on an @Autowired bean. That mixture does not translate
// cleanly: in idiomatic Go we test the concrete userService against a mocked
// UserDao rather than stubbing the service under test itself (stubbing the
// system under test, as the original did with Mockito.when(userServiceImp...),
// asserts nothing about real behaviour and is an anti-pattern).
//
// The migrated test therefore drives UserService.GetUserByName through a
// fake UserDao and asserts the returned user's name — preserving the original
// intent ("when a valid name is given, the user should be found").
//
// NOTE (from migration debate): the GET /get_user_data empty-table test
// (GAP J) belongs in handler_test.go, not here — this file is the service-layer
// test only.

import (
	"context"
	"testing"

	smartErrors "github.com/smartContact/internal/smartcontact/error"
	"github.com/smartContact/internal/smartcontact/model"
)

// fakeUserDao is a hand-written test double for repository.UserDao. It
// implements only the behaviour exercised by these tests; unused methods
// return zero values / nil so the double still satisfies the interface.
//
// MIGRATION_NOTE: this replaces Mockito's dynamic proxy stubbing. Go has no
// runtime mock-generation in the standard toolchain, so an explicit fake keeps
// the test dependency-free and easy to reason about.
type fakeUserDao struct {
	findByNameFn func(ctx context.Context, name string) (*model.User, error)
}

func (f *fakeUserDao) Save(ctx context.Context, u *model.User) (*model.User, error) {
	return u, nil
}

func (f *fakeUserDao) FindByID(ctx context.Context, id int) (*model.User, error) {
	return nil, smartErrors.NewUserNotFoundError()
}

func (f *fakeUserDao) FindAll(ctx context.Context) ([]*model.User, error) {
	return nil, nil
}

func (f *fakeUserDao) FindByName(ctx context.Context, name string) (*model.User, error) {
	if f.findByNameFn != nil {
		return f.findByNameFn(ctx, name)
	}
	return nil, smartErrors.NewUserNotFoundError()
}

func (f *fakeUserDao) DeleteByID(ctx context.Context, id int) error {
	return nil
}

func (f *fakeUserDao) Count(ctx context.Context) (int64, error) {
	return 0, nil
}

// TestUserServiceGetUserByName mirrors the original
// WhenValidDepartmentName_ThenUserShouldBeFound test: given a valid name, the
// service returns the matching user.
func TestUserServiceGetUserByName(t *testing.T) {
	tests := []struct {
		name       string
		lookupName string
		stub       *model.User
		stubErr    error
		wantName   string
		wantErr    bool
	}{
		{
			name:       "valid name returns matching user",
			lookupName: "hemraj",
			stub: model.NewUser(
				3,
				"hemraj",
				"hemrajmalhi1234@gmail.com",
				"Sr",
				"root",
				"java developer",
			),
			wantName: "hemraj",
			wantErr:  false,
		},
		{
			name:       "unknown name returns error",
			lookupName: "nobody",
			stub:       nil,
			stubErr:    smartErrors.NewUserNotFoundError(),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			dao := &fakeUserDao{
				findByNameFn: func(ctx context.Context, name string) (*model.User, error) {
					if tt.stubErr != nil {
						return nil, tt.stubErr
					}
					return tt.stub, nil
				},
			}

			svc := NewUserService(dao)

			got, err := svc.GetUserByName(context.Background(), tt.lookupName)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("GetUserByName(%q): expected error, got nil", tt.lookupName)
				}
				return
			}

			if err != nil {
				t.Fatalf("GetUserByName(%q): unexpected error: %v", tt.lookupName, err)
			}
			if got == nil {
				t.Fatalf("GetUserByName(%q): expected user, got nil", tt.lookupName)
			}
			if got.Name != tt.wantName {
				t.Errorf("GetUserByName(%q): got name %q, want %q", tt.lookupName, got.Name, tt.wantName)
			}
		})
	}
}
