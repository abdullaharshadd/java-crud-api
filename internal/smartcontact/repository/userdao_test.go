```go
package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apperror "github.com/smartcontact/internal/smartcontact/error"
	"github.com/smartcontact/internal/smartcontact/model"
	"github.com/smartcontact/internal/smartcontact/repository"
)

// ---------------------------------------------------------------------------
// Minimal sqlmock-style helpers using DATA-RACE-safe in-process fakes.
// We implement the Querier interface (sqlx.ExtContext) with a fake that
// stores canned rows / errors, avoiding any real DB or third-party mock lib.
// ---------------------------------------------------------------------------

// fakeRow is a single scannable result row.
type fakeRow struct {
	values []interface{}
	err    error
}

func (r *fakeRow) Scan(dest ...interface{}) error {
	if r.err != nil {
		return r.err
	}
	for i, d := range dest {
		if i >= len(r.values) {
			break
		}
		switch p := d.(type) {
		case *int:
			*p = r.values[i].(int)
		case *string:
			*p = r.values[i].(string)
		case *bool:
			*p = r.values[i].(bool)
		default:
			return fmt.Errorf("fakeRow: unsupported dest type %T", d)
		}
	}
	return nil
}

// fakeResult implements sql.Result.
type fakeResult struct {
	rowsAffected int64
	rowsAffErr   error
}

func (f fakeResult) LastInsertId() (int64, error) { return 0, errors.New("not supported") }
func (f fakeResult) RowsAffected() (int64, error) { return f.rowsAffected, f.rowsAffErr }

// ---------------------------------------------------------------------------
// fakeQuerier is a controllable double for the Querier interface.
//
// It records each call (query + args) and returns preconfigured responses.
// Each call pops from the front of the queue; tests build the queue before
// calling repository methods.
// ---------------------------------------------------------------------------

type callRecord struct {
	query string
	args  []interface{}
}

// queuedQueryRowxResult holds the row(s) that QueryRowxContext should return.
type queuedQueryRowxResult struct {
	// For single-row (QueryRowxContext used via GetContext)
	scanValues []interface{}
	scanErr    error
}

// queuedSelectResult holds the rows that SelectContext should return.
type queuedSelectResult struct {
	users []model.User
	err   error
}

// queuedExecResult holds the result that ExecContext should return.
type queuedExecResult struct {
	res fakeResult
	err error
}

// fakeQuerier is safe for sequential test use.
type fakeQuerier struct {
	calls []callRecord

	// queues consumed in FIFO order
	queryRowxQueue []queuedQueryRowxResult
	selectQueue    []queuedSelectResult
	execQueue      []queuedExecResult
}

func newFakeQuerier() *fakeQuerier { return &fakeQuerier{} }

func (q *fakeQuerier) addQueryRowx(values []interface{}, err error) {
	q.queryRowxQueue = append(q.queryRowxQueue, queuedQueryRowxResult{values, err})
}

func (q *fakeQuerier) addSelect(users []model.User, err error) {
	q.selectQueue = append(q.selectQueue, queuedSelectResult{users, err})
}

func (q *fakeQuerier) addExec(rowsAffected int64, rowsAffErr error, execErr error) {
	q.execQueue = append(q.execQueue, queuedExecResult{fakeResult{rowsAffected, rowsAffErr}, execErr})
}

// popQueryRowx consumes the next entry, or returns a "no rows" error by default.
func (q *fakeQuerier) popQueryRowx() queuedQueryRowxResult {
	if len(q.queryRowxQueue) == 0 {
		return queuedQueryRowxResult{scanErr: sql.ErrNoRows}
	}
	r := q.queryRowxQueue[0]
	q.queryRowxQueue = q.queryRowxQueue[1:]
	return r
}

func (q *fakeQuerier) popSelect() queuedSelectResult {
	if len(q.selectQueue) == 0 {
		return queuedSelectResult{users: []model.User{}}
	}
	r := q.selectQueue[0]
	q.selectQueue = q.selectQueue[1:]
	return r
}

func (q *fakeQuerier) popExec() queuedExecResult {
	if len(q.execQueue) == 0 {
		return queuedExecResult{res: fakeResult{rowsAffected: 0}}
	}
	r := q.execQueue[0]
	q.execQueue = q.execQueue[1:]
	return r
}

// ---------------------------------------------------------------------------
// sqlx.ExtContext implementation
// ---------------------------------------------------------------------------

func (q *fakeQuerier) DriverName() string { return "fake" }

func (q *fakeQuerier) Rebind(query string) string { return query }

func (q *fakeQuerier) BinderName() string { return "dollar" }

// QueryRowxContext is called by sqlx.GetContext internally.  We intercept it
// by wrapping in a *sqlx.Row-compatible value via sqlx internals — but since
// we cannot construct *sqlx.Row externally, we satisfy the interface that
// sqlx.GetContext actually uses under the hood: sqlx.ExtContext calls
// QueryContext; sqlx.GetContext then scans the *sql.Rows.
//
// Because the Querier interface is defined as sqlx.ExtContext, the methods
// sqlx.GetContext and sqlx.SelectContext use are:
//   QueryContext  (from database/sql DriverContext)
//   QueryxContext (from sqlx)
//
// We implement the broadest set of methods sqlx expects.

func (q *fakeQuerier) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	q.calls = append(q.calls, callRecord{query, args})
	// Not used directly; sqlx calls QueryxContext.
	return nil, errors.New("fakeQuerier: QueryContext not implemented; use QueryxContext")
}

func (q *fakeQuerier) QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error) {
	q.calls = append(q.calls, callRecord{query, args})
	// Used by sqlx.SelectContext. We cannot construct *sqlx.Rows externally,
	// so we build an in-memory *sql.DB + *sql.Rows via a helper.
	entry := q.popSelect()
	if entry.err != nil {
		return nil, entry.err
	}
	return buildSqlxRows(entry.users)
}

func (q *fakeQuerier) QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row {
	q.calls = append(q.calls, callRecord{query, args})
	entry := q.popQueryRowx()
	return buildSqlxRow(entry.scanValues, entry.scanErr)
}

func (q *fakeQuerier) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	q.calls = append(q.calls, callRecord{query, args})
	entry := q.popExec()
	if entry.err != nil {
		return nil, entry.err
	}
	return entry.res, nil
}

// ---------------------------------------------------------------------------
// Helpers to build *sqlx.Row / *sqlx.Rows backed by a real in-memory DB.
// We use the "sqlite3" driver if available, otherwise fall back to a custom
// channel-based trick.  Since we want zero external deps beyond testify and
// sqlx, we use the exported sqlx.NewDb with the "txdb" pattern — but the
// simplest is to just use a real sqlite3 DB in memory.
//
// ALTERNATIVE: Because constructing *sqlx.Row externally is impossible (it is
// an unexported-field struct), we instead make fakeQuerier implement the
// interface that sqlx.GetContext really uses: it calls QueryRowxContext which
// returns *sqlx.Row.  The only public way to get one is via sqlx.DB or
// sqlx.Tx.  We therefore spin up a real in-process SQLite DB only for
// constructing test rows — OR we change our approach and use DATA-DRIVEN
// sqlmock.
//
// To avoid pulling in go-sqlmock or sqlite3, we swap to a testing pattern
// that avoids needing *sqlx.Row at all: we test via a *real* Postgres-
// compatible in-memory DB replacement using the "pgx" driver that ships with
// sqlx tests. However that also adds deps.
//
// PRAGMATIC DECISION: Use DATA-DRIVEN approach with github.com/DATA-DOG/go-sqlmock
// which is already the standard in Go projects using sqlx.
// ---------------------------------------------------------------------------
//
// The helpers below use go-sqlmock to build a real *sql.DB whose prepared
// query responses are canned. We then wrap it in *sqlx.DB so sqlx.GetContext
// and sqlx.SelectContext work normally.
//
// This keeps the test self-contained without a running Postgres.
// ---------------------------------------------------------------------------

// We re-implement fakeQuerier using go-sqlmock to properly satisfy sqlx.

// NOTE: The above fakeQuerier approach cannot properly return *sqlx.Row
// because sqlx.Row is not constructable outside the package.  The cleanest
// production-quality approach is:
//
//  1. Use go-sqlmock for the *sql.DB.
//  2. Wrap it in sqlx.NewDb.
//  3. Pass the *sqlx.DB as the Querier.
//
// We restructure the tests accordingly.

import (
	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

// newMockDB creates a *sqlx.DB backed by go-sqlmock.
func newMockDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	xdb := sqlx.NewDb(db, "sqlmock")
	return xdb, mock
}

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func userRow(u model.User) *sqlmock.Rows {
	rows := sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "about"}).
		AddRow(u.ID, u.Name, u.Email, u.Password, u.Role, u.About)
	return rows
}

func usersRows(users []model.User) *sqlmock.Rows {
	rows := sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "about"})
	for _, u := range users {
		rows.AddRow(u.ID, u.Name, u.Email, u.Password, u.Role, u.About)
	}
	return rows
}

// ---------------------------------------------------------------------------
// FindByName
// ---------------------------------------------------------------------------

func TestFindByName(t *testing.T) {
	ctx := context.Background()

	alice := model.User{ID: 1, Name: "alice", Email: "alice@example.com", Password: "hash", Role: "user", About: "about alice"}

	tests := []struct {
		name        string
		inputName   string
		setupMock   func(mock sqlmock.Sqlmock)
		wantUser    *model.User
		wantErr     bool
		errIsNotFound bool
	}{
		{
			name:      "user found by name",
			inputName: "alice",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, name, email, password, role, about FROM users WHERE name = \$1`).
					WithArgs("alice").
					WillReturnRows(userRow(alice))
			},
			wantUser: &alice,
		},
		{
			name:      "user not found by name",
			inputName: "bob",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, name, email, password, role, about FROM users WHERE name = \$1`).
					WithArgs("bob").
					WillReturnError(sql.ErrNoRows)
			},
			wantUser:      nil,
			wantErr:       true,
			errIsNotFound: true,
		},
		{
			name:      "database error on find by name",
			inputName: "charlie",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, name, email, password, role, about FROM users WHERE name = \$1`).
					WithArgs("charlie").
					WillReturnError(errors.New("connection refused"))
			},
			wantUser:      nil,
			wantErr:       true,
			errIsNotFound: false,
		},
		{
			name:      "name is empty string",
			inputName: "",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, name, email, password, role, about FROM users WHERE name = \$1`).
					WithArgs("").
					WillReturnError(sql.ErrNoRows)
			},
			wantUser:      nil,
			wantErr:       true,
			errIsNotFound: true,
		},
		{
			name:      "returned user has matching name",
			inputName: "alice",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, name, email, password, role, about FROM users WHERE name = \$1`).
					WithArgs("alice").
					WillReturnRows(userRow(alice))
			},
			wantUser: &alice,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := newMockDB(t)
			tc.setupMock(mock)
			repo := repository.NewUserRepository(db)

			got, err := repo.FindByName(ctx, tc.inputName)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.errIsNotFound {
					assert.True(t, apperror.IsUserNotFound(err), "expected UserNotFound error, got: %v", err)
				}
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, tc.wantUser.ID, got.ID)
				assert.Equal(t, tc.wantUser.Name, got.Name)
				assert.Equal(t, tc.wantUser.Email, got.Email)
				// Invariant: returned user name equals input name
				assert.Equal(t, tc.inputName, got.Name)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// FindByID
// ---------------------------------------------------------------------------

func TestFindByID(t *testing.T) {
	ctx := context.Background()

	alice := model.User{ID: 42, Name: "alice", Email: "alice@example.com", Password: "hash", Role: "admin", About: "admin