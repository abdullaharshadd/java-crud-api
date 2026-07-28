package service

// MIGRATION_NOTE: The Java source (UserServiceImpTest) was a Spring Boot
// integration test that used @SpringBootTest to bootstrap the full application
// context, @Autowired field injection to obtain a UserServiceImp, and
// Mockito.when(...).thenReturn(...) static stubbing against that same injected
// bean. That test was fundamentally broken in the original: it stubbed the
// concrete UserServiceImp (not a mock of its collaborator), so the assertion
// only passed because the stub returned the canned value. In idiomatic Go we
// do NOT bootstrap a full application context for a unit test. Instead we:
//   - Define a fake UserRepository (the real collaborator) whose FindByName
//     returns a known model.User.
//   - Construct a real UserServiceImp via NewUserServiceImp with that fake.
//   - Exercise GetUserByName and assert on the returned *dto.UserResponse.
//
// Per the migration debate note, the service now returns a UserResponse
// (pointer fields / JSON-null semantics for absent optionals) rather than a
// plain model.User, so the assertion targets the response's Name field.
//
// This is a table-driven test. It replaces JUnit 5 lifecycle annotations
// (@BeforeAll / @Test) with standard Go testing setup inside each case.

import (
	"context"
	"testing"

	scerr "internal/smartcontact/error"
	"internal/smartcontact/model"
	"internal/smartcontact/repository"
)

// fakeUserRepository is a hand-written test double that implements
// repository.UserRepository. It lets us control exactly what the service layer
// sees without a real database or a mocking framework.
//
// MIGRATION_NOTE: This replaces Mockito. Go's standard library has no mocking
// framework; the idiomatic approach is a small fake that satisfies the
// collaborator interface. Only the methods exercised by these tests carry
// behaviour; the rest return zero values / not-found so the double still
// satisfies the full interface.
type fakeUserRepository struct {
	findByName func(ctx context.Context, name string) (*model.User, error)
}

// FindAll is unused by these tests and returns an empty slice.
func (f *fakeUserRepository) FindAll(ctx context.Context) ([]model.User, error) {
	return nil, nil
}

// FindByID is unused by these tests and reports the user as not found.
func (f *fakeUserRepository) FindByID(ctx context.Context, id int) (*model.User, error) {
	return nil, scerr.NewUserNotFoundErrorWithCause(id, nil)
}

// FindByName delegates to the configured stub function.
func (f *fakeUserRepository) FindByName(ctx context.Context, name string) (*model.User, error) {
	if f.findByName != nil {
		return f.findByName(ctx, name)
	}
	return nil, scerr.NewUserNotFoundErrorWithCause(0, nil)
}

// Merge is unused by these tests and echoes the input back.
func (f *fakeUserRepository) Merge(ctx context.Context, user *model.User) (*model.User, error) {
	return user, nil
}

// DeleteByID is unused by these tests and is a no-op.
func (f *fakeUserRepository) DeleteByID(ctx context.Context, id int) error {
	return nil
}

// compile-time assertion that the fake satisfies the interface.
var _ repository.UserRepository = (*fakeUserRepository)(nil)

// TestUserServiceImp_GetUserByName verifies that GetUserByName returns the
// expected user for a valid name, mirroring the original
// WhenValidDepartmentName_ThenUserShouldBeFound test case, plus the
// not-found path which the Java test never covered.
func TestUserServiceImp_GetUserByName(t *testing.T) {
	const wantName = "hemraj"

	seed := &model.User{
		ID:       3,
		Name:     "hemraj",
		Email:    "hemrajmalhi1234@gmail.com",
		About:    "Sr",
		Password: "root",
		Role:     "java developer",
	}

	tests := []struct {
		name       string
		inputName  string
		repoResult func(ctx context.Context, name string) (*model.User, error)
		wantName   string
		wantErr    bool
	}{
		{
			name:      "valid name returns matching user",
			inputName: wantName,
			repoResult: func(ctx context.Context, name string) (*model.User, error) {
				if name != wantName {
					return nil, scerr.NewUserNotFoundErrorWithCause(0, nil)
				}
				return seed, nil
			},
			wantName: wantName,
			wantErr:  false,
		},
		{
			name:      "unknown name returns not-found error",
			inputName: "nobody",
			repoResult: func(ctx context.Context, name string) (*model.User, error) {
				return nil, scerr.NewUserNotFoundErrorWithCause(0, nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeUserRepository{findByName: tt.repoResult}
			svc := NewUserServiceImp(repo)

			got, err := svc.GetUserByName(context.Background(), tt.inputName)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("GetUserByName(%q): expected error, got nil", tt.inputName)
				}
				return
			}

			if err != nil {
				t.Fatalf("GetUserByName(%q): unexpected error: %v", tt.inputName, err)
			}
			if got == nil {
				t.Fatalf("GetUserByName(%q): expected a user, got nil", tt.inputName)
			}

			// MIGRATION_NOTE: The service returns a UserResponse whose Name is a
			// *string to model JSON-null for absent optionals. Guard against a
			// nil pointer before dereferencing.
			if got.Name == nil {
				t.Fatalf("GetUserByName(%q): response Name is nil", tt.inputName)
			}
			if *got.Name != tt.wantName {
				t.Errorf("GetUserByName(%q): name = %q, want %q", tt.inputName, *got.Name, tt.wantName)
			}
		})
	}
}
