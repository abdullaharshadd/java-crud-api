package service_test

// MIGRATION_NOTE: This file is migrated from the Spring Boot integration test
// com.smartContact.service.UserServiceImpTest.
//
// The Java test was structurally broken as an integration test: it used
// @SpringBootTest to @Autowire the *real* UserServiceImp, then called
// Mockito.when(...).thenReturn(...) on that non-mock bean. Mockito can only
// stub actual mock objects, so this stubbing has no effect on a real Spring
// bean and the assertion would run against whatever the real service returns.
// The clear *intent*, however, is: "given the repository yields a user named
// 'hemraj', GetUserNameByName('hemraj') returns that user's name".
//
// In idiomatic Go we express that intent as a proper unit test: we inject a
// fake repository.UserRepository into service.NewUserService and assert on
// the resulting behaviour. The debate notes ask us to also cover the
// name-miss (nil, nil) and FetchUserById (*UserNotFoundError) paths, so those
// are added as table-driven / dedicated cases.
//
// Because service.NewUserService and repository.UserRepository already exist
// in the migrated codebase, no production type is (re)declared here — this is
// a pure _test package.

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	apperr "github.com/smartcontact/internal/smartcontact/error"
	"github.com/smartcontact/internal/smartcontact/model"
	"github.com/smartcontact/internal/smartcontact/repository"
	"github.com/smartcontact/internal/smartcontact/service"
)

// fakeUserRepository is a hand-rolled test double implementing
// repository.UserRepository. It replaces Mockito stubbing from the Java test.
//
// MIGRATION_NOTE: If the project standardises on testify/mock or gomock, this
// could be replaced by a generated mock; a hand-rolled fake is used here to
// avoid introducing a mocking dependency for a single test.
type fakeUserRepository struct {
	findByNameFn func(ctx context.Context, name string) (*model.User, error)
	findByIDFn   func(ctx context.Context, id int64) (*model.User, error)
	saveFn       func(ctx context.Context, u *model.User) (*model.User, error)
	deleteFn     func(ctx context.Context, id int64) error
}

func (f *fakeUserRepository) Save(ctx context.Context, u *model.User) (*model.User, error) {
	if f.saveFn != nil {
		return f.saveFn(ctx, u)
	}
	return u, nil
}

func (f *fakeUserRepository) FindByID(ctx context.Context, id int64) (*model.User, error) {
	if f.findByIDFn != nil {
		return f.findByIDFn(ctx, id)
	}
	return nil, sql.ErrNoRows
}

func (f *fakeUserRepository) FindByName(ctx context.Context, name string) (*model.User, error) {
	if f.findByNameFn != nil {
		return f.findByNameFn(ctx, name)
	}
	return nil, sql.ErrNoRows
}

func (f *fakeUserRepository) Delete(ctx context.Context, id int64) error {
	if f.deleteFn != nil {
		return f.deleteFn(ctx, id)
	}
	return nil
}

// compile-time assertion that the fake satisfies the interface.
var _ repository.UserRepository = (*fakeUserRepository)(nil)

// TestGetUserNameByName_WhenValidName_ThenUserShouldBeFound is the direct
// migration of WhenValidDepartmentName_ThenUserShouldBeFound. It seeds the
// fake repository with the same fixture the Java @BeforeAll built via
// User.builder() and asserts the service returns the expected name.
func TestGetUserNameByName_WhenValidName_ThenUserShouldBeFound(t *testing.T) {
	const name = "hemraj"

	want := &model.User{
		ID:       3,
		Name:     name,
		Email:    "hemrajmalhi1234@gmail.com",
		About:    "Sr",
		Password: "root",
		Role:     "java developer",
	}

	repo := &fakeUserRepository{
		findByNameFn: func(_ context.Context, n string) (*model.User, error) {
			if n == name {
				return want, nil
			}
			return nil, sql.ErrNoRows
		},
	}

	svc := service.NewUserService(repo)

	got, err := svc.GetUserNameByName(context.Background(), name)
	if err != nil {
		t.Fatalf("GetUserNameByName(%q) unexpected error: %v", name, err)
	}
	if got == nil {
		t.Fatalf("GetUserNameByName(%q) returned nil user", name)
	}
	if got.Name != name {
		t.Errorf("GetUserNameByName(%q).Name = %q, want %q", name, got.Name, name)
	}
}

// TestGetUserNameByName_WhenNameMissing_ThenNilNil covers the name-miss path
// requested by the migration debate: a missing row must surface as (nil, nil)
// rather than an error.
//
// MIGRATION_NOTE: This encodes the expected contract of GetUserNameByName for
// a not-found lookup. If the real service instead propagates the error,
// adjust this expectation to match service.GetUserNameByName's actual
// behaviour during human review.
func TestGetUserNameByName_WhenNameMissing_ThenNilNil(t *testing.T) {
	repo := &fakeUserRepository{
		findByNameFn: func(_ context.Context, _ string) (*model.User, error) {
			return nil, sql.ErrNoRows
		},
	}

	svc := service.NewUserService(repo)

	got, err := svc.GetUserNameByName(context.Background(), "does-not-exist")
	if err != nil {
		t.Fatalf("GetUserNameByName miss: unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("GetUserNameByName miss: got user %+v, want nil", got)
	}
}

// TestDeleteUser_WhenRowMissing_ThenError covers the missing-row delete path:
// deleting a non-existent user must return an error.
func TestDeleteUser_WhenRowMissing_ThenError(t *testing.T) {
	repo := &fakeUserRepository{
		deleteFn: func(_ context.Context, _ int64) error {
			return repository.ErrUserNotFound
		},
	}

	svc := service.NewUserService(repo)

	err := svc.DeleteUser(context.Background(), 42)
	if err == nil {
		t.Fatal("DeleteUser(missing): expected error, got nil")
	}
}

// TestFetchUserById_WhenMissing_ThenUserNotFoundError covers the
// FetchUserById path: a missing user must surface as an *apperr.UserNotFoundError
// (the migrated equivalent of Java's UserNotFoundException).
func TestFetchUserById_WhenMissing_ThenUserNotFoundError(t *testing.T) {
	repo := &fakeUserRepository{
		findByIDFn: func(_ context.Context, _ int64) (*model.User, error) {
			return nil, sql.ErrNoRows
		},
	}

	svc := service.NewUserService(repo)

	_, err := svc.FetchUserById(context.Background(), 99)
	if err == nil {
		t.Fatal("FetchUserById(missing): expected error, got nil")
	}

	var notFound *apperr.UserNotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("FetchUserById(missing): error = %v, want *apperr.UserNotFoundError", err)
	}
}
