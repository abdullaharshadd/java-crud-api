package service

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	smartError "migrated-app/internal/smartcontact/error"
	"migrated-app/internal/smartcontact/model"
	"migrated-app/internal/smartcontact/repository"
)

// UserService defines the user-related business operations exposed to the
// transport (HTTP handler) layer.
//
// It mirrors the original Spring UserService contract: create, list, fetch by
// id, delete, update, and lookup by name.
type UserService interface {
	// Save persists a new user and returns the stored representation.
	//
	// MIGRATION_NOTE: Java `User saveUser(User user)`.
	Save(ctx context.Context, user model.User) (model.User, error)

	// List returns all users.
	//
	// MIGRATION_NOTE: Java `List<User> fetchUserList()`.
	List(ctx context.Context) ([]model.User, error)

	// GetByID fetches a single user by its id.
	//
	// MIGRATION_NOTE: Java `User fetchUserById(int id) throws
	// UserNotFoundException`. The checked exception becomes a wrapped
	// error.ErrUserNotFound, matchable with errors.Is.
	GetByID(ctx context.Context, id int) (model.User, error)

	// Delete removes the user with the given id.
	//
	// MIGRATION_NOTE: Java `void deleteUser(int id)`.
	Delete(ctx context.Context, id int) error

	// Update replaces the user identified by id with the supplied user.
	//
	// MIGRATION_NOTE: Java `void updateUser(int id, User user)`.
	Update(ctx context.Context, id int, user model.User) error

	// GetByName fetches a single user by name.
	//
	// MIGRATION_NOTE: Java `User getUserNameByName(String name)`.
	GetByName(ctx context.Context, name string) (model.User, error)
}

// userService is the concrete UserService implementation backed by a
// repository.UserRepository.
//
// MIGRATION_NOTE: this is the Go analogue of the (unmigrated) Spring
// UserServiceImpl. Dependencies are injected explicitly via NewUserService
// rather than through Spring's @Autowired.
type userService struct {
	repo     repository.UserRepository
	validate *validator.Validate
}

// NewUserService constructs a UserService backed by the supplied repository.
func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo, validate: validator.New()}
}

// Save validates and persists a new user.
func (s *userService) Save(ctx context.Context, user model.User) (model.User, error) {
	if err := user.Validate(s.validate); err != nil {
		return model.User{}, fmt.Errorf("save user: %w", err)
	}
	saved, err := s.repo.Save(ctx, &user)
	if err != nil {
		return model.User{}, fmt.Errorf("save user: %w", err)
	}
	return *saved, nil
}

// List returns all users.
func (s *userService) List(ctx context.Context) ([]model.User, error) {
	users, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	return users, nil
}

// GetByID fetches a single user by id, returning error.ErrUserNotFound when no
// matching user exists.
func (s *userService) GetByID(ctx context.Context, id int) (model.User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return model.User{}, fmt.Errorf("get user by id %d: %w", id, err)
	}
	return *user, nil
}

// Delete removes the user with the given id.
func (s *userService) Delete(ctx context.Context, id int) error {
	if err := s.repo.DeleteByID(ctx, id); err != nil {
		return fmt.Errorf("delete user %d: %w", id, err)
	}
	return nil
}

// Update replaces the user identified by id with the supplied user.
//
// MIGRATION_NOTE: The Java `updateUser` interface method carried no body. Here
// we make the update semantics explicit: verify the target exists (surfacing
// ErrUserNotFound if not), then persist. The id from the path takes precedence
// over any id embedded in the payload.
func (s *userService) Update(ctx context.Context, id int, user model.User) error {
	if err := user.Validate(s.validate); err != nil {
		return fmt.Errorf("update user %d: %w", id, err)
	}
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return fmt.Errorf("update user %d: %w", id, err)
	}
	user.ID = id
	if _, err := s.repo.Save(ctx, &user); err != nil {
		return fmt.Errorf("update user %d: %w", id, err)
	}
	return nil
}

// GetByName fetches a single user by name, returning error.ErrUserNotFound when
// no matching user exists.
func (s *userService) GetByName(ctx context.Context, name string) (model.User, error) {
	user, err := s.repo.FindByName(ctx, name)
	if err != nil {
		return model.User{}, fmt.Errorf("get user by name %q: %w", name, err)
	}
	return *user, nil
}

// Ensure the sentinel error remains referenced so intent stays documented for
// callers using errors.Is against a not-found result surfaced by this service.
var _ = smartError.ErrUserNotFound