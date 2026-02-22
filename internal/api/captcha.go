package api

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"
	"web-app/internal/database"
	"web-app/internal/utils"
)

type MathCaptcha struct {
	ID       string `json:"captcha_id"`
	Question string `json:"question"`
	Num1     int    `json:"num1"`
	Num2     int    `json:"num2"`
	Operator string `json:"operator"`
}

func HandleCaptcha(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		num1 := rand.Intn(20) + 1
		num2 := rand.Intn(20) + 1
		operators := []string{"+", "-", "*"}
		op := operators[rand.Intn(len(operators))]

		var result int
		switch op {
		case "+":
			result = num1 + num2
		case "-":
			result = num1 - num2
		case "*":
			result = num1 * num2
		}

		captchaID, err := utils.GenerateSecureToken(16)
		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}

		expiresAt := time.Now().Add(5 * time.Minute)
		_, err = db.Exec("INSERT INTO captchas (id, answer, expires_at) VALUES (?, ?, ?)",
			captchaID, fmt.Sprintf("%d", result), expiresAt)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		resp := MathCaptcha{
			ID:       captchaID,
			Question: "What is the answer to this equation?",
			Num1:     num1,
			Num2:     num2,
			Operator: op,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
