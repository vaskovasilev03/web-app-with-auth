package database

import (
	"web-app/internal/models"
)

func (db *DB) CreateSession(session *models.Session) error {
	query := "INSERT INTO sessions (token, user_id, expires_at) VALUES (?, ?, ?)"
	_, err := db.Exec(query, session.Token, session.UserID, session.ExpiresAt)
	return err
}
