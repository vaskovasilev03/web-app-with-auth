package database

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

type DB struct {
	*sql.DB
}

func InitDB(dsn string) (*DB, error) {
	db, err := sql.Open("mqsql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{db}, nil
}

func (db *DB) Close() error {
	return db.DB.Close()
}
