```go
package model

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// TestNewUser – constructor (all-args replacement)
// ---------------------------------------------------------------------------

func TestNewUser(t *testing.T) {
	tests := []struct {
		name     string
		inName   string
		inEmail  string
		inPwd    string
		inRole   string
		inAbout  string
		wantUser *User
	}{
		{
			name:    "all fields populated",
			inName:  "Alice",
			inEmail: "alice@example.com",
			inPwd:   "s3cr3t",
			inRole:  "ADMIN",
			inAbout: "Test user",
			wantUser: &User{
				ID:       0,
				Name:     "Alice",
				Email:    "alice@example.com",
				Password: "s3cr3t",
				Role:     "ADMIN",
				About:    "Test user",
			},
		},
		{
			name:    "empty strings (mirrors Java no-arg + setters)",
			inName:  "",
			inEmail: "",
			inPwd:   "",
			inRole:  "",
			inAbout: "",
			wantUser: &User{
				ID:       0,
				Name:     "",
				Email:    "",
				Password: "",
				Role:     "",
				About:    "",
			},
		},
		{
			name:    "about at max length boundary",
			inName:  "Bob",
			inEmail: "bob@example.com",
			inPwd:   "pw",
			inRole:  "USER",
			inAbout: strings.Repeat("x", AboutMaxLength),
			wantUser: &User{
				ID:       0,
				Name:     "Bob",
				Email:    "bob@example.com",
				Password: "pw",
				Role:     "USER",
				About:    strings.Repeat("x", AboutMaxLength),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := NewUser(tc.inName, tc.inEmail, tc.inPwd, tc.inRole, tc.inAbout)

			assert.NotNil(t, got)
			assert.Equal(t, tc.wantUser.ID, got.ID, "ID should default to 0")
			assert.Equal(t, tc.wantUser.Name, got.Name)
			assert.Equal(t, tc.wantUser.Email, got.Email)
			assert.Equal(t, tc.wantUser.Password, got.Password)
			assert.Equal(t, tc.wantUser.Role, got.Role)
			assert.Equal(t, tc.wantUser.About, got.About)
		})
	}
}

// ---------------------------------------------------------------------------
// TestUserZeroValue – mirrors Java no-arg constructor
// ---------------------------------------------------------------------------

func TestUserZeroValue(t *testing.T) {
	t.Run("zero value has id=0 and all string fields empty", func(t *testing.T) {
		var u User
		assert.Equal(t, 0, u.ID)
		assert.Equal(t, "", u.Name)
		assert.Equal(t, "", u.Email)
		assert.Equal(t, "", u.Password)
		assert.Equal(t, "", u.Role)
		assert.Equal(t, "", u.About)
	})
}

// ---------------------------------------------------------------------------
// TestUserFieldMutation – mirrors Java setters / getters
// ---------------------------------------------------------------------------

func TestUserFieldMutation(t *testing.T) {
	tests := []struct {
		name     string
		mutate   func(u *User)
		validate func(t *testing.T, u *User)
	}{
		{
			name:   "setId / getId",
			mutate: func(u *User) { u.ID = 42 },
			validate: func(t *testing.T, u *User) {
				assert.Equal(t, 42, u.ID)
			},
		},
		{
			name:   "setName / getName",
			mutate: func(u *User) { u.Name = "Charlie" },
			validate: func(t *testing.T, u *User) {
				assert.Equal(t, "Charlie", u.Name)
			},
		},
		{
			name:   "setEmail / getEmail",
			mutate: func(u *User) { u.Email = "charlie@example.com" },
			validate: func(t *testing.T, u *User) {
				assert.Equal(t, "charlie@example.com", u.Email)
			},
		},
		{
			name:   "setPassword / getPassword",
			mutate: func(u *User) { u.Password = "newpassword" },
			validate: func(t *testing.T, u *User) {
				assert.Equal(t, "newpassword", u.Password)
			},
		},
		{
			name:   "setRole / getRole",
			mutate: func(u *User) { u.Role = "MANAGER" },
			validate: func(t *testing.T, u *User) {
				assert.Equal(t, "MANAGER", u.Role)
			},
		},
		{
			name:   "setAbout / getAbout",
			mutate: func(u *User) { u.About = "some description" },
			validate: func(t *testing.T, u *User) {
				assert.Equal(t, "some description", u.About)
			},
		},
		{
			name: "set all fields then read back",
			mutate: func(u *User) {
				u.ID = 7
				u.Name = "Dana"
				u.Email = "dana@example.com"
				u.Password = "p@ss"
				u.Role = "VIEWER"
				u.About = "viewer account"
			},
			validate: func(t *testing.T, u *User) {
				assert.Equal(t, 7, u.ID)
				assert.Equal(t, "Dana", u.Name)
				assert.Equal(t, "dana@example.com", u.Email)
				assert.Equal(t, "p@ss", u.Password)
				assert.Equal(t, "VIEWER", u.Role)
				assert.Equal(t, "viewer account", u.About)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			u := &User{}
			tc.mutate(u)
			tc.validate(t, u)
		})
	}
}

// ---------------------------------------------------------------------------
// TestUserValidate – Validate() method / @NotBlank replacement
// ---------------------------------------------------------------------------

func TestUserValidate(t *testing.T) {
	tests := []struct {
		name    string
		user    *User
		wantErr error
	}{
		{
			name:    "valid name returns nil",
			user:    NewUser("Alice", "a@b.com", "pw", "ADMIN", "about"),
			wantErr: nil,
		},
		{
			name:    "empty name returns ErrUserNameBlank",
			user:    NewUser("", "a@b.com", "pw", "ADMIN", "about"),
			wantErr: ErrUserNameBlank,
		},
		{
			name:    "whitespace-only name returns ErrUserNameBlank",
			user:    NewUser("   ", "a@b.com", "pw", "ADMIN", "about"),
			wantErr: ErrUserNameBlank,
		},
		{
			name:    "tab-only name returns ErrUserNameBlank",
			user:    NewUser("\t", "a@b.com", "pw", "ADMIN", "about"),
			wantErr: ErrUserNameBlank,
		},
		{
			name:    "newline-only name returns ErrUserNameBlank",
			user:    NewUser("\n", "a@b.com", "pw", "ADMIN", "about"),
			wantErr: ErrUserNameBlank,
		},
		{
			name:    "name with leading and trailing whitespace is valid",
			user:    NewUser("  Alice  ", "a@b.com", "pw", "ADMIN", "about"),
			wantErr: nil,
		},
		{
			name: "name set after construction – blank",
			user: func() *User {
				u := NewUser("ValidName", "x@y.com", "pw", "U", "")
				u.Name = ""
				return u
			}(),
			wantErr: ErrUserNameBlank,
		},
		{
			name: "name set after construction – non-blank",
			user: func() *User {
				u := NewUser("", "x@y.com", "pw", "U", "")
				u.Name = "Bob"
				return u
			}(),
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.user.Validate()
			if tc.wantErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr))
				assert.Equal(t, tc.wantErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestErrUserNameBlank – sentinel error properties
// ---------------------------------------------------------------------------

func TestErrUserNameBlank(t *testing.T) {
	t.Run("error message preserves original Java message", func(t *testing.T) {
		assert.Equal(t, "please Add the department Name", ErrUserNameBlank.Error())
	})

	t.Run("Validate returns the exact sentinel error", func(t *testing.T) {
		u := &User{Name: ""}
		err := u.Validate()
		assert.True(t, errors.Is(err, ErrUserNameBlank))
	})
}

// ---------------------------------------------------------------------------
// TestAboutMaxLength – constant value
// ---------------------------------------------------------------------------

func TestAboutMaxLength(t *testing.T) {
	t.Run("AboutMaxLength equals 500", func(t *testing.T) {
		assert.Equal(t, 500, AboutMaxLength)
	})
}

// ---------------------------------------------------------------------------
// TestUserEquality – mirrors Java @Data equals / hashCode (via struct compare)
// ---------------------------------------------------------------------------

func TestUserEquality(t *testing.T) {
	tests := []struct {
		name      string
		a         User
		b         User
		wantEqual bool
	}{
		{
			name: "identical users are equal",
			a:    User{ID: 1, Name: "Eve", Email: "e@e.com", Password: "p", Role: "R", About: "A"},
			b:    User{ID: 1, Name: "Eve", Email: "e@e.com", Password: "p", Role: "R", About: "A"},
			wantEqual: true,
		},
		{
			name: "differing ID makes users unequal",
			a:    User{ID: 1, Name: "Eve", Email: "e@e.com", Password: "p", Role: "R", About: "A"},
			b:    User{ID: 2, Name: "Eve", Email: "e@e.com", Password: "p", Role: "R", About: "A"},
			wantEqual: false,
		},
		{
			name: "differing Name makes users unequal",
			a:    User{ID: 1, Name: "Eve", Email: "e@e.com", Password: "p", Role: "R", About: "A"},
			b:    User{ID: 1, Name: "Frank", Email: "e@e.com", Password: "p", Role: "R", About: "A"},
			wantEqual: false,
		},
		{
			name: "differing Email makes users unequal",
			a:    User{ID: 1, Name: "Eve", Email: "e@e.com", Password: "p", Role: "R", About: "A"},
			b:    User{ID: 1, Name: "Eve", Email: "f@f.com", Password: "p", Role: "R", About: "A"},
			wantEqual: false,
		},
		{
			name: "differing Password makes users unequal",
			a:    User{ID: 1, Name: "Eve", Email: "e@e.com", Password: "p", Role: "R", About: "A"},
			b:    User{ID: 1, Name: "Eve", Email: "e@e.com", Password: "q", Role: "R", About: "A"},
			wantEqual: false,
		},
		{
			name: "differing Role makes users unequal",
			a:    User{ID: 1, Name: "Eve", Email: "e@e.com", Password: "p", Role: "R", About: "A"},
			b:    User{ID: 1, Name: "Eve", Email: "e@e.com", Password: "p", Role: "S", About: "A"},
			wantEqual: false,
		},
		{
			name: "differing About makes users unequal",
			a:    User{ID: 1, Name: "Eve", Email: "e@e.com", Password: "p", Role: "R", About: "A"},
			b:    User{ID: 1, Name: "Eve", Email: "e@e.com", Password: "p", Role: "R", About: "B"},
			wantEqual: false,
		},
		{
			name:      "zero-value users are equal",
			a:         User{},
			b:         User{},
			wantEqual: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.wantEqual {
				assert.Equal(t, tc.a, tc.b)
			} else {
				assert.NotEqual(t, tc.a, tc.b)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestUserReflexiveEquality – reflexive property
// ---------------------------------------------------------------------------

func TestUserReflexiveEquality(t *testing.T) {
	u := User{ID: 5, Name: "Grace", Email: "g@g.com", Password: "pw", Role: "U", About: "hi"}
	assert.Equal(t, u, u, "a user must equal itself")
}

// ---------------------------------------------------------------------------
// TestUserToString – mirrors Java @Data toString
// ---------------------------------------------------------------------------

func TestUserToString(t *testing.T) {
	tests := []struct {
		name            string
		user            User
		expectedSubstrs []string
	}{
		{
			name: "populated user contains all field values in fmt output",
			user: User{
				ID:       99,
				Name:     "Hank",
				Email:    "hank@example.com",
				Password: "secret",
				Role:     "ADMIN",
				About:    "senior admin",
			},
			// Use %+v which Go's fmt package produces: {ID:99 Name:Hank ...}
			expectedSubstrs: []string{
				"99",
				"Hank",
				"hank@example.com",
				"secret",
				"ADMIN",
				"senior admin",
			},
		},
		{
			name:            "zero-value user contains zero values",
			user:            User{},
			expectedSubstrs: []string{"0"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Go structs have a standard string representation via fmt.Sprintf("%+v", ...).
			// We test that each field value appears in that representation.
			import_fmt_sprintf := func(u User) string {
				// inline helper to avoid import cycle concerns – use fmt directly
				// We cannot call fmt here without importing it, so we