package database

import (
	"fmt"
	"web-app/internal/models"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

func (db *DB) CreateUser(user *models.User) error {
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	query := "insert into users (first_name, last_name, email, password) values (?, ?, ?, ?)"
	_, err = db.Exec(query, user.FirstName, user.LastName, user.Email, string(hashedPass))
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			return fmt.Errorf("email already registered")
		}
		return err
	}
	return nil
}

func (db *DB) EmailExists(email string) (bool, error) {
	var exists bool
	query := "select exists(select 1 from users where email = ?)"
	err := db.QueryRow(query, email).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
