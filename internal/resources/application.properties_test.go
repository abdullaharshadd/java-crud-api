package resources

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/spf13/viper"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) SaveUser(user *repository.User) (*repository.User, error) {
	args := m.Called(user)
	return args.Get(0).(*repository.User), args.Error(1)
}

func (m *MockUserRepository) FetchUserList() ([]*repository.User, error) {
	args := m.Called()
	return args.Get(0).([]*repository.User), args.Error(1)
}

func (m *MockUserRepository) FetchUserByID(id int) (*repository.User, error) {
	args := m.Called(id)
	return args.Get(0).(*repository.User), args.Error(1)
}

func (m *MockUserRepository) UpdateUser(user *repository.User) (*repository.User, error) {
	args := m.Called(user)
	return args.Get(0).(*repository.User), args.Error(1)
}

func (m *MockUserRepository) DeleteUser(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) FindUserByName(name string) (*repository.User, error) {
	args := m.Called(name)
	return args.Get(0).(*repository.User), args.Error(1)
}

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(user *repository.User) (*repository.User, error) {
	args := m.Called(user)
	return args.Get(0).(*repository.User), args.Error(1)
}

func (m *MockUserService) ListUsers() ([]*repository.User, error) {
	args := m.Called()
	return args.Get(0).([]*repository.User), args.Error(1)
}

func (m *MockUserService) GetUserByID(id int) (*repository.User, error) {
	args := m.Called(id)
	return args.Get(0).(*repository.User), args.Error(1)
}

func (m *MockUserService) UpdateUser(user *repository.User) (*repository.User, error) {
	args := m.Called(user)
	return args.Get(0).(*repository.User), args.Error(1)
}

func (m *MockUserService) DeleteUser(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserService) FindUserByName(name string) (*repository.User, error) {
	args := m.Called(name)
	return args.Get(0).(*repository.User), args.Error(1)
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name          string
		expectedError bool
	}{
		{"ValidConfig", false},
		{"InvalidConfig", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			viper.SetConfigName("application")
			viper.SetConfigType("properties")
			viper.AddConfigPath("./test-fixtures/")
			viper.AutomaticEnv()

			err := LoadConfig()
			if tt.expectedError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestBuildApplication(t *testing.T) {
	tests := []struct {
		name                string
		expectedPort        string
		expectedDBURL       string
		expectedDBUsername  string
		expectedDBPassword  string
		expectedJDBCClass   string
		expectedHibernateDDL string
		expectedHibernateDia string
	}{
		{
			name:                "ValidConfiguration",
			expectedPort:        "8082",
			expectedDBURL:       "jdbc:mysql://localhost:3306/barcode",
			expectedDBUsername:  "root",
			expectedDBPassword:  "root",
			expectedJDBCClass:   "com.mysql.cj.jdbc.Driver",
			expectedHibernateDDL: "update",
			expectedHibernateDia: "MySQL8Dialect",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			viper.SetConfigName("application")
			viper.SetConfigType("properties")
			viper.AddConfigPath("./test-fixtures/")
			viper.AutomaticEnv()

			viper.Set("server.port", tt.expectedPort)
			viper.Set("spring.datasource.url", tt.expectedDBURL)
			viper.Set("spring.datasource.username", tt.expectedDBUsername)
			viper.Set("spring.datasource.password", tt.expectedDBPassword)
			viper.Set("spring.datasource.driverClassName", tt.expectedJDBCClass)
			viper.Set("spring.jpa.hibernate.ddl-auto", tt.expectedHibernateDDL)
			viper.Set("spring.jpa.properties.hibernate.dialect", tt.expectedHibernateDia)

			mockDB := &sqlx.DB{}
			mockRepo := new(MockUserRepository)
			mockService := new(MockUserService)
			mockController := handler.NewUserController(mockService)

			r := chi.NewRouter()
			r.Use(middleware.Logger)
			r.Use(middleware.Recoverer)

			r.Route("/admin", func(r chi.Router) {
				r.Post("/users", mockController.SaveUserHandler)
				r.Get("/users", mockController.FetchUserListHandler)
				r.Get("/users/{id:[0-9]+}", mockController.FetchUserByIDHandler)
				r.Put("/users/{id:[0-9]+}", mockController.UpdateUserHandler)
				r.Delete("/users/{id:[0-9]+}", mockController.DeleteUserHandler)
				r.Get("/users/name/{name}", mockController.FindUserByNameHandler)
			})

			server := &http.Server{
				Addr:           ":" + tt.expectedPort,
				ReadTimeout:    10 * time.Second,
				WriteTimeout:   10 * time.Second,
				MaxHeaderBytes: 1 << 20,
				Handler:        r,
			}

			go func() {
				require.NoError(t, server.ListenAndServe())
			}()

			resp, err := http.Get("http://localhost:" + tt.expectedPort + "/admin/users")
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			require.NoError(t, server.Shutdown(ctx))
		})
	}
}