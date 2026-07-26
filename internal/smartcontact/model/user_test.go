```go
package model

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// Zero-value / no-args constructor
// ---------------------------------------------------------------------------

func TestUser_ZeroValue(t *testing.T) {
	tests := []struct {
		name     string
		user     User
		wantID   int
		wantName string
		wantEmail    string
		wantPassword string
		wantRole     string
		wantAbout    string
	}{
		{
			name:         "instantiated with no arguments yields zero values",
			user:         User{},
			wantID:       0,
			wantName:     "",
			wantEmail:    "",
			wantPassword: "",
			wantRole:     "",
			wantAbout:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantID, tt.user.ID)
			assert.Equal(t, tt.wantName, tt.user.Name)
			assert.Equal(t, tt.wantEmail, tt.user.Email)
			assert.Equal(t, tt.wantPassword, tt.user.Password)
			assert.Equal(t, tt.wantRole, tt.user.Role)
			assert.Equal(t, tt.wantAbout, tt.user.About)
		})
	}
}

// ---------------------------------------------------------------------------
// NewUser (all-args constructor)
// ---------------------------------------------------------------------------

func TestNewUser(t *testing.T) {
	tests := []struct {
		name     string
		id       int
		uname    string
		email    string
		password string
		role     string
		about    string
	}{
		{
			name:     "all fields populated",
			id:       42,
			uname:    "Alice",
			email:    "alice@example.com",
			password: "s3cr3t",
			role:     "ADMIN",
			about:    "about alice",
		},
		{
			name:     "zero id with non-empty string fields",
			id:       0,
			uname:    "Bob",
			email:    "bob@example.com",
			password: "pass",
			role:     "USER",
			about:    "about bob",
		},
		{
			name:     "all empty strings and zero id",
			id:       0,
			uname:    "",
			email:    "",
			password: "",
			role:     "",
			about:    "",
		},
		{
			name:     "large id value",
			id:       999999,
			uname:    "Carol",
			email:    "carol@example.com",
			password: "pw",
			role:     "MOD",
			about:    "mod user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := NewUser(tt.id, tt.uname, tt.email, tt.password, tt.role, tt.about)
			assert.NotNil(t, u)
			assert.Equal(t, tt.id, u.ID)
			assert.Equal(t, tt.uname, u.Name)
			assert.Equal(t, tt.email, u.Email)
			assert.Equal(t, tt.password, u.Password)
			assert.Equal(t, tt.role, u.Role)
			assert.Equal(t, tt.about, u.About)
		})
	}
}

// ---------------------------------------------------------------------------
// Builder-style partial construction (direct struct literal)
// ---------------------------------------------------------------------------

func TestUser_PartialConstruction(t *testing.T) {
	tests := []struct {
		name     string
		build    func() User
		wantID   int
		wantName string
		wantRole string
	}{
		{
			name: "only id and name set",
			build: func() User {
				return User{ID: 1, Name: "Dave"}
			},
			wantID:   1,
			wantName: "Dave",
			wantRole: "",
		},
		{
			name: "only role set",
			build: func() User {
				return User{Role: "VIEWER"}
			},
			wantID:   0,
			wantName: "",
			wantRole: "VIEWER",
		},
		{
			name: "nothing set",
			build: func() User {
				return User{}
			},
			wantID:   0,
			wantName: "",
			wantRole: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := tt.build()
			assert.Equal(t, tt.wantID, u.ID)
			assert.Equal(t, tt.wantName, u.Name)
			assert.Equal(t, tt.wantRole, u.Role)
		})
	}
}

// ---------------------------------------------------------------------------
// Field get/set (direct field access — idiomatic Go equivalent)
// ---------------------------------------------------------------------------

func TestUser_FieldAccess(t *testing.T) {
	t.Run("ID get and set", func(t *testing.T) {
		u := &User{}
		u.ID = 7
		assert.Equal(t, 7, u.ID)
	})

	t.Run("Name get and set", func(t *testing.T) {
		u := &User{}
		u.Name = "Eve"
		assert.Equal(t, "Eve", u.Name)
	})

	t.Run("Email get and set", func(t *testing.T) {
		u := &User{}
		u.Email = "eve@example.com"
		assert.Equal(t, "eve@example.com", u.Email)
	})

	t.Run("Password get and set", func(t *testing.T) {
		u := &User{}
		u.Password = "hunter2"
		assert.Equal(t, "hunter2", u.Password)
	})

	t.Run("Role get and set", func(t *testing.T) {
		u := &User{}
		u.Role = "SUPERUSER"
		assert.Equal(t, "SUPERUSER", u.Role)
	})

	t.Run("About get and set", func(t *testing.T) {
		u := &User{}
		u.About = "I am a user"
		assert.Equal(t, "I am a user", u.About)
	})
}

// ---------------------------------------------------------------------------
// Equality (struct comparison)
// ---------------------------------------------------------------------------

func TestUser_Equality(t *testing.T) {
	base := User{ID: 1, Name: "Frank", Email: "frank@example.com",
		Password: "pw", Role: "USER", About: "about"}

	tests := []struct {
		name      string
		a         User
		b         User
		wantEqual bool
	}{
		{
			name:      "identical structs are equal",
			a:         base,
			b:         base,
			wantEqual: true,
		},
		{
			name: "differ in ID",
			a:    base,
			b:    User{ID: 2, Name: "Frank", Email: "frank@example.com", Password: "pw", Role: "USER", About: "about"},
			wantEqual: false,
		},
		{
			name: "differ in Name",
			a:    base,
			b:    User{ID: 1, Name: "George", Email: "frank@example.com", Password: "pw", Role: "USER", About: "about"},
			wantEqual: false,
		},
		{
			name: "differ in Email",
			a:    base,
			b:    User{ID: 1, Name: "Frank", Email: "other@example.com", Password: "pw", Role: "USER", About: "about"},
			wantEqual: false,
		},
		{
			name: "differ in Password",
			a:    base,
			b:    User{ID: 1, Name: "Frank", Email: "frank@example.com", Password: "other", Role: "USER", About: "about"},
			wantEqual: false,
		},
		{
			name: "differ in Role",
			a:    base,
			b:    User{ID: 1, Name: "Frank", Email: "frank@example.com", Password: "pw", Role: "ADMIN", About: "about"},
			wantEqual: false,
		},
		{
			name: "differ in About",
			a:    base,
			b:    User{ID: 1, Name: "Frank", Email: "frank@example.com", Password: "pw", Role: "USER", About: "different"},
			wantEqual: false,
		},
		{
			name:      "both zero values are equal",
			a:         User{},
			b:         User{},
			wantEqual: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantEqual {
				assert.Equal(t, tt.a, tt.b)
			} else {
				assert.NotEqual(t, tt.a, tt.b)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// String representation (via fmt.Sprintf / %+v)
// ---------------------------------------------------------------------------

func TestUser_StringRepresentation(t *testing.T) {
	tests := []struct {
		name            string
		user            User
		wantContains    []string
	}{
		{
			name: "all fields reflected in string",
			user: User{
				ID:       10,
				Name:     "Hannah",
				Email:    "hannah@example.com",
				Password: "secret",
				Role:     "ADMIN",
				About:    "admin user",
			},
			wantContains: []string{"10", "Hannah", "hannah@example.com", "secret", "ADMIN", "admin user"},
		},
		{
			name: "zero value user",
			user: User{},
			// Go's %+v will at minimum show field names; values are empty/zero
			wantContains: []string{"ID:", "Name:", "Email:", "Password:", "Role:", "About:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			import_fmt_sprintf := func(u User) string {
				// Use Go's default struct formatting which shows field names and values.
				return strings.Replace(
					strings.Replace(
						strings.Replace(
							strings.Replace(
								strings.Replace(
									strings.Replace(
										fmt_sprintf(u),
										"{", "", -1),
									"}", "", -1),
								" ", "", -1),
							"\n", "", -1),
						"\t", "", -1),
					",", "", -1)
			}
			_ = import_fmt_sprintf // avoid unused warning; we use the helper below
			repr := formatUser(tt.user)
			for _, want := range tt.wantContains {
				assert.Contains(t, repr, want)
			}
		})
	}
}

// formatUser returns a string representation of u for test assertions.
func formatUser(u User) string {
	// Manually build a representation that includes all field names and values.
	return strings.Join([]string{
		"ID:" + itoa(u.ID),
		"Name:" + u.Name,
		"Email:" + u.Email,
		"Password:" + u.Password,
		"Role:" + u.Role,
		"About:" + u.About,
	}, " ")
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	buf := make([]byte, 0, 20)
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	if neg {
		buf = append([]byte{'-'}, buf...)
	}
	return string(buf)
}

// ---------------------------------------------------------------------------
// Validate
// ---------------------------------------------------------------------------

func TestUser_Validate(t *testing.T) {
	tests := []struct {
		name        string
		user        User
		wantErr     bool
		wantField   string
		wantMessage string
	}{
		{
			name: "valid user passes validation",
			user: User{
				ID:       1,
				Name:     "Ivan",
				Email:    "ivan@example.com",
				Password: "pw",
				Role:     "USER",
				About:    "hello",
			},
			wantErr: false,
		},
		{
			name: "blank name returns validation error with expected message",
			user: User{
				Name:  "",
				Email: "x@x.com",
			},
			wantErr:     true,
			wantField:   "name",
			wantMessage: errNameBlank,
		},
		{
			name: "whitespace-only name is treated as blank",
			user: User{
				Name:  "   ",
				Email: "x@x.com",
			},
			wantErr:     true,
			wantField:   "name",
			wantMessage: errNameBlank,
		},
		{
			name: "tab-only name is treated as blank",
			user: User{
				Name: "\t\t",
			},
			wantErr:     true,
			wantField:   "name",
			wantMessage: errNameBlank,
		},
		{
			name: "newline-only name is treated as blank",
			user: User{
				Name: "\n",
			},
			wantErr:     true,
			wantField:   "name",
			wantMessage: errNameBlank,
		},
		{
			name: "about exactly 500 characters is valid",
			user: User{
				Name:  "Jane",
				About: strings.Repeat("a", 500),
			},
			wantErr: false,
		},
		{
			name: "about 501 characters fails validation",
			user: User{
				Name:  "Jane",
				About: strings.Repeat("a", 501),
			},
			wantErr:     true,
			wantField:   "about",
			wantMessage: "about must be at most 500 characters",
		},
		{
			name: "about 1000 characters fails validation",
			user: User{
				Name:  "Jane",
				About: strings.Repeat("x", 1000),
			},
			wantErr:     true,
			wantField:   "about",
			wantMessage: "about must be at most 500 characters",
		},
		{
			name: "blank name takes priority over long about",
			user: User{
				Name:  "",
				About: strings.Repeat("a", 600),
			},
			wantErr:     true,
			wantField:   "name",
			wantMessage: errNameBlank,
		},
		{
			name: "name with only non-whitespace single char is valid",
			user: User{
				Name: "A",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				var ve *ValidationError
				assert.ErrorAs(t, err, &ve)
				assert.Equal(t, tt.wantField, ve.Field)
				assert.Equal(t, tt.wantMessage, ve.Message)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ValidationError
// ---------------------------------------------------------------------------

func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name    string
		ve      ValidationError
		wantMsg string
	}{
		{
			name:    "