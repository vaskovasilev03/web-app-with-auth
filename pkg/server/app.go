package server

import (
	"web-app/internal/database"
)

type App struct {
	DB *database.DB
}

func NewApp(db *database.DB) *App {
	return &App{DB: db}
}
