```go
package repository

import (
	"migrated-app/internal/smartcontact/model"
	"database/sql"
	"fmt"
	"log"
)

// UserDao represents a user data access object.
type UserDao struct {
	DB *sql.DB
}

// NewUserDao creates a new instance of UserDao.
func NewUserDao(db *sql.DB) *UserDao {
	return &UserDao{DB: db}
}

// CreateUser inserts a new user into the database.
func (ud *UserDao) CreateUser(user model.User) (model.User, error) {
	sqlStatement := `
	INSERT INTO users (name, email, password, role, about)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id`
	err := ud.DB.QueryRow(sqlStatement, user.Name, user.Email, user.Password, user.Role, user.About).Scan(&user.ID)
	if err != nil {
		log.Println(fmt.Sprintf("Unable to execute query. %v", err))
		return user, err
	}
	return user, nil
}

// GetUser retrieves a user by ID.
func (ud *UserDao) GetUser(id string) (model.User, error) {
	sqlStatement := `SELECT id, name, email, password, role, about FROM users WHERE id=$1;`
	row := ud.DB.QueryRow(sqlStatement, id)
	var user model.User
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Role, &user.About)
	if err == sql.ErrNoRows {
		return user, fmt.Errorf("user with id %s not found", id)
	} else if err != nil {
		log.Println(fmt.Sprintf("Unable to scan row. %v", err))
		return user, err
	}
	return user, nil
}
```