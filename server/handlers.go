package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
	"web-app/internal/models"
	"web-app/internal/utils"
	"web-app/internal/validator"
)

func (app *App) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var data struct {
		models.User
		CaptchaID     string `json:"captcha_id"`
		CaptchaAnswer string `json:"captcha_answer"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if !validator.IsValidName(data.FirstName) || !validator.IsValidName(data.LastName) {
		http.Error(w, "Invalid name format", http.StatusBadRequest)
		return
	}

	if !validator.IsValidEmail(data.Email) {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	if !validator.IsValidPassword(data.Password) {
		http.Error(w, "Invalid password format", http.StatusBadRequest)
		return
	}

	var dbCaptchaAnswer string
	err := app.DB.QueryRow("select answer from captchas where id = ? and expires_at > NOW()",
		data.CaptchaID).Scan(&dbCaptchaAnswer)
	if err != nil || dbCaptchaAnswer != data.CaptchaAnswer {
		log.Printf("DEBUG: Captcha DB Lookup Error: %v", err)
		http.Error(w, "Invalid captcha answer", http.StatusUnauthorized)
		return
	}

	exists, err := app.DB.EmailExists(data.Email)
	if err != nil || exists {
		log.Printf("DEBUG: EmailExists Error: %v", err)
		http.Error(w, "Email already registered", http.StatusConflict)
		return
	}

	userID, err := app.DB.CreateUser(&data.User)
	if err != nil {
		log.Printf("DEBUG: CreateUser Error: %v", err)
		http.Error(w, "Could not create user", http.StatusInternalServerError)
		return
	}

	sessionToken, err := utils.GenerateSecureToken(24)
	if err != nil {
		log.Printf("DEBUG: GenerateSecureToken Error: %v", err)
		http.Error(w, "Failed to generate session token", http.StatusInternalServerError)
		return
	}

	session := &models.Session{
		Token:     sessionToken,
		UserID:    int(userID),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := app.DB.CreateSession(session); err != nil {
		log.Printf("DEBUG: CreateSession Error: %v", err)
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    session.Token,
		Path:     "/",
		HttpOnly: true,
		Expires:  session.ExpiresAt,
	})

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
}
