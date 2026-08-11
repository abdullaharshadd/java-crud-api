```go
package model

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// TestUser_ZeroValue – "no-args constructor" spec
// A zero-value User (equivalent of no-arg constructor) must have id=0 and all
// string fields equal to their zero value ("").
// ---------------------------------------------------------------------------

func TestUser_ZeroValue(t *testing.T) {
	tests := []struct {
		name        string
		expectedID  int
		expectedStr string // every string field should be this value
	}{
		{
			name:        "instantiated with no arguments has id=0 and empty string fields",
			expectedID:  0,
			expectedStr: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var u User
			assert.Equal(t, tc.expectedID, u.ID)
			assert.Equal(t, tc.expectedStr, u.Name)
			assert.Equal(t, tc.expectedStr, u.Email)
			assert.Equal(t, tc.expectedStr, u.Password)
			assert.Equal(t, tc.expectedStr, u.Role)
			assert.Equal(t, tc.expectedStr, u.About)
		})
	}
}

// ---------------------------------------------------------------------------
// TestUser_AllArgsConstructor – "all-args constructor" spec
// NewUser stores all provided values verbatim. ID is omitted (DB-assigned).
// Direct struct literal covers the "set every field including ID" case.
// ---------------------------------------------------------------------------

func TestUser_AllArgsConstructor(t *testing.T) {
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
			password: "secret",
			role:     "ADMIN",
			about:    "About Alice",
		},
		{
			name:     "empty string values (no validation at construction time)",
			id:       0,
			uname:    "",
			email:    "",
			password: "",
			role:     "",
			about:    "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Use struct literal to simulate all-args constructor (including ID).
			u := User{
				ID:       tc.id,
				Name:     tc.uname,
				Email:    tc.email,
				Password: tc.password,
				Role:     tc.role,
				About:    tc.about,
			}
			assert.Equal(t, tc.id, u.ID)
			assert.Equal(t, tc.uname, u.Name)
			assert.Equal(t, tc.email, u.Email)
			assert.Equal(t, tc.password, u.Password)
			assert.Equal(t, tc.role, u.Role)
			assert.Equal(t, tc.about, u.About)
		})
	}
}

// ---------------------------------------------------------------------------
// TestNewUser – "builder" / NewUser constructor spec
// NewUser omits ID (left at default 0). Unset fields stay at zero value.
// ---------------------------------------------------------------------------

func TestNewUser(t *testing.T) {
	tests := []struct {
		name     string
		uname    string
		email    string
		password string
		role     string
		about    string
		wantID   int // always 0 – DB-assigned
	}{
		{
			name:     "all fields set via NewUser",
			uname:    "Bob",
			email:    "bob@example.com",
			password: "hash",
			role:     "USER",
			about:    "About Bob",
			wantID:   0,
		},
		{
			name:     "subset of fields – empty role and about",
			uname:    "Carol",
			email:    "carol@example.com",
			password: "pw",
			role:     "",
			about:    "",
			wantID:   0,
		},
		{
			name:     "all empty strings – no validation at construction",
			uname:    "",
			email:    "",
			password: "",
			role:     "",
			about:    "",
			wantID:   0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			u := NewUser(tc.uname, tc.email, tc.password, tc.role, tc.about)
			assert.NotNil(t, u)
			assert.Equal(t, tc.wantID, u.ID, "ID must default to 0 (DB-assigned)")
			assert.Equal(t, tc.uname, u.Name)
			assert.Equal(t, tc.email, u.Email)
			assert.Equal(t, tc.password, u.Password)
			assert.Equal(t, tc.role, u.Role)
			assert.Equal(t, tc.about, u.About)
		})
	}
}

// ---------------------------------------------------------------------------
// TestUser_FieldAccessorMutator – getId/setId, getName/setName, etc.
// Go uses direct field access; these tests mirror the getter/setter behaviour.
// ---------------------------------------------------------------------------

func TestUser_FieldAccessorMutator(t *testing.T) {
	t.Run("ID get and set", func(t *testing.T) {
		tests := []struct {
			name  string
			value int
		}{
			{"positive id", 7},
			{"zero id", 0},
			{"large id", 999999},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				var u User
				u.ID = tc.value
				assert.Equal(t, tc.value, u.ID)
			})
		}
	})

	t.Run("Name get and set", func(t *testing.T) {
		tests := []struct {
			name  string
			value string
		}{
			{"non-empty name", "Alice"},
			{"empty name", ""},
			{"whitespace name", "   "},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				var u User
				u.Name = tc.value
				assert.Equal(t, tc.value, u.Name)
			})
		}
	})

	t.Run("Email get and set", func(t *testing.T) {
		tests := []struct {
			name  string
			value string
		}{
			{"valid email", "user@example.com"},
			{"empty email", ""},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				var u User
				u.Email = tc.value
				assert.Equal(t, tc.value, u.Email)
			})
		}
	})

	t.Run("Password get and set", func(t *testing.T) {
		tests := []struct {
			name  string
			value string
		}{
			{"hashed password", "$2a$10$abc"},
			{"empty password", ""},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				var u User
				u.Password = tc.value
				assert.Equal(t, tc.value, u.Password)
			})
		}
	})

	t.Run("Role get and set", func(t *testing.T) {
		tests := []struct {
			name  string
			value string
		}{
			{"ADMIN role", "ADMIN"},
			{"USER role", "USER"},
			{"empty role", ""},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				var u User
				u.Role = tc.value
				assert.Equal(t, tc.value, u.Role)
			})
		}
	})

	t.Run("About get and set", func(t *testing.T) {
		long := strings.Repeat("x", 500)
		tests := []struct {
			name  string
			value string
		}{
			{"short about", "Hello"},
			{"empty about", ""},
			{"500-char about (max DB length)", long},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				var u User
				u.About = tc.value
				assert.Equal(t, tc.value, u.About)
			})
		}
	})
}

// ---------------------------------------------------------------------------
// TestUser_Validate – Validate() method spec
// ---------------------------------------------------------------------------

func TestUser_Validate(t *testing.T) {
	tests := []struct {
		name    string
		user    *User
		wantErr error
	}{
		{
			name:    "valid non-blank name returns nil",
			user:    &User{Name: "Alice"},
			wantErr: nil,
		},
		{
			name:    "empty name returns ErrUserNameBlank",
			user:    &User{Name: ""},
			wantErr: ErrUserNameBlank,
		},
		{
			name:    "whitespace-only name returns ErrUserNameBlank",
			user:    &User{Name: "   "},
			wantErr: ErrUserNameBlank,
		},
		{
			name:    "tab-only name returns ErrUserNameBlank",
			user:    &User{Name: "\t\n"},
			wantErr: ErrUserNameBlank,
		},
		{
			name:    "name with leading/trailing spaces but non-blank content is valid",
			user:    &User{Name: "  Bob  "},
			wantErr: nil,
		},
		{
			name: "user with all fields set and valid name",
			user: &User{
				ID:       1,
				Name:     "Carol",
				Email:    "carol@example.com",
				Password: "pw",
				Role:     "USER",
				About:    "About Carol",
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.user.Validate()
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestUser_Equality – "equals" spec
// Go structs support == comparison; we also test reflect.DeepEqual semantics.
// ---------------------------------------------------------------------------

func TestUser_Equality(t *testing.T) {
	base := User{
		ID:       1,
		Name:     "Alice",
		Email:    "alice@example.com",
		Password: "hash",
		Role:     "ADMIN",
		About:    "About Alice",
	}

	tests := []struct {
		name      string
		a         User
		b         User
		wantEqual bool
	}{
		{
			name:      "identical users are equal",
			a:         base,
			b:         base,
			wantEqual: true,
		},
		{
			name: "different ID makes users not equal",
			a:    base,
			b: User{
				ID:       2,
				Name:     "Alice",
				Email:    "alice@example.com",
				Password: "hash",
				Role:     "ADMIN",
				About:    "About Alice",
			},
			wantEqual: false,
		},
		{
			name: "different Name makes users not equal",
			a:    base,
			b: User{
				ID:       1,
				Name:     "Bob",
				Email:    "alice@example.com",
				Password: "hash",
				Role:     "ADMIN",
				About:    "About Alice",
			},
			wantEqual: false,
		},
		{
			name: "different Email makes users not equal",
			a:    base,
			b: User{
				ID:       1,
				Name:     "Alice",
				Email:    "other@example.com",
				Password: "hash",
				Role:     "ADMIN",
				About:    "About Alice",
			},
			wantEqual: false,
		},
		{
			name: "different Password makes users not equal",
			a:    base,
			b: User{
				ID:       1,
				Name:     "Alice",
				Email:    "alice@example.com",
				Password: "other",
				Role:     "ADMIN",
				About:    "About Alice",
			},
			wantEqual: false,
		},
		{
			name: "different Role makes users not equal",
			a:    base,
			b: User{
				ID:       1,
				Name:     "Alice",
				Email:    "alice@example.com",
				Password: "hash",
				Role:     "USER",
				About:    "About Alice",
			},
			wantEqual: false,
		},
		{
			name: "different About makes users not equal",
			a:    base,
			b: User{
				ID:       1,
				Name:     "Alice",
				Email:    "alice@example.com",
				Password: "hash",
				Role:     "ADMIN",
				About:    "Other",
			},
			wantEqual: false,
		},
		{
			name:      "two zero-value users are equal",
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

	// Reflexivity
	t.Run("reflexive: a user equals itself", func(t *testing.T) {
		u := base
		assert.Equal(t, u, u)
	})
}

// ---------------------------------------------------------------------------
// TestUser_HashCode – "hashCode" spec
// In Go the natural "hash code" for a struct is its comparable value; two
// equal structs will always occupy the same map bucket key. We verify this
// by using User as a map key.
// ---------------------------------------------------------------------------

func TestUser_HashCode(t *testing.T) {
	tests := []struct {
		name string
		a    User
		b    User
	}{
		{
			name: "two equal users have the same map key",
			a: User{
				ID:       1,
				Name:     "Alice",
				Email:    "alice@example.com",
				Password: "hash",
				Role:     "ADMIN",
				About:    "bio",
			},
			b: User{
				ID:       1,
				Name:     "Alice",
				Email:    "alice@example.com",
				Password: "hash",
				Role:     "ADMIN",
				About:    "bio",
			},
		},
		{
			name: "two zero-value users share the same map key",
			a:    User{},
			b:    User{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := make(map[User]bool)
			m[tc.a] = true
			// If a and b are equal their keys collide → both lookups return true.
			assert.True(t, m[tc.b], "equal users must hash to the same map key")
		})
	}
}

// ---------------------------------------------------------------------------
// TestUser_ToString – "toString" spec
// Go's equivalent is the JSON / fmt representation. We verify that a string
// representation containing the relevant field values is non-empty.
// ---------------------------------------------------------------------------

func TestUser_ToString(t *testing.T) {
	tests := []struct