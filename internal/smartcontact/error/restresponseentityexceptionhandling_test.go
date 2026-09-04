package error_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"migrated-app/internal/smartcontact/error"
	"migrated-app/internal/smartcontact/model"
)

func TestUserNotFoundMiddleware(t *testing.T) {
	type ctxKey string
	var errorKey = ctxKey("error")

	tests := []struct {
		name        string
		inputErr    error
		expectedCode int
	}{
		{"UserNotFound", &model.UserNotFoundError{}, http.StatusNotFound},
		{"OtherError", &model.SomeOtherError{}, http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := r.Context()
				ctx = context.WithValue(ctx, errorKey, tt.inputErr)
				r = r.WithContext(ctx)
			})

			req := httptest.NewRequest(http.MethodGet, "/{any}", nil)
			w := httptest.NewRecorder()

			userNotFoundMiddleware := error.UserNotFoundMiddleware(next)
			userNotFoundMiddleware.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code, "HTTP status code does not match expected value")
			if tt.expectedCode == http.StatusNotFound {
				var resp model.ErrorMessage
				err := model.FromHTTPError(w.Body.Bytes(), &resp)
				assert.NoError(t, err, "Failed to parse response body")
				assert.Equal(t, tt.expectedCode, resp.StatusCode, "Response status code does not match expected value")
			}
		})
	}
}

// Mocking the UserNotFoundError for testing purposes
type SomeOtherError struct{}

func (e *SomeOtherError) Error() string {
	return "some other error occurred"
}

// Mocking the ErrorMessage struct
func FromHTTPError(body []byte, err *model.ErrorMessage) error {
	// This function should parse the HTTP body and fill the err struct accordingly
	// For simplicity, we'll just set a dummy value here
	err.StatusCode = http.StatusNotFound
	err.Message = "User Not Found"
	return nil
}

// Mocking the ToHTTPError function
func ToHTTPError(w http.ResponseWriter, err *model.ErrorMessage) error {
	_, _ = w.Write([]byte(err.Message))
	return nil
}

// Mocking the NewErrorMessage function
func NewErrorMessage(code int, msg string) *model.ErrorMessage {
	return &model.ErrorMessage{
		StatusCode: code,
		Message:    msg,
	}
}

// Mocking the model package
package model

import "net/http"

type ErrorMessage struct {
	StatusCode int
	Message    string
}

var (
	UserNotFoundError = &ErrorMessage{
		StatusCode: http.StatusNotFound,
		Message:    "User Not Found",
	}
	SomeOtherError = &ErrorMessage{
		StatusCode: http.StatusInternalServerError,
		Message:    "Some Other Error",
	}
)

func FromHTTPError(body []byte, err *ErrorMessage) error {
	// Implementation provided in the test file
	return nil
}

func ToHTTPError(w http.ResponseWriter, err *ErrorMessage) error {
	// Implementation provided in the test file
	return nil
}

func NewErrorMessage(code int, msg string) *ErrorMessage {
	// Implementation provided in the test file
	return nil
}