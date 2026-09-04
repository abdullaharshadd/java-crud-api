```go
package repository

import (
	"errors"
)

var ErrUserNotFound = errors.New("user not found")

func IsUserNotFoundError(err error) bool {
	return errors.Is(err, ErrUserNotFound)
}
```