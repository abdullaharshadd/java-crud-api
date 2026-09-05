package repository

import (
	"migrated-app/internal/smartcontact/model"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

var db *sqlx.DB

func InitDB(dataSourceName string) error {
	var err error
	db, err = sqlx.Open("postgres", dataSourceName)
	if err != nil {
		return err
	}
	return db.Ping()
}

func GetUser(id string) (*model.User, error) {
	var user model.User
	err := db.Get(&user, "SELECT * FROM users WHERE id = $1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, pq.Error{Code: "404"}
		}
		return nil, err
	}
	return &user, nil
}

func UpdateUser(user model.User) error {
	_, err := db.NamedExec("UPDATE users SET name=:name, email=:email WHERE id=:id", user)
	return err
}

func DeleteUser(id string) error {
	_, err := db.Exec("DELETE FROM users WHERE id=$1", id)
	return err
}