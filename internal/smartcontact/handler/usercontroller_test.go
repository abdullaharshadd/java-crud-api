package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"migrated-app/internal/smartcontact/error"
	"migrated-app/internal/smartcontact/model"
	"migrated-app/internal/smartcontact/service"
)

type mockUserService struct {
	service.UserService
	saveErr   error
	fetchErr  error
	updateErr error
	deleteErr error
	findErr   error
}

func (ms *mockUserService) SaveUser(ctx context.Context, user *model.User) (*model.User, error) {
	if ms.saveErr != nil {
		return nil, ms.saveErr
	}
	return user, nil
}

func (ms *mockUserService) FetchUserList(ctx context.Context) ([]*model.User, error) {
	if ms.fetchErr != nil {
		return nil, ms.fetchErr
	}
	return []*model.User{{ID: 1, Name: "Test User"}}, nil
}

func (ms *mockUserService) FetchUserByID(ctx context.Context, id int) (*model.User, error) {
	if ms.fetchErr == usererror.NewUserNotFoundError() {
		return nil, ms.fetchErr
	}
	if ms.fetchErr != nil {
		return nil, ms.fetchErr
	}
	return &model.User{ID: id, Name: "Test User"}, nil
}

func (ms *mockUserService) DeleteUser(ctx context.Context, id int) error {
	if ms.deleteErr != nil {
		return ms.deleteErr
	}
	return nil
}

func (ms *mockUserService) UpdateUser(ctx context.Context, id int, user *model.User) error {
	if ms.updateErr != nil {
		return ms.updateErr
	}
	return nil
}

func (ms *mockUserService) FindByName(ctx context.Context, name string) (*model.User, error) {
	if ms.findErr == usererror.NewUserNotFoundError() {
		return nil, ms.findErr
	}
	if ms.findErr != nil {
		return nil, ms.findErr
	}
	return &model.User{Name: name}, nil
}

func TestSaveUserHandler(t *testing.T) {
	tests := []struct {
		name         string
		input        model.User
		expectedCode int
		expectedBody string
		mock         func() service.UserService
	}{
		{
			name:         "Valid User",
			input:        model.User{Name: "John Doe"},
			expectedCode: http.StatusOK,
			expectedBody: `{"message":"User data saved successfully!"}`,
			mock: func() service.UserService {
				return &mockUserService{}
			},
		},
		{
			name:         "Invalid JSON",
			input:        model.User{},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"error":"EOF"}`,
			mock: func() service.UserService {
				return &mockUserService{}
			},
		},
		{
			name:         "Database Error",
			input:        model.User{Name: "John Doe"},
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"error":"database error"}`,
			mock: func() service.UserService {
				return &mockUserService{saveErr: fmt.Errorf("database error")}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewUserController(tt.mock())
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest("POST", "/save_user_data", bytes.NewBuffer(json.Marshal(tt.input)))

			uc.SaveUserHandler(c)

			assert.Equal(t, tt.expectedCode, w.Code)
			assert.Equal(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestFetchUserListHandler(t *testing.T) {
	tests := []struct {
		name         string
		expectedCode int
		expectedBody string
		mock         func() service.UserService
	}{
		{
			name:         "Users Present",
			expectedCode: http.StatusOK,
			expectedBody: `[{"ID":1,"Name":"Test User"}]`,
			mock: func() service.UserService {
				return &mockUserService{}
			},
		},
		{
			name:         "No Users",
			expectedCode: http.StatusOK,
			expectedBody: "[]",
			mock: func() service.UserService {
				return &mockUserService{fetchErr: fmt.Errorf("no users found")}
			},
		},
		{
			name:         "Internal Server Error",
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"error":"internal server error"}`,
			mock: func() service.UserService {
				return &mockUserService{fetchErr: fmt.Errorf("internal server error")}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewUserController(tt.mock())
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest("GET", "/get_user_data", nil)

			uc.FetchUserListHandler(c)

			assert.Equal(t, tt.expectedCode, w.Code)
			assert.Equal(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestFetchUserByIDHandler(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedCode int
		expectedBody string
		mock         func() service.UserService
	}{
		{
			name:         "Valid ID",
			input:        "1",
			expectedCode: http.StatusOK,
			expectedBody: `{"ID":1,"Name":"Test User"}`,
			mock: func() service.UserService {
				return &mockUserService{}
			},
		},
		{
			name:         "Invalid ID",
			input:        "abc",
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"error":"Invalid user ID"}`,
			mock: func() service.UserService {
				return &mockUserService{}
			},
		},
		{
			name:         "User Not Found",
			input:        "2",
			expectedCode: http.StatusNotFound,
			expectedBody: `{"error":"user not found"}`,
			mock: func() service.UserService {
				return &mockUserService{fetchErr: usererror.NewUserNotFoundError()}
			},
		},
		{
			name:         "Internal Server Error",
			input:        "3",
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"error":"internal server error"}`,
			mock: func() service.UserService {
				return &mockUserService{fetchErr: fmt.Errorf("internal server error")}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewUserController(tt.mock())
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = append(c.Params, gin.Param{"id", tt.input})

			uc.FetchUserByIDHandler(c)

			assert.Equal(t, tt.expectedCode, w.Code)
			assert.Equal(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestDeleteUserHandler(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedCode int
		expectedBody string
		mock         func() service.UserService
	}{
		{
			name:         "Valid ID",
			input:        "1",
			expectedCode: http.StatusOK,
			expectedBody: `{"message":"User data deleted successfully!"}`,
			mock: func() service.UserService {
				return &mockUserService{}
			},
		},
		{
			name:         "Invalid ID",
			input:        "abc",
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"error":"Invalid user ID"}`,
			mock: func() service.UserService {
				return &mockUserService{}
			},
		},
		{
			name:         "Database Error",
			input:        "2",
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"error":"database error"}`,
			mock: func() service.UserService {
				return &mockUserService{deleteErr: fmt.Errorf("database error")}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewUserController(tt.mock())
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = append(c.Params, gin.Param{"id", tt.input})

			uc.DeleteUserHandler(c)

			assert.Equal(t, tt.expectedCode, w.Code)
			assert.Equal(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestUpdateUserHandler(t *testing.T) {
	tests := []struct {
		name         string
		inputID      string
		inputUser    model.User
		expectedCode int
		expectedBody string
		mock         func() service.UserService
	}{
		{
			name:         "Valid Input",
			inputID:      "1",
			inputUser:    model.User{Name: "Updated User"},
			expectedCode: http.StatusOK,
			expectedBody: `{"ID":1,"Name":"Updated User"}`,
			mock: func() service.UserService {
				return &mockUserService{}
			},
		},
		{
			name:         "Invalid ID",
			inputID:      "abc",
			inputUser:    model.User{Name: "Updated User"},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"error":"Invalid user ID"}`,
			mock: func() service.UserService {
				return &mockUserService{}
			},
		},
		{
			name:         "Database Error",
			inputID:      "2",
			inputUser:    model.User{Name: "Updated User"},
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"error":"database error"}`,
			mock: func() service.UserService {
				return &mockUserService{updateErr: fmt.Errorf("database error")}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewUserController(tt.mock())
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = append(c.Params, gin.Param{"id", tt.inputID})
			c.Request, _ = http.NewRequest("PUT", "/update_user_data/"+tt.inputID, bytes.NewBuffer(json.Marshal(tt.inputUser)))

			uc.UpdateUserHandler(c)

			assert.Equal(t, tt.expectedCode, w.Code)
			assert.Equal(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestFindUserByNameHandler(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedCode int
		expectedBody string
		mock         func() service.UserService
	}{
		{
			name:         "User Found",
			input:        "John Doe",
			expectedCode: http.StatusOK,
			expectedBody: `{"Name":"John Doe"}`,
			mock: func() service.UserService {
				return &mockUserService{}
			},
		},
		{
			name:         "User Not Found",
			input:        "Non Existing User",
			expectedCode: http.StatusNotFound,
			expectedBody: `{"error":"user not found"}`,
			mock: func() service.UserService {
				return &mockUserService{findErr: usererror.NewUserNotFoundError()}
			},
		},
		{
			name:         "Internal Server Error",
			input:        "Error User",
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"error":"internal server error"}`,
			mock: func() service.UserService {
				return &mockUserService{findErr: fmt.Errorf("internal server error")}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewUserController(tt.mock())
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = append(c.Params, gin.Param{"name", tt.input})

			uc.FindUserByNameHandler(c)

			assert.Equal(t, tt.expectedCode, w.Code)
			assert.Equal(t, tt.expectedBody, w.Body.String())
		})
	}
}