package service

// Package service provides the concrete service-layer implementation for the
// SmartContact application. This file corresponds to the original Spring Java
// class com.smartContact.service.UserServiceImp.
//
// MIGRATION_NOTE: In Java, UserService was an interface and UserServiceImp its
// @Service-annotated implementation, wired together by Spring's DI container
// via field injection (@Autowired UserDao). Go has no interface/implementation
// file split of that kind: the UserService interface and its constructor
// (NewUserService) were already migrated in userservice.go, and the concrete
// type (userService) plus its methods live there too. To avoid redeclaring the
// same concrete type in the same package (which would fail to compile), this
// file does NOT re-declare userService, its constructor, or the CRUD methods.
//
// The only behaviour present in UserServiceImp that is not already expressed by
// the interface contract is the getUserNameByName helper (mapped to
// GetUserByName in the interface). That method IS part of the UserService
// interface, so its implementation belongs with the other methods in
// userservice.go rather than here.
//
// This file therefore contributes a small, real, non-duplicate piece: a
// compile-time assertion that the concrete implementation satisfies the
// UserService interface, plus documentation of the update-semantics caveat
// noted during migration.
//
// MIGRATION_NOTE (NOTES #H): The original updateUser set the ID on the supplied
// User and persisted it via save, returning nothing (the Java method was void).
// The migrated UpdateUser returns the request struct verbatim rather than
// re-SELECTing the row after upsert. This is faithful today because SmartContact
// has no DB-side mutations (no triggers, generated columns, or server-side
// defaults applied on write). If DB-side defaults are added later, UpdateUser
// must re-SELECT the row after the upsert to return the authoritative,
// database-materialized state instead of the caller-supplied struct.

// interfaceGuard is a compile-time assertion that the concrete service type
// declared in userservice.go implements the UserService interface. It exists so
// that a mismatch between the interface contract and its implementation is
// caught at build time rather than at runtime.
//
// It references the unexported concrete type (*userService) and the exported
// NewUserService constructor's return type indirectly: because both live in
// this same package, the assertion below is valid without additional imports.
var _ UserService = (*userService)(nil)
