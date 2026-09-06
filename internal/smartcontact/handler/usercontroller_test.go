package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockUserService struct {
	UserService
}

func (m *MockUserService) SaveUser(user User) error {
	return nil
}

func (m *MockUserService) FetchUserList() ([]User, error) {
	return []User{}, nil
}

func (m *MockUserService) FetchUserById(id int) (*User, error) {
	return nil, ErrUserNotFound
}

func (m *MockUserService) DeleteUser(id int) error {
	return nil
}

func (m *MockUserService) UpdateUser(id int, user User) (*User, error) {
	return &user, nil
}

func (m *MockUserService) GetUserNameByName(name string) (*User, error) {
	return nil, ErrUserNotFound
}

var testUserService = &MockUserService{}

func TestSaveUser(t *testing.T) {
	type args struct {
		body []byte
	}
	tests := []struct {
		name     string
		args     args
		wantCode int
		wantBody string
	}{
		{
			name: "valid user",
			args: args{
				body: []byte(`{"name": "John Doe", "email": "john@example.com"}`),
			},
			wantCode: http.StatusOK,
			wantBody: `{"message":"User data saved successfully!"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userController := NewUserController(testUserService)
			r := userController.Router()
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/save_user_data", json.NewReader(tt.args.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantCode, w.Code)
			assert.JSONEq(t, tt.wantBody, w.Body.String())
		})
	}
}

func TestFetchUserList(t *testing.T) {
	tests := []struct {
		name     string
		wantCode int
		wantBody string
	}{
		{
			name: "users exist",
			wantCode: http.StatusOK,
			wantBody: `[]`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userController := NewUserController(testUserService)
			r := userController.Router()
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/get_user_data", nil)
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantCode, w.Code)
			assert.JSONEq(t, tt.wantBody, w.Body.String())
		})
	}
}

func TestFetchUserById(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		wantCode     int
		wantBody     string
		wantErr      bool
		expectedErr  error
	}{
		{
			name:    "invalid user id",
			url:     "/get_user_data/1",
			wantCode: http.StatusNotFound,
			wantBody: `{"message":"User not found"}`,
			wantErr: true,
			expectedErr: ErrUserNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userController := NewUserController(testUserService)
			r := userController.Router()
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tt.url, nil)
			r.ServeHTTP(w, req)
			if tt.wantErr {
				assert.Equal(t, tt.wantCode, w.Code)
				assert.JSONEq(t, tt.wantBody, w.Body.String())
			} else {
				assert.Equal(t, tt.wantCode, w.Code)
			}
		})
	}
}

func TestDeleteUser(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		wantCode int
		wantBody string
	}{
		{
			name:     "valid user id",
			url:      "/delete_user_data/1",
			wantCode: http.StatusOK,
			wantBody: `{"message":"user data deleted Successfully"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userController := NewUserController(testUserService)
			r := userController.Router()
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("DELETE", tt.url, nil)
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantCode, w.Code)
			assert.JSONEq(t, tt.wantBody, w.Body.String())
		})
	}
}

func TestUpdateUser(t *testing.T) {
	type args struct {
		url  string
		body []byte
	}
	tests := []struct {
		name     string
		args     args
		wantCode int
		wantBody string
	}{
		{
			name: "valid user update",
			args: args{
				url:  "/update_user_data/1",
				body: []byte(`{"name": "John Doe", "email": "john@example.com"}`),
			},
			wantCode: http.StatusOK,
			wantBody: `{"name":"John Doe","email":"john@example.com"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userController := NewUserController(testUserService)
			r := userController.Router()
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("PUT", tt.args.url, json.NewReader(tt.args.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantCode, w.Code)
			assert.JSONEq(t, tt.wantBody, w.Body.String())
		})
	}
}

func TestGetUserNameByName(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		wantCode     int
		wantBody     string
		wantErr      bool
		expectedErr  error
	}{
		{
			name:    "user name does not exist",
			url:     "/get_user_name/name/John Doe",
			wantCode: http.StatusNotFound,
			wantBody: `null`,
			wantErr: true,
			expectedErr: ErrUserNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userController := NewUserController(testUserService)
			r := userController.Router()
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tt.url, nil)
			r.ServeHTTP(w, req)
			if tt.wantErr {
				assert.Equal(t, tt.wantCode, w.Code)
				assert.JSONEq(t, tt.wantBody, w.Body.String())
			} else {
				assert.Equal(t, tt.wantCode, w.Code)
			}
		})
	}
}