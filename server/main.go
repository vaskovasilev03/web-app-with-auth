package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

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

	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			if err := customDB.CleanupExpired(time.Now()); err != nil {
				log.Printf("Error cleaning up expired rows: %v", err)
			}
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", app.HomeHandler)
	mux.Handle("GET /api/session", app.SessionLoader(http.HandlerFunc(app.SessionHandler)))

	mux.Handle("GET /register", app.SessionLoader(app.RedirectIfAuthenticated(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/register.html")
	}))))
	mux.Handle("GET /login", app.SessionLoader(app.RedirectIfAuthenticated(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/login.html")
	}))))

	fileServer := http.FileServer(http.Dir("./web/static"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fileServer))
	mux.HandleFunc("GET /captcha", api.HandleCaptcha(customDB))

	mux.HandleFunc("POST /register", app.handleRegister)
	mux.HandleFunc("POST /login", app.handleLogin)
	mux.HandleFunc("POST /logout", app.handleLogout)

	mux.Handle("GET /profile", app.SessionLoader(app.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/profile.html")
	}))))
	mux.Handle("PUT /profile/updateName", app.SessionLoader(app.RequireAuth(http.HandlerFunc(app.handleUpdateName))))
	mux.Handle("PUT /profile/updatePassword", app.SessionLoader(app.RequireAuth(http.HandlerFunc(app.handleUpdatePassword))))

	log.Printf("Server is running on port %s", port)
	http.ListenAndServe(":"+port, mux)

}
