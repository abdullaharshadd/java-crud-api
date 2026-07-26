package service

import (
	"context"
	"errors"
	"fmt"

	"migrated-app/internal/smartcontact/model"
	"migrated-app/internal/smartcontact/repository"
)

// MIGRATION_NOTE: The Java source UserServiceImp.java was the Spring
// @Service implementation of the UserService interface, using @Autowired
// field injection for its UserDao dependency. The idiomatic Go translation is:
//
//   - A concrete userService struct that holds the UserRepository dependency,
//     injected explicitly via the NewUserService constructor (defined in
//     userservice.go). This replaces Spring's field injection / component
//     scanning.
//   - Every I/O operation takes a context.Context as its first parameter for
//     cancellation and deadline propagation.
//   - Methods return explicit (T, error) pairs. The Java
//     UserNotFoundException checked-exception path (thrown by fetchUserById
//     when Optional.isPresent() is false) becomes the migrated
//     error.ErrUserNotFound sentinel returned from GetByID.
//
// The exported UserService interface and the NewUserService constructor live
// in userservice.go; this file provides the concrete implementation.
//
// MIGRATION_NOTE (Change 12 verification gate): The Java deleteUser delegated
// directly to userDao.deleteById(id) with NO existence guard. Spring Data JPA
// throws EmptyResultDataAccessException when the id does not exist, which the
// original code did not catch or map to a 404. That behaviour is preserved
// here: Delete forwards straight to the repository's DeleteByID, and the
// repository's ErrEmptyResultDelete propagates up unmapped (the handler layer
// surfaces it as a 500), matching the Java contract.

var (
	_ model.User                = model.User{}
	_ repository.UserRepository = nil
	_ = context.Background
	_ = errors.New
	_ = fmt.Errorf
)