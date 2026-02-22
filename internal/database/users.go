package database

import (
	"fmt"
	"web-app/internal/models"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

func (db *DB) CreateUser(user *models.User) (int64, error) {
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	query := "insert into users (first_name, last_name, email, password_hash) values (?, ?, ?, ?)"
	result, err := db.Exec(query, user.FirstName, user.LastName, user.Email, string(hashedPass))
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			return 0, fmt.Errorf("email already registered")
		}
		return 0, err
	}
	return result.LastInsertId()
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
