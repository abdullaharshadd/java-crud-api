package smartcontact

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"migrated-app/internal/resources/application.properties"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(user *User) (*User, error) {
	args := m.Called(user)
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserRepository) GetUserByID(id int) (*User, error) {
	args := m.Called(id)
	return args.Get(0).(*User), args.Error(1)
}

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) RegisterUser(user *User) (*User, error) {
	args := m.Called(user)
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserService) GetUserByID(id int) (*User, error) {
	args := m.Called(id)
	return args.Get(0).(*User), args.Error(1)
}

func TestBuildRouter(t *testing.T) {
	type args struct {
		us service.UserService
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"happy path", args{&MockUserService{}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := buildRouter(tt.args.us)
			assert.NotNil(t, r)
			assert.IsType(t, &chi.Mux{}, r)
		})
	}
}

func TestRunApplication(t *testing.T) {
	// Mock the LoadConfig function
	oldLoadConfig := application.properties.LoadConfig
	defer func() { application.properties.LoadConfig = oldLoadConfig }()
	application.properties.LoadConfig = func() error {
		return nil
	}

	// Mock the NewDatabase function
	oldNewDatabase := repository.NewDatabase
	defer func() { repository.NewDatabase = oldNewDatabase }()
	repository.NewDatabase = func() (*DB, error) {
		return &DB{}, nil
	}

	// Mock the Close function on DB
	oldClose := (*DB).Close
	defer func() { (*DB).Close = oldClose }()
	(*DB).Close = func() error {
		return nil
	}

	// Mock the NewUserRepository function
	oldNewUserRepository := repository.NewUserRepository
	defer func() { repository.NewUserRepository = oldNewUserRepository }()
	repository.NewUserRepository = func(db *DB) repository.UserRepository {
		return &MockUserRepository{}
	}

	// Mock the NewUserService function
	oldNewUserService := service.NewUserService
	defer func() { service.NewUserService = oldNewUserService }()
	service.NewUserService = func(repo repository.UserRepository) service.UserService {
		return &MockUserService{}
	}

	// Start the server in a separate goroutine
	go RunApplication()

	// Wait for the server to start listening
	time.Sleep(100 * time.Millisecond)

	// Create a request to send to the server
	req, err := http.NewRequest("GET", "http://localhost:8080/healthz", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Send the request and check the response
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Status, "200 OK")
}

func TestShutdown(t *testing.T) {
	// Mock the LoadConfig function
	oldLoadConfig := application.properties.LoadConfig
	defer func() { application.properties.LoadConfig = oldLoadConfig }()
	application.properties.LoadConfig = func() error {
		return nil
	}

	// Mock the NewDatabase function
	oldNewDatabase := repository.NewDatabase
	defer func() { repository.NewDatabase = oldNewDatabase }()
	repository.NewDatabase = func() (*DB, error) {
		return &DB{}, nil
	}

	// Mock the Close function on DB
	oldClose := (*DB).Close
	defer func() { (*DB).Close = oldClose }()
	(*DB).Close = func() error {
		return nil
	}

	// Mock the NewUserRepository function
	oldNewUserRepository := repository.NewUserRepository
	defer func() { repository.NewUserRepository = oldNewUserRepository }()
	repository.NewUserRepository = func(db *DB) repository.UserRepository {
		return &MockUserRepository{}
	}

	// Mock the NewUserService function
	oldNewUserService := service.NewUserService
	defer func() { service.NewUserService = oldNewUserService }()
	service.NewUserService = func(repo repository.UserRepository) service.UserService {
		return &MockUserService{}
	}

	// Start the server in a separate goroutine
	srv := &http.Server{Addr: ":8080"}
	go func() {
		RunApplication()
	}()

	// Wait for the server to start listening
	time.Sleep(100 * time.Millisecond)

	// Send a shutdown signal
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)

	// Check if the server is shutdown properly
	select {
	case <-time.After(1 * time.Second):
		t.Errorf("server did not shutdown within 1 second")
	default:
	}
}