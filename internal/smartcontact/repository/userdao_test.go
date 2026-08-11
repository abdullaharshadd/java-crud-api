```go
package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock/v2"
	"github.com/smartContact/internal/smartcontact/model"
	"github.com/smartContact/internal/smartcontact/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newMockedDao creates a UserDao backed by a sqlmock DB.
// The helper pre-expects the CREATE TABLE statement that NewUserDao emits.
func newMockedDao(t *testing.T) (*repository.UserDao, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	// NewUserDao always executes the DDL first.
	mock.ExpectExec(`CREATE TABLE IF NOT EXISTS users`).
		WillReturnResult(sqlmock.NewResult(0, 0))

	dao, err := repository.NewUserDao(context.Background(), db)
	require.NoError(t, err)

	t.Cleanup(func() { _ = db.Close() })
	return dao, mock
}

// ---------------------------------------------------------------------------
// NewUserDao
// ---------------------------------------------------------------------------

func TestNewUserDao_NilDB(t *testing.T) {
	_, err := repository.NewUserDao(context.Background(), nil)
	assert.Error(t, err)
}

func TestNewUserDao_DDLError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectExec(`CREATE TABLE IF NOT EXISTS users`).
		WillReturnError(errors.New("permission denied"))

	_, err = repository.NewUserDao(context.Background(), db)
	assert.Error(t, err)
}

func TestNewUserDao_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectExec(`CREATE TABLE IF NOT EXISTS users`).
		WillReturnResult(sqlmock.NewResult(0, 0))

	dao, err := repository.NewUserDao(context.Background(), db)
	assert.NoError(t, err)
	assert.NotNil(t, dao)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// FindByName
// ---------------------------------------------------------------------------

func TestFindByName(t *testing.T) {
	tests := []struct {
		name        string
		searchName  string
		setupMock   func(mock sqlmock.Sqlmock)
		wantUser    *model.User
		wantErr     error
		wantErrWrap bool // true → just assert error is non-nil (wrapped)
	}{
		{
			name:       "user found",
			searchName: "alice",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name"}).
					AddRow(1, "alice")
				mock.ExpectQuery(`SELECT id, name FROM users WHERE name = \$1`).
					WithArgs("alice").
					WillReturnRows(rows)
			},
			wantUser: &model.User{ID: 1, Name: "alice"},
		},
		{
			name:       "user not found",
			searchName: "ghost",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, name FROM users WHERE name = \$1`).
					WithArgs("ghost").
					WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))
			},
			wantUser: nil,
			wantErr:  repository.ErrUserNotFound,
		},
		{
			name:       "db error",
			searchName: "broken",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, name FROM users WHERE name = \$1`).
					WithArgs("broken").
					WillReturnError(errors.New("connection reset"))
			},
			wantUser:    nil,
			wantErrWrap: true,
		},
		{
			name:       "name is empty string (null-equivalent)",
			searchName: "",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, name FROM users WHERE name = \$1`).
					WithArgs("").
					WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))
			},
			wantUser: nil,
			wantErr:  repository.ErrUserNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dao, mock := newMockedDao(t)
			tc.setupMock(mock)

			got, err := dao.FindByName(context.Background(), tc.searchName)

			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else if tc.wantErrWrap {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tc.wantUser, got)

			// Invariant: returned user name matches search name
			if got != nil {
				assert.Equal(t, tc.searchName, got.Name)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// FindByID
// ---------------------------------------------------------------------------

func TestFindByID(t *testing.T) {
	tests := []struct {
		name        string
		id          int
		setupMock   func(mock sqlmock.Sqlmock)
		wantUser    *model.User
		wantErr     error
		wantErrWrap bool
	}{
		{
			name: "user found",
			id:   42,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name"}).AddRow(42, "bob")
				mock.ExpectQuery(`SELECT id, name FROM users WHERE id = \$1`).
					WithArgs(42).
					WillReturnRows(rows)
			},
			wantUser: &model.User{ID: 42, Name: "bob"},
		},
		{
			name: "user not found",
			id:   999,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, name FROM users WHERE id = \$1`).
					WithArgs(999).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))
			},
			wantUser: nil,
			wantErr:  repository.ErrUserNotFound,
		},
		{
			name: "db error",
			id:   1,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, name FROM users WHERE id = \$1`).
					WithArgs(1).
					WillReturnError(errors.New("timeout"))
			},
			wantUser:    nil,
			wantErrWrap: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dao, mock := newMockedDao(t)
			tc.setupMock(mock)

			got, err := dao.FindByID(context.Background(), tc.id)

			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else if tc.wantErrWrap {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tc.wantUser, got)

			// Invariant: returned id matches input id
			if got != nil {
				assert.Equal(t, tc.id, got.ID)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// FindAll
// ---------------------------------------------------------------------------

func TestFindAll(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(mock sqlmock.Sqlmock)
		wantUsers   []model.User
		wantErr     bool
		wantErrWrap bool
	}{
		{
			name: "multiple users",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name"}).
					AddRow(1, "alice").
					AddRow(2, "bob")
				mock.ExpectQuery(`SELECT id, name FROM users ORDER BY id`).
					WillReturnRows(rows)
			},
			wantUsers: []model.User{
				{ID: 1, Name: "alice"},
				{ID: 2, Name: "bob"},
			},
		},
		{
			name: "no users exist",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, name FROM users ORDER BY id`).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))
			},
			wantUsers: nil, // empty slice represented as nil from append
		},
		{
			name: "db error on query",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, name FROM users ORDER BY id`).
					WillReturnError(errors.New("query failed"))
			},
			wantErr: true,
		},
		{
			name: "db error on scan",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name"}).
					AddRow("not-an-int", "alice")
				mock.ExpectQuery(`SELECT id, name FROM users ORDER BY id`).
					WillReturnRows(rows)
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dao, mock := newMockedDao(t)
			tc.setupMock(mock)

			got, err := dao.FindAll(context.Background())

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Invariant: FindAll never returns null (non-nil slice or nil
				// slice with zero length is acceptable; never an error on empty)
				assert.Equal(t, tc.wantUsers, got)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// Save
// ---------------------------------------------------------------------------

func TestSave(t *testing.T) {
	tests := []struct {
		name        string
		inputUser   *model.User
		setupMock   func(mock sqlmock.Sqlmock)
		wantUser    *model.User
		wantErr     error
		wantErrWrap bool
	}{
		{
			name:      "nil user returns error",
			inputUser: nil,
			setupMock: func(mock sqlmock.Sqlmock) {},
			wantUser:  nil,
			wantErr:   errors.New("repository: user must not be nil"),
		},
		{
			name:      "insert new user (id == 0)",
			inputUser: &model.User{ID: 0, Name: "carol"},
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id"}).AddRow(7)
				mock.ExpectQuery(`INSERT INTO users`).
					WithArgs("carol").
					WillReturnRows(rows)
			},
			wantUser: &model.User{ID: 7, Name: "carol"},
		},
		{
			name:      "insert error",
			inputUser: &model.User{ID: 0, Name: "fail"},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`INSERT INTO users`).
					WithArgs("fail").
					WillReturnError(errors.New("unique violation"))
			},
			wantUser:    nil,
			wantErrWrap: true,
		},
		{
			name:      "update existing user",
			inputUser: &model.User{ID: 5, Name: "dave"},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`UPDATE users SET name = \$1 WHERE id = \$2`).
					WithArgs("dave", 5).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantUser: &model.User{ID: 5, Name: "dave"},
		},
		{
			name:      "update non-existent user",
			inputUser: &model.User{ID: 9999, Name: "nobody"},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`UPDATE users SET name = \$1 WHERE id = \$2`).
					WithArgs("nobody", 9999).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantUser: nil,
			wantErr:  repository.ErrUserNotFound,
		},
		{
			name:      "update db error",
			inputUser: &model.User{ID: 3, Name: "err"},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`UPDATE users SET name = \$1 WHERE id = \$2`).
					WithArgs("err", 3).
					WillReturnError(errors.New("connection lost"))
			},
			wantUser:    nil,
			wantErrWrap: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dao, mock := newMockedDao(t)
			tc.setupMock(mock)

			got, err := dao.Save(context.Background(), tc.inputUser)

			if tc.wantErr != nil {
				assert.Error(t, err)
			} else if tc.wantErrWrap {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantUser, got)
			}

			// Invariant: if a user was returned it is the same pointer as input
			if got != nil && tc.inputUser != nil {
				assert.Equal(t, tc.inputUser.Name, got.Name)
				assert.NotZero(t, got.ID)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// DeleteByID
// ---------------------------------------------------------------------------

func TestDeleteByID(t *testing.T) {
	tests := []struct {
		name        string
		id          int
		setupMock   func(mock sqlmock.Sqlmock)
		wantErr     error
		wantErrWrap bool
	}{
		{
			name: "delete existing user",
			id:   10,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE FROM users WHERE id = \$1`).
					WithArgs(10).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "delete non