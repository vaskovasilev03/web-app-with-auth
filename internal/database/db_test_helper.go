package database

import (
	"database/sql"
	"os"
	"testing"

	"web-app/internal/models"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

func NewTestDB(t *testing.T) *DB {

	err := godotenv.Load("../../.env")
	if err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	testDSN := os.Getenv("TEST_DB_DSN")
	if testDSN == "" {
		t.Fatal("TEST_DB_DSN not set in .env file")
	}

	db, err := sql.Open("mysql", testDSN)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	_, err = db.Exec("drop tables if exists users, sessions, captchas")
	if err != nil {
		t.Fatalf("Failed to drop tables: %v", err)
	}

	createTables(db, t)

	return &DB{DB: db}
}

func createTables(db *sql.DB, t *testing.T) {
	queries := []string{
		`CREATE TABLE users (
			id INT AUTO_INCREMENT PRIMARY KEY,
			first_name VARCHAR(255) NOT NULL,
			last_name VARCHAR(255) NOT NULL,
			email VARCHAR(255) NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)`,
		`CREATE TABLE sessions (
			session_token VARCHAR(255) PRIMARY KEY,
			user_id INT NOT NULL,
			expires_at TIMESTAMP NOT NULL
			)`,
		`CREATE TABLE captchas (
			id VARCHAR(255) PRIMARY KEY,
			answer VARCHAR(255) NOT NULL,
			expires_at TIMESTAMP NOT NULL
			)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			t.Fatalf("Failed to create tables: %v", err)
		}
	}

}

func (db *DB) SeedUser(t *testing.T, user *models.User) int64 {

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	user.Password = string(hashedPassword)

	id, err := db.CreateUser(user)
	if err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}
	return id
}

func (db *DB) SeedCaptcha(t *testing.T, id, answer string) {
	_, err := db.Exec("INSERT INTO captchas (id, answer, expires_at) VALUES (?, ?, NOW() + INTERVAL 5 MINUTE)",
		id, answer)
	if err != nil {
		t.Fatalf("failed to seed captcha: %v", err)
	}
}
