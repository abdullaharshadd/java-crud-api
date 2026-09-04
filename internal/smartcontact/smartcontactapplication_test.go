package smartcontact

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/jmoiron/sqlx"
)

type mockDB struct {
	mock.Mock
}

func (m *mockDB) PingContext(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return m.Called(query, args).Get(1).(error)
}

func TestBuildRouter(t *testing.T) {
	tests := []struct {
		name     string
		db       *mockDB
		expected int
	}{
		{"Success", &mockDB{}, http.StatusOK},
		{"DB Error", &mockDB{Mock: mock.Mock{On: func(method string, ctx context.Context) *mock.Call { return mock.Mocker(m).On(method, ctx).Return(fmt.Errorf("db error")) }}}, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.db.On("PingContext", mock.Anything).Return(nil)
			tt.db.On("Exec", mock.Anything, mock.Anything).Return(nil, nil)

			router := buildRouter(tt.db)

			req, err := http.NewRequest("GET", "/admin/users", nil)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expected, w.Code)
			tt.db.AssertExpectations(t)
		})
	}
}

func TestInitDatabase(t *testing.T) {
	tests := []struct {
		name     string
		expected *sqlx.DB
		err      error
	}{
		{"Success", &sqlx.DB{}, nil},
		{"Error", nil, fmt.Errorf("failed to connect")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &mockDB{}
			mockDB.On("PingContext", mock.Anything).Return(tt.err)
			mockDB.On("Exec", mock.Anything, mock.Anything).Return(nil, tt.err)

			db, err := initDatabase()
			assert.Equal(t, tt.expected, db)
			assert.Equal(t, tt.err, err)

			mockDB.AssertExpectations(t)
		})
	}
}

func TestStartServer(t *testing.T) {
	tests := []struct {
		name        string
		routerSetup func(*chi.Mux)
		expectedErr error
	}{
		{"Success", func(r *chi.Mux) {}, nil},
		{"Error", func(r *chi.Mux) { r.NotFound(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusInternalServerError) })) }, fmt.Errorf("server forced to shutdown")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := chi.NewRouter()
			tt.routerSetup(router)

			err := startServer(router)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}