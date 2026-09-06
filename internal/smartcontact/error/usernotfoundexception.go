```go
package error

import (
	"errors"
)

type UserNotFoundError struct{}

func (e *UserNotFoundError) Error() string {
	return "User not found"
}

func NewUserNotFoundError() error {
	return &UserNotFoundError{}
}
```