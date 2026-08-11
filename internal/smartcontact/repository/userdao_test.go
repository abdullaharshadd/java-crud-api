```go
package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	smartErr "github.com/smartContact/internal/smartcontact/error"
	"github.com/smartContact/internal/smartcontact/model"
	"github.com/smartContact/internal/smartcontact/repository"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func newMockRepo(t *testing.T) (*repository.PostgresUserRepository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo, err := repository.NewPostgresUserRepository(db)
	require.NoError(t, err)
	return repo, mock
}

func isUserNotFound(err error) bool {
	var e *smartErr.UserNotFoundError
	return errors.As(err, &e)
}

// ---------------------------------------------------------------------------
// NewPostgresUserRepository
// ---------------------------------------------------------------------------

func TestNewPostgresUserRepository(t *testing.T) {
	t.Parallel()
	t.Run("nil db returns error", func(t *testing.T) {
		repo, err := repository.NewPostgresUserRepository(nil)
		assert.Nil(t, repo)
		assert.Error(t, err)
	})
	t.Run("valid db returns repo", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()
		repo, err := repository.NewPostgresUserRepository(db)
		assert.NoError(t, err)
		assert.NotNil(t, repo)
	})
}

// ---------------------------------------------------------------------------
// Save
// ---------------------------------------------------------------------------

func TestSave(t *testing.T) {
	t.Parallel()

	insertRE := regexp.MustCompile(`INSERT INTO users`)
	updateRE := regexp.MustCompile(`UPDATE users`)

	tests := []struct {
		name      string
		user      *model.User
		setup     func(mock sqlmock.Sqlmock)
		wantID    int
		wantErr   bool
		wantNFErr bool
	}{
		{
			name: "insert new user – zero id – returns user with generated id",
			user: &model.User{Name: "alice", Password: "secret"},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(insertRE.String()).
					WithArgs("alice", "secret").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(42))
			},
			wantID:  42,
			wantErr: false,
		},
		{
			name: "update existing user – non-zero id – returns updated user",
			user: &model.User{ID: 7, Name: "bob", Password: "pass"},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(updateRE.String()).
					WithArgs("bob", "pass", 7).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantID:  7,
			wantErr: false,
		},
		{
			name: "update non-existent user – returns UserNotFoundError",
			user: &model.User{ID: 99, Name: "ghost", Password: "pass"},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(updateRE.String()).
					WithArgs("ghost", "pass", 99).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr:   true,
			wantNFErr: true,
		},
		{
			name:    "nil user returns error",
			user:    nil,
			setup:   func(_ sqlmock.Sqlmock) {},
			wantErr: true,
		},
		{
			name: "db error on insert returns wrapped error",
			user: &model.User{Name: "carol", Password: "pw"},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(insertRE.String()).
					WithArgs("carol", "pw").
					WillReturnError(errors.New("connection refused"))
			},
			wantErr: true,
		},
		{
			name: "db error on update returns wrapped error",
			user: &model.User{ID: 3, Name: "dave", Password: "pw"},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(updateRE.String()).
					WithArgs("dave", "pw", 3).
					WillReturnError(errors.New("connection refused"))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			repo, mock := newMockRepo(t)
			tc.setup(mock)

			got, err := repo.Save(context.Background(), tc.user)
			if tc.wantErr {
				require.Error(t, err)
				if tc.wantNFErr {
					assert.True(t, isUserNotFound(err), "expected UserNotFoundError, got %T: %v", err, err)
				}
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, tc.wantID, got.ID)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// FindByID
// ---------------------------------------------------------------------------

func TestFindByID(t *testing.T) {
	t.Parallel()

	queryRE := regexp.MustCompile(`SELECT id, name, password`)

	tests := []struct {
		name      string
		id        int
		setup     func(mock sqlmock.Sqlmock)
		wantUser  *model.User
		wantErr   bool
		wantNFErr bool
	}{
		{
			name: "existing id returns user",
			id:   1,
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "password"}).
					AddRow(1, "alice", "secret")
				mock.ExpectQuery(queryRE.String()).
					WithArgs(1).
					WillReturnRows(rows)
			},
			wantUser: &model.User{ID: 1, Name: "alice", Password: "secret"},
		},
		{
			name: "non-existent id returns UserNotFoundError",
			id:   999,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(queryRE.String()).
					WithArgs(999).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr:   true,
			wantNFErr: true,
		},
		{
			name: "db error returns wrapped error",
			id:   2,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(queryRE.String()).
					WithArgs(2).
					WillReturnError(errors.New("timeout"))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			repo, mock := newMockRepo(t)
			tc.setup(mock)

			got, err := repo.FindByID(context.Background(), tc.id)
			if tc.wantErr {
				require.Error(t, err)
				if tc.wantNFErr {
					assert.True(t, isUserNotFound(err), "expected UserNotFoundError, got %T: %v", err, err)
				}
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, tc.wantUser.ID, got.ID)
			assert.Equal(t, tc.wantUser.Name, got.Name)
			assert.Equal(t, tc.wantUser.Password, got.Password)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// FindByName
// ---------------------------------------------------------------------------

func TestFindByName(t *testing.T) {
	t.Parallel()

	queryRE := regexp.MustCompile(`SELECT id, name, password`)

	tests := []struct {
		name      string
		lookupName string
		setup     func(mock sqlmock.Sqlmock)
		wantUser  *model.User
		wantErr   bool
		wantNFErr bool
	}{
		{
			name:       "exact name match returns user",
			lookupName: "alice",
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "password"}).
					AddRow(1, "alice", "secret")
				mock.ExpectQuery(queryRE.String()).
					WithArgs("alice").
					WillReturnRows(rows)
			},
			wantUser: &model.User{ID: 1, Name: "alice", Password: "secret"},
		},
		{
			name:       "no user with provided name returns UserNotFoundError",
			lookupName: "nobody",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(queryRE.String()).
					WithArgs("nobody").
					WillReturnError(sql.ErrNoRows)
			},
			wantErr:   true,
			wantNFErr: true,
		},
		{
			name:       "null name lookup – no matching row returns UserNotFoundError",
			lookupName: "",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(queryRE.String()).
					WithArgs("").
					WillReturnError(sql.ErrNoRows)
			},
			wantErr:   true,
			wantNFErr: true,
		},
		{
			name:       "null name lookup – row exists returns user",
			lookupName: "",
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "password"}).
					AddRow(5, "", "pw")
				mock.ExpectQuery(queryRE.String()).
					WithArgs("").
					WillReturnRows(rows)
			},
			wantUser: &model.User{ID: 5, Name: "", Password: "pw"},
		},
		{
			name:       "db error returns wrapped error",
			lookupName: "error-case",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(queryRE.String()).
					WithArgs("error-case").
					WillReturnError(errors.New("network error"))
			},
			wantErr: true,
		},
		// Multiple-row scenario: QueryRowContext only scans the first row; the
		// driver mock delivers multiple rows but sql.Row.Scan reads only the
		// first one. Document that FindByName returns the first result when
		// duplicates exist (the spec says the query "may fail" – in this
		// implementation it does NOT panic, it silently returns the first row).
		{
			name:       "multiple users share same name – returns first result",
			lookupName: "dup",
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "password"}).
					AddRow(10, "dup", "pw1").
					AddRow(11, "dup", "pw2")
				mock.ExpectQuery(queryRE.String()).
					WithArgs("dup").
					WillReturnRows(rows)
			},
			// The implementation uses QueryRowContext which returns only the
			// first row; no error is raised at the driver level.
			wantUser: &model.User{ID: 10, Name: "dup", Password: "pw1"},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			repo, mock := newMockRepo(t)
			tc.setup(mock)

			got, err := repo.FindByName(context.Background(), tc.lookupName)
			if tc.wantErr {
				require.Error(t, err)
				if tc.wantNFErr {
					assert.True(t, isUserNotFound(err), "expected UserNotFoundError, got %T: %v", err, err)
				}
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, tc.wantUser.ID, got.ID)
			assert.Equal(t, tc.wantUser.Name, got.Name)
			assert.Equal(t, tc.wantUser.Password, got.Password)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// FindAll
// ---------------------------------------------------------------------------

func TestFindAll(t *testing.T) {
	t.Parallel()

	queryRE := regexp.MustCompile(`SELECT id, name, password`)

	tests := []struct {
		name      string
		setup     func(mock sqlmock.Sqlmock)
		wantCount int
		wantErr   bool
	}{
		{
			name: "multiple persisted users returns all",
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "password"}).
					AddRow(1, "alice", "pw1").
					AddRow(2, "bob", "pw2").
					AddRow(3, "carol", "pw3")
				mock.ExpectQuery(queryRE.String()).
					WillReturnRows(rows)
			},
			wantCount: 3,
		},
		{
			name: "no persisted users returns empty slice (not nil)",
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "password"})
				mock.ExpectQuery(queryRE.String()).
					WillReturnRows(rows)
			},
			wantCount: 0,
		},
		{
			name: "db error on query returns wrapped error",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(queryRE.String()).
					WillReturnError(errors.New("db down"))
			},
			wantErr: true,
		},
		{
			name: "row scan error returns wrapped error",
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "password"}).
					AddRow("not-an-int", "alice", "pw")
				mock.ExpectQuery(queryRE.String()).
					WillReturnRows(rows)
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			repo, mock := newMockRepo(t)
			tc.setup(mock)

			got, err := repo.FindAll(context.Background())
			if tc.wantErr {
				require.Error(t, err