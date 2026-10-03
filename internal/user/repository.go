package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// Repository is the persistence port for User entities
// (source: com.smartContact.repository.UserDao, a Spring Data
// JpaRepository<User, Integer>). Every method takes a context so callers can
// propagate deadlines and cancellation down to the database driver.
//
// MIGRATION_NOTE: only the JpaRepository operations the application actually
// relies on are ported (save, findAll, findById, deleteById, delete,
// findByName). Paging/sorting overloads, batch operations and flush/
// persistence-context features (first-level cache, dirty checking) have no
// equivalent here: entities are plain values and every change must go
// through Save explicitly.
type Repository interface {
	// Save inserts u when it is new (ID == 0) or when no row with u.ID
	// exists, otherwise overwrites every column of the existing row. It
	// returns the persisted state, including the allocated id.
	Save(ctx context.Context, u *User) (*User, error)
	// FindAll returns every user in database order. The result is never nil.
	FindAll(ctx context.Context) ([]User, error)
	// FindByID returns the user with the given id, or an error matching
	// ErrNotFound when no such row exists.
	FindByID(ctx context.Context, id int) (*User, error)
	// DeleteByID removes the user with the given id. It returns an error
	// matching ErrEmptyResult when no such row exists.
	DeleteByID(ctx context.Context, id int) error
	// Delete removes the given user. A new (ID == 0) or already-absent user
	// is silently ignored, as in SimpleJpaRepository.delete.
	Delete(ctx context.Context, u *User) error
	// FindByName returns the single user whose name equals name. The bool is
	// false when no user matches. When several users match, the error
	// matches ErrIncorrectResultSize.
	FindByName(ctx context.Context, name string) (*User, bool, error)
}

// MySQLRepository is the database/sql implementation of Repository for MySQL.
// It is safe for concurrent use.
type MySQLRepository struct {
	db *sql.DB
}

// Compile-time check that MySQLRepository satisfies Repository.
var _ Repository = (*MySQLRepository)(nil)

// NewMySQLRepository returns a Repository backed by db. The schema must have
// been created beforehand with EnsureSchema.
func NewMySQLRepository(db *sql.DB) *MySQLRepository {
	return &MySQLRepository{db: db}
}

const selectColumns = "SELECT `user_id`, `user_name`, `user_email`, `user_password`, `user_role`, `user_about` FROM `user`"

// rowScanner is satisfied by both *sql.Row and *sql.Rows.
type rowScanner interface {
	Scan(dest ...any) error
}

func scanUser(s rowScanner) (*User, error) {
	var (
		u                                  User
		name, email, password, role, about sql.NullString
	)
	if err := s.Scan(&u.ID, &name, &email, &password, &role, &about); err != nil {
		return nil, err
	}
	u.Name = nullToPtr(name)
	u.Email = nullToPtr(email)
	u.Password = nullToPtr(password)
	u.Role = nullToPtr(role)
	u.About = nullToPtr(about)
	return &u, nil
}

func nullToPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	s := ns.String
	return &s
}

func ptrToNull(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}

// Save implements Repository.
//
// Semantics mirror SimpleJpaRepository.save / Hibernate merge:
//   - ID == 0: a fresh id is allocated from hibernate_sequence and the row
//     is inserted; u.ID is updated in place (persist mutates the entity).
//   - ID != 0 and the row exists: all five data columns are overwritten,
//     nil fields become NULL.
//   - ID != 0 and the row is missing: the row is inserted under a FRESH
//     sequence id, not the caller-supplied one (Hibernate merge behaviour).
//
// Driver errors (e.g. duplicate e-mail, about longer than 500 characters)
// are wrapped and returned unchanged in kind.
func (r *MySQLRepository) Save(ctx context.Context, u *User) (*User, error) {
	if u == nil {
		return nil, errors.New("user: save: entity must not be nil")
	}

	if u.ID != 0 {
		updated, err := r.updateIfExists(ctx, u)
		if err != nil {
			return nil, err
		}
		if updated {
			out := *u
			return &out, nil
		}
	}

	id, err := r.nextID(ctx)
	if err != nil {
		return nil, err
	}
	out := *u
	out.ID = id
	if _, err := r.db.ExecContext(ctx,
		"INSERT INTO `user` (`user_about`, `user_email`, `user_name`, `user_password`, `user_role`, `user_id`) VALUES (?, ?, ?, ?, ?, ?)",
		ptrToNull(out.About), ptrToNull(out.Email), ptrToNull(out.Name),
		ptrToNull(out.Password), ptrToNull(out.Role), out.ID,
	); err != nil {
		return nil, fmt.Errorf("user: insert: %w", err)
	}
	if u.ID == 0 {
		u.ID = id
	}
	return &out, nil
}

// updateIfExists overwrites the row for u.ID inside a transaction and reports
// whether the row existed. The existence check uses SELECT ... FOR UPDATE so
// that an UPDATE which changes no values is still recognised as a hit
// (MySQL's RowsAffected counts only changed rows).
func (r *MySQLRepository) updateIfExists(ctx context.Context, u *User) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("user: update: begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // no-op after Commit

	var exists int
	err = tx.QueryRowContext(ctx,
		"SELECT 1 FROM `user` WHERE `user_id` = ? FOR UPDATE", u.ID).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("user: update: lock row %d: %w", u.ID, err)
	}

	if _, err := tx.ExecContext(ctx,
		"UPDATE `user` SET `user_about` = ?, `user_email` = ?, `user_name` = ?, `user_password` = ?, `user_role` = ? WHERE `user_id` = ?",
		ptrToNull(u.About), ptrToNull(u.Email), ptrToNull(u.Name),
		ptrToNull(u.Password), ptrToNull(u.Role), u.ID,
	); err != nil {
		return false, fmt.Errorf("user: update %d: %w", u.ID, err)
	}
	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("user: update %d: commit: %w", u.ID, err)
	}
	return true, nil
}

// nextID allocates the next id from the Hibernate `hibernate_sequence` table
// in its own transaction, exactly like Hibernate's TableStructure with
// increment 1: read next_val under a row lock, advance it, use the old value.
func (r *MySQLRepository) nextID(ctx context.Context) (int, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("user: allocate id: begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // no-op after Commit

	var next int64
	err = tx.QueryRowContext(ctx,
		"SELECT `next_val` FROM `hibernate_sequence` FOR UPDATE").Scan(&next)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, errors.New("user: allocate id: hibernate_sequence is empty (was EnsureSchema called?)")
	}
	if err != nil {
		return 0, fmt.Errorf("user: allocate id: read sequence: %w", err)
	}
	if _, err := tx.ExecContext(ctx,
		"UPDATE `hibernate_sequence` SET `next_val` = ? WHERE `next_val` = ?", next+1, next); err != nil {
		return 0, fmt.Errorf("user: allocate id: advance sequence: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("user: allocate id: commit: %w", err)
	}
	return int(next), nil
}

// FindAll implements Repository. No ORDER BY is applied, matching the
// unsorted JpaRepository.findAll().
func (r *MySQLRepository) FindAll(ctx context.Context) ([]User, error) {
	rows, err := r.db.QueryContext(ctx, selectColumns)
	if err != nil {
		return nil, fmt.Errorf("user: find all: %w", err)
	}
	defer rows.Close()

	users := make([]User, 0)
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, fmt.Errorf("user: find all: scan: %w", err)
		}
		users = append(users, *u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("user: find all: %w", err)
	}
	return users, nil
}

// FindByID implements Repository.
func (r *MySQLRepository) FindByID(ctx context.Context, id int) (*User, error) {
	u, err := scanUser(r.db.QueryRowContext(ctx, selectColumns+" WHERE `user_id` = ?", id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("user: find by id %d: %w", id, err)
	}
	return u, nil
}

// DeleteByID implements Repository.
func (r *MySQLRepository) DeleteByID(ctx context.Context, id int) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM `user` WHERE `user_id` = ?", id)
	if err != nil {
		return fmt.Errorf("user: delete %d: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("user: delete %d: rows affected: %w", id, err)
	}
	if n == 0 {
		return fmt.Errorf("user: delete %d: %w", id, ErrEmptyResult)
	}
	return nil
}

// Delete implements Repository.
func (r *MySQLRepository) Delete(ctx context.Context, u *User) error {
	if u == nil {
		return errors.New("user: delete: entity must not be nil")
	}
	if u.ID == 0 {
		return nil
	}
	if _, err := r.db.ExecContext(ctx, "DELETE FROM `user` WHERE `user_id` = ?", u.ID); err != nil {
		return fmt.Errorf("user: delete %d: %w", u.ID, err)
	}
	return nil
}

// FindByName implements Repository (source: UserDao.findByName, a derived
// query "WHERE name = ?"). Equality follows the column collation, as it did
// under Hibernate.
func (r *MySQLRepository) FindByName(ctx context.Context, name string) (*User, bool, error) {
	rows, err := r.db.QueryContext(ctx, selectColumns+" WHERE `user_name` = ? LIMIT 2", name)
	if err != nil {
		return nil, false, fmt.Errorf("user: find by name: %w", err)
	}
	defer rows.Close()

	var found *User
	for rows.Next() {
		if found != nil {
			return nil, false, fmt.Errorf("user: find by name %q: %w", name, ErrIncorrectResultSize)
		}
		u, err := scanUser(rows)
		if err != nil {
			return nil, false, fmt.Errorf("user: find by name: scan: %w", err)
		}
		found = u
	}
	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("user: find by name: %w", err)
	}
	if found == nil {
		return nil, false, nil
	}
	return found, true, nil
}
