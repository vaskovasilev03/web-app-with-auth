package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
	"web-app/internal/models"
	"web-app/internal/utils"
	"web-app/internal/validator"
)

func (app *App) HandleRegister(w http.ResponseWriter, r *http.Request) {
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

func (app *App) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	userID, err := app.DB.Authenticate(input.Email, input.Password)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := app.DB.GetValidSessionToken(userID)
	expiresAt := time.Now().Add(24 * time.Hour)

	if err == nil && token != "" {
		if err := app.DB.UpdateSessionExpiry(token, expiresAt); err != nil {
			log.Printf("DEBUG: UpdateSessionExpiry Error: %v", err)
			http.Error(w, "Failed to update session", http.StatusInternalServerError)
			return
		}
	} else {
		token, _ = utils.GenerateSecureToken(32)
		session := &models.Session{
			Token:     token,
			UserID:    userID,
			ExpiresAt: expiresAt,
		}
		if err := app.DB.CreateSession(session); err != nil {
			log.Printf("DEBUG: CreateSession Error: %v", err)
			http.Error(w, "Failed to create session", http.StatusInternalServerError)
			return
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "User logged in successfully"})
}

func (app *App) HomeHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./web/index.html")
}

func (app *App) SessionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		json.NewEncoder(w).Encode(map[string]interface{}{"authenticated": false})
		return
	}

	user, err := app.DB.GetUserByID(userID)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"authenticated": false})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"authenticated": true,
		"firstName":     user.FirstName,
		"lastName":      user.LastName,
	})
}

func (app *App) HandleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Logged out")
}

func (app *App) HandleUpdateName(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var input struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if !validator.IsValidName(input.FirstName) || !validator.IsValidName(input.LastName) {
		http.Error(w, "Invalid name format", http.StatusBadRequest)
		return
	}

	if err := app.DB.UpdateUser(userID, input.FirstName, input.LastName); err != nil {
		log.Printf("DEBUG: UpdateUser Error: %v", err)
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Profile updated successfully"})
}

func (app *App) HandleUpdatePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var input struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if !validator.IsValidPassword(input.NewPassword) {
		http.Error(w, "Invalid password format", http.StatusBadRequest)
		return
	}

	if err := app.DB.VerifyPassword(userID, input.CurrentPassword); err != nil {
		http.Error(w, "Incorrect current password", http.StatusUnauthorized)
		return
	}

	if err := app.DB.UpdatePassword(userID, input.NewPassword); err != nil {
		log.Printf("DEBUG: UpdatePassword Error: %v", err)
		http.Error(w, "Failed to update password", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Password updated successfully"})
}
