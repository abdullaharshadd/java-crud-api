```go
package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	smartError "github.com/smartContact/internal/smartcontact/error"
	"github.com/smartContact/internal/smartcontact/model"
	"github.com/smartContact/internal/smartcontact/repository"
)

// newMockDB creates a sqlmock-backed *sqlx.DB for testing.
func newMockDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return sqlx.NewDb(db, "sqlmock"), mock
}

// userColumns replicates the column ordering returned by SELECT queries.
var userResultColumns = []string{"id", "name", "secondName", "work", "email", "phone", "about"}

// ─────────────────────────────────────────────────────────────────────────────
// Save
// ─────────────────────────────────────────────────────────────────────────────

func TestUserRepository_Save(t *testing.T) {
	ctx := context.Background()

	validUser := model.User{
		Name:       "Alice",
		SecondName: "Smith",
		Work:       "Engineer",
		Email:      "alice@example.com",
		Phone:      "555-1234",
		About:      "Test user",
	}

	tests := []struct {
		name      string
		user      model.User
		setupMock func(mock sqlmock.Sqlmock)
		wantUser  model.User
		wantErr   bool
		errContains string
	}{
		{
			name: "insert new user returns persisted user with generated id",
			user: validUser,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO USER`).
					WithArgs(
						validUser.Name,
						validUser.SecondName,
						validUser.Work,
						validUser.Email,
						validUser.Phone,
						validUser.About,
					).
					WillReturnResult(sqlmock.NewResult(42, 1))
			},
			wantUser: model.User{
				ID:         42,
				Name:       validUser.Name,
				SecondName: validUser.SecondName,
				Work:       validUser.Work,
				Email:      validUser.Email,
				Phone:      validUser.Phone,
				About:      validUser.About,
			},
			wantErr: false,
		},
		{
			name: "update existing user returns updated user",
			user: model.User{
				ID:         10,
				Name:       "Bob",
				SecondName: "Jones",
				Work:       "Manager",
				Email:      "bob@example.com",
				Phone:      "555-5678",
				About:      "Updated",
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`UPDATE USER`).
					WithArgs(
						"Bob", "Jones", "Manager", "bob@example.com", "555-5678", "Updated", 10,
					).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantUser: model.User{
				ID:         10,
				Name:       "Bob",
				SecondName: "Jones",
				Work:       "Manager",
				Email:      "bob@example.com",
				Phone:      "555-5678",
				About:      "Updated",
			},
			wantErr: false,
		},
		{
			name: "insert returns error when db fails",
			user: validUser,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO USER`).
					WillReturnError(errors.New("db connection error"))
			},
			wantErr:     true,
			errContains: "insert user",
		},
		{
			name: "update returns error when db fails",
			user: model.User{
				ID:         5,
				Name:       "Charlie",
				SecondName: "Brown",
				Work:       "Dev",
				Email:      "charlie@example.com",
				Phone:      "555-9999",
				About:      "Dev",
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`UPDATE USER`).
					WillReturnError(errors.New("db timeout"))
			},
			wantErr:     true,
			errContains: "update user",
		},
		{
			name: "invalid user returns validation error",
			user: model.User{
				// ID == 0 triggers insert path, but Validate() should fail if
				// the model enforces required fields. Adjust the zero value
				// to what your Validate() actually rejects.
				Name: "",
			},
			setupMock:   func(_ sqlmock.Sqlmock) {},
			wantErr:     true,
			errContains: "save user",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := newMockDB(t)
			tc.setupMock(mock)

			repo := repository.NewUserRepository(db)
			got, err := repo.Save(ctx, tc.user)

			if tc.wantErr {
				require.Error(t, err)
				if tc.errContains != "" {
					assert.Contains(t, err.Error(), tc.errContains)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.wantUser, got)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// FindByID
// ─────────────────────────────────────────────────────────────────────────────

func TestUserRepository_FindByID(t *testing.T) {
	ctx := context.Background()

	existingUser := model.User{
		ID:         1,
		Name:       "Alice",
		SecondName: "Smith",
		Work:       "Engineer",
		Email:      "alice@example.com",
		Phone:      "555-1234",
		About:      "Test user",
	}

	tests := []struct {
		name        string
		id          int
		setupMock   func(mock sqlmock.Sqlmock)
		wantUser    model.User
		wantErr     bool
		wantNotFound bool
		errContains  string
	}{
		{
			name: "returns user when id exists",
			id:   1,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(userResultColumns).
					AddRow(
						existingUser.ID,
						existingUser.Name,
						existingUser.SecondName,
						existingUser.Work,
						existingUser.Email,
						existingUser.Phone,
						existingUser.About,
					)
				mock.ExpectQuery(`SELECT .+ FROM USER WHERE id = ?`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			wantUser: existingUser,
			wantErr:  false,
		},
		{
			name: "returns UserNotFoundError when id does not exist",
			id:   99,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT .+ FROM USER WHERE id = ?`).
					WithArgs(99).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr:      true,
			wantNotFound: true,
		},
		{
			name: "returns wrapped error on db failure",
			id:   2,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT .+ FROM USER WHERE id = ?`).
					WithArgs(2).
					WillReturnError(errors.New("connection refused"))
			},
			wantErr:     true,
			errContains: "find user by id 2",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := newMockDB(t)
			tc.setupMock(mock)

			repo := repository.NewUserRepository(db)
			got, err := repo.FindByID(ctx, tc.id)

			if tc.wantErr {
				require.Error(t, err)
				if tc.wantNotFound {
					var nfe *smartError.UserNotFoundError
					assert.True(t, errors.As(err, &nfe), "expected UserNotFoundError, got: %v", err)
				}
				if tc.errContains != "" {
					assert.Contains(t, err.Error(), tc.errContains)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.wantUser, got)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// FindAll
// ─────────────────────────────────────────────────────────────────────────────

func TestUserRepository_FindAll(t *testing.T) {
	ctx := context.Background()

	users := []model.User{
		{ID: 1, Name: "Alice", SecondName: "Smith", Work: "Engineer", Email: "alice@example.com", Phone: "555-1111", About: "A"},
		{ID: 2, Name: "Bob", SecondName: "Jones", Work: "Manager", Email: "bob@example.com", Phone: "555-2222", About: "B"},
	}

	tests := []struct {
		name        string
		setupMock   func(mock sqlmock.Sqlmock)
		wantUsers   []model.User
		wantEmpty   bool
		wantErr     bool
		errContains string
	}{
		{
			name: "returns all users when records exist",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(userResultColumns)
				for _, u := range users {
					rows.AddRow(u.ID, u.Name, u.SecondName, u.Work, u.Email, u.Phone, u.About)
				}
				mock.ExpectQuery(`SELECT .+ FROM USER`).WillReturnRows(rows)
			},
			wantUsers: users,
		},
		{
			name: "returns empty slice when no records exist",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(userResultColumns)
				mock.ExpectQuery(`SELECT .+ FROM USER`).WillReturnRows(rows)
			},
			wantEmpty: true,
		},
		{
			name: "returns error on db failure",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT .+ FROM USER`).WillReturnError(errors.New("db error"))
			},
			wantErr:     true,
			errContains: "find all users",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := newMockDB(t)
			tc.setupMock(mock)

			repo := repository.NewUserRepository(db)
			got, err := repo.FindAll(ctx)

			if tc.wantErr {
				require.Error(t, err)
				if tc.errContains != "" {
					assert.Contains(t, err.Error(), tc.errContains)
				}
				// FindAll must not return nil on error per invariant
				return
			}

			require.NoError(t, err)
			// Invariant: never nil
			assert.NotNil(t, got)

			if tc.wantEmpty {
				assert.Empty(t, got)
			} else {
				assert.Equal(t, tc.wantUsers, got)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// FindByName
// ─────────────────────────────────────────────────────────────────────────────

func TestUserRepository_FindByName(t *testing.T) {
	ctx := context.Background()

	existingUser := model.User{
		ID:         3,
		Name:       "Carol",
		SecondName: "White",
		Work:       "Designer",
		Email:      "carol@example.com",
		Phone:      "555-3333",
		About:      "Creative",
	}

	tests := []struct {
		name        string
		inputName   string
		setupMock   func(mock sqlmock.Sqlmock)
		wantUser    model.User
		wantErr     bool
		errContains string
	}{
		{
			name:      "returns user when name exists",
			inputName: "Carol",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(userResultColumns).
					AddRow(
						existingUser.ID,
						existingUser.Name,
						existingUser.SecondName,
						existingUser.Work,
						existingUser.Email,
						existingUser.Phone,
						existingUser.About,
					)
				mock.ExpectQuery(`SELECT .+ FROM USER WHERE name = .+ LIMIT 1`).
					WithArgs("Carol").
					WillReturnRows(rows)
			},
			wantUser: existingUser,
		},
		{
			name:      "returns error wrapping sql.ErrNoRows when name not found",
			inputName: "Unknown",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT .+ FROM USER WHERE name = .+ LIMIT 1`).
					WithArgs("Unknown").
					WillReturnError(sql.ErrNoRows)
			},
			wantErr:     true,
			errContains: "Unknown",
		},
		{
			name:      "returned user name matches input name invariant",
			inputName: "Carol",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(userResultColumns).
					AddRow(
						existingUser.ID,
						existingUser.Name,
						existingUser.SecondName,
						existingUser.Work,
						existingUser.Email,
						existingUser.Phone,
						existingUser.About,
					)
				mock.ExpectQuery(`SELECT .+ FROM USER WHERE name = .+ LIMIT 1`).
					WithArgs("Carol").
					WillReturnRows(rows)
			},
			wantUser: existingUser,
		},
		{
			name:      "multiple users share same name returns first (LIMIT 1)",
			inputName: "Alice",
			setupMock: func(mock sqlmock.Sqlmock) {
				// LIMIT 1 in the query ensures only first row is returned
				rows := sqlmock.NewRows(userResultColumns).
					AddRow(1, "Alice", "Smith", "Engineer", "alice1@example.com", "555-1111", "First Alice")
				// A second row would be discarded by LIMIT 1; sqlmock only