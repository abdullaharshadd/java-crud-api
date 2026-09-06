package repository

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"migrated-app/internal/smartcontact/model"
)

type mockDB struct {
	mockQueryRow func(query string, args ...interface{}) *mockRow
}

type mockRow struct {
	scanErr error
	user    *model.User
}

func (mr *mockRow) Scan(dest ...interface{}) error {
	if mr.scanErr != nil {
		return mr.scanErr
	}
	if mr.user == nil {
		return sql.ErrNoRows
	}
	dest[0].(*int64) = &mr.user.ID
	dest[1].(*string) = &mr.user.Name
	dest[2].(*string) = &mr.user.Email
	dest[3].(*string) = &mr.user.Password
	dest[4].(*string) = &mr.user.Role
	dest[5].(*string) = &mr.user.About
	return nil
}

func (mdb *mockDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *mockRow {
	return mdb.mockQueryRow(query, args...)
}

var findByNameTests = []struct {
	name           string
	inputName      string
	mockQueryRow   func(query string, args ...interface{}) *mockRow
	expectedOutput *model.User
	expectedError  bool
}{
	{
		name:       "Existing user",
		inputName:  "john_doe",
		mockQueryRow: func(query string, args ...interface{}) *mockRow {
			return &mockRow{user: &model.User{ID: 1, Name: "john_doe", Email: "john@example.com", Password: "hashed_password", Role: "user", About: "About John Doe"}}
		},
		expectedOutput: &model.User{ID: 1, Name: "john_doe", Email: "john@example.com", Password: "hashed_password", Role: "user", About: "About John Doe"},
		expectedError:  false,
	},
	{
		name:       "Non-existing user",
		inputName:  "non_existent_user",
		mockQueryRow: func(query string, args ...interface{}) *mockRow {
			return &mockRow{scanErr: sql.ErrNoRows}
		},
		expectedOutput: nil,
		expectedError:  true,
	},
	{
		name:       "Database error",
		inputName:  "error_user",
		mockQueryRow: func(query string, args ...interface{}) *mockRow {
			return &mockRow{scanErr: errors.New("db error")}
		},
		expectedOutput: nil,
		expectedError:  true,
	},
}

func TestUserRepository_FindByName(t *testing.T) {
	for _, tt := range findByNameTests {
		t.Run(tt.name, func(t *testing.T) {
			mdb := &mockDB{mockQueryRow: tt.mockQueryRow}
			ur := newUserRepository(mdb)
			user, err := ur.FindByName(context.Background(), tt.inputName)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectedOutput != nil {
				assert.Equal(t, tt.expectedOutput, user)
			} else {
				assert.Nil(t, user)
			}
		})
	}
}