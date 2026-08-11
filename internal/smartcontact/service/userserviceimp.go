package service

import (
	"context"
	"errors"

	smarterror "migrated-app/internal/smartcontact/error"
	"migrated-app/internal/smartcontact/model"
	"migrated-app/internal/smartcontact/repository"
)

// FetchUserByIDStrict fetches a single user by id and returns a
// *smarterror.UserNotFoundError when no matching user exists.
//
// MIGRATION_NOTE: This mirrors the Java UserServiceImp.fetchUserById logic,
// which threw UserNotFoundException("User are not available") when the JPA
// Optional was empty. In Go the repository signals "no rows" via
// sql.ErrNoRows, which we translate into the domain-specific not-found error.
func FetchUserByIDStrict(ctx context.Context, repo repository.UserRepository, id int) (*model.User, error) {
	user, err := repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, smarterror.ErrUserNotFound) {
			return nil, smarterror.NewUserNotFoundError("User are not available")
		}
		return nil, err
	}
	if user == nil {
		return nil, smarterror.NewUserNotFoundError("User are not available")
	}
	return user, nil
}

// UpdateUserByID assigns id to the supplied user and persists it, matching the
// Java UserServiceImp.updateUser behavior (user.setId(id); userDao.save(user)).
//
// MIGRATION_NOTE: The Java method returned void, confirming callers discard the
// saved entity; this helper likewise ignores the value returned by Save.
func UpdateUserByID(ctx context.Context, repo repository.UserRepository, id int, user *model.User) error {
	if user == nil {
		return errors.New("service: user must not be nil")
	}
	user.ID = id
	if _, err := repo.Save(ctx, user); err != nil {
		return err
	}
	return nil
}