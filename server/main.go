package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"web-app/internal/api"
	"web-app/internal/database"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

type App struct {
	DB *database.DB
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dsn := os.Getenv("DB_DSN")
	port := os.Getenv("PORT")

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Connected to database successfully.")

	customDB := &database.DB{DB: db}
	app := &App{
		DB: customDB,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /register", app.handleRegister)
	mux.HandleFunc("GET /captcha", api.HandleCaptcha(customDB))

	log.Printf("Server is running on port %s", port)
	http.ListenAndServe(":"+port, mux)

}
