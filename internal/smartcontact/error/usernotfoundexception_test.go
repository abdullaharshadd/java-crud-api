package error_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"internal/smartcontact/error"
)

var testData = []struct {
	name           string
	message        string
	cause          error
	expectedOutput string
}{
	{"no arguments provided", "", nil, "user not found"},
	{"a string message provided", "custom message", nil, "custom message"},
	{"a string message and a Throwable cause provided", "custom message", errors.New("some cause"), "custom message: some cause"},
	{"only a Throwable cause provided", "", errors.New("some cause"), "user not found: some cause"},
	{"a string message, a Throwable cause, and boolean flags for suppression and stack trace provided", "custom message", errors.New("some cause"), "custom message: some cause"},
}

func TestNewUserNotFoundError(t *testing.T) {
	for _, td := range testData {
		t.Run(td.name, func(t *testing.T) {
			unfe := error.NewUserNotFoundError(td.message, td.cause)
			assert.Equal(t, td.expectedOutput, unfe.Error())
			if td.cause != nil {
				assert.Equal(t, td.cause, unfe.Unwrap())
			} else {
				assert.Nil(t, unfe.Unwrap())
			}
		})
	}
}

func TestUserNotFoundError(t *testing.T) {
	defaultMessage := "user not found"
	unfe := error.NewUserNotFoundError("", nil)
	assert.Equal(t, defaultMessage, unfe.Error())
	assert.Nil(t, unfe.Unwrap())
}

func TestUserNotFoundErrorWithCause(t *testing.T) {
	message := "user not found"
	cause := errors.New("database error")
	unfe := error.NewUserNotFoundError(message, cause)
	assert.Equal(t, message+": "+cause.Error(), unfe.Error())
	assert.Equal(t, cause, unfe.Unwrap())
}

func TestUserNotFoundErrorWithMessageOnly(t *testing.T) {
	message := "specific user not found"
	unfe := error.NewUserNotFoundError(message, nil)
	assert.Equal(t, message, unfe.Error())
	assert.Nil(t, unfe.Unwrap())
}

// Assuming there's an HTTP handler that might throw UserNotFoundError
type MockUserService struct{}

func (m *MockUserService) GetUserByID(id int) (*User, error) {
	return nil, error.NewUserNotFoundError("user not found", nil)
}

type User struct {
	ID int
}

func TestHTTPHandlerWithError(t *testing.T) {
	mockService := &MockUserService{}
	handler := func(w http.ResponseWriter, r *http.Request) {
		user, err := mockService.GetUserByID(1)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(user)
	}

	req, _ := http.NewRequest("GET", "/users/1", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "user not found")
}