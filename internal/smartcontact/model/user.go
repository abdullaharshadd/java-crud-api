package model

import "fmt"

// User represents a user entity in the system.
type User struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Password string `json:"password"`
	Role   string `json:"role"`
	About  string `json:"about"`
}

// Validate checks if the user's required fields are set.
func (u *User) Validate() error {
	if u.Name == "" {
		return NewUserNotFoundError("Please add the user name", nil)
	}
	return nil
}
