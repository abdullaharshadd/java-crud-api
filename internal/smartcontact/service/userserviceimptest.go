package service

// MIGRATION_NOTE: The Java source UserServiceImpTest.java was a Spring Boot
// integration test (@SpringBootTest) that attempted to verify
// UserServiceImp.getUserNameByName. The original test was flawed in a way that
// cannot (and should not) be carried over verbatim:
//
//   - It @Autowired the *real* UserServiceImp bean and then called
//     Mockito.when(...).thenReturn(...) on it. Mockito stubbing only works on
//     Mockito mocks, not on real Spring-managed beans, so the stub had no
//     effect and the assertion relied on undefined behaviour.
//   - @BeforeAll on a non-static method without @TestInstance(PER_CLASS) is
//     itself invalid in JUnit 5.
//
// The idiomatic Go translation replaces all of this with table-driven unit
// tests that exercise the concrete service (via NewUserService) against a
// fake, in-memory UserRepository. This gives us deterministic control over
// repository behaviour without a mocking framework.
//
// Per the migration change-set the tests also cover:
//   - GetByName happy-path (the original test's intent).
//   - GetByID not-found, which must surface the ErrUserNotFound sentinel
//     (Change 13).
//   - Delete of a missing id, which must surface ErrEmptyResultDelete
//     (Change 12).

import (
	"context"
	"errors"
	"testing"

	resterr "github.com/smartcontact/internal/smartcontact/error/restresponseentityexceptionhandling"
	"github.com/smartcontact/internal/smartcontact/model"
	"github.com/smartcontact/internal/smartcontact/repository"
)

// fakeUserRepository is an in-memory implementation of repository.UserRepository
// used to drive the service under test without a real database.
//
// MIGRATION_NOTE: This replaces Mockito stubbing. Instead of recording
// expectations, we return canned data / errors directly from a fake that
// satisfies the same interface the production service depends on.
type fakeUserRepository struct {
	saveFn       func(ctx context.Context, u model.User) (model.User, error)
	findByIDFn   func(ctx context.Context, id int64) (model.User, error)
	findAllFn    func(ctx context.Context) ([]model.User, error)
	deleteByIDFn func(ctx context.Context, id int64) error
	findByNameFn func(ctx context.Context, name string) (model.User, error)
}

func (f *fakeUserRepository) Save(ctx context.Context, u model.User) (model.User, error) {
	if f.saveFn != nil {
		return f.saveFn(ctx, u)
	}
	return model.User{}, errors.New("Save not implemented")
}

func (f *fakeUserRepository) FindByID(ctx context.Context, id int64) (model.User, error) {
	if f.findByIDFn != nil {
		return f.findByIDFn(ctx, id)
	}
	return model.User{}, errors.New("FindByID not implemented")
}

func (f *fakeUserRepository) FindAll(ctx context.Context) ([]model.User, error) {
	if f.findAllFn != nil {
		return f.findAllFn(ctx)
	}
	return nil, errors.New("FindAll not implemented")
}

func (f *fakeUserRepository) DeleteByID(ctx context.Context, id int64) error {
	if f.deleteByIDFn != nil {
		return f.deleteByIDFn(ctx, id)
	}
	return errors.New("DeleteByID not implemented")
}

func (f *fakeUserRepository) FindByName(ctx context.Context, name string) (model.User, error) {
	if f.findByNameFn != nil {
		return f.findByNameFn(ctx, name)
	}
	return model.User{}, errors.New("FindByName not implemented")
}

// TestUserService_GetByName mirrors the original test's intent: a valid name
// should return the matching user. It is table-driven to also cover the
// not-found path.
func TestUserService_GetByName(t *testing.T) {
	hemraj := model.User{
		ID:       3,
		Name:     "hemraj",
		Email:    "hemrajmalhi1234@gmail.com",
		About:    "Sr",
		Password: "root",
		Role:     "java developer",
	}

	tests := []struct {
		name       string
		input      string
		repo       *fakeUserRepository
		wantName   string
		wantErr    bool
		wantNotFnd bool
	}{
		{
			name:  "valid name returns user",
			input: "hemraj",
			repo: &fakeUserRepository{
				findByNameFn: func(ctx context.Context, n string) (model.User, error) {
					if n == "hemraj" {
						return hemraj, nil
					}
					return model.User{}, resterr.ErrUserNotFound
				},
			},
			wantName: "hemraj",
		},
		{
			name:  "unknown name returns ErrUserNotFound",
			input: "nobody",
			repo: &fakeUserRepository{
				findByNameFn: func(ctx context.Context, n string) (model.User, error) {
					return model.User{}, resterr.ErrUserNotFound
				},
			},
			wantErr:    true,
			wantNotFnd: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			svc := NewUserService(tt.repo)
			got, err := svc.GetByName(context.Background(), tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("GetByName(%q) expected error, got nil", tt.input)
				}
				if tt.wantNotFnd && !errors.Is(err, resterr.ErrUserNotFound) {
					t.Fatalf("GetByName(%q) expected ErrUserNotFound, got %v", tt.input, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("GetByName(%q) unexpected error: %v", tt.input, err)
			}
			if got.Name != tt.wantName {
				t.Errorf("GetByName(%q) name = %q, want %q", tt.input, got.Name, tt.wantName)
			}
		})
	}
}

// TestUserService_GetByID_NotFound verifies that fetching a non-existent id
// surfaces the ErrUserNotFound sentinel (Change 13).
func TestUserService_GetByID_NotFound(t *testing.T) {
	repo := &fakeUserRepository{
		findByIDFn: func(ctx context.Context, id int64) (model.User, error) {
			return model.User{}, resterr.ErrUserNotFound
		},
	}

	svc := NewUserService(repo)

	_, err := svc.GetByID(context.Background(), 999)
	if err == nil {
		t.Fatalf("GetByID expected error for missing id, got nil")
	}
	if !errors.Is(err, resterr.ErrUserNotFound) {
		t.Fatalf("GetByID expected ErrUserNotFound, got %v", err)
	}
}

// TestUserService_Delete_MissingID verifies that deleting a non-existent id
// surfaces the ErrEmptyResultDelete sentinel (Change 12).
func TestUserService_Delete_MissingID(t *testing.T) {
	repo := &fakeUserRepository{
		deleteByIDFn: func(ctx context.Context, id int64) error {
			return repository.ErrEmptyResultDelete
		},
	}

	svc := NewUserService(repo)

	err := svc.Delete(context.Background(), 999)
	if err == nil {
		t.Fatalf("Delete expected error for missing id, got nil")
	}
	if !errors.Is(err, repository.ErrEmptyResultDelete) {
		t.Fatalf("Delete expected ErrEmptyResultDelete, got %v", err)
	}
}
