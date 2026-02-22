package internal_tests

import (
	"os"
	"testing"
	"time"
	"web-app/internal/database"
	"web-app/internal/models"
	"web-app/internal/utils"

	"golang.org/x/crypto/bcrypt"
)

func TestDatabaseInternalLogic(t *testing.T) {
	db := database.NewTestDB(t)
	defer db.Close()

	t.Run("CreateUser", func(t *testing.T) {
		user := &models.User{
			FirstName: "Test",
			LastName:  "User",
			Email:     "logic@test.com",
			Password:  "pSss122222!",
		}
		_, err := db.CreateUser(user)
		if err != nil {
			t.Fatalf("CreateUser failed: %v", err)
		}
	})

	t.Run("CreateUserDuplicateError", func(t *testing.T) {
		user := &models.User{
			FirstName: "Duplicate",
			LastName:  "User",
			Email:     "duplicate@test.com",
			Password:  "Duplicate%123!",
		}
		_, err := db.CreateUser(user)
		if err != nil {
			t.Fatalf("CreateUser failed: %v", err)
		}

		_, err = db.CreateUser(user)
		if err == nil {
			t.Error("Expected an error for duplicate email, but got nil")
		}
	})

	t.Run("CheckEmailExists", func(t *testing.T) {
		user := &models.User{
			FirstName: "Existing",
			LastName:  "User",
			Email:     "already_here@test.com",
			Password:  "Secret123!",
		}
		db.SeedUser(t, user)
		exists, err := db.EmailExists(user.Email)
		if err != nil {
			t.Errorf("EmailExists errored: %v", err)
		}
		if !exists {
			t.Error("Expected EmailExists to return true for seeded user")
		}
	})

	t.Run("CheckEmailExistsFalse", func(t *testing.T) {
		email := "not_in_db@test.com"
		exists, err := db.EmailExists(email)
		if err != nil {
			t.Errorf("EmailExists errored: %v", err)
		}
		if exists {
			t.Error("Expected EmailExists to return false for non-existent email")
		}
	})

	t.Run("CreateUser_GenericDatabaseError", func(t *testing.T) {
		tempDB := database.NewTestDB(t)
		tempDB.Close()

		user := &models.User{
			FirstName: "Error",
			LastName:  "Test",
			Email:     "error@test.com",
			Password:  "Ppassword123!",
		}

		_, err := tempDB.CreateUser(user)

		if err == nil {
			t.Error("Expected connection error, but got nil")
		}
	})

	t.Run("EmailExists_DatabaseError", func(t *testing.T) {
		tempDB := database.NewTestDB(t)

		tempDB.Close()

		_, err := tempDB.EmailExists("test@test.com")

		if err == nil {
			t.Error("Expected error from closed database, but got nil")
		}
	})

	t.Run("CreateUser_BcryptError", func(t *testing.T) {
		longPassword := make([]byte, 73)
		for i := range longPassword {
			longPassword[i] = 'a'
		}

		user := &models.User{
			FirstName: "Long",
			LastName:  "Pass",
			Email:     "long@test.com",
			Password:  string(longPassword),
		}

		_, err := db.CreateUser(user)
		if err == nil {
			t.Error("Expected error for password over 72 chars, but got nil")
		}
	})

	t.Run("CaptchaLogic", func(t *testing.T) {
		captchaID := "test-math-id"
		answer := "25"

		db.SeedCaptcha(t, captchaID, answer)

		var dbAnswer string
		err := db.QueryRow("SELECT answer FROM captchas WHERE id = ? AND expires_at > NOW()", captchaID).Scan(&dbAnswer)
		if err != nil || dbAnswer != answer {
			t.Errorf("Captcha retrieval failed: %v", err)
		}
	})

	t.Run("ExpiredCaptcha", func(t *testing.T) {
		id := "expired-id"
		_, err := db.Exec("INSERT INTO captchas (id, answer, expires_at) VALUES (?, ?, NOW() - INTERVAL 1 MINUTE)", id, "10")
		if err != nil {
			t.Fatal(err)
		}

		var answer string
		err = db.QueryRow("SELECT answer FROM captchas WHERE id = ? AND expires_at > NOW()", id).Scan(&answer)
		if err == nil {
			t.Error("Expected error for expired captcha, but found none")
		}
	})

	t.Run("TokenLogic", func(t *testing.T) {
		token1, _ := utils.GenerateSecureToken(16)
		token2, _ := utils.GenerateSecureToken(16)

		if token1 == token2 {
			t.Error("GenerateSecureToken produced duplicate tokens")
		}
		if len(token1) != 32 {
			t.Errorf("Expected token length 32, got %d", len(token1))
		}
	})

	t.Run("VerifyPasswordHash", func(t *testing.T) {
		user := &models.User{
			FirstName: "Hash",
			LastName:  "User",
			Email:     "check_hash@test.com",
			Password:  "SecurePass123!",
		}
		db.SeedUser(t, user)

		email := user.Email
		pass := user.Password

		var storedHash string
		err := db.QueryRow("SELECT password_hash FROM users WHERE email = ?", email).Scan(&storedHash)
		if err != nil {
			t.Fatalf("Could not fetch seeded user: %v", err)
		}

		err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(pass))
		if err != nil {
			t.Errorf("Password verification failed for seeded user: %v", err)
		}
	})

	t.Run("SessionLogic", func(t *testing.T) {
		token, _ := utils.GenerateSecureToken(32)
		session := &models.Session{
			Token:     token,
			UserID:    1,
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}

		err := db.CreateSession(session)
		if err != nil {
			t.Errorf("CreateSession failed: %v", err)
		}
	})

	t.Run("NewDBConnection", func(t *testing.T) {
		dsn := os.Getenv("MYSQLSERVER") + "web_app_test"
		db, err := database.InitDB("mysql", dsn)
		if err != nil {
			t.Errorf("NewDB failed with valid DSN: %v", err)
		}
		if db != nil {
			db.Close()
		}

		_, err = database.InitDB("mysql", "wrong_user:wrong_pass@tcp(localhost:9999)/invalid")
		if err == nil {
			t.Error("Expected error from InitDB with invalid DSN, but got nil")
		}

		_, err = database.InitDB("unsupported_db", dsn)
		if err == nil {
			t.Error("Expected error from InitDB with unsupported driver, but got nil")
		}
	})

	t.Run("InitDB_ErrorPaths", func(t *testing.T) {
		badDSN := "root:totally_wrong_password@tcp(127.0.0.1:3306)/web_app_test"
		_, err := database.InitDB("mysql", badDSN)
		if err == nil {
			t.Error("Expected error from db.Ping() with bad credentials, but got nil")
		}
	})
}
