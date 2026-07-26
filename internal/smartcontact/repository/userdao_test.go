```go
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	smartErr "github.com/smartContact/internal/smartcontact/error"
	"github.com/smartContact/internal/smartcontact/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newMockDB is a helper that returns a *sql.DB backed by sqlmock and the mock
// controller. Tests call require.NoError on the error to fail fast.
func newMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db, mock
}

// isUserNotFoundError is a small helper that unwraps to *smartErr.UserNotFoundError.
func isUserNotFoundError(err error) bool {
	var target *smartErr.UserNotFoundError
	return errors.As(err, &target)
}

// ---------------------------------------------------------------------------
// Save
// ---------------------------------------------------------------------------

func TestUserDao_Save(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		user      *model.User
		setupMock func(mock sqlmock.Sqlmock, user *model.User)
		wantErr   bool
		errCheck  func(t *testing.T, err error)
		check     func(t *testing.T, got *model.User, input *model.User)
	}{
		{
			name: "insert new user (ID==0) returns user with generated ID",
			user: &model.User{Name: "Alice"},
			setupMock: func(mock sqlmock.Sqlmock, user *model.User) {
				mock.ExpectQuery(`INSERT INTO users \(name\) VALUES \(\$1\) RETURNING id`).
					WithArgs("Alice").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(42))
			},
			wantErr: false,
			check: func(t *testing.T, got *model.User, _ *model.User) {
				assert.Equal(t, 42, got.ID)
				assert.Equal(t, "Alice", got.Name)
			},
		},
		{
			name: "upsert existing user (ID!=0) returns updated user",
			user: &model.User{ID: 7, Name: "Bob"},
			setupMock: func(mock sqlmock.Sqlmock, user *model.User) {
				mock.ExpectExec(`INSERT INTO users \(id, name\) VALUES \(\$1, \$2\)`).
					WithArgs(7, "Bob").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
			check: func(t *testing.T, got *model.User, input *model.User) {
				assert.Equal(t, 7, got.ID)
				assert.Equal(t, "Bob", got.Name)
			},
		},
		{
			name:      "nil user returns error",
			user:      nil,
			setupMock: func(mock sqlmock.Sqlmock, user *model.User) {},
			wantErr:   true,
			errCheck: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "nil user")
			},
		},
		{
			name: "DB error on insert returns wrapped error",
			user: &model.User{Name: "Charlie"},
			setupMock: func(mock sqlmock.Sqlmock, user *model.User) {
				mock.ExpectQuery(`INSERT INTO users \(name\) VALUES \(\$1\) RETURNING id`).
					WithArgs("Charlie").
					WillReturnError(fmt.Errorf("connection refused"))
			},
			wantErr: true,
			errCheck: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "insert user")
			},
		},
		{
			name: "DB error on upsert returns wrapped error",
			user: &model.User{ID: 3, Name: "Dave"},
			setupMock: func(mock sqlmock.Sqlmock, user *model.User) {
				mock.ExpectExec(`INSERT INTO users \(id, name\) VALUES \(\$1, \$2\)`).
					WithArgs(3, "Dave").
					WillReturnError(fmt.Errorf("unique violation"))
			},
			wantErr: true,
			errCheck: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "upsert user")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := newMockDB(t)
			tc.setupMock(mock, tc.user)

			dao := NewUserDao(db)
			got, err := dao.Save(ctx, tc.user)

			if tc.wantErr {
				require.Error(t, err)
				if tc.errCheck != nil {
					tc.errCheck(t, err)
				}
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				if tc.check != nil {
					tc.check(t, got, tc.user)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// FindByID
// ---------------------------------------------------------------------------

func TestUserDao_FindByID(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		id        int
		setupMock func(mock sqlmock.Sqlmock)
		wantErr   bool
		errCheck  func(t *testing.T, err error)
		check     func(t *testing.T, got *model.User)
	}{
		{
			name: "existing ID returns user",
			id:   1,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "Alice")
				mock.ExpectQuery(`SELECT id, name FROM users WHERE id = \$1`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			wantErr: false,
			check: func(t *testing.T, got *model.User) {
				assert.Equal(t, 1, got.ID)
				assert.Equal(t, "Alice", got.Name)
			},
		},
		{
			name: "non-existent ID returns UserNotFoundError",
			id:   999,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, name FROM users WHERE id = \$1`).
					WithArgs(999).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
			errCheck: func(t *testing.T, err error) {
				assert.True(t, isUserNotFoundError(err), "expected UserNotFoundError, got %T: %v", err, err)
				assert.Contains(t, err.Error(), "999")
			},
		},
		{
			name: "DB error returns wrapped error",
			id:   5,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, name FROM users WHERE id = \$1`).
					WithArgs(5).
					WillReturnError(fmt.Errorf("connection lost"))
			},
			wantErr: true,
			errCheck: func(t *testing.T, err error) {
				assert.False(t, isUserNotFoundError(err))
				assert.Contains(t, err.Error(), "find user by id")
			},
		},
		{
			name: "returned user ID matches requested ID",
			id:   2,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name"}).AddRow(2, "Bob")
				mock.ExpectQuery(`SELECT id, name FROM users WHERE id = \$1`).
					WithArgs(2).
					WillReturnRows(rows)
			},
			wantErr: false,
			check: func(t *testing.T, got *model.User) {
				assert.Equal(t, 2, got.ID)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := newMockDB(t)
			tc.setupMock(mock)

			dao := NewUserDao(db)
			got, err := dao.FindByID(ctx, tc.id)

			if tc.wantErr {
				require.Error(t, err)
				if tc.errCheck != nil {
					tc.errCheck(t, err)
				}
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				if tc.check != nil {
					tc.check(t, got)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// FindAll
// ---------------------------------------------------------------------------

func TestUserDao_FindAll(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		setupMock func(mock sqlmock.Sqlmock)
		wantErr   bool
		errCheck  func(t *testing.T, err error)
		check     func(t *testing.T, got []model.User)
	}{
		{
			name: "multiple users returns all users",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name"}).
					AddRow(1, "Alice").
					AddRow(2, "Bob").
					AddRow(3, "Charlie")
				mock.ExpectQuery(`SELECT id, name FROM users`).WillReturnRows(rows)
			},
			wantErr: false,
			check: func(t *testing.T, got []model.User) {
				require.Len(t, got, 3)
				assert.Equal(t, "Alice", got[0].Name)
				assert.Equal(t, "Bob", got[1].Name)
				assert.Equal(t, "Charlie", got[2].Name)
			},
		},
		{
			name: "empty table returns non-nil empty slice",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name"})
				mock.ExpectQuery(`SELECT id, name FROM users`).WillReturnRows(rows)
			},
			wantErr: false,
			check: func(t *testing.T, got []model.User) {
				assert.NotNil(t, got, "slice must be non-nil for JSON []")
				assert.Len(t, got, 0)
			},
		},
		{
			name: "DB query error returns wrapped error",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, name FROM users`).
					WillReturnError(fmt.Errorf("timeout"))
			},
			wantErr: true,
			errCheck: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "find all users")
			},
		},
		{
			name: "row scan error returns wrapped error",
			setupMock: func(mock sqlmock.Sqlmock) {
				// Return a row with wrong column count to force a scan error.
				rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
				mock.ExpectQuery(`SELECT id, name FROM users`).WillReturnRows(rows)
			},
			wantErr: true,
			errCheck: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "scan user row")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := newMockDB(t)
			tc.setupMock(mock)

			dao := NewUserDao(db)
			got, err := dao.FindAll(ctx)

			if tc.wantErr {
				require.Error(t, err)
				if tc.errCheck != nil {
					tc.errCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				if tc.check != nil {
					tc.check(t, got)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// FindByName
// ---------------------------------------------------------------------------

func TestUserDao_FindByName(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		inputName string
		setupMock func(mock sqlmock.Sqlmock)
		wantErr   bool
		errCheck  func(t *testing.T, err error)
		check     func(t *testing.T, got *model.User)
	}{
		{
			name:      "existing name returns matching user",
			inputName: "Alice",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "Alice")
				mock.ExpectQuery(`SELECT id, name FROM users WHERE name = \$1`).
					WithArgs("Alice").
					WillReturnRows(rows)
			},
			wantErr: false,
			check: func(t *testing.T, got *model.User) {
				assert.Equal(t, "Alice", got.Name)
				assert.Equal(t, 1, got.ID)
			},
		},
		{
			name:      "non-existent name returns UserNotFoundError",
			inputName: "Nobody",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, name FROM users WHERE name = \$1`).
					WithArgs("Nobody").
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
			errCheck: func(t *testing.T, err error) {
				assert.True(t, isUserNotFoundError(err), "expected UserNotFoundError, got %T: %v", err, err)
				assert.Contains(t, err.Error(), "Nobody")
			},
		},
		{
			name:      "DB error returns wrapped error",
			inputName: "Alice",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, name FROM users WHERE name = \$1`).
					WithArgs("Alice").
					WillReturnError(fmt.Errorf("I/O error"))
			},
			wantErr: true,
			errCheck: func(t *testing.T, err error) {
				assert.False(t, isUserNotFoundError(err))
				assert.Contains(t, err.Error(), "find user by name")
			},
		},
		{
			name:      "returned user name equals input name (invariant)",
			inputName: "Bob",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name"}).AddRow(10, "Bob")
				mock.ExpectQuery(`SELECT id,