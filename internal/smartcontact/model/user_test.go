```go
package model

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// TestUser_ZeroValue – mirrors the "no-args constructor" spec:
// a freshly declared User must have ID==0 and all string fields empty.
// ---------------------------------------------------------------------------

func TestUser_ZeroValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{"instantiated with no arguments"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var u User
			assert.Equal(t, 0, u.ID, "ID should default to 0")
			assert.Equal(t, "", u.Name, "Name should default to empty string")
			assert.Equal(t, "", u.Email, "Email should default to empty string")
			assert.Equal(t, "", u.Password, "Password should default to empty string")
			assert.Equal(t, "", u.Role, "Role should default to empty string")
			assert.Equal(t, "", u.About, "About should default to empty string")
		})
	}
}

// ---------------------------------------------------------------------------
// TestUser_AllArgsConstructor – mirrors the "all-args constructor" spec.
// ---------------------------------------------------------------------------

func TestUser_AllArgsConstructor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		id       int
		userName string
		email    string
		password string
		role     string
		about    string
	}{
		{
			name:     "all fields populated",
			id:       42,
			userName: "Alice",
			email:    "alice@example.com",
			password: "s3cr3t",
			role:     "ADMIN",
			about:    "I am Alice.",
		},
		{
			name:     "zero id with non-empty strings",
			id:       0,
			userName: "Bob",
			email:    "bob@example.com",
			password: "hunter2",
			role:     "USER",
			about:    "Bob's profile",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			u := User{
				ID:       tc.id,
				Name:     tc.userName,
				Email:    tc.email,
				Password: tc.password,
				Role:     tc.role,
				About:    tc.about,
			}
			assert.Equal(t, tc.id, u.ID)
			assert.Equal(t, tc.userName, u.Name)
			assert.Equal(t, tc.email, u.Email)
			assert.Equal(t, tc.password, u.Password)
			assert.Equal(t, tc.role, u.Role)
			assert.Equal(t, tc.about, u.About)
		})
	}
}

// ---------------------------------------------------------------------------
// TestNewUser – mirrors the "builder / NewUser" spec.
// ---------------------------------------------------------------------------

func TestNewUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		userName string
		email    string
		password string
		role     string
		about    string
	}{
		{
			name:     "all fields specified",
			userName: "Carol",
			email:    "carol@example.com",
			password: "pass",
			role:     "MOD",
			about:    "Moderator Carol",
		},
		{
			name:     "subset – empty role and about",
			userName: "Dave",
			email:    "dave@example.com",
			password: "pw",
			role:     "",
			about:    "",
		},
		{
			name:     "all empty strings",
			userName: "",
			email:    "",
			password: "",
			role:     "",
			about:    "",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			u := NewUser(tc.userName, tc.email, tc.password, tc.role, tc.about)
			assert.NotNil(t, u, "NewUser must never return nil")
			// ID must stay 0 (unsaved)
			assert.Equal(t, 0, u.ID, "NewUser should leave ID as 0")
			assert.Equal(t, tc.userName, u.Name)
			assert.Equal(t, tc.email, u.Email)
			assert.Equal(t, tc.password, u.Password)
			assert.Equal(t, tc.role, u.Role)
			assert.Equal(t, tc.about, u.About)
		})
	}
}

// ---------------------------------------------------------------------------
// TestUser_FieldMutation – covers getId/setId, getName/setName, … specs.
// ---------------------------------------------------------------------------

func TestUser_FieldMutation(t *testing.T) {
	t.Parallel()

	t.Run("ID get/set", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			scenario string
			initial  int
			newVal   int
		}{
			{"set positive", 0, 7},
			{"set zero", 5, 0},
			{"set negative", 1, -1},
		}
		for _, tc := range tests {
			tc := tc
			t.Run(tc.scenario, func(t *testing.T) {
				t.Parallel()
				u := User{ID: tc.initial}
				u.ID = tc.newVal
				assert.Equal(t, tc.newVal, u.ID)
			})
		}
	})

	t.Run("Name get/set", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			scenario string
			newVal   string
		}{
			{"non-empty name", "Eve"},
			{"empty name", ""},
		}
		for _, tc := range tests {
			tc := tc
			t.Run(tc.scenario, func(t *testing.T) {
				t.Parallel()
				u := User{}
				u.Name = tc.newVal
				assert.Equal(t, tc.newVal, u.Name)
			})
		}
	})

	t.Run("Email get/set", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			scenario string
			newVal   string
		}{
			{"valid email", "test@test.org"},
			{"empty email", ""},
		}
		for _, tc := range tests {
			tc := tc
			t.Run(tc.scenario, func(t *testing.T) {
				t.Parallel()
				u := User{}
				u.Email = tc.newVal
				assert.Equal(t, tc.newVal, u.Email)
			})
		}
	})

	t.Run("Password get/set", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			scenario string
			newVal   string
		}{
			{"hashed password", "$2a$10$abcdef"},
			{"empty password", ""},
		}
		for _, tc := range tests {
			tc := tc
			t.Run(tc.scenario, func(t *testing.T) {
				t.Parallel()
				u := User{}
				u.Password = tc.newVal
				assert.Equal(t, tc.newVal, u.Password)
			})
		}
	})

	t.Run("Role get/set", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			scenario string
			newVal   string
		}{
			{"admin role", "ADMIN"},
			{"empty role", ""},
		}
		for _, tc := range tests {
			tc := tc
			t.Run(tc.scenario, func(t *testing.T) {
				t.Parallel()
				u := User{}
				u.Role = tc.newVal
				assert.Equal(t, tc.newVal, u.Role)
			})
		}
	})

	t.Run("About get/set", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			scenario string
			newVal   string
		}{
			{"short about", "Software engineer"},
			{"empty about", ""},
			{"exactly 500 chars", strings.Repeat("x", 500)},
		}
		for _, tc := range tests {
			tc := tc
			t.Run(tc.scenario, func(t *testing.T) {
				t.Parallel()
				u := User{}
				u.About = tc.newVal
				assert.Equal(t, tc.newVal, u.About)
			})
		}
	})
}

// ---------------------------------------------------------------------------
// TestUser_Equality – mirrors the "equals / hashCode" spec.
// In Go, struct equality is value-based by default.
// ---------------------------------------------------------------------------

func TestUser_Equality(t *testing.T) {
	t.Parallel()

	base := User{ID: 1, Name: "Frank", Email: "frank@example.com", Password: "pw", Role: "USER", About: "hi"}

	tests := []struct {
		name     string
		a        User
		b        User
		wantEqual bool
	}{
		{
			name:      "identical field values",
			a:         base,
			b:         base,
			wantEqual: true,
		},
		{
			name:      "differ in ID",
			a:         base,
			b:         User{ID: 2, Name: "Frank", Email: "frank@example.com", Password: "pw", Role: "USER", About: "hi"},
			wantEqual: false,
		},
		{
			name:      "differ in Name",
			a:         base,
			b:         User{ID: 1, Name: "Other", Email: "frank@example.com", Password: "pw", Role: "USER", About: "hi"},
			wantEqual: false,
		},
		{
			name:      "differ in Email",
			a:         base,
			b:         User{ID: 1, Name: "Frank", Email: "other@example.com", Password: "pw", Role: "USER", About: "hi"},
			wantEqual: false,
		},
		{
			name:      "differ in Password",
			a:         base,
			b:         User{ID: 1, Name: "Frank", Email: "frank@example.com", Password: "different", Role: "USER", About: "hi"},
			wantEqual: false,
		},
		{
			name:      "differ in Role",
			a:         base,
			b:         User{ID: 1, Name: "Frank", Email: "frank@example.com", Password: "pw", Role: "ADMIN", About: "hi"},
			wantEqual: false,
		},
		{
			name:      "differ in About",
			a:         base,
			b:         User{ID: 1, Name: "Frank", Email: "frank@example.com", Password: "pw", Role: "USER", About: "bye"},
			wantEqual: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if tc.wantEqual {
				assert.Equal(t, tc.a, tc.b, "expected users to be equal")
			} else {
				assert.NotEqual(t, tc.a, tc.b, "expected users to be unequal")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestUser_ToString – mirrors the "toString" spec.
// We verify via JSON marshalling that all fields appear in the output.
// ---------------------------------------------------------------------------

func TestUser_ToString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		user User
		keys []string
	}{
		{
			name: "populated user JSON contains all field keys",
			user: User{ID: 5, Name: "Grace", Email: "grace@example.com", Password: "pw123", Role: "MOD", About: "greetings"},
			keys: []string{"id", "name", "email", "password", "role", "about"},
		},
		{
			name: "zero-value user JSON still contains all field keys",
			user: User{},
			keys: []string{"id", "name", "email", "password", "role", "about"},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			b, err := json.Marshal(tc.user)
			assert.NoError(t, err, "marshalling must not error")
			assert.NotEmpty(t, b, "JSON output must not be empty")
			s := string(b)
			for _, key := range tc.keys {
				assert.Contains(t, s, `"`+key+`"`, "JSON must contain key %q", key)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestUser_Validate – covers the Validate / @NotBlank spec.
// ---------------------------------------------------------------------------

func TestUser_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		user        User
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid user – non-blank name",
			user:    User{Name: "Helen", Email: "helen@example.com", Password: "pw", Role: "USER", About: ""},
			wantErr: false,
		},
		{
			name:        "empty name returns error",
			user:        User{Name: "", Email: "nobody@example.com"},
			wantErr:     true,
			errContains: "please Add the department Name",
		},
		{
			name:        "whitespace-only name returns error",
			user:        User{Name: "   ", Email: "nobody@example.com"},
			wantErr:     true,
			errContains: "please Add the department Name",
		},
		{
			name:        "tab-only name returns error",
			user:        User{Name: "\t\t", Email: "nobody@example.com"},
			wantErr:     true,
			errContains: "please Add the department Name",
		},
		{
			name:        "newline-only name returns error",
			user:        User{Name: "\n\r", Email: "nobody@example.com"},
			wantErr:     true,
			errContains: "please Add the department Name",
		},
		{
			name:    "name with surrounding whitespace but non-blank content",
			user:    User{Name: "  Ivan  "},
			wantErr: false,
		},
		{
			name:    "single non-whitespace character is valid",
			user:    User{Name: "X"},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := tc.user.Validate()
			if tc.wantErr {
				assert.NotNil(t, result, "Validate should return an error message")
				assert.Contains(t, result.Message, tc.errContains)
				assert.Equal(t, 400, result.StatusCode)
			} else {
				assert.Nil(t, result, "Validate should return nil for a valid user")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestIsBlank – unit tests for the internal isBlank helper.
// ---------------------------------------------------------------------------

func TestIsBlank(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"empty string", "", true},
		{"single space", " ", true},
		{"multiple spaces", "   ", true},
		{"tab", "\t", true},
		{"newline