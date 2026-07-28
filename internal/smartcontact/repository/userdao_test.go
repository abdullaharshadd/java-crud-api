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

	smartcontacterror "github.com/smartContact/internal/smartcontact/error"
	"github.com/smartContact/internal/smartcontact/model"
	"github.com/smartContact/internal/smartcontact/repository"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func newMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db, mock
}

// userColumns mirrors the selectColumns constant so the mock rows match.
var userColumns = []string{
	"id", "name", "second_name", "work", "email",
	"phone_number", "image", "description", "user_name", "password",
}

func addUserRow(rows *sqlmock.Rows, u model.UserResponse, password string) *sqlmock.Rows {
	return rows.AddRow(
		u.ID, u.Name, u.SecondName, u.Work, u.Email,
		u.PhoneNumber, u.Image, u.Description, u.UserName, password,
	)
}

// sampleResponse returns a fully-populated UserResponse for test assertions.
func sampleResponse(id int) model.UserResponse {
	return model.UserResponse{
		ID:          id,
		Name:        "Alice",
		SecondName:  "Smith",
		Work:        "Engineer",
		Email:       "alice@example.com",
		PhoneNumber: "555-0100",
		Image:       "avatar.png",
		Description: "Test user",
		UserName:    "asmith",
	}
}

// sampleUser returns a model.User that mirrors sampleResponse.
func sampleUser(id int) model.User {
	return model.User{
		ID:          id,
		Name:        "Alice",
		SecondName:  "Smith",
		Work:        "Engineer",
		Email:       "alice@example.com",
		PhoneNumber: "555-0100",
		Image:       "avatar.png",
		Description: "Test user",
		UserName:    "asmith",
		Password:    "secret",
	}
}

// isUserNotFound returns true when err carries a UserNotFoundError.
func isUserNotFound(err error) bool {
	var nfe *smartcontacterror.UserNotFoundError
	return errors.As(err, &nfe)
}

// ---------------------------------------------------------------------------
// FindAll
// ---------------------------------------------------------------------------

func TestUserDao_FindAll(t *testing.T) {
	selectQ := "SELECT id, name, second_name, work, email, phone_number, image, description, user_name, password FROM users ORDER BY id"

	tests := []struct {
		name      string
		setup     func(mock sqlmock.Sqlmock)
		wantLen   int
		wantErr   bool
		wantUsers []model.UserResponse
	}{
		{
			name: "users exist – returns all",
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(userColumns)
				addUserRow(rows, sampleResponse(1), "secret1")
				addUserRow(rows, sampleResponse(2), "secret2")
				mock.ExpectQuery(selectQ).WillReturnRows(rows)
			},
			wantLen: 2,
		},
		{
			name: "no users – returns empty slice (not nil)",
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(userColumns)
				mock.ExpectQuery(selectQ).WillReturnRows(rows)
			},
			wantLen: 0,
		},
		{
			name: "db error – propagates error",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(selectQ).WillReturnError(errors.New("connection lost"))
			},
			wantErr: true,
		},
		{
			name: "scan error – propagates error",
			setup: func(mock sqlmock.Sqlmock) {
				// Return a row with one column instead of ten to force a scan error.
				rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
				mock.ExpectQuery(selectQ).WillReturnRows(rows)
			},
			wantErr: true,
		},
		{
			name: "NULL columns – tolerated by NullString",
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(userColumns).AddRow(
					3, nil, nil, nil, nil, nil, nil, nil, nil, nil,
				)
				mock.ExpectQuery(selectQ).WillReturnRows(rows)
			},
			wantLen: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := newMock(t)
			tc.setup(mock)

			dao := repository.NewUserDao(db)
			got, err := dao.FindAll(context.Background())

			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, got, tc.wantLen)
			// Invariant: never returns nil slice when no error
			assert.NotNil(t, got)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// FindByID
// ---------------------------------------------------------------------------

func TestUserDao_FindByID(t *testing.T) {
	selectQ := "SELECT id, name, second_name, work, email, phone_number, image, description, user_name, password FROM users WHERE id = \\$1"

	tests := []struct {
		name        string
		id          int
		setup       func(mock sqlmock.Sqlmock)
		wantErr     bool
		wantNotFound bool
		wantResp    model.UserResponse
	}{
		{
			name: "user found",
			id:   1,
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(userColumns)
				addUserRow(rows, sampleResponse(1), "secret")
				mock.ExpectQuery(selectQ).WithArgs(1).WillReturnRows(rows)
			},
			wantResp: sampleResponse(1),
		},
		{
			name: "user not found – returns UserNotFoundError",
			id:   99,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(selectQ).WithArgs(99).WillReturnError(sql.ErrNoRows)
			},
			wantErr:      true,
			wantNotFound: true,
		},
		{
			name: "db error – propagates error",
			id:   1,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(selectQ).WithArgs(1).WillReturnError(errors.New("timeout"))
			},
			wantErr: true,
		},
		{
			name: "NULL columns – tolerated",
			id:   5,
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(userColumns).AddRow(
					5, nil, nil, nil, nil, nil, nil, nil, nil, nil,
				)
				mock.ExpectQuery(selectQ).WithArgs(5).WillReturnRows(rows)
			},
			wantResp: model.UserResponse{ID: 5},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := newMock(t)
			tc.setup(mock)

			dao := repository.NewUserDao(db)
			got, err := dao.FindByID(context.Background(), tc.id)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tc.wantNotFound, isUserNotFound(err),
					"UserNotFoundError mismatch")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantResp, got)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// FindByName
// ---------------------------------------------------------------------------

func TestUserDao_FindByName(t *testing.T) {
	selectQ := "SELECT id, name, second_name, work, email, phone_number, image, description, user_name, password FROM users WHERE name = \\$1"

	tests := []struct {
		name         string
		inputName    string
		setup        func(mock sqlmock.Sqlmock)
		wantErr      bool
		wantNotFound bool
		wantResp     model.UserResponse
	}{
		{
			name:      "user exists with given name",
			inputName: "Alice",
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(userColumns)
				addUserRow(rows, sampleResponse(1), "secret")
				mock.ExpectQuery(selectQ).WithArgs("Alice").WillReturnRows(rows)
			},
			wantResp: sampleResponse(1),
		},
		{
			name:      "no user with given name – UserNotFoundError",
			inputName: "Ghost",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(selectQ).WithArgs("Ghost").WillReturnError(sql.ErrNoRows)
			},
			wantErr:      true,
			wantNotFound: true,
		},
		{
			name:      "db error – propagates",
			inputName: "Alice",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(selectQ).WithArgs("Alice").WillReturnError(errors.New("disk full"))
			},
			wantErr: true,
		},
		{
			name:      "name is empty string – still queries DB",
			inputName: "",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(selectQ).WithArgs("").WillReturnError(sql.ErrNoRows)
			},
			wantErr:      true,
			wantNotFound: true,
		},
		{
			name:      "returned user name matches input – invariant",
			inputName: "Alice",
			setup: func(mock sqlmock.Sqlmock) {
				resp := sampleResponse(7)
				resp.Name = "Alice"
				rows := sqlmock.NewRows(userColumns)
				addUserRow(rows, resp, "pw")
				mock.ExpectQuery(selectQ).WithArgs("Alice").WillReturnRows(rows)
			},
			wantResp: func() model.UserResponse {
				r := sampleResponse(7)
				r.Name = "Alice"
				return r
			}(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := newMock(t)
			tc.setup(mock)

			dao := repository.NewUserDao(db)
			got, err := dao.FindByName(context.Background(), tc.inputName)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tc.wantNotFound, isUserNotFound(err),
					"UserNotFoundError mismatch")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantResp, got)
			// Invariant: returned name equals input name
			assert.Equal(t, tc.inputName, got.Name)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// Merge (insert + update)
// ---------------------------------------------------------------------------

func TestUserDao_Merge_Insert(t *testing.T) {
	insertQ := `INSERT INTO users`
	selectQ := "SELECT id, name, second_name, work, email, phone_number, image, description, user_name, password FROM users WHERE id = \\$1"

	tests := []struct {
		name     string
		user     model.User
		setup    func(mock sqlmock.Sqlmock)
		wantErr  bool
		wantResp model.UserResponse
	}{
		{
			name: "insert new user (ID==0) – returns persisted user with generated id",
			user: sampleUser(0),
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(insertQ).
					WithArgs(
						"Alice", "Smith", "Engineer", "alice@example.com",
						"555-0100", "avatar.png", "Test user", "asmith", "secret",
					).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(42))

				// FindByID call after insert
				rows := sqlmock.NewRows(userColumns)
				addUserRow(rows, sampleResponse(42), "secret")
				mock.ExpectQuery(selectQ).WithArgs(42).WillReturnRows(rows)
			},
			wantResp: sampleResponse(42),
		},
		{
			name: "insert db error – propagates",
			user: sampleUser(0),
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(insertQ).
					WillReturnError(errors.New("unique violation"))
			},
			wantErr: true,
		},
		{
			name: "insert ok but subsequent FindByID fails – propagates error",
			user: sampleUser(0),
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(insertQ).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(99))
				mock.ExpectQuery(selectQ).WithArgs(99).WillReturnError(errors.New("network"))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := newMock(t)
			tc.setup(mock)

			dao := repository.NewUserDao(db)
			got, err := dao.Merge(context.Background(), tc.user)

			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantResp, got)
			// Invariant: returned entity reflects persisted state
			assert.NotZero(t, got.ID)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserDao_Merge_Update(t *testing.T) {
	updateQ := `UPDATE users SET`
	selectQ := "SELECT id, name, second_name, work, email, phone_number, image, description, user_name, password FROM users WHERE id = \\$1"

	tests := []struct {
		name         string
		user         model.User
		setup        func(mock sqlmock.Sqlmock)
		wantErr      bool
		wantNotFound bool
		wantResp     model.UserResponse
	}{
		{
			name: "update existing user – returns updated user",
			user: sampleUser(10),
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec