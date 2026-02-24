package database

import (
	"log"
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
	query := "SELECT session_token FROM sessions WHERE user_id = ? AND expires_at > NOW()"
	err := db.QueryRow(query, userID).Scan(&token)
	return token, err
}

func (db *DB) UpdateSessionExpiry(token string) error {
	_, err := db.Exec("UPDATE sessions SET expires_at = DATE_ADD(NOW(), INTERVAL 24 HOUR) WHERE session_token = ?", token)
	return err
}

func (db *DB) CleanupExpired() error {
	sessionsResult, err := db.Exec("DELETE FROM sessions WHERE expires_at < NOW()")
	if err != nil {
		return err
	}

	captchasResult, err := db.Exec("DELETE FROM captchas WHERE expires_at < NOW()")
	if err != nil {
		return err
	}

	sessionsDeleted, _ := sessionsResult.RowsAffected()
	captchasDeleted, _ := captchasResult.RowsAffected()
	log.Printf("CleanupExpired completed: sessions=%d, captchas=%d", sessionsDeleted, captchasDeleted)

	return nil
}
