package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockUserDAO struct {
	users map[string]User
}

func (m *mockUserDAO) FindByName(name string) (*User, error) {
	user, exists := m.users[name]
	if !exists {
		return nil, nil
	}
	return &user, nil
}

func TestFindByName(t *testing.T) {
	testCases := []struct {
		name           string
		input          string
		expectedOutput *User
	}{
		{
			name:           "User exists",
			input:          "John Doe",
			expectedOutput: &User{Name: "John Doe", ID: 1},
		},
		{
			name:           "User does not exist",
			input:          "Non Existent User",
			expectedOutput: nil,
		},
	}

	mockData := map[string]User{
		"John Doe": {Name: "John Doe", ID: 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDAO := &mockUserDAO{users: mockData}
			user, _ := mockDAO.FindByName(tc.input)
			assert.Equal(t, tc.expectedOutput, user)
		})
	}
}