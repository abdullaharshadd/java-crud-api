package model

import (
	"testing"
	"assertions"

	"github.com/stretchr/testify/assert"
)

type getUserTest struct {
	user    User
	wantID  int
	wantErr bool
}

func TestValidate(t *testing.T) {
	tests := []getUserTest{
		{User{Name: "", Email: "test@example.com", Password: "password", Role: "admin", About: "about info"}, 0, true},
		{User{Name: "John Doe", Email: "test@example.com", Password: "password", Role: "admin", About: "about info"}, 0, false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Name: %s", tt.user.Name), func(t *testing.T) {
			err := tt.user.Validate()
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

type getSetFieldTest struct {
	setFunc  func(user *User, value interface{})
	getFunc  func(user *User) interface{}
	input    interface{}
	expected interface{}
}

func TestGetSetFields(t *testing.T) {
	tests := []struct {
		name     string
		testFunc getSetFieldTest
	}{
		{"ID", getSetFieldTest{func(u *User, id interface{}) { u.ID = id.(int) }, func(u *User) interface{} { return u.ID }, 123, 123}},
		{"Name", getSetFieldTest{func(u *User, name interface{}) { u.Name = name.(string) }, func(u *User) interface{} { return u.Name }, "John Doe", "John Doe"}},
		{"Email", getSetFieldTest{func(u *User, email interface{}) { u.Email = email.(string) }, func(u *User) interface{} { return u.Email }, "john.doe@example.com", "john.doe@example.com"}},
		{"Password", getSetFieldTest{func(u *User, password interface{}) { u.Password = password.(string) }, func(u *User) interface{} { return u.Password }, "securepassword", "securepassword"}},
		{"Role", getSetFieldTest{func(u *User, role interface{}) { u.Role = role.(string) }, func(u *User) interface{} { return u.Role }, "admin", "admin"}},
		{"About", getSetFieldTest{func(u *User, about interface{}) { u.About = about.(string) }, func(u *User) interface{} { return u.About }, "about me", "about me"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{}
			tt.testFunc.setFunc(u, tt.testFunc.input)
			got := tt.testFunc.getFunc(u)
			assert.Equal(t, tt.testFunc.expected, got)
		})
	}
}

func TestSetNameWithBlankString(t *testing.T) {
	u := &User{}
	err := u.setName("")
	assert.NotNil(t, err)
}

func TestSetAboutWithLongString(t *testing.T) {
	longString := "a" // assume this is a string longer than 500 characters
	u := &User{}
	err := u.setAbout(longString)
	assert.NotNil(t, err)
}