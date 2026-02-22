package database

import (
	"time"
	"web-app/internal/models"
)

func (db *DB) CreateSession(session *models.Session) error {
	query := "INSERT INTO sessions (session_token, user_id, expires_at) VALUES (?, ?, ?)"
	_, err := db.Exec(query, session.Token, session.UserID, session.ExpiresAt)
	return err
}

func (db *DB) GetUserIDByToken(token string) (int, error) {
	var userID int
	query := "SELECT user_id FROM sessions WHERE session_token = ? AND expires_at > NOW()"

	err := db.QueryRow(query, token).Scan(&userID)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

func (db *DB) GetValidSessionToken(userID int) (string, error) {
	var token string
	query := "SELECT session_token FROM sessions WHERE user_id = ? AND expires_at > ?"
	err := db.QueryRow(query, userID, time.Now()).Scan(&token)
	return token, err
}

func (db *DB) UpdateSessionExpiry(token string, expiry time.Time) error {
	_, err := db.Exec("UPDATE sessions SET expires_at = ? WHERE session_token = ?", expiry, token)
	return err
}

func (db *DB) CleanupExpired(currentTime time.Time) error {
	if _, err := db.Exec("DELETE FROM sessions WHERE expires_at < ?", currentTime); err != nil {
		return err
	}
	_, err := db.Exec("DELETE FROM captchas WHERE expires_at < ?", currentTime)
	return err
}
