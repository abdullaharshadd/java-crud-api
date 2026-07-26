```go
// Package service_test contains tests for the UserService interface contract.
//
// Because UserService is an interface, we test it through a mock/stub
// implementation that satisfies the contract, verifying that the behavioural
// specs described in the migration notes are enforced at the interface boundary.
package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/smartContact/internal/smartcontact/model"
	"github.com/smartContact/internal/smartcontact/service"
)

// ─── sentinel errors ──────────────────────────────────────────────────────────

// ErrUserNotFound is used by the stub to signal a missing user, mirroring the
// smartcontacterror.ErrUserNotFound sentinel that real implementations must use.
var ErrUserNotFound = errors.New("user not found")

// errPersistence simulates a generic persistence failure.
var errPersistence = errors.New("persistence error")

// ─── stub implementation ──────────────────────────────────────────────────────

// stubUserService is a minimal, in-memory implementation of service.UserService
// that satisfies the interface contract and enforces the behavioural asymmetry
// documented in the migration notes.
type stubUserService struct {
	// store maps user ID → *model.User
	store map[int]*model.User
	// nextID is the auto-increment counter used by SaveUser.
	nextID int

	// Hooks – when non-nil they override normal behaviour so error paths can be
	// exercised without polluting the happy-path store.
	saveErr        error
	fetchListErr   error
	fetchByIDErr   error
	deleteErr      error
	updateErr      error
	getUserNameErr error
}

func newStub() *stubUserService {
	return &stubUserService{
		store:  make(map[int]*model.User),
		nextID: 1,
	}
}

// Ensure stubUserService satisfies the interface at compile time.
var _ service.UserService = (*stubUserService)(nil)

func (s *stubUserService) SaveUser(_ context.Context, user *model.User) (*model.User, error) {
	if s.saveErr != nil {
		return nil, s.saveErr
	}
	saved := *user // shallow copy
	saved.ID = s.nextID
	s.nextID++
	s.store[saved.ID] = &saved
	return &saved, nil
}

func (s *stubUserService) FetchUserList(_ context.Context) ([]*model.User, error) {
	if s.fetchListErr != nil {
		return nil, s.fetchListErr
	}
	list := make([]*model.User, 0, len(s.store))
	for _, u := range s.store {
		list = append(list, u)
	}
	return list, nil
}

func (s *stubUserService) FetchUserByID(_ context.Context, id int) (*model.User, error) {
	if s.fetchByIDErr != nil {
		return nil, s.fetchByIDErr
	}
	u, ok := s.store[id]
	if !ok {
		// MIGRATION_NOTE: must convert absence into ErrUserNotFound.
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (s *stubUserService) DeleteUser(_ context.Context, id int) error {
	if s.deleteErr != nil {
		return s.deleteErr
	}
	delete(s.store, id)
	return nil
}

func (s *stubUserService) UpdateUser(_ context.Context, id int, user *model.User) error {
	if s.updateErr != nil {
		return s.updateErr
	}
	if _, ok := s.store[id]; !ok {
		// contract says behaviour is unspecified; stub silently returns nil.
		return nil
	}
	updated := *user
	updated.ID = id
	s.store[id] = &updated
	return nil
}

func (s *stubUserService) GetUserNameByName(_ context.Context, name string) (*model.User, error) {
	if s.getUserNameErr != nil {
		return nil, s.getUserNameErr
	}
	for _, u := range s.store {
		if u.Name == name {
			return u, nil
		}
	}
	// MIGRATION_NOTE: must NOT return ErrUserNotFound – return (nil, nil).
	return nil, nil
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func newUserWith(id int, name string) *model.User {
	return &model.User{ID: id, Name: name}
}

// ─── SaveUser tests ───────────────────────────────────────────────────────────

func TestSaveUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setup       func(s *stubUserService)
		input       *model.User
		wantErr     bool
		wantNonNil  bool
		wantName    string
		wantIDGTZero bool
	}{
		{
			name:         "valid user is persisted and returned with generated ID",
			input:        &model.User{Name: "Alice"},
			wantErr:      false,
			wantNonNil:   true,
			wantName:     "Alice",
			wantIDGTZero: true,
		},
		{
			name:       "persistence error propagated",
			setup:      func(s *stubUserService) { s.saveErr = errPersistence },
			input:      &model.User{Name: "Bob"},
			wantErr:    true,
			wantNonNil: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svc := newStub()
			if tc.setup != nil {
				tc.setup(svc)
			}

			got, err := svc.SaveUser(context.Background(), tc.input)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				return
			}

			assert.NoError(t, err)
			if tc.wantNonNil {
				assert.NotNil(t, got)
			}
			if tc.wantName != "" {
				assert.Equal(t, tc.wantName, got.Name)
			}
			if tc.wantIDGTZero {
				assert.Greater(t, got.ID, 0, "saved user should have a positive generated ID")
			}
		})
	}
}

// SaveUser invariant: the saved user should be retrievable afterwards.
func TestSaveUser_Retrievable(t *testing.T) {
	t.Parallel()

	svc := newStub()
	input := &model.User{Name: "Carol"}

	saved, err := svc.SaveUser(context.Background(), input)
	assert.NoError(t, err)
	assert.NotNil(t, saved)

	// Verify retrieval
	fetched, err := svc.FetchUserByID(context.Background(), saved.ID)
	assert.NoError(t, err)
	assert.Equal(t, saved.ID, fetched.ID)
	assert.Equal(t, saved.Name, fetched.Name)
}

// ─── FetchUserList tests ──────────────────────────────────────────────────────

func TestFetchUserList(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setup     func(s *stubUserService)
		wantErr   bool
		wantCount int
		wantEmpty bool
	}{
		{
			name: "returns all stored users",
			setup: func(s *stubUserService) {
				_, _ = s.SaveUser(context.Background(), &model.User{Name: "Dave"})
				_, _ = s.SaveUser(context.Background(), &model.User{Name: "Eve"})
			},
			wantErr:   false,
			wantCount: 2,
		},
		{
			name:      "returns empty slice when no users exist",
			wantErr:   false,
			wantEmpty: true,
		},
		{
			name:    "retrieval failure propagated",
			setup:   func(s *stubUserService) { s.fetchListErr = errPersistence },
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svc := newStub()
			if tc.setup != nil {
				tc.setup(svc)
			}

			got, err := svc.FetchUserList(context.Background())

			if tc.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			// Invariant: never nil
			assert.NotNil(t, got, "FetchUserList must never return a nil slice")

			if tc.wantEmpty {
				assert.Empty(t, got)
			}
			if tc.wantCount > 0 {
				assert.Len(t, got, tc.wantCount)
			}
		})
	}
}

// FetchUserList invariant: read-only – list call must not alter stored data.
func TestFetchUserList_ReadOnly(t *testing.T) {
	t.Parallel()

	svc := newStub()
	_, _ = svc.SaveUser(context.Background(), &model.User{Name: "Frank"})

	before := len(svc.store)
	_, _ = svc.FetchUserList(context.Background())
	after := len(svc.store)

	assert.Equal(t, before, after, "FetchUserList must not modify the store")
}

// ─── FetchUserByID tests ──────────────────────────────────────────────────────

func TestFetchUserByID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setup          func(s *stubUserService) int // returns the ID to look up
		overrideID     *int                         // if non-nil, use this ID instead
		wantErr        bool
		wantNotFound   bool
		wantUserName   string
	}{
		{
			name: "existing user returned",
			setup: func(s *stubUserService) int {
				u, _ := s.SaveUser(context.Background(), &model.User{Name: "Grace"})
				return u.ID
			},
			wantErr:      false,
			wantUserName: "Grace",
		},
		{
			name: "absent user causes UserNotFound error",
			setup: func(s *stubUserService) int {
				return 9999 // non-existent
			},
			wantErr:      true,
			wantNotFound: true,
		},
		{
			name: "retrieval infrastructure error propagated",
			setup: func(s *stubUserService) int {
				s.fetchByIDErr = errPersistence
				return 1
			},
			wantErr:      true,
			wantNotFound: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svc := newStub()
			id := tc.setup(svc)

			got, err := svc.FetchUserByID(context.Background(), id)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				if tc.wantNotFound {
					// MIGRATION_NOTE: must satisfy errors.Is(err, ErrUserNotFound)
					assert.True(t, errors.Is(err, ErrUserNotFound),
						"expected ErrUserNotFound, got %v", err)
				}
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, got)
			if tc.wantUserName != "" {
				assert.Equal(t, tc.wantUserName, got.Name)
			}
		})
	}
}

// FetchUserByID invariant: read-only.
func TestFetchUserByID_ReadOnly(t *testing.T) {
	t.Parallel()

	svc := newStub()
	saved, _ := svc.SaveUser(context.Background(), &model.User{Name: "Hank"})

	before := len(svc.store)
	_, _ = svc.FetchUserByID(context.Background(), saved.ID)
	after := len(svc.store)

	assert.Equal(t, before, after, "FetchUserByID must not modify the store")
}

// ─── DeleteUser tests ─────────────────────────────────────────────────────────

func TestDeleteUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setup         func(s *stubUserService) int
		wantErr       bool
		verifyDeleted bool
	}{
		{
			name: "existing user is removed",
			setup: func(s *stubUserService) int {
				u, _ := s.SaveUser(context.Background(), &model.User{Name: "Ivy"})
				return u.ID
			},
			wantErr:       false,
			verifyDeleted: true,
		},
		{
			name: "non-existent id returns no error (unspecified contract)",
			setup: func(s *stubUserService) int {
				return 9999
			},
			wantErr: false,
		},
		{
			name: "persistence error propagated",
			setup: func(s *stubUserService) int {
				s.deleteErr = errPersistence
				return 1
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svc := newStub()
			id := tc.setup(svc)

			err := svc.DeleteUser(context.Background(), id)

			if tc.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			if tc.verifyDeleted {
				// Invariant: after deletion the user is no longer retrievable.
				_, fetchErr := svc.FetchUserByID(context.Background(), id)
				assert.True(t, errors.Is(fetchErr, ErrUserNotFound),
					"deleted user must not be retrievable")
			}
		})
	}
}

// ─── UpdateUser tests ─────────────────────────────────────────────────────────

func TestUpdateUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setup          func(s *stubUserService) int
		updateData     *model.User
		wantErr        bool
		wantUpdatedName string
		verifyUpdate   bool
	}{
		{
			name: "existing user is updated and subsequent read reflects new data",
			setup: func(s *stubUserService) int {
				u, _ := s.SaveUser(context.Background(), &model.User{Name: "Jack"})
				return u.ID
			},
			updateData:      &model.User{Name: "Jack-Updated"},
			wantErr:         false,
			wantUpdatedName: "Jack-Updated",
			verifyUpdate:    true,
		},
		{
			name: "non-existent id handled without error (unspecified contract)",
			setup: func(s *stubUserService) int {
				return 9999
			},
			updateData: &model.User{Name: "Ghost"},
			wantErr:    false,
		},
		{
			name: "persistence error propagated",
			setup: func(s *stubUserService)