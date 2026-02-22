package server_tests

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"web-app/internal/database"
	"web-app/pkg/server"

	"database/sql"

	"github.com/joho/godotenv"
)

var app *server.App

func TestMain(m *testing.M) {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	serverDSN := os.Getenv("MYSQLSERVER")
	if serverDSN == "" {
		absPath, _ := filepath.Abs("../../.env")
		log.Fatalf("MYSQLSERVER not set in .env file %s", absPath)
	}

	tmpDB, err := sql.Open("mysql", serverDSN)
	if err != nil {
		log.Fatalf("Could not connect to MySQL for setup: %v", err)
	}

	_, err = tmpDB.Exec("CREATE DATABASE IF NOT EXISTS simple_web_app_test")
	if err != nil {
		log.Fatalf("Could not create test database: %v", err)
	}
	tmpDB.Close()

	db := database.NewTestDB(&testing.T{})
	app = server.NewApp(db)

	exitCode := m.Run()

	db.Close()
	os.Exit(exitCode)
}
