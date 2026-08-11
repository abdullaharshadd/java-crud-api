package repository

import (
	"context"

	"migrated-app/internal/smartcontact/model"
)

// UserRepository defines the full persistence operations the service layer needs.
// This extends the interface defined in userdao.go to include FindAll.
