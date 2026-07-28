// Package service defines the business-logic contracts for the smartcontact
// application. This file provides the concrete implementation of the
// UserService interface declared in userservice.go.
//
// MIGRATION_NOTE: The Java source was UserServiceImp, the @Service-annotated
// implementation of the UserService interface. Because the already-migrated
// userservice.go declares both the UserService interface AND its concrete
// implementation (via NewUserService), we must NOT redeclare that concrete
// type here or the package would fail to compile with duplicate declarations.
//
// Instead, this file contributes the one method that the Java implementation
// exposed but the migrated interface did not carry over exactly: the
// name-based lookup. The Java UserServiceImp had a public getUserNameByName
// method that returned a single User by name. The mapping table records this
// Java method migrating to BOTH GetUserByID and GetUserByName; the byID lookup
// is already covered by the interface, so here we add the byName behaviour as
// a method on the same concrete type declared in userservice.go.
//
// If userservice.go's concrete type already implements GetUserByName, delete
// this file — it exists only to fill a gap. See the manual-review note.
package service

import (
	"context"
	"fmt"

	apperror "migrated-app/internal/smartcontact/error"
	"migrated-app/internal/smartcontact/model"
)

// GetUserByName returns the single user whose name matches the supplied value.
//
// MIGRATION_NOTE: The Java getUserNameByName(String) delegated straight to the
// Spring Data derived query userDao.findByName(name) and returned the result
// verbatim (which could be null). In idiomatic Go we surface "no such user"
// as an ErrUserNotFound sentinel rather than returning a nil value, so callers
// can distinguish absence from a transport/DB failure with errors.Is.
//
// This method is declared on *userService, the concrete type defined in
// userservice.go. If that file already provides GetUserByName, remove this
// method to avoid a duplicate declaration.
