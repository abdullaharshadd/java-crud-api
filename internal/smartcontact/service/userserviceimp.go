// Package service defines the business-logic layer contracts and
// implementations migrated from com.smartContact.service.
//
// MIGRATION_NOTE: The Java source split UserService (interface) and
// UserServiceImp (implementation) into two files. In Go there is no need
// for a separate implementation class: the concrete type (userService)
// and its constructor (NewUserService) plus all six methods (SaveUser,
// GetAllUsers, FetchUserById, DeleteUser, UpdateUser, GetUserNameByName)
// are already fully declared in userservice.go in this same package.
//
// Redeclaring that concrete type here would cause a duplicate-declaration
// compile error, so this file intentionally adds no new type or method.
// It documents the mapping from the Java UserServiceImp members to their
// already-migrated Go counterparts so a reader looking for the
// "implementation half" of the pair is pointed to the right place.
//
// Java UserServiceImp member -> Go equivalent (in userservice.go):
//
//	saveUser(User)             -> (*userService).SaveUser(ctx, *model.User) (*model.User, error)
//	fetchUserList()            -> (*userService).GetAllUsers(ctx) ([]*model.User, error)
//	fetchUserById(int)         -> (*userService).FetchUserById(ctx, id) (*model.User, error)
//	                              wraps a missing row as *apperr.UserNotFoundError
//	deleteUser(int)            -> (*userService).DeleteUser(ctx, id) error
//	                              propagates a generic error on a missing row
//	updateUser(int, User)      -> (*userService).UpdateUser(ctx, id, *model.User) error
//	                              sets id then performs a full-row upsert with no validation
//	getUserNameByName(String)  -> (*userService).GetUserNameByName(ctx, name) (*model.User, error)
//
// See userservice.go for the actual implementations.
package service
