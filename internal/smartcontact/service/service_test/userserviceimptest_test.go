```go
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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

// hemrajFixture returns the canonical test user fixture that mirrors the Java
// @BeforeAll User.builder() setup.
func hemrajFixture() *model.User {
	return &model.User{
		ID:       3,
		Name:     "hemraj",
		Email:    "hemrajmalhi1234@gmail.com",
		About:    "Sr",
		Password: "root",
		Role:     "java developer",
	}
}

// ---------------------------------------------------------------------------
// GetUserNameByName — table-driven tests
// ---------------------------------------------------------------------------

func TestGetUserNameByName(t *testing.T) {
	fixture := hemrajFixture()

	tests := []struct {
		name string // subtest name

		// repository behaviour
		repoUser  *model.User
		repoErr   error
		inputName string

		// expected outcomes
		wantUser    *model.User
		wantErr     bool
		wantNilUser bool

		// fine-grained field assertions (non-nil user path)
		wantName     string
		wantEmail    string
		wantAbout    string
		wantPassword string
		wantRole     string
		wantID       int64
	}{
		{
			// spec: "given a valid existing user name — returns a User object whose
			// name matches the requested name"
			name:         "given_valid_existing_name_returns_matching_user",
			repoUser:     fixture,
			repoErr:      nil,
			inputName:    "hemraj",
			wantUser:     fixture,
			wantErr:      false,
			wantNilUser:  false,
			wantName:     "hemraj",
			wantEmail:    "hemrajmalhi1234@gmail.com",
			wantAbout:    "Sr",
			wantPassword: "root",
			wantRole:     "java developer",
			wantID:       3,
		},
		{
			// spec: "given a name that maps to a stubbed user 'hemraj' — returns a
			// User with name 'hemraj', email 'hemrajmalhi1234@gmail.com', about 'Sr',
			// password 'root', role 'java developer', id 3"
			name:         "given_hemraj_name_returns_full_fixture_fields",
			repoUser:     fixture,
			repoErr:      nil,
			inputName:    "hemraj",
			wantUser:     fixture,
			wantErr:      false,
			wantNilUser:  false,
			wantName:     "hemraj",
			wantEmail:    "hemrajmalhi1234@gmail.com",
			wantAbout:    "Sr",
			wantPassword: "root",
			wantRole:     "java developer",
			wantID:       3,
		},
		{
			// spec: "given a name that does not exist — returns null or no matching
			// user (implementation-defined); error case: may return null when no
			// user is found"
			name:        "given_missing_name_returns_nil_user_no_error",
			repoUser:    nil,
			repoErr:     sql.ErrNoRows,
			inputName:   "does-not-exist",
			wantErr:     false,
			wantNilUser: true,
		},
		{
			// additional safety: a completely different name must not return the
			// hemraj fixture — read-only / non-mutating invariant check.
			name:        "given_different_name_repository_returns_no_rows",
			repoUser:    nil,
			repoErr:     sql.ErrNoRows,
			inputName:   "alice",
			wantErr:     false,
			wantNilUser: true,
		},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeUserRepository{
				findByNameFn: func(_ context.Context, n string) (*model.User, error) {
					if n == tc.inputName && tc.repoUser != nil {
						return tc.repoUser, tc.repoErr
					}
					return nil, sql.ErrNoRows
				},
			}

			svc := service.NewUserService(repo)
			got, err := svc.GetUserNameByName(context.Background(), tc.inputName)

			if tc.wantErr {
				assert.Error(t, err, "expected an error")
			} else {
				assert.NoError(t, err, "expected no error")
			}

			if tc.wantNilUser {
				assert.Nil(t, got, "expected nil user for missing name")
				return
			}

			require.NotNil(t, got, "expected non-nil user")

			// invariant: returned user's name equals the input name
			assert.Equal(t, tc.wantName, got.Name,
				"User.Name should equal the requested name")

			// full field assertions (spec: hemraj fixture)
			assert.Equal(t, tc.wantEmail, got.Email, "User.Email mismatch")
			assert.Equal(t, tc.wantAbout, got.About, "User.About mismatch")
			assert.Equal(t, tc.wantPassword, got.Password, "User.Password mismatch")
			assert.Equal(t, tc.wantRole, got.Role, "User.Role mismatch")
			assert.Equal(t, tc.wantID, got.ID, "User.ID mismatch")
		})
	}
}

// ---------------------------------------------------------------------------
// DeleteUser — table-driven tests
// ---------------------------------------------------------------------------

func TestDeleteUser(t *testing.T) {
	tests := []struct {
		name       string
		deleteFn   func(ctx context.Context, id int64) error
		inputID    int64
		wantErr    bool
		errChecker func(t *testing.T, err error)
	}{
		{
			name: "given_existing_user_delete_succeeds",
			deleteFn: func(_ context.Context, _ int64) error {
				return nil
			},
			inputID: 3,
			wantErr: false,
		},
		{
			name: "given_missing_user_delete_returns_error",
			deleteFn: func(_ context.Context, _ int64) error {
				return repository.ErrUserNotFound
			},
			inputID: 42,
			wantErr: true,
			errChecker: func(t *testing.T, err error) {
				t.Helper()
				assert.ErrorIs(t, err, repository.ErrUserNotFound,
					"expected ErrUserNotFound sentinel")
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeUserRepository{
				deleteFn: tc.deleteFn,
			}

			svc := service.NewUserService(repo)
			err := svc.DeleteUser(context.Background(), tc.inputID)

			if tc.wantErr {
				require.Error(t, err, "expected an error from DeleteUser")
				if tc.errChecker != nil {
					tc.errChecker(t, err)
				}
			} else {
				assert.NoError(t, err, "expected no error from DeleteUser")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// FetchUserById — table-driven tests
// ---------------------------------------------------------------------------

func TestFetchUserById(t *testing.T) {
	fixture := hemrajFixture()

	tests := []struct {
		name       string
		findByIDFn func(ctx context.Context, id int64) (*model.User, error)
		inputID    int64
		wantErr    bool
		wantUser   *model.User
		errChecker func(t *testing.T, err error)
	}{
		{
			name: "given_existing_id_returns_user",
			findByIDFn: func(_ context.Context, id int64) (*model.User, error) {
				if id == fixture.ID {
					return fixture, nil
				}
				return nil, sql.ErrNoRows
			},
			inputID:  fixture.ID,
			wantErr:  false,
			wantUser: fixture,
		},
		{
			// spec: FetchUserById with missing id must surface as
			// *apperr.UserNotFoundError
			name: "given_missing_id_returns_UserNotFoundError",
			findByIDFn: func(_ context.Context, _ int64) (*model.User, error) {
				return nil, sql.ErrNoRows
			},
			inputID: 99,
			wantErr: true,
			errChecker: func(t *testing.T, err error) {
				t.Helper()
				var notFound *apperr.UserNotFoundError
				assert.True(t, errors.As(err, &notFound),
					"error should be *apperr.UserNotFoundError, got %T: %v", err, err)
			},
		},
		{
			// repository returns an unexpected internal error — service should
			// propagate it (or wrap it); either way, an error is expected.
			name: "given_repository_internal_error_returns_error",
			findByIDFn: func(_ context.Context, _ int64) (*model.User, error) {
				return nil, errors.New("db connection lost")
			},
			inputID: 7,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeUserRepository{
				findByIDFn: tc.findByIDFn,
			}

			svc := service.NewUserService(repo)
			got, err := svc.FetchUserById(context.Background(), tc.inputID)

			if tc.wantErr {
				require.Error(t, err, "expected an error from FetchUserById")
				if tc.errChecker != nil {
					tc.errChecker(t, err)
				}
				assert.Nil(t, got, "expected nil user on error path")
				return
			}

			require.NoError(t, err, "expected no error from FetchUserById")
			require.NotNil(t, got, "expected non-nil user")
			assert.Equal(t, tc.wantUser.ID, got.ID, "User.ID mismatch")
			assert.Equal(t, tc.wantUser.Name, got.Name, "User.Name mismatch")
			assert.Equal(t, tc.wantUser.Email, got.Email, "User.Email mismatch")
		})
	}
}

// ---------------------------------------------------------------------------
// GetUserNameByName — read-only / non-mutating invariant
// ---------------------------------------------------------------------------

// TestGetUserNameByName_IsNonMutating verifies that calling GetUserNameByName
// does not trigger Save or Delete on the repository (global invariant:
// "Retrieving a user is a non-mutating operation").
func TestGetUserNameByName_IsNonMutating(t *testing.T) {
	fixture := hemrajFixture()
	saveCalled := false
	deleteCalled := false

	repo := &fakeUserRepository{
		findByNameFn: func(_ context.Context, n string) (*model.User, error) {
			if n == fixture.Name {
				return fixture, nil
			}
			return nil, sql.ErrNoRows
		},
		saveFn: func(_ context.Context, u *model.User) (*model.User, error) {
			saveCalled = true
			return u, nil
		},
		deleteFn: func(_ context.Context, _ int64) error {
			deleteCalled = true
			return nil
		},
	}

	svc := service.NewUserService(repo)
	_, err := svc.GetUserNameByName(context.Background(), fixture.Name)
	require.NoError(t, err)

	assert.False(t, saveCalled, "Save must not be called by GetUserNameByName")
	assert.False(t, deleteCalled, "Delete must not be called by GetUserNameByName")
}

// ---------------------------------------------------------------------------
// GetUserNameByName — User fields invariant
// ---------------------------------------------------------------------------

// TestGetUserNameByName_UserFieldsExposed verifies that a returned User always
// exposes name, email, about, password, role, and id fields (global invariant).
func TestGetUserNameBy