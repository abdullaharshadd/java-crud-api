package service

// MIGRATION_NOTE: The Java source (UserServiceImpTest) was a Spring Boot
// integration test that bootstrapped the full application context
// (@SpringBootTest), field-injected the concrete UserServiceImp via @Autowired,
// and then attempted to stub that same injected bean with Mockito
// (Mockito.when(userServiceImp.getUserNameByName("hemraj")).thenReturn(user)).
//
// MIGRATION_NOTE: That original test is fundamentally broken and is NOT
// preserved verbatim. You cannot stub a real @Autowired bean with Mockito.when
// unless it is a @MockBean/@Spy — calling when() on a real object actually
// invokes the real method. So in Java this either threw or exercised the real
// service against no database. We translate the *intent* ("a valid name returns
// the matching user") into an idiomatic, table-driven Go test that injects a
// fake repository (test double) into the real UserService, which is the correct
// Go analogue of "mock the collaborator, test the real unit under test".
//
// MIGRATION_NOTE: In Go, tests live in *_test.go files with a TestXxx(t *testing.T)
// signature and are run by `go test`. The target path requested
// (userserviceimptest.go) is not a valid test file name, so the build tag /
// filename convention means this must be renamed to service_test.go to actually
// execute as a test. It is kept in the `service` package as an internal test so
// it can construct the concrete service directly.

import (
	"context"
	"errors"
	"testing"

	smartError "github.com/smartContact/internal/smartcontact/error"
	"github.com/smartContact/internal/smartcontact/model"
	"github.com/smartContact/internal/smartcontact/repository"
)

// fakeQuerier is a hand-written test double implementing repository.Querier.
// It lets us inject deterministic behaviour into the real UserService without a
// live PostgreSQL database, replacing the Mockito stub from the Java source.
type fakeQuerier struct {
	findByNameFn func(ctx context.Context, name string) (model.User, error)
	findByIDFn   func(ctx context.Context, id int64) (model.User, error)
	findAllFn    func(ctx context.Context) ([]model.User, error)
	saveFn       func(ctx context.Context, u model.User) (model.User, error)
	deleteByIDFn func(ctx context.Context, id int64) error
}

func (f *fakeQuerier) FindByName(ctx context.Context, name string) (model.User, error) {
	if f.findByNameFn != nil {
		return f.findByNameFn(ctx, name)
	}
	return model.User{}, smartError.NewUserNotFoundErrorf("user %q not found", name)
}

func (f *fakeQuerier) FindByID(ctx context.Context, id int64) (model.User, error) {
	if f.findByIDFn != nil {
		return f.findByIDFn(ctx, id)
	}
	return model.User{}, smartError.NewUserNotFoundErrorf("user %d not found", id)
}

func (f *fakeQuerier) FindAll(ctx context.Context) ([]model.User, error) {
	if f.findAllFn != nil {
		return f.findAllFn(ctx)
	}
	return nil, nil
}

func (f *fakeQuerier) Save(ctx context.Context, u model.User) (model.User, error) {
	if f.saveFn != nil {
		return f.saveFn(ctx, u)
	}
	return u, nil
}

func (f *fakeQuerier) DeleteByID(ctx context.Context, id int64) error {
	if f.deleteByIDFn != nil {
		return f.deleteByIDFn(ctx, id)
	}
	return nil
}

// TestGetUserByName_WhenValidName_ThenUserShouldBeFound is the direct migration
// of the Java WhenValidDepartmentName_ThenUserShouldBeFound test: given a valid
// name, the service returns the matching user.
func TestGetUserByName_WhenValidName_ThenUserShouldBeFound(t *testing.T) {
	const wantName = "hemraj"

	seeded := model.NewUser(
		3,
		wantName,
		"hemrajmalhi1234@gmail.com",
		"Sr",
		"root",
		"java developer",
	)

	tests := []struct {
		name       string
		lookupName string
		repo       *fakeQuerier
		wantName   string
		wantErr    bool
		wantNotFnd bool
	}{
		{
			name:       "valid name returns matching user",
			lookupName: wantName,
			repo: &fakeQuerier{
				findByNameFn: func(_ context.Context, name string) (model.User, error) {
					if name != wantName {
						return model.User{}, smartError.NewUserNotFoundErrorf("user %q not found", name)
					}
					return seeded, nil
				},
			},
			wantName: wantName,
		},
		{
			name:       "unknown name returns not-found error",
			lookupName: "nobody",
			repo: &fakeQuerier{
				findByNameFn: func(_ context.Context, name string) (model.User, error) {
					return model.User{}, smartError.NewUserNotFoundErrorf("user %q not found", name)
				},
			},
			wantErr:    true,
			wantNotFnd: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repository.NewUserRepository(tt.repo)
			svc := NewUserService(repo)

			got, err := svc.GetUserByName(context.Background(), tt.lookupName)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("GetUserByName(%q): expected error, got nil", tt.lookupName)
				}
				if tt.wantNotFnd && !smartError.IsUserNotFound(err) {
					t.Fatalf("GetUserByName(%q): expected user-not-found error, got %v", tt.lookupName, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("GetUserByName(%q): unexpected error: %v", tt.lookupName, err)
			}
			if got.Name != tt.wantName {
				t.Errorf("GetUserByName(%q): got name %q, want %q", tt.lookupName, got.Name, tt.wantName)
			}
		})
	}
}

// TestGetUserByName_PropagatesRepositoryError verifies that a non-not-found
// repository failure is surfaced to the caller rather than being swallowed or
// misclassified as a not-found condition.
func TestGetUserByName_PropagatesRepositoryError(t *testing.T) {
	sentinel := errors.New("connection reset by peer")

	repo := repository.NewUserRepository(&fakeQuerier{
		findByNameFn: func(_ context.Context, _ string) (model.User, error) {
			return model.User{}, sentinel
		},
	})
	svc := NewUserService(repo)

	_, err := svc.GetUserByName(context.Background(), "hemraj")
	if err == nil {
		t.Fatal("GetUserByName: expected error, got nil")
	}
	if smartError.IsUserNotFound(err) {
		t.Fatalf("GetUserByName: infrastructure error misclassified as not-found: %v", err)
	}
	if !errors.Is(err, sentinel) {
		t.Fatalf("GetUserByName: expected wrapped sentinel error, got %v", err)
	}
}
