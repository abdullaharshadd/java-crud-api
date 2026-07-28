package service

// MIGRATION_NOTE: The Java source UserServiceImpTest.java was a broken
// @SpringBootTest integration test. It attempted to stub the real,
// @Autowired UserServiceImp with Mockito.when(...) — but you cannot stub a
// concrete Spring bean that way (only Mockito mocks/spies), so the setup
// never took effect and the test was functionally dead. It also used
// @BeforeAll on a non-static method without @TestInstance(PER_CLASS), which
// would have failed at runtime.
//
// Rather than faithfully reproduce a broken test, this file migrates the
// clear INTENT: verify that UserService.GetUserByName returns the user whose
// name matches the requested name. Idiomatically in Go we do this as a fast
// unit test with a stub UserDao (satisfying repository.UserDao) injected via
// NewUserService — no live Spring context, no static mocking framework.

import (
	"context"
	"errors"
	"testing"

	"migrated-app/internal/smartcontact/model"
	"migrated-app/internal/smartcontact/repository"
)

// stubUserDao is a hand-written test double implementing repository.UserDao.
// Only the methods exercised by the tests below return meaningful values; the
// rest satisfy the interface and fail loudly if unexpectedly called.
//
// MIGRATION_NOTE: This replaces Mockito static stubbing. In idiomatic Go we
// prefer explicit, compile-checked fakes over reflection-based mocks.
type stubUserDao struct {
	findByNameFn func(ctx context.Context, name string) (*model.User, error)
}

var _ repository.UserDao = (*stubUserDao)(nil)

// FindByName delegates to the configured stub function.
func (s *stubUserDao) FindByName(ctx context.Context, name string) (*model.User, error) {
	if s.findByNameFn != nil {
		return s.findByNameFn(ctx, name)
	}
	return nil, errors.New("FindByName not stubbed")
}

// Save is unused in these tests.
func (s *stubUserDao) Save(ctx context.Context, u *model.User) (*model.User, error) {
	return nil, errors.New("Save not stubbed")
}

// Update is unused in these tests.
func (s *stubUserDao) Update(ctx context.Context, u *model.User) error {
	return errors.New("Update not stubbed")
}

// FindByID is unused in these tests.
func (s *stubUserDao) FindByID(ctx context.Context, id int) (*model.User, error) {
	return nil, errors.New("FindByID not stubbed")
}

// FindAll is unused in these tests.
func (s *stubUserDao) FindAll(ctx context.Context) ([]model.User, error) {
	return nil, errors.New("FindAll not stubbed")
}

// DeleteByID is unused in these tests.
func (s *stubUserDao) DeleteByID(ctx context.Context, id int) error {
	return errors.New("DeleteByID not stubbed")
}

// TestUserService_GetUserByName verifies GetUserByName returns the user whose
// name matches the requested name (the intent of the original
// WhenValidDepartmentName_ThenUserShouldBeFound test).
func TestUserService_GetUserByName(t *testing.T) {
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
		name     string
		input    string
		stub     func(ctx context.Context, name string) (*model.User, error)
		wantName string
		wantErr  bool
	}{
		{
			name:  "valid name returns matching user",
			input: wantName,
			stub: func(ctx context.Context, name string) (*model.User, error) {
				if name != wantName {
					return nil, errors.New("user not found")
				}
				return seed, nil
			},
			wantName: wantName,
			wantErr:  false,
		},
		{
			name:  "unknown name surfaces error",
			input: "nobody",
			stub: func(ctx context.Context, name string) (*model.User, error) {
				return nil, errors.New("user not found")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			dao := &stubUserDao{findByNameFn: tt.stub}
			svc := NewUserService(dao)

			got, err := svc.GetUserByName(context.Background(), tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("GetUserByName(%q): expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("GetUserByName(%q): unexpected error: %v", tt.input, err)
			}
			if got.Name != tt.wantName {
				t.Errorf("GetUserByName(%q): got name %q, want %q", tt.input, got.Name, tt.wantName)
			}
		})
	}
}
