package repository

import (
	"context"
	"fmt"

	"migrated-app/internal/smartcontact/model"
)

// FindAll retrieves all users from the database.
// This method extends the UserRepository interface implementation on userDao.
func (d *userDao) FindAll(ctx context.Context) ([]*model.User, error) {
	const querySQL = `SELECT id, name, email, password, role, about FROM users ORDER BY id`
	rows, err := d.db.QueryContext(ctx, querySQL)
	if err != nil {
		return nil, fmt.Errorf("repository: find all users: %w", err)
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		u, err := d.scanUser(rows)
		if err != nil {
			return nil, fmt.Errorf("repository: scan user row: %w", err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: find all users rows: %w", err)
	}
	return users, nil
}