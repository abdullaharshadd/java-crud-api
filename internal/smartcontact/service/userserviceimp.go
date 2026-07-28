package service

// MIGRATION_NOTE: The Java source file UserServiceImp.java was the
// implementation half of an interface/implementation pair:
//
//	@Service
//	public class UserServiceImp implements UserService { ... }
//
// In idiomatic Go this split is unnecessary and, worse, harmful: the
// UserService interface was already collapsed into a single concrete
// UserService type (see userservice.go) whose constructor NewUserService
// injects a repository.UserDao. Every method the Java UserServiceImp
// provided (saveUser, fetchUserList, fetchUserById, deleteUser, updateUser,
// getUserNameByName) is already implemented there as SaveUser, FetchUserList,
// FetchUserByID, DeleteUser, UpdateUser and GetUserByName.
//
// Re-declaring that type (or its methods) here would produce a duplicate
// declaration in the same package and fail to compile. Therefore this file
// contributes only NEW code: a small compile-time assertion documenting the
// contract the migrated concrete type is expected to satisfy, plus a package
// alias interface capturing the exact behaviour set that the original
// @Service class exposed. This preserves the intent of the Java
// "implements UserService" declaration (a verifiable behavioural contract)
// without redeclaring the concrete type.

import (
	"context"

	"migrated-app/internal/smartcontact/model"
)

// UserServicer describes the behaviour originally declared by the Java
// UserService interface and implemented by UserServiceImp. It exists so that
// consumers (e.g. HTTP handlers) can depend on the narrow contract rather than
// the concrete *UserService type, and so that the migration preserves the
// explicit "this type fulfils the user-service contract" guarantee that Java's
// `implements UserService` provided.
//
// MIGRATION_NOTE: In Java, saveUser/updateUser returned or mutated User and
// several methods declared checked exceptions (UserNotFoundException). In Go
// every I/O method takes a context.Context and returns an explicit error;
// the not-found case surfaces as error.ErrUserNotFound rather than a thrown
// exception. Absent-value handling that Java expressed with Optional is folded
// into the (value, error) return convention.
type UserServicer interface {
	// SaveUser persists a new user and returns the stored representation.
	SaveUser(ctx context.Context, req model.CreateUserRequest) (model.UserResponse, error)
	// FetchUserList returns all users.
	FetchUserList(ctx context.Context) ([]model.UserResponse, error)
	// FetchUserByID returns the user with the given id, or ErrUserNotFound.
	FetchUserByID(ctx context.Context, id int) (model.UserResponse, error)
	// DeleteUser removes the user with the given id.
	DeleteUser(ctx context.Context, id int) error
	// UpdateUser updates the user identified by id.
	UpdateUser(ctx context.Context, id int, req model.CreateUserRequest) (model.UserResponse, error)
	// GetUserByName returns the user matching the given name.
	GetUserByName(ctx context.Context, name string) (model.UserResponse, error)
}

// Ensure the concrete *UserService (declared in userservice.go) satisfies the
// migrated contract. This compile-time assertion is the Go equivalent of the
// Java `class UserServiceImp implements UserService` declaration: if the
// concrete type ever drifts from the contract, the build fails here.
//
// MIGRATION_NOTE: If the concrete method signatures in userservice.go differ
// from those declared above (for example, a different request/response type or
// a missing context parameter), this assertion will not compile and the
// signatures must be reconciled by hand. This is intentional and flags the
// single point that needs manual review after the interface/implementation
// merge.
var _ UserServicer = (*UserService)(nil)
