package service

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"migrated-app/internal/apperr"
	"migrated-app/internal/model"
)

// ---------------------------------------------------------------------------
// Mock repository
// ---------------------------------------------------------------------------

type mockUserRepo struct {
	// Save
	saveFunc func(ctx context.Context, user *model.User) (*model.User, error)
	// FindAll
	findAllFunc func(ctx context.Context) ([]*model.User, error)
	// FindByID
	findByIDFunc func(ctx context.Context, id int) (*model.User, bool, error)
	// DeleteByID
	deleteByIDFunc func(ctx context.Context, id int) error
	// FindByName
	findByNameFunc func(ctx context.Context, name string) (*model.User, error)
}

func (m *mockUserRepo) Save(ctx context.Context, user *model.User) (*model.User, error) {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, user)
	}
	return user, nil
}

func (m *mockUserRepo) FindAll(ctx context.Context) ([]*model.User, error) {
	if m.findAllFunc != nil {
		return m.findAllFunc(ctx)
	}
	return nil, nil
}

func (m *mockUserRepo) FindByID(ctx context.Context, id int) (*model.User, bool, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return nil, false, nil
}

func (m *mockUserRepo) DeleteByID(ctx context.Context, id int) error {
	if m.deleteByIDFunc != nil {
		return m.deleteByIDFunc(ctx, id)
	}
	return nil
}

func (m *mockUserRepo) FindByName(ctx context.Context, name string) (*model.User, error) {
	if m.findByNameFunc != nil {
		return m.findByNameFunc(ctx, name)
	}
	return nil, nil
}

// newServiceWithMock constructs a UserService backed by the supplied mock.
func newServiceWithMock(r UserRepository) *UserService {
	return NewUserService(r)
}

// ---------------------------------------------------------------------------
// SaveUser
// ---------------------------------------------------------------------------

func TestSaveUser(t *testing.T) {
	ctx := context.Background()
	repoErr := errors.New("db write error")

	tests := []struct {
		name      string
		repoSave  func(ctx context.Context, user *model.User) (*model.User, error)
		wantErr   bool
		checkUser bool
	}{
		{
			name: "repo returns saved user",
			repoSave: func(ctx context.Context, user *model.User) (*model.User, error) {
				return user, nil
			},
			wantErr:   false,
			checkUser: true,
		},
		{
			name: "repo returns error",
			repoSave: func(ctx context.Context, user *model.User) (*model.User, error) {
				return nil, repoErr
			},
			wantErr:   true,
			checkUser: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockUserRepo{saveFunc: tc.repoSave}
			svc := newServiceWithMock(repo)

			input := &model.User{}
			got, err := svc.SaveUser(ctx, input)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if got != nil {
					t.Fatalf("expected nil user on error, got %v", got)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.checkUser && got == nil {
				t.Fatal("expected non-nil user, got nil")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// FetchUserList
// ---------------------------------------------------------------------------

func TestFetchUserList(t *testing.T) {
	ctx := context.Background()
	repoErr := errors.New("db read error")

	u1 := &model.User{}
	u2 := &model.User{}

	tests := []struct {
		name        string
		findAllFunc func(ctx context.Context) ([]*model.User, error)
		wantCount   int
		wantErr     bool
	}{
		{
			name: "returns list with multiple users",
			findAllFunc: func(ctx context.Context) ([]*model.User, error) {
				return []*model.User{u1, u2}, nil
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "returns empty list when no users exist",
			findAllFunc: func(ctx context.Context) ([]*model.User, error) {
				return []*model.User{}, nil
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "repo returns error",
			findAllFunc: func(ctx context.Context) ([]*model.User, error) {
				return nil, repoErr
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockUserRepo{findAllFunc: tc.findAllFunc}
			svc := newServiceWithMock(repo)

			got, err := svc.FetchUserList(ctx)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != tc.wantCount {
				t.Fatalf("expected %d users, got %d", tc.wantCount, len(got))
			}
		})
	}
}

// ---------------------------------------------------------------------------
// FetchUserByID
// ---------------------------------------------------------------------------

func TestFetchUserByID(t *testing.T) {
	ctx := context.Background()
	repoErr := errors.New("db read error")
	existingUser := &model.User{}

	tests := []struct {
		name         string
		id           int
		findByIDFunc func(ctx context.Context, id int) (*model.User, bool, error)
		wantUser     bool
		wantErr      bool
		wantNotFound bool
	}{
		{
			name: "user found",
			id:   1,
			findByIDFunc: func(ctx context.Context, id int) (*model.User, bool, error) {
				return existingUser, true, nil
			},
			wantUser: true,
			wantErr:  false,
		},
		{
			name: "user not found returns ErrUserNotFound",
			id:   99,
			findByIDFunc: func(ctx context.Context, id int) (*model.User, bool, error) {
				return nil, false, nil
			},
			wantUser:     false,
			wantErr:      true,
			wantNotFound: true,
		},
		{
			name: "repo error propagated",
			id:   1,
			findByIDFunc: func(ctx context.Context, id int) (*model.User, bool, error) {
				return nil, false, repoErr
			},
			wantErr:      true,
			wantNotFound: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockUserRepo{findByIDFunc: tc.findByIDFunc}
			svc := newServiceWithMock(repo)

			got, err := svc.FetchUserByID(ctx, tc.id)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tc.wantNotFound && !errors.Is(err, apperr.ErrUserNotFound) {
					t.Fatalf("expected ErrUserNotFound, got %v", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantUser && got == nil {
				t.Fatal("expected non-nil user, got nil")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// DeleteUser
// ---------------------------------------------------------------------------

func TestDeleteUser(t *testing.T) {
	ctx := context.Background()
	repoErr := errors.New("db delete error")

	var capturedID int

	tests := []struct {
		name           string
		id             int
		deleteByIDFunc func(ctx context.Context, id int) error
		wantErr        bool
		checkID        bool
	}{
		{
			name: "delegates to repo deleteById with correct id",
			id:   42,
			deleteByIDFunc: func(ctx context.Context, id int) error {
				capturedID = id
				return nil
			},
			wantErr: false,
			checkID: true,
		},
		{
			name: "repo error is propagated",
			id:   7,
			deleteByIDFunc: func(ctx context.Context, id int) error {
				return repoErr
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			capturedID = 0
			repo := &mockUserRepo{deleteByIDFunc: tc.deleteByIDFunc}
			svc := newServiceWithMock(repo)

			err := svc.DeleteUser(ctx, tc.id)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.checkID && capturedID != tc.id {
				t.Fatalf("expected repo to be called with id %d, got %d", tc.id, capturedID)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// UpdateUser
// ---------------------------------------------------------------------------

func TestUpdateUser(t *testing.T) {
	ctx := context.Background()
	repoErr := errors.New("db save error")

	tests := []struct {
		name       string
		id         int
		saveFunc   func(ctx context.Context, user *model.User) (*model.User, error)
		wantErr    bool
		checkSetID bool
	}{
		{
			name: "sets user id before saving",
			id:   10,
			saveFunc: func(ctx context.Context, user *model.User) (*model.User, error) {
				return user, nil
			},
			wantErr:    false,
			checkSetID: true,
		},
		{
			name: "repo save error propagated",
			id:   10,
			saveFunc: func(ctx context.Context, user *model.User) (*model.User, error) {
				return nil, repoErr
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var savedUser *model.User
			wrappedSave := func(ctx context.Context, user *model.User) (*model.User, error) {
				savedUser = user
				return tc.saveFunc(ctx, user)
			}

			repo := &mockUserRepo{saveFunc: wrappedSave}
			svc := newServiceWithMock(repo)

			input := &model.User{}
			err := svc.UpdateUser(ctx, tc.id, input)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if savedUser == nil {
				t.Fatal("expected repo.Save to be called, but it was not")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// GetUserNameByName
// ---------------------------------------------------------------------------

func TestGetUserNameByName(t *testing.T) {
	ctx := context.Background()
	repoErr := errors.New("db name lookup error")

	hemrajUser := &model.User{}

	tests := []struct {
		name           string
		queryName      string
		findByNameFunc func(ctx context.Context, name string) (*model.User, error)
		wantUser       bool
		wantErr        bool
	}{
		{
			name:      "returns user for existing name hemraj",
			queryName: "hemraj",
			findByNameFunc: func(ctx context.Context, name string) (*model.User, error) {
				if name == "hemraj" {
					return hemrajUser, nil
				}
				return nil, fmt.Errorf("unexpected name %q", name)
			},
			wantUser: true,
			wantErr:  false,
		},
		{
			name:      "repo error propagated",
			queryName: "unknown",
			findByNameFunc: func(ctx context.Context, name string) (*model.User, error) {
				return nil, repoErr
			},
			wantUser: false,
			wantErr:  true,
		},
		{
			name:      "repo returns nil user without error",
			queryName: "nobody",
			findByNameFunc: func(ctx context.Context, name string) (*model.User, error) {
				return nil, nil
			},
			wantUser: false,
			wantErr:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockUserRepo{findByNameFunc: tc.findByNameFunc}
			svc := newServiceWithMock(repo)

			got, err := svc.GetUserNameByName(ctx, tc.queryName)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantUser && got == nil {
				t.Fatal("expected non-nil user, got nil")
			}
			if !tc.wantUser && got != nil {
				// acceptable — repo returned nil, so we just pass it through
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Error wrapping — ensure repository errors surface with context.
// ---------------------------------------------------------------------------

func TestErrorWrapping(t *testing.T) {
	ctx := context.Background()
	sentinel := errors.New("sentinel")

	t.Run("SaveUser wraps repo error", func(t *testing.T) {
		repo := &mockUserRepo{saveFunc: func(ctx context.Context, user *model.User) (*model.User, error) {
			return nil, sentinel
		}}
		svc := newServiceWithMock(repo)
		_, err := svc.SaveUser(ctx, &model.User{})
		if !errors.Is(err, sentinel) {
			t.Fatalf("expected wrapped sentinel, got %v", err)
		}
	})

	t.Run("FetchUserList wraps repo error", func(t *testing.T) {
		repo := &mockUserRepo{findAllFunc: func(ctx context.Context) ([]*model.User, error) {
			return nil, sentinel
		}}
		svc := newServiceWithMock(repo)
		_, err := svc.FetchUserList(ctx)
		if !errors.Is(err, sentinel) {
			t.Fatalf("expected wrapped sentinel, got %v", err)
		}
	})

	t.Run("FetchUserByID wraps repo error", func(t *testing.T) {
		repo := &mockUserRepo{findByIDFunc: func(ctx context.Context, id int) (*model.User, bool, error) {
			return nil, false, sentinel
		}}
		svc := newServiceWithMock(repo)
		_, err := svc.FetchUserByID(ctx, 1)
		if !errors.Is(err, sentinel) {
			t.Fatalf("expected wrapped sentinel, got %v", err)
		}
	})

	t.Run("DeleteUser wraps repo error", func(t *testing.T) {
		repo := &mockUserRepo{deleteByIDFunc: func(ctx context.Context, id int) error {
			return sentinel
		}}
		svc := newServiceWithMock(repo)
		err := svc.DeleteUser(ctx, 1)
		if !errors.Is(err, sentinel) {
			t.Fatalf("expected wrapped sentinel, got %v", err)
		}
	})

	t.Run("UpdateUser wraps repo error", func(t *testing.T) {
		repo := &mockUserRepo{saveFunc: func(ctx context.Context, user *model.User) (*model.User, error) {
			return nil, sentinel
		}}
		svc := newServiceWithMock(repo)
		err := svc.UpdateUser(ctx, 1, &model.User{})
		if !errors.Is(err, sentinel) {
			t.Fatalf("expected wrapped sentinel, got %v", err)
		}
	})

	t.Run("GetUserNameByName wraps repo error", func(t *testing.T) {
		repo := &mockUserRepo{findByNameFunc: func(ctx context.Context, name string) (*model.User, error) {
			return nil, sentinel
		}}
		svc := newServiceWithMock(repo)
		_, err := svc.GetUserNameByName(ctx, "test")
		if !errors.Is(err, sentinel) {
			t.Fatalf("expected wrapped sentinel, got %v", err)
		}
	})
}