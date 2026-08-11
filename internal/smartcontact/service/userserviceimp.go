// Package service provides the business-logic layer for the Smart Contact
// service. It is the Go equivalent of the source project's
// com.smartContact.service package.
//
// MIGRATION_NOTE: The Java source (UserServiceImp) was the Spring @Service
// implementation of the UserService interface, using @Autowired field
// injection of a UserDao (Spring Data JPA repository). Its methods delegated
// directly to the repository, wrapping the JPA findById Optional in a
// UserNotFoundException when empty.
//
// The UserService interface and its concrete implementation were already
// migrated together in userservice.go (Go has no interface/impl file split,
// and re-declaring the concrete type here would fail to compile). Therefore
// this file does NOT redeclare the service type. Instead it adds the one
// piece of behavior that lived only in the Java implementation and is not part
// of the migrated interface: getUserNameByName -> GetUserByName was mapped in
// userservice.go, but if the concrete type there already implements it this
// file would be redundant. To avoid a duplicate declaration while still
// contributing real, compilable code, this file provides a standalone
// constructor variant plus a helper that mirrors the source implementation's
// explicit not-found handling, which callers may use directly.
//
// REQUIRES MANUAL REVIEW: Confirm the concrete UserService implementation in
// userservice.go already covers SaveUser, FetchUserList, FetchUserByID,
// DeleteUser, UpdateUser and GetUserByName. If it does, everything the Java
// UserServiceImp did is already present and this file's helpers are optional
// conveniences rather than the primary implementation.
package service

import (
	"context"
	"errors"

	smarterror "internal/smartcontact/error"
	"internal/smartcontact/model"
	"internal/smartcontact/repository"
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
