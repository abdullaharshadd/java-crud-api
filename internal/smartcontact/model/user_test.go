```go
package model_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"internal/smartcontact/model"
)

// helpers

func strPtr(s string) *string { return &s }

// ---------------------------------------------------------------------------
// NewUser / zero-value construction
// ---------------------------------------------------------------------------

func TestUser_Construction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		buildUser    func() *model.User
		wantID       int64
		wantName     *string
		wantEmail    *string
		wantPassword *string
		wantRole     *string
		wantAbout    *string
	}{
		{
			name: "no-args constructor (zero value)",
			buildUser: func() *model.User {
				return &model.User{}
			},
			wantID:       0,
			wantName:     nil,
			wantEmail:    nil,
			wantPassword: nil,
			wantRole:     nil,
			wantAbout:    nil,
		},
		{
			name: "all-args constructor via NewUser",
			buildUser: func() *model.User {
				return model.NewUser(
					strPtr("Alice"),
					strPtr("alice@example.com"),
					strPtr("s3cr3t"),
					strPtr("admin"),
					strPtr("about alice"),
				)
			},
			wantID:       0, // ID not provided through NewUser
			wantName:     strPtr("Alice"),
			wantEmail:    strPtr("alice@example.com"),
			wantPassword: strPtr("s3cr3t"),
			wantRole:     strPtr("admin"),
			wantAbout:    strPtr("about alice"),
		},
		{
			name: "builder pattern – subset of fields",
			buildUser: func() *model.User {
				return model.NewUser(
					strPtr("Bob"),
					nil,
					nil,
					strPtr("viewer"),
					nil,
				)
			},
			wantID:       0,
			wantName:     strPtr("Bob"),
			wantEmail:    nil,
			wantPassword: nil,
			wantRole:     strPtr("viewer"),
			wantAbout:    nil,
		},
		{
			name: "setting id explicitly",
			buildUser: func() *model.User {
				u := model.NewUser(
					strPtr("Charlie"),
					strPtr("charlie@example.com"),
					strPtr("pass"),
					strPtr("user"),
					strPtr("about charlie"),
				)
				u.ID = 42
				return u
			},
			wantID:       42,
			wantName:     strPtr("Charlie"),
			wantEmail:    strPtr("charlie@example.com"),
			wantPassword: strPtr("pass"),
			wantRole:     strPtr("user"),
			wantAbout:    strPtr("about charlie"),
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			u := tc.buildUser()

			assert.Equal(t, tc.wantID, u.ID)
			assert.Equal(t, tc.wantName, u.Name)
			assert.Equal(t, tc.wantEmail, u.Email)
			assert.Equal(t, tc.wantPassword, u.Password)
			assert.Equal(t, tc.wantRole, u.Role)
			assert.Equal(t, tc.wantAbout, u.About)
		})
	}
}

// ---------------------------------------------------------------------------
// Getter / setter semantics (direct field access in Go)
// ---------------------------------------------------------------------------

func TestUser_GettersSetters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func(u *model.User)
		validate func(t *testing.T, u *model.User)
	}{
		{
			name: "getter on never-set string fields returns nil",
			setup: func(u *model.User) {
				// do nothing – zero value
			},
			validate: func(t *testing.T, u *model.User) {
				assert.Equal(t, int64(0), u.ID)
				assert.Nil(t, u.Name)
				assert.Nil(t, u.Email)
				assert.Nil(t, u.Password)
				assert.Nil(t, u.Role)
				assert.Nil(t, u.About)
			},
		},
		{
			name: "set then get id",
			setup: func(u *model.User) {
				u.ID = 99
			},
			validate: func(t *testing.T, u *model.User) {
				assert.Equal(t, int64(99), u.ID)
			},
		},
		{
			name: "set then get name",
			setup: func(u *model.User) {
				u.Name = strPtr("Dave")
			},
			validate: func(t *testing.T, u *model.User) {
				require.NotNil(t, u.Name)
				assert.Equal(t, "Dave", *u.Name)
			},
		},
		{
			name: "set then get email",
			setup: func(u *model.User) {
				u.Email = strPtr("dave@example.com")
			},
			validate: func(t *testing.T, u *model.User) {
				require.NotNil(t, u.Email)
				assert.Equal(t, "dave@example.com", *u.Email)
			},
		},
		{
			name: "set then get password",
			setup: func(u *model.User) {
				u.Password = strPtr("hunter2")
			},
			validate: func(t *testing.T, u *model.User) {
				require.NotNil(t, u.Password)
				assert.Equal(t, "hunter2", *u.Password)
			},
		},
		{
			name: "set then get role",
			setup: func(u *model.User) {
				u.Role = strPtr("moderator")
			},
			validate: func(t *testing.T, u *model.User) {
				require.NotNil(t, u.Role)
				assert.Equal(t, "moderator", *u.Role)
			},
		},
		{
			name: "set then get about",
			setup: func(u *model.User) {
				u.About = strPtr("some description")
			},
			validate: func(t *testing.T, u *model.User) {
				require.NotNil(t, u.About)
				assert.Equal(t, "some description", *u.About)
			},
		},
		{
			name: "overwrite field preserves latest value",
			setup: func(u *model.User) {
				u.Name = strPtr("first")
				u.Name = strPtr("second")
			},
			validate: func(t *testing.T, u *model.User) {
				require.NotNil(t, u.Name)
				assert.Equal(t, "second", *u.Name)
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			u := &model.User{}
			tc.setup(u)
			tc.validate(t, u)
		})
	}
}

// ---------------------------------------------------------------------------
// Validate
// ---------------------------------------------------------------------------

func TestUser_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		user      *model.User
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "valid user – non-blank name",
			user:    model.NewUser(strPtr("Alice"), strPtr("a@b.com"), strPtr("pw"), strPtr("user"), strPtr("about")),
			wantErr: false,
		},
		{
			name:      "nil name fails validation",
			user:      model.NewUser(nil, strPtr("a@b.com"), strPtr("pw"), strPtr("user"), strPtr("about")),
			wantErr:   true,
			errSubstr: "please Add the department Name",
		},
		{
			name:      "empty string name fails validation",
			user:      model.NewUser(strPtr(""), strPtr("a@b.com"), strPtr("pw"), strPtr("user"), strPtr("about")),
			wantErr:   true,
			errSubstr: "please Add the department Name",
		},
		{
			name:      "whitespace-only name fails validation",
			user:      model.NewUser(strPtr("   "), strPtr("a@b.com"), strPtr("pw"), strPtr("user"), strPtr("about")),
			wantErr:   true,
			errSubstr: "please Add the department Name",
		},
		{
			name:      "tab-only name fails validation",
			user:      model.NewUser(strPtr("\t"), strPtr("a@b.com"), strPtr("pw"), strPtr("user"), strPtr("about")),
			wantErr:   true,
			errSubstr: "please Add the department Name",
		},
		{
			name:      "newline-only name fails validation",
			user:      model.NewUser(strPtr("\n"), strPtr("a@b.com"), strPtr("pw"), strPtr("user"), strPtr("about")),
			wantErr:   true,
			errSubstr: "please Add the department Name",
		},
		{
			name:    "name with leading/trailing spaces but non-blank content passes",
			user:    model.NewUser(strPtr("  Alice  "), nil, nil, nil, nil),
			wantErr: false,
		},
		{
			name:    "all optional fields nil with valid name",
			user:    model.NewUser(strPtr("OnlyName"), nil, nil, nil, nil),
			wantErr: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.user.Validate()
			if tc.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errSubstr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// EnsureUserSchema (uses sqlmock)
// ---------------------------------------------------------------------------

func TestEnsureUserSchema(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		mockSetup func(mock sqlmock.Sqlmock)
		wantErr   bool
		errSubstr string
	}{
		{
			name: "schema creation succeeds",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`CREATE TABLE IF NOT EXISTS users`).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: false,
		},
		{
			name: "database error propagates",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`CREATE TABLE IF NOT EXISTS users`).
					WillReturnError(assert.AnError)
			},
			wantErr:   true,
			errSubstr: "create users table",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
			require.NoError(t, err)
			defer db.Close()

			tc.mockSetup(mock)

			err = model.EnsureUserSchema(context.Background(), db)
			if tc.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errSubstr)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// DB-level constraint simulations (unique email, about length, auto-id)
// ---------------------------------------------------------------------------

func TestUser_DBConstraints(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		mockSetup func(mock sqlmock.Sqlmock)
		action    func(db interface{ ExecContext(context.Context, string, ...interface{}) (interface{ LastInsertId() (int64, error) }, error) }) error
		run       func(t *testing.T, mock sqlmock.Sqlmock)
	}{
		{
			name: "duplicate email causes unique constraint violation",
			run: func(t *testing.T, mock sqlmock.Sqlmock) {
				db, mk, err := sqlmock.New()
				require.NoError(t, err)
				defer db.Close()

				// First insert succeeds
				mk.ExpectExec(`INSERT INTO users`).
					WithArgs("Alice", "alice@example.com", "pw", "user", "about").
					WillReturnResult(sqlmock.NewResult(1, 1))

				// Second insert with same email fails with unique constraint error
				mk.ExpectExec(`INSERT INTO users`).
					WithArgs("Bob", "alice@example.com", "pw2", "user", "about2").
					WillReturnError(fmt.Errorf("ERROR: duplicate key value violates unique constraint \"users_email_key\""))

				_, err = db.ExecContext(context.Background(),
					`INSERT INTO users (name, email, password, role, about) VALUES (?, ?, ?, ?, ?)`,
					"Alice", "alice@example.com", "pw", "user", "about")
				assert.NoError(t, err)

				_, err = db.ExecContext(context.Background(),
					`INSERT INTO users (name, email, password, role, about) VALUES (?, ?, ?, ?, ?)`,
					"Bob", "alice@example.com", "pw2", "user", "about2")
				require.Error(t, err)
				assert.Contains(t, err.Error(), "duplicate key value")

				assert.NoError(t, mk.ExpectationsWereMet())
			},
		},
		{
			name: "persisting new user without explicit id auto-generates id",
			run: func(t *testing.T, mock sqlmock.Sqlmock) {
				db, mk, err := sqlmock.New()
				require.NoError(t, err)
				defer db.Close()

				mk.ExpectExec(`INSERT INTO users`).
					WithArgs("Carol", "carol@example.com", "pw", "user", "about").
					WillReturnResult(sqlmock.NewResult(7, 1)) // auto-generated id = 7

				res, err := db.ExecContext(context.Background(),
					`INSERT INTO users (name, email, password, role, about) VALUES (?, ?, ?, ?, ?)`,
					"Carol", "carol@example.com", "pw", "user", "about")
				require.NoError(t, err)

				lastID, err := res.LastInsertId()
				require.NoError(t, err)
				assert.Equal(t, int64(7), lastID, "auto-generated id should be assigned by persistence layer")

				assert.NoError(t, mk.ExpectationsWereMet())
			},
		},
		{
			name: "about field exceeding 500 chars causes column length constraint violation",
			run: func(t *testing.T, mock sqlmock.Sqlmock) {
				db, mk, err