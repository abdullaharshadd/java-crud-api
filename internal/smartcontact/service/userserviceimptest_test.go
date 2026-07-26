```go
package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	resterr "github.com/smartcontact/internal/smartcontact/error/restresponseentityexceptionhandling"
	"github.com/smartcontact/internal/smartcontact/model"
	"github.com/smartcontact/internal/smartcontact/repository"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func newFakeRepo() *fakeUserRepository { return &fakeUserRepository{} }

// ---------------------------------------------------------------------------
// TestUserService_GetByName_TableDriven
// Covers the behavioral spec "getUserNameByName" in full.
// ---------------------------------------------------------------------------

func TestUserService_GetByName_TableDriven(t *testing.T) {
	t.Parallel()

	hemraj := model.User{
		ID:       3,
		Name:     "hemraj",
		Email:    "hemrajmalhi1234@gmail.com",
		About:    "Sr",
		Password: "root",
		Role:     "java developer",
	}

	tests := []struct {
		name        string
		inputName   string
		setupRepo   func() *fakeUserRepository
		wantUser    model.User
		wantErr     bool
		wantErrIs   error
		wantNilUser bool
	}{
		{
			// spec: "given a valid existing user name" →
			//       "returns a User object whose name equals the provided name"
			name:      "valid existing name returns correct user",
			inputName: "hemraj",
			setupRepo: func() *fakeUserRepository {
				return &fakeUserRepository{
					findByNameFn: func(_ context.Context, n string) (model.User, error) {
						if n == "hemraj" {
							return hemraj, nil
						}
						return model.User{}, resterr.ErrUserNotFound
					},
				}
			},
			wantUser: hemraj,
			wantErr:  false,
		},
		{
			// invariant: returned user's name must equal the input name
			name:      "returned user name equals input name",
			inputName: "hemraj",
			setupRepo: func() *fakeUserRepository {
				return &fakeUserRepository{
					findByNameFn: func(_ context.Context, n string) (model.User, error) {
						return model.User{ID: 1, Name: n}, nil
					},
				}
			},
			wantUser: model.User{ID: 1, Name: "hemraj"},
			wantErr:  false,
		},
		{
			// spec: "given a name that does not match any user" →
			//       "returns null or no matching user is found"
			// error case: "may return null when no user matches the name"
			name:        "unknown name returns ErrUserNotFound",
			inputName:   "nobody",
			setupRepo: func() *fakeUserRepository {
				return &fakeUserRepository{
					findByNameFn: func(_ context.Context, _ string) (model.User, error) {
						return model.User{}, resterr.ErrUserNotFound
					},
				}
			},
			wantErr:     true,
			wantErrIs:   resterr.ErrUserNotFound,
			wantNilUser: true,
		},
		{
			// edge: empty string name also yields not-found
			name:      "empty name returns ErrUserNotFound",
			inputName: "",
			setupRepo: func() *fakeUserRepository {
				return &fakeUserRepository{
					findByNameFn: func(_ context.Context, _ string) (model.User, error) {
						return model.User{}, resterr.ErrUserNotFound
					},
				}
			},
			wantErr:     true,
			wantErrIs:   resterr.ErrUserNotFound,
			wantNilUser: true,
		},
		{
			// edge: repository returns an unexpected internal error
			name:      "repository internal error is propagated",
			inputName: "hemraj",
			setupRepo: func() *fakeUserRepository {
				return &fakeUserRepository{
					findByNameFn: func(_ context.Context, _ string) (model.User, error) {
						return model.User{}, errors.New("db connection refused")
					},
				}
			},
			wantErr:     true,
			wantNilUser: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := NewUserService(tt.setupRepo())
			got, err := svc.GetByName(context.Background(), tt.inputName)

			if tt.wantErr {
				require.Error(t, err, "expected an error but got nil")

				if tt.wantErrIs != nil {
					assert.True(t,
						errors.Is(err, tt.wantErrIs),
						"expected error %v, got %v", tt.wantErrIs, err,
					)
				}

				if tt.wantNilUser {
					assert.Equal(t, model.User{}, got,
						"expected zero-value user on error path")
				}
				return
			}

			require.NoError(t, err)

			// invariant: returned user's Name must equal the input name
			assert.Equal(t, tt.inputName, got.Name,
				"returned user Name must equal input name")

			// full struct equality when we have a canonical fixture
			if tt.wantUser != (model.User{}) {
				assert.Equal(t, tt.wantUser, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestUserService_GetByID_TableDriven
// Change 13: not-found path must surface ErrUserNotFound.
// ---------------------------------------------------------------------------

func TestUserService_GetByID_TableDriven(t *testing.T) {
	t.Parallel()

	alice := model.User{ID: 1, Name: "alice", Email: "alice@example.com"}

	tests := []struct {
		name      string
		inputID   int64
		setupRepo func() *fakeUserRepository
		wantUser  model.User
		wantErr   bool
		wantErrIs error
	}{
		{
			name:    "existing id returns user",
			inputID: 1,
			setupRepo: func() *fakeUserRepository {
				return &fakeUserRepository{
					findByIDFn: func(_ context.Context, id int64) (model.User, error) {
						if id == 1 {
							return alice, nil
						}
						return model.User{}, resterr.ErrUserNotFound
					},
				}
			},
			wantUser: alice,
		},
		{
			name:    "missing id returns ErrUserNotFound",
			inputID: 999,
			setupRepo: func() *fakeUserRepository {
				return &fakeUserRepository{
					findByIDFn: func(_ context.Context, _ int64) (model.User, error) {
						return model.User{}, resterr.ErrUserNotFound
					},
				}
			},
			wantErr:   true,
			wantErrIs: resterr.ErrUserNotFound,
		},
		{
			name:    "zero id returns ErrUserNotFound",
			inputID: 0,
			setupRepo: func() *fakeUserRepository {
				return &fakeUserRepository{
					findByIDFn: func(_ context.Context, _ int64) (model.User, error) {
						return model.User{}, resterr.ErrUserNotFound
					},
				}
			},
			wantErr:   true,
			wantErrIs: resterr.ErrUserNotFound,
		},
		{
			name:    "repository internal error is propagated",
			inputID: 1,
			setupRepo: func() *fakeUserRepository {
				return &fakeUserRepository{
					findByIDFn: func(_ context.Context, _ int64) (model.User, error) {
						return model.User{}, errors.New("timeout")
					},
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := NewUserService(tt.setupRepo())
			got, err := svc.GetByID(context.Background(), tt.inputID)

			if tt.wantErr {
				require.Error(t, err)
				if tt.wantErrIs != nil {
					assert.True(t, errors.Is(err, tt.wantErrIs),
						"expected %v, got %v", tt.wantErrIs, err)
				}
				assert.Equal(t, model.User{}, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantUser, got)
		})
	}
}

// ---------------------------------------------------------------------------
// TestUserService_Delete_TableDriven
// Change 12: deleting a missing id must surface ErrEmptyResultDelete.
// ---------------------------------------------------------------------------

func TestUserService_Delete_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		inputID   int64
		setupRepo func() *fakeUserRepository
		wantErr   bool
		wantErrIs error
	}{
		{
			name:    "existing id is deleted without error",
			inputID: 1,
			setupRepo: func() *fakeUserRepository {
				return &fakeUserRepository{
					deleteByIDFn: func(_ context.Context, _ int64) error {
						return nil
					},
				}
			},
			wantErr: false,
		},
		{
			name:    "missing id returns ErrEmptyResultDelete",
			inputID: 999,
			setupRepo: func() *fakeUserRepository {
				return &fakeUserRepository{
					deleteByIDFn: func(_ context.Context, _ int64) error {
						return repository.ErrEmptyResultDelete
					},
				}
			},
			wantErr:   true,
			wantErrIs: repository.ErrEmptyResultDelete,
		},
		{
			name:    "zero id returns ErrEmptyResultDelete",
			inputID: 0,
			setupRepo: func() *fakeUserRepository {
				return &fakeUserRepository{
					deleteByIDFn: func(_ context.Context, _ int64) error {
						return repository.ErrEmptyResultDelete
					},
				}
			},
			wantErr:   true,
			wantErrIs: repository.ErrEmptyResultDelete,
		},
		{
			name:    "repository internal error is propagated",
			inputID: 1,
			setupRepo: func() *fakeUserRepository {
				return &fakeUserRepository{
					deleteByIDFn: func(_ context.Context, _ int64) error {
						return errors.New("disk full")
					},
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := NewUserService(tt.setupRepo())
			err := svc.Delete(context.Background(), tt.inputID)

			if tt.wantErr {
				require.Error(t, err)
				if tt.wantErrIs != nil {
					assert.True(t, errors.Is(err, tt.wantErrIs),
						"expected %v, got %v", tt.wantErrIs, err)
				}
				return
			}

			assert.NoError(t, err)
		})
	}
}

// ---------------------------------------------------------------------------
// TestUserService_Save_TableDriven
// ---------------------------------------------------------------------------

func TestUserService_Save_TableDriven(t *testing.T) {
	t.Parallel()

	bob := model.User{ID: 10, Name: "bob", Email: "bob@example.com"}

	tests := []struct {
		name      string
		input     model.User
		setupRepo func() *fakeUserRepository
		wantUser  model.User
		wantErr   bool
	}{
		{
			name:  "save new user returns saved user",
			input: bob,
			setupRepo: func() *fakeUserRepository {
				return &fakeUserRepository{
					saveFn: func(_ context.Context, u model.User) (model.User, error) {
						return u, nil
					},
				}
			},
			wantUser: bob,
		},
		{
			name:  "repository save error is propagated",
			input: bob,
			setupRepo: func() *fakeUserRepository {
				return &fakeUserRepository{
					saveFn: func(_ context.Context, _ model.User) (model.User, error) {
						return model.User{}, errors.New("unique constraint violation")
					},
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := NewUserService(tt.setupRepo())
			got, err := svc.Save(context.Background(), tt.input)

			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, model.User{}, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantUser, got)
		})
	}
}

// ---------------------------------------------------------------------------
// TestUserService_GetAll_TableDriven
// ---------------------------------------------------------------------------

func TestUserService_GetAll_TableDriven(t *testing.T) {
	t.Parallel()

	users := []model.User{
		{ID: 1, Name: "alice"},
		{ID: 2, Name: "bob"},
	}

	tests := []struct {
		name      string
		setupRepo func() *fakeUserRepository
		wantUsers []model.User
		wantErr   bool
	}{
		{
			name: "returns all users",
			setupRepo: func() *fakeUserRepository {
				return &fakeUserRepository{
					findAllFn: func(_ context.Context) ([]model.User, error) {
						return users, nil
					},
				}
			},
			wantUsers: users,
		},
		{
			name: "empty table returns empty slice",
			setupRepo: func() *fakeUserRepository {
				return &fakeUserRepository{
					findAllFn: func(_ context.Context) ([]model.User, error) {
						return []model.User{}, nil
					},
				}
			},
			wantUsers: []model.User{},
		},
		{
			name: "repository error is propagated",
			setupRepo: func() *fakeUserRepository {
				return &fakeUserRepository{
					findAllFn: func(_ context.Context) ([]model.User, error) {
						return nil, errors.New("connection lost")
					},
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := NewUserService(tt.setupRepo())
			got, err := svc.GetAll(context.Background())

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantUsers, got)
		})
	}
}

// ---------------------------------------------------------------------------
// TestUserService_GlobalInvariants
// Exercises the global invariants stated in the spec:
//   - "A retrieved user, when present