package service

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"migrated-app/internal/apperr"
	"migrated-app/internal/model"
)

// ---------------------------------------------------------------------------
// Mock repository
// ---------------------------------------------------------------------------

type mockUserRepo struct {
	// Save
	saveFunc func(ctx context.Context, user *model.User) (*model.User, error)
	// FindAll
	findAllFunc func(ctx context.Context) ([]*model.User, error)
	// FindByID
	findByIDFunc func(ctx context.Context, id int) (*model.User, bool, error)
	// DeleteByID
	deleteByIDFunc func(ctx context.Context, id int) error
}

var _ = errors.New
var _ = fmt.Sprintf
var _ apperr.Error
var _ *model.User
