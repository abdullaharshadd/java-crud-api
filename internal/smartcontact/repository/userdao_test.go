```go
package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	smartErr "github.com/smartContact/internal/smartcontact/error"
	"github.com/smartContact/internal/smartcontact/model"
	"github.com/smartContact/internal/smartcontact/repository"
)

// ─────────────────────────────────────────────
// helpers
// ─────────────────────────────────────────────

func newMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db, mock
}

func isUserNotFound(err error) bool {
	var nf *smartErr.UserNotFoundError
	return errors.As(err, &nf)
}

// ─────────────────────────────────────────────
// Save
// ─────────────────────────────────────────────

func TestMySQLUserRepo_Save(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		inputUser   *model.User
		setupMock   func(mock sqlmock.Sqlmock)
		wantUser    *model.User
		wantErr     bool
		errContains string
	}{
		{
			name:      "save new user without existing id – generated id populated",
			inputUser: &model.User{Name: "Alice"},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO user \(name\) VALUES \(\?\)`).
					WithArgs("Alice").
					WillReturnResult(sqlmock.NewResult(42, 1))
			},
			wantUser: &model.User{ID: 42, Name: "Alice"},
		},
		{
			name:      "save nil user – returns error immediately",
			inputUser: nil,
			setupMock: func(mock sqlmock.Sqlmock) {
				// no DB calls expected
			},
			wantErr:     true,
			errContains: "nil user",
		},
		{
			name:      "DB exec error propagated",
			inputUser: &model.User{Name: "Bob"},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO user \(name\) VALUES \(\?\)`).
					WithArgs("Bob").
					WillReturnError(errors.New("connection refused"))
			},
			wantErr:     true,
			errContains: "connection refused",
		},
		{
			name:      "LastInsertId error propagated",
			inputUser: &model.User{Name: "Carol"},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO user \(name\) VALUES \(\?\)`).
					WithArgs("Carol").
					WillReturnResult(sqlmock.NewErrorResult(errors.New("last insert id unavailable")))
			},
			wantErr:     true,
			errContains: "last insert id unavailable",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := newMockDB(t)
			tc.setupMock(mock)

			repo := repository.NewMySQLUserRepo(db)
			got, err := repo.Save(ctx, tc.inputUser)

			if tc.wantErr {
				require.Error(t, err)
				if tc.errContains != "" {
					assert.Contains(t, err.Error(), tc.errContains)
				}
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, tc.wantUser.ID, got.ID)
				assert.Equal(t, tc.wantUser.Name, got.Name)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ─────────────────────────────────────────────
// FindByID
// ─────────────────────────────────────────────

func TestMySQLUserRepo_FindByID(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		inputID     int
		setupMock   func(mock sqlmock.Sqlmock)
		wantUser    *model.User
		wantErr     bool
		wantNotFound bool
		errContains string
	}{
		{
			name:    "user with given id exists – returned",
			inputID: 1,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "Alice")
				mock.ExpectQuery(`SELECT id, name FROM user WHERE id = \?`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			wantUser: &model.User{ID: 1, Name: "Alice"},
		},
		{
			name:    "no user with given id – ErrUserNotFound",
			inputID: 99,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, name FROM user WHERE id = \?`).
					WithArgs(99).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr:      true,
			wantNotFound: true,
		},
		{
			name:    "DB query error propagated",
			inputID: 5,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, name FROM user WHERE id = \?`).
					WithArgs(5).
					WillReturnError(errors.New("network timeout"))
			},
			wantErr:     true,
			errContains: "network timeout",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := newMockDB(t)
			tc.setupMock(mock)

			repo := repository.NewMySQLUserRepo(db)
			got, err := repo.FindByID(ctx, tc.inputID)

			if tc.wantErr {
				require.Error(t, err)
				if tc.wantNotFound {
					assert.True(t, isUserNotFound(err), "expected UserNotFoundError, got: %v", err)
				}
				if tc.errContains != "" {
					assert.Contains(t, err.Error(), tc.errContains)
				}
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, tc.wantUser.ID, got.ID)
				assert.Equal(t, tc.wantUser.Name, got.Name)
				// invariant: returned user id equals input id
				assert.Equal(t, tc.inputID, got.ID)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ─────────────────────────────────────────────
// FindAll
// ─────────────────────────────────────────────

func TestMySQLUserRepo_FindAll(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		setupMock   func(mock sqlmock.Sqlmock)
		wantUsers   []model.User
		wantErr     bool
		errContains string
	}{
		{
			name: "one or more users exist – all returned",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name"}).
					AddRow(1, "Alice").
					AddRow(2, "Bob")
				mock.ExpectQuery(`SELECT id, name FROM user`).
					WillReturnRows(rows)
			},
			wantUsers: []model.User{
				{ID: 1, Name: "Alice"},
				{ID: 2, Name: "Bob"},
			},
		},
		{
			name: "no users exist – empty slice returned (not nil error)",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name"})
				mock.ExpectQuery(`SELECT id, name FROM user`).
					WillReturnRows(rows)
			},
			wantUsers: nil, // append never called, nil slice is fine
		},
		{
			name: "DB query error propagated",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, name FROM user`).
					WillReturnError(errors.New("table not found"))
			},
			wantErr:     true,
			errContains: "table not found",
		},
		{
			name: "row scan error propagated",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name"}).
					AddRow("not-an-int", "Alice") // type mismatch triggers scan error
				mock.ExpectQuery(`SELECT id, name FROM user`).
					WillReturnRows(rows)
			},
			wantErr: true,
		},
		{
			name: "rows.Err propagated",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name"}).
					AddRow(1, "Alice").
					RowError(0, errors.New("streaming error"))
				mock.ExpectQuery(`SELECT id, name FROM user`).
					WillReturnRows(rows)
			},
			wantErr:     true,
			errContains: "streaming error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := newMockDB(t)
			tc.setupMock(mock)

			repo := repository.NewMySQLUserRepo(db)
			got, err := repo.FindAll(ctx)

			if tc.wantErr {
				require.Error(t, err)
				if tc.errContains != "" {
					assert.Contains(t, err.Error(), tc.errContains)
				}
			} else {
				require.NoError(t, err)
				if tc.wantUsers == nil {
					assert.Empty(t, got)
				} else {
					assert.Equal(t, tc.wantUsers, got)
					// invariant: result count equals stored records
					assert.Len(t, got, len(tc.wantUsers))
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ─────────────────────────────────────────────
// DeleteByID
// ─────────────────────────────────────────────

func TestMySQLUserRepo_DeleteByID(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name                 string
		inputID              int
		setupMock            func(mock sqlmock.Sqlmock)
		wantErr              bool
		wantEmptyResultDelete bool
		errContains          string
	}{
		{
			name:    "user with given id exists – deleted successfully",
			inputID: 1,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE FROM user WHERE id = \?`).
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name:    "no user with given id – ErrEmptyResultDelete returned",
			inputID: 99,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE FROM user WHERE id = \?`).
					WithArgs(99).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr:               true,
			wantEmptyResultDelete: true,
		},
		{
			name:    "DB exec error propagated",
			inputID: 3,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE FROM user WHERE id = \?`).
					WithArgs(3).
					WillReturnError(errors.New("deadlock detected"))
			},
			wantErr:     true,
			errContains: "deadlock detected",
		},
		{
			name:    "RowsAffected error propagated",
			inputID: 4,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE FROM user WHERE id = \?`).
					WithArgs(4).
					WillReturnResult(sqlmock.NewErrorResult(errors.New("rows affected unavailable")))
			},
			wantErr:     true,
			errContains: "rows affected unavailable",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := newMockDB(t)
			tc.setupMock(mock)

			repo := repository.NewMySQLUserRepo(db)
			err := repo.DeleteByID(ctx, tc.inputID)

			if tc.wantErr {
				require.Error(t, err)
				if tc.wantEmptyResultDelete {
					assert.ErrorIs(t, err, repository.ErrEmptyResultDelete)
					// invariant: ErrEmptyResultDelete is NOT a UserNotFoundError
					assert.False(t, isUserNotFound(err),
						"ErrEmptyResultDelete must not be a UserNotFoundError")
				}
				if tc.errContains != "" {
					assert.Contains(t, err.Error(), tc.errContains)
				}
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ─────────────────────────────────────────────
// FindByName
// ─────────────────────────────────────────────

func TestMySQLUserRepo_FindByName(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		inputName    string
		setupMock    func(mock sqlmock.Sqlmock)
		wantUser     *model.User
		wantErr      bool
		wantNotFound bool
		errContains  string
	}{
		{
			name:      "user with exact matching name exists – returned",
			inputName: "Alice",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name"}).AddRow(7, "Alice")
				mock.ExpectQuery(`SELECT id, name FROM user WHERE name = \?`).
					WithArgs("Alice").
					WillReturnRows(rows)
			},
			wantUser: &model.User{ID: 7, Name: "Alice"},
		},
		{
			name:      "no user with given name – ErrUserNotFound",
			inputName: "Ghost",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, name FROM user WHERE name = \?`).
					WithArgs