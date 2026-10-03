// Package store contains the database/sql-backed persistence layer.
package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"migrated-app/internal/model"
)

// ErrMultipleResults is returned by FindByName when more than one row matches
// (Spring Data's IncorrectResultSizeDataAccessException; surfaces as 500).
var ErrMultipleResults = errors.New("store: query did not return a unique result")

// ErrNoRowsDeleted is returned by DeleteByID when no row has the given id
// (Spring Data's EmptyResultDataAccessException; surfaces as 500, not 404).
var ErrNoRowsDeleted = errors.New("store: no user entity with the given id exists")

// ErrInvalidPage is returned by FindPage for a negative page or non-positive size.
var ErrInvalidPage = errors.New("store: invalid page request")

const userSelectColumns = "`user_id`, `user_name`, `user_email`, `user_password`, `user_role`, `user_about`"

// MySQLUserRepository is the database/sql implementation of the source's
// Spring Data JPA UserDao (JpaRepository<User, Integer> + findByName).
type MySQLUserRepository struct {
	db *sql.DB
}

// NewMySQLUserRepository returns a repository backed by db.
func NewMySQLUserRepository(db *sql.DB) *MySQLUserRepository {
	return &MySQLUserRepository{db: db}
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanUser(rs rowScanner) (model.User, error) {
	var (
		u                                   model.User
		name, email, password, role, about sql.NullString
	)
	if err := rs.Scan(&u.ID, &name, &email, &password, &role, &about); err != nil {
		return model.User{}, err
	}
	u.Name = nullToPtr(name)
	u.Email = nullToPtr(email)
	u.Password = nullToPtr(password)
	u.Role = nullToPtr(role)
	u.About = nullToPtr(about)
	return u, nil
}

func nullToPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	s := ns.String
	return &s
}

// ptrToArg converts a nullable string into a driver argument (nil -> NULL).
func ptrToArg(s *string) any {
	if s == nil {
		return nil
	}
	return *s
}

// Save persists u with JPA merge semantics and returns the managed copy.
// If u.ID != 0 and a row with that id exists, every column is updated (nil
// fields are written as NULL). Otherwise a new id is allocated from
// hibernate_sequence and the row is inserted; any supplied id is ignored.
// The whole operation runs in one transaction.
func (r *MySQLUserRepository) Save(ctx context.Context, u *model.User) (*model.User, error) {
	if u == nil {
		return nil, errors.New("store: save user: entity must not be nil")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("store: save user: begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	saved := *u
	exists := false
	if u.ID != 0 {
		var one int
		err := tx.QueryRowContext(ctx,
			"SELECT 1 FROM `user` WHERE `user_id` = ? FOR UPDATE", u.ID).Scan(&one)
		switch {
		case err == nil:
			exists = true
		case errors.Is(err, sql.ErrNoRows):
		default:
			return nil, fmt.Errorf("store: save user: lookup id %d: %w", u.ID, err)
		}
	}

	if exists {
		if _, err := tx.ExecContext(ctx,
			"UPDATE `user` SET `user_name` = ?, `user_email` = ?, `user_password` = ?, "+
				"`user_role` = ?, `user_about` = ? WHERE `user_id` = ?",
			ptrToArg(u.Name), ptrToArg(u.Email), ptrToArg(u.Password),
			ptrToArg(u.Role), ptrToArg(u.About), u.ID); err != nil {
			return nil, fmt.Errorf("store: save user: update id %d: %w", u.ID, err)
		}
	} else {
		id, err := nextSequenceValue(ctx, tx)
		if err != nil {
			return nil, fmt.Errorf("store: save user: %w", err)
		}
		saved.ID = id
		if _, err := tx.ExecContext(ctx,
			"INSERT INTO `user` (`user_id`, `user_name`, `user_email`, `user_password`, `user_role`, `user_about`) "+
				"VALUES (?, ?, ?, ?, ?, ?)",
			saved.ID, ptrToArg(u.Name), ptrToArg(u.Email), ptrToArg(u.Password),
			ptrToArg(u.Role), ptrToArg(u.About)); err != nil {
			return nil, fmt.Errorf("store: save user: insert: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("store: save user: commit: %w", err)
	}
	return &saved, nil
}

// nextSequenceValue allocates the next id from hibernate_sequence, mirroring
// Hibernate's TableStructure-based generator for GenerationType.AUTO on MySQL.
//
// MIGRATION_NOTE: Hibernate allocates in a separate (isolated) transaction;
// here allocation shares the save transaction, so a rolled-back save does not
// consume an id. Ids remain unique because the row is locked FOR UPDATE.
func nextSequenceValue(ctx context.Context, tx *sql.Tx) (int32, error) {
	var next sql.NullInt64
	err := tx.QueryRowContext(ctx,
		"SELECT `next_val` FROM `hibernate_sequence` LIMIT 1 FOR UPDATE").Scan(&next)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.New("hibernate_sequence is not initialized")
		}
		return 0, fmt.Errorf("read hibernate_sequence: %w", err)
	}
	if !next.Valid {
		return 0, errors.New("hibernate_sequence.next_val is NULL")
	}
	if _, err := tx.ExecContext(ctx,
		"UPDATE `hibernate_sequence` SET `next_val` = ? WHERE `next_val` = ?",
		next.Int64+1, next.Int64); err != nil {
		return 0, fmt.Errorf("advance hibernate_sequence: %w", err)
	}
	return int32(next.Int64), nil
}

// FindByID returns the user with the given id. ok is false when no row exists.
func (r *MySQLUserRepository) FindByID(ctx context.Context, id int32) (model.User, bool, error) {
	row := r.db.QueryRowContext(ctx,
		"SELECT "+userSelectColumns+" FROM `user` WHERE `user_id` = ?", id)
	u, err := scanUser(row)
	if errors.Is(err, sql.ErrNoRows) {
		return model.User{}, false, nil
	}
	if err != nil {
		return model.User{}, false, fmt.Errorf("store: find user by id %d: %w", id, err)
	}
	return u, true, nil
}

// FindByName returns the single user whose name exactly equals name
// (derived query findByName). ok is false when no user matches; more than one
// match yields ErrMultipleResults.
func (r *MySQLUserRepository) FindByName(ctx context.Context, name string) (model.User, bool, error) {
	users, err := r.query(ctx,
		"SELECT "+userSelectColumns+" FROM `user` WHERE `user_name` = ? LIMIT 2", name)
	if err != nil {
		return model.User{}, false, fmt.Errorf("store: find user by name: %w", err)
	}
	switch len(users) {
	case 0:
		return model.User{}, false, nil
	case 1:
		return users[0], true, nil
	default:
		return model.User{}, false, fmt.Errorf("store: find user by name %q: %w", name, ErrMultipleResults)
	}
}

// FindAll returns every user in database order (no ORDER BY, like the
// unsorted JpaRepository.findAll). The result is never nil.
func (r *MySQLUserRepository) FindAll(ctx context.Context) ([]model.User, error) {
	users, err := r.query(ctx, "SELECT "+userSelectColumns+" FROM `user`")
	if err != nil {
		return nil, fmt.Errorf("store: find all users: %w", err)
	}
	return users, nil
}

// FindPage returns one page of users ordered by id (the paged variant of
// findAll). page is zero-based, as in Spring's PageRequest. It also returns
// the total number of users.
func (r *MySQLUserRepository) FindPage(ctx context.Context, page, size int) ([]model.User, int64, error) {
	if page < 0 || size < 1 {
		return nil, 0, fmt.Errorf("store: find user page %d size %d: %w", page, size, ErrInvalidPage)
	}
	total, err := r.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	users, err := r.query(ctx,
		"SELECT "+userSelectColumns+" FROM `user` ORDER BY `user_id` LIMIT ? OFFSET ?",
		size, int64(page)*int64(size))
	if err != nil {
		return nil, 0, fmt.Errorf("store: find user page: %w", err)
	}
	return users, total, nil
}

// Count returns the number of users.
func (r *MySQLUserRepository) Count(ctx context.Context) (int64, error) {
	var n int64
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM `user`").Scan(&n); err != nil {
		return 0, fmt.Errorf("store: count users: %w", err)
	}
	return n, nil
}

// ExistsByID reports whether a user with the given id exists.
func (r *MySQLUserRepository) ExistsByID(ctx context.Context, id int32) (bool, error) {
	var one int
	err := r.db.QueryRowContext(ctx,
		"SELECT 1 FROM `user` WHERE `user_id` = ? LIMIT 1", id).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("store: exists user id %d: %w", id, err)
	}
	return true, nil
}

// DeleteByID removes the user with the given id. It returns ErrNoRowsDeleted
// when no such user exists (Spring Data 2.x deleteById behaviour).
func (r *MySQLUserRepository) DeleteByID(ctx context.Context, id int32) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM `user` WHERE `user_id` = ?", id)
	if err != nil {
		return fmt.Errorf("store: delete user id %d: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("store: delete user id %d: rows affected: %w", id, err)
	}
	if n == 0 {
		return fmt.Errorf("store: delete user id %d: %w", id, ErrNoRowsDeleted)
	}
	return nil
}

// Delete removes the given entity. Like SimpleJpaRepository.delete, a new
// (id 0) or no-longer-existing entity is silently ignored.
func (r *MySQLUserRepository) Delete(ctx context.Context, u *model.User) error {
	if u == nil {
		return errors.New("store: delete user: entity must not be nil")
	}
	if u.ID == 0 {
		return nil
	}
	if _, err := r.db.ExecContext(ctx, "DELETE FROM `user` WHERE `user_id` = ?", u.ID); err != nil {
		return fmt.Errorf("store: delete user id %d: %w", u.ID, err)
	}
	return nil
}

func (r *MySQLUserRepository) query(ctx context.Context, q string, args ...any) ([]model.User, error) {
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]model.User, 0)
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate users: %w", err)
	}
	return users, nil
}
