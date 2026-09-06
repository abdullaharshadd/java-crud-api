package model

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

type User struct {
	ID      int
	Name    string
	Email   string
	Password string
	Role    string
	About   string
}

func (u *User) getId() int {
	return u.ID
}

func (u *User) setId(id int) {
	u.ID = id
}

func (u *User) getName() string {
	return u.Name
}

func (u *User) setName(name string) error {
	if name == "" {
		return ErrNameBlank
	}
	u.Name = name
	return nil
}

func (u *User) getEmail() string {
	return u.Email
}

func (u *User) setEmail(email string) error {
	// Assume there's a check for uniqueness and validity here.
	if !isValidEmail(email) || !isUniqueEmail(email) {
		return ErrEmailInvalidOrNotUnique
	}
	u.Email = email
	return nil
}

func (u *User) getPassword() string {
	return u.Password
}

func (u *User) setPassword(password string) error {
	// Assume there's a check for password validity here.
	if !isValidPassword(password) {
		return ErrPasswordInvalid
	}
	u.Password = password
	return nil
}

func (u *User) getRole() string {
	return u.Role
}

func (u *User) setRole(role string) error {
	// Assume there's a check for role validity here.
	if !isValidRole(role) {
		return ErrRoleInvalid
	}
	u.Role = role
	return nil
}

func (u *User) getAbout() string {
	return u.About
}

func (u *User) setAbout(about string) error {
	if len(about) > 500 {
		return ErrAboutTooLong
	}
	// Assume there's a check for about validity here.
	if !isValidAbout(about) {
		return ErrAboutInvalid
	}
	u.About = about
	return nil
}

var ErrNameBlank = errors.New("name cannot be blank")
var ErrEmailInvalidOrNotUnique = errors.New("email is invalid or not unique")
var ErrPasswordInvalid = errors.New("password is invalid")
var ErrRoleInvalid = errors.New("role is invalid")
var ErrAboutTooLong = errors.New("about cannot exceed 500 characters")
var ErrAboutInvalid = errors.New("about is invalid")

func TestGetId(t *testing.T) {
	tests := []struct {
		name     string
		user     User
		expected int
	}{
		{"with ID set", User{ID: 123}, 123},
		{"without ID set", User{}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.user.getId())
		})
	}
}

func TestSetId(t *testing.T) {
	tests := []struct {
		name     string
		user     User
		id       int
		expected User
	}{
		{"valid ID", User{}, 456, User{ID: 456}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.user.setId(tt.id)
			assert.Equal(t, tt.expected, tt.user)
		})
	}
}

func TestGetName(t *testing.T) {
	tests := []struct {
		name     string
		user     User
		expected string
	}{
		{"with name set", User{Name: "Alice"}, "Alice"},
		{"without name set", User{}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.user.getName())
		})
	}
}

func TestSetName(t *testing.T) {
	tests := []struct {
		name        string
		user        User
		newName     string
		expected    User
		expectError bool
	}{
		{"valid name", User{}, "Bob", User{Name: "Bob"}, false},
		{"blank name", User{}, "", User{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.setName(tt.newName)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, tt.user)
			}
		})
	}
}

func TestGetEmail(t *testing.T) {
	tests := []struct {
		name     string
		user     User
		expected string
	}{
		{"with email set", User{Email: "alice@example.com"}, "alice@example.com"},
		{"without email set", User{}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.user.getEmail())
		})
	}
}

func TestSetEmail(t *testing.T) {
	tests := []struct {
		name        string
		user        User
		newEmail    string
		expected    User
		expectError bool
	}{
		{"unique valid email", User{}, "bob@example.com", User{Email: "bob@example.com"}, false},
		{"invalid email", User{}, "not-an-email", User{}, true},
		{"non-unique email", User{}, "alice@example.com", User{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.setEmail(tt.newEmail)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, tt.user)
			}
		})
	}
}

func TestGetPassword(t *testing.T) {
	tests := []struct {
		name     string
		user     User
		expected string
	}{
		{"with password set", User{Password: "secure123"}, "secure123"},
		{"without password set", User{}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.user.getPassword())
		})
	}
}

func TestSetPassword(t *testing.T) {
	tests := []struct {
		name        string
		user        User
		newPassword string
		expected    User
		expectError bool
	}{
		{"valid password", User{}, "SecurePassword123!", User{Password: "SecurePassword123!"}, false},
		{"invalid password", User{}, "weak", User{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.setPassword(tt.newPassword)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, tt.user)
			}
		})
	}
}

func TestGetRole(t *testing.T) {
	tests := []struct {
		name     string
		user     User
		expected string
	}{
		{"with role set", User{Role: "admin"}, "admin"},
		{"without role set", User{}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.user.getRole())
		})
	}
}

func TestSetRole(t *testing.T) {
	tests := []struct {
		name        string
		user        User
		newRole     string
		expected    User
		expectError bool
	}{
		{"valid role", User{}, "user", User{Role: "user"}, false},
		{"invalid role", User{}, "invalid", User{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.setRole(tt.newRole)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, tt.user)
			}
		})
	}
}

func TestGetAbout(t *testing.T) {
	tests := []struct {
		name     string
		user     User
		expected string
	}{
		{"with about set", User{About: "About Alice"}, "About Alice"},
		{"without about set", User{}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.user.getAbout())
		})
	}
}

func TestSetAbout(t *testing.T) {
	tests := []struct {
		name        string
		user        User
		newAbout    string
		expected    User
		expectError bool
	}{
		{"valid about", User{}, "About Bob", User{About: "About Bob"}, false},
		{"too long about", User{}, "x" + "About Bob", User{}, true},
		{"invalid about", User{}, "invalid", User{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.setAbout(tt.newAbout)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, tt.user)
			}
		})
	}
}

func isValidEmail(email string) bool {
	// Placeholder for email validation logic.
	return email != ""
}

func isUniqueEmail(email string) bool {
	// Placeholder for email uniqueness check logic.
	return email != "alice@example.com"
}

func isValidPassword(password string) bool {
	// Placeholder for password validation logic.
	return len(password) >= 8
}

func isValidRole(role string) bool {
	// Placeholder for role validation logic.
	return role == "admin" || role == "user"
}

func isValidAbout(about string) bool {
	// Placeholder for about validation logic.
	return about != "invalid"
}