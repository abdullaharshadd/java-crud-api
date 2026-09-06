package errorHandling

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"migrated-app/internal/smartcontact/model"
)

type mockHandler struct{}

func (mh *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	panic(&model.UserNotFoundException{Message: "User not found"})
}

func TestErrorMiddleware(t *testing.T) {
	tests := []struct {
		name         string
		handler      http.Handler
		expectedCode int
		expectedBody string
	}{
		{
			name:         "UserNotFoundException",
			handler:      &mockHandler{},
			expectedCode: http.StatusNotFound,
			expectedBody: `{"status":404,"message":"User not found"}`,
		},
		{
			name:         "OtherException",
			handler:      http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { panic("Some other error") }),
			expectedCode: http.StatusInternalServerError,
			expectedBody: http.StatusText(http.StatusInternalServerError),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/test", nil)
			recorder := httptest.NewRecorder()

			ErrorMiddleware(tt.handler).ServeHTTP(recorder, req)

			assert.Equal(t, tt.expectedCode, recorder.Code)
			assert.JSONEq(t, tt.expectedBody, recorder.Body.String())
		})
	}
}

func TestHandleUserNotFoundException(t *testing.T) {
	err := &model.UserNotFoundException{Message: "User not found"}
	expectedBody := `{"status":404,"message":"User not found"}`
	expectedCode := http.StatusNotFound

	testCases := []struct {
		name          string
		inputErr      *model.UserNotFoundException
		expectedCode  int
		expectedBody  string
		expectedPanic bool
	}{
		{
			name:          "User not found",
			inputErr:      err,
			expectedCode:  expectedCode,
			expectedBody:  expectedBody,
			expectedPanic: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			handleUserNotFoundException(w, tc.inputErr)

			assert.Equal(t, tc.expectedCode, w.Code)
			assert.JSONEq(t, tc.expectedBody, w.Body.String())
		})
	}
}