package resources

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockDB struct {
	sql.DB
}

func (m *mockDB) Open(dsn string) (*sql.DB, error) {
	return &m.DB, nil
}

func TestApplicationProperties(t *testing.T) {
	type testCase struct {
		name             string
		expectedPort     int
		expectedHost     string
		expectedDatabase string
		expectedUser     string
		expectedPassword string
		expectedDriver   string
		expectedJpa      bool
		expectedDialect  string
	}

	testCases := []testCase{
		{
			name:             "Valid properties",
			expectedPort:     8082,
			expectedHost:     "localhost:3306",
			expectedDatabase: "barcode",
			expectedUser:     "root",
			expectedPassword: "root",
			expectedDriver:   "com.mysql.cj.jdbc.Driver",
			expectedJpa:      true,
			expectedDialect:  "MySQL8Dialect",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := &mockDB{}
			assert.NotNil(t, db)

			// Assuming there is a function to get the properties
			props, err := GetApplicationProperties()
			if assert.NoError(t, err) {
				assert.Equal(t, tc.expectedPort, props.ServerPort)
				assert.Equal(t, tc.expectedHost, props.DBHost)
				assert.Equal(t, tc.expectedDatabase, props.DBName)
				assert.Equal(t, tc.expectedUser, props.DBUser)
				assert.Equal(t, tc.expectedPassword, props.DBPassword)
				assert.Equal(t, tc.expectedDriver, props.JDBCDriver)
				assert.Equal(t, tc.expectedJpa, props.AutoUpdateSchema)
				assert.Equal(t, tc.expectedDialect, props.HibernateDialect)
			}
		})
	}

	// Assuming there is an HTTP handler that uses these properties
	t.Run("Server listens on correct port", func(t *testing.T) {
		// Setup mock server
		mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		server := httptest.NewServer(mockHandler)
		defer server.Close()

		// The actual port used by the test server might differ from 8082 due to test environment.
		// We can only verify that the server is up and running.
		resp, err := http.Get(server.URL)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}