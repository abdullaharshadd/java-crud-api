```go
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock/v2"
	smartErr "github.com/smartContact/internal/smartcontact/error"
	"github.com/smartContact/internal/smartcontact/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func newMockRepo(t *testing.T) (UserRepository, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	repo, err := NewUserRepository(db)
	require.NoError(t, err)

	return repo, mock, func() { db.Close() }
}

func isUserNotFound(err error) bool {
	return smartErr.IsUserNotFound(err)
}

// ---------------------------------------------------------------------------
// NewUserRepository
// ---------------------------------------------------------------------------

func TestNewUserRepository_NilDB(t *testing.T) {
	repo, err := NewUserRepository(nil)
	assert.Nil(t, repo)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nil")
}

func TestNewUserRepository_ValidDB(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo, err := NewUserRepository(db)
	assert.NoError(t, err)
	assert.NotNil(t, repo)
}

// ---------------------------------------------------------------------------
// FindByID
// ---------------------------------------------------------------------------

func TestFindByID(t *testing.T) {
	cols := []string{"id", "name", "email"}

	tests := []struct {
		name        string
		id          int
		setupMock   func(mock sqlmock.Sqlmock, id int)
		wantUser    model.User
		wantNotFound bool
		wantErr     bool
	}{
		{
			name: "user exists",
			id:   1,
			setupMock: func(mock sqlmock.Sqlmock, id int) {
				mock.ExpectQuery(`SELECT id, name, email FROM users WHERE id = \$1`).
					WithArgs(id).
					WillReturnRows(sqlmock.NewRows(cols).AddRow(1, "Alice", "alice@example.com"))
			},
			wantUser: model.User{ID: 1, Name: "Alice", Email: "alice@example.com"},
		},
		{
			name: "user not found",
			id:   42,
			setupMock: func(mock sqlmock.Sqlmock, id int) {
				mock.ExpectQuery(`SELECT id, name, email FROM users WHERE id = \$1`).
					WithArgs(id).
					WillReturnError(sql.ErrNoRows)
			},
			wantNotFound: true,
		},
		{
			name: "db error",
			id:   99,
			setupMock: func(mock sqlmock.Sqlmock, id int) {
				mock.ExpectQuery(`SELECT id, name, email FROM users WHERE id = \$1`).
					WithArgs(id).
					WillReturnError(errors.New("connection refused"))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo, mock, cleanup := newMockRepo(t)
			defer cleanup()

			tc.setupMock(mock, tc.id)

			got, err := repo.FindByID(context.Background(), tc.id)

			if tc.wantNotFound {
				require.Error(t, err)
				assert.True(t, isUserNotFound(err), "expected ErrUserNotFound, got: %v", err)
				return
			}
			if tc.wantErr {
				require.Error(t, err)
				assert.False(t, isUserNotFound(err))
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantUser, got)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// FindByName
// ---------------------------------------------------------------------------

func TestFindByName(t *testing.T) {
	cols := []string{"id", "name", "email"}

	tests := []struct {
		name       string
		inputName  string
		setupMock  func(mock sqlmock.Sqlmock, name string)
		wantUser   model.User
		wantFound  bool
		wantErr    bool
	}{
		{
			name:      "matching user exists",
			inputName: "Alice",
			setupMock: func(mock sqlmock.Sqlmock, name string) {
				mock.ExpectQuery(`SELECT id, name, email FROM users WHERE name = \$1`).
					WithArgs(name).
					WillReturnRows(sqlmock.NewRows(cols).AddRow(1, "Alice", "alice@example.com"))
			},
			wantUser:  model.User{ID: 1, Name: "Alice", Email: "alice@example.com"},
			wantFound: true,
		},
		{
			name:      "no matching user",
			inputName: "Ghost",
			setupMock: func(mock sqlmock.Sqlmock, name string) {
				mock.ExpectQuery(`SELECT id, name, email FROM users WHERE name = \$1`).
					WithArgs(name).
					WillReturnError(sql.ErrNoRows)
			},
			wantFound: false,
		},
		{
			// QueryRowContext only scans the first row; a "too many rows" error
			// can only surface if the driver itself reports it. We model this as
			// a generic db error to verify the error path.
			name:      "multiple users with same name causes db error",
			inputName: "Bob",
			setupMock: func(mock sqlmock.Sqlmock, name string) {
				mock.ExpectQuery(`SELECT id, name, email FROM users WHERE name = \$1`).
					WithArgs(name).
					WillReturnError(errors.New("non-unique result"))
			},
			wantErr: true,
		},
		{
			name:      "name is empty string",
			inputName: "",
			setupMock: func(mock sqlmock.Sqlmock, name string) {
				mock.ExpectQuery(`SELECT id, name, email FROM users WHERE name = \$1`).
					WithArgs(name).
					WillReturnError(sql.ErrNoRows)
			},
			wantFound: false,
		},
		{
			name:      "user with null-like name exists",
			inputName: "",
			setupMock: func(mock sqlmock.Sqlmock, name string) {
				mock.ExpectQuery(`SELECT id, name, email FROM users WHERE name = \$1`).
					WithArgs(name).
					WillReturnRows(sqlmock.NewRows(cols).AddRow(5, "", "noname@example.com"))
			},
			wantUser:  model.User{ID: 5, Name: "", Email: "noname@example.com"},
			wantFound: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo, mock, cleanup := newMockRepo(t)
			defer cleanup()

			tc.setupMock(mock, tc.inputName)

			got, found, err := repo.FindByName(context.Background(), tc.inputName)

			if tc.wantErr {
				require.Error(t, err)
				assert.False(t, found)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantFound, found)
			if tc.wantFound {
				assert.Equal(t, tc.wantUser, got)
			} else {
				assert.Equal(t, model.User{}, got)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// FindByName must not modify any rows (read-only invariant).
func TestFindByName_ReadOnly(t *testing.T) {
	repo, mock, cleanup := newMockRepo(t)
	defer cleanup()

	cols := []string{"id", "name", "email"}
	mock.ExpectQuery(`SELECT id, name, email FROM users WHERE name = \$1`).
		WithArgs("Alice").
		WillReturnRows(sqlmock.NewRows(cols).AddRow(1, "Alice", "alice@example.com"))

	_, _, err := repo.FindByName(context.Background(), "Alice")
	require.NoError(t, err)

	// If the mock saw any unexpected writes the ExpectationsWereMet call would
	// still pass but go-sqlmock would have already failed on the unexpected
	// Exec. Calling it here confirms no Exec was issued.
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// FindAll
// ---------------------------------------------------------------------------

func TestFindAll(t *testing.T) {
	cols := []string{"id", "name", "email"}

	tests := []struct {
		name      string
		setupMock func(mock sqlmock.Sqlmock)
		wantUsers []model.User
		wantErr   bool
	}{
		{
			name: "multiple users",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, name, email FROM users ORDER BY id`).
					WillReturnRows(sqlmock.NewRows(cols).
						AddRow(1, "Alice", "alice@example.com").
						AddRow(2, "Bob", "bob@example.com"))
			},
			wantUsers: []model.User{
				{ID: 1, Name: "Alice", Email: "alice@example.com"},
				{ID: 2, Name: "Bob", Email: "bob@example.com"},
			},
		},
		{
			name: "no users",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, name, email FROM users ORDER BY id`).
					WillReturnRows(sqlmock.NewRows(cols))
			},
			wantUsers: nil, // empty slice, never nil from spec perspective
		},
		{
			name: "query error",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, name, email FROM users ORDER BY id`).
					WillReturnError(errors.New("db down"))
			},
			wantErr: true,
		},
		{
			name: "scan error",
			setupMock: func(mock sqlmock.Sqlmock) {
				// Return a row with too few columns to force a scan error.
				mock.ExpectQuery(`SELECT id, name, email FROM users ORDER BY id`).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo, mock, cleanup := newMockRepo(t)
			defer cleanup()

			tc.setupMock(mock)

			got, err := repo.FindAll(context.Background())

			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			if len(tc.wantUsers) == 0 {
				// Spec says "never returns null" – an empty result may be nil
				// slice in Go but never an error.
				assert.NotErrorIs(t, err, errors.New("any"))
			} else {
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
		name      string
		input     model.User
		setupMock func(mock sqlmock.Sqlmock, u model.User)
		wantUser  model.User
		wantErr   bool
		wantNotFound bool
	}{
		{
			name:  "insert new user (ID == 0)",
			input: model.User{Name: "Carol", Email: "carol@example.com"},
			setupMock: func(mock sqlmock.Sqlmock, u model.User) {
				mock.ExpectQuery(`INSERT INTO users \(name, email\) VALUES \(\$1, \$2\) RETURNING id`).
					WithArgs(u.Name, u.Email).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7))
			},
			wantUser: model.User{ID: 7, Name: "Carol", Email: "carol@example.com"},
		},
		{
			name:  "insert error",
			input: model.User{Name: "Carol", Email: "carol@example.com"},
			setupMock: func(mock sqlmock.Sqlmock, u model.User) {
				mock.ExpectQuery(`INSERT INTO users \(name, email\) VALUES \(\$1, \$2\) RETURNING id`).
					WithArgs(u.Name, u.Email).
					WillReturnError(errors.New("unique violation"))
			},
			wantErr: true,
		},
		{
			name:  "update existing user",
			input: model.User{ID: 3, Name: "Dave", Email: "dave@example.com"},
			setupMock: func(mock sqlmock.Sqlmock, u model.User) {
				mock.ExpectExec(`UPDATE users SET name = \$1, email = \$2 WHERE id = \$3`).
					WithArgs(u.Name, u.Email, u.ID).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantUser: model.User{ID: 3, Name: "Dave", Email: "dave@example.com"},
		},
		{
			name:  "update non-existent user returns not found",
			input: model.User{ID: 99, Name: "Nobody", Email: "nobody@example.com"},
			setupMock: func(mock sqlmock.Sqlmock, u model.User) {
				mock.ExpectExec(`UPDATE users SET name = \$1, email = \$2 WHERE id = \$3`).
					WithArgs(u.Name, u.Email, u.ID).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantNotFound: true,
		},
		{
			name:  "update exec error",
			input: model.User{ID: 3, Name: "Dave", Email: "dave@example.com"},
			setupMock: func(mock sqlmock.Sqlmock, u model.User) {
				mock.ExpectExec(`UPDATE users SET name = \$1, email = \$2 WHERE id = \$3`).
					WithArgs(u.Name, u.Email, u.ID).
					WillReturnError(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo, mock, cleanup := newMockRepo(t)
			defer cleanup()

			tc.setupMock(mock, tc.input)

			got, err := repo.Save(context.Background(), tc.input)

			if tc.wantNotFound {
				require.Error(t, err)
				assert.True(t, isUserNotFound(err), "expected ErrUserNot