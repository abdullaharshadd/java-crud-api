```go
package repository

import (
	"database/sql"
	"migrated-app/internal/smartcontact/model"
	"migrated-app/internal/smartcontact/errors"
)

type UserRepository struct {
	DB *sql.DB
}

func (r *UserRepository) GetUserByID(id int) (*model.User, error) {
	// implementation
	return nil, errors.NewAppError("GetUserByID", "User not found", nil)
}
```