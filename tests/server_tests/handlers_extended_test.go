package server_tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"web-app/internal/models"
)

func TestHandleLogin_MethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	rr := httptest.NewRecorder()

	http.HandlerFunc(app.HandleLogin).ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestHandleLogin_BadRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("{invalid-json"))
	rr := httptest.NewRecorder()

	http.HandlerFunc(app.HandleLogin).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestHandleLogin_InvalidCredentials(t *testing.T) {
	body := `{"email":"missing@test.com", "password":"Password123!"}`
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
	rr := httptest.NewRecorder()

	http.HandlerFunc(app.HandleLogin).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestHandleLogin_ReusesExistingSession(t *testing.T) {
	user := &models.User{
		FirstName: "Existing",
		LastName:  "Session",
		Email:     uniqueEmail("login_existing"),
		Password:  "Password123!",
	}
	userID := app.DB.SeedUser(t, user)

	token := fmt.Sprintf("existing_session_%d", time.Now().UnixNano())
	err := app.DB.CreateSession(&models.Session{
		Token:     token,
		UserID:    int(userID),
		ExpiresAt: time.Now().Add(30 * time.Minute),
	})
	if err != nil {
		t.Fatalf("failed to seed existing session: %v", err)
	}

	body := fmt.Sprintf(`{"email":"%s", "password":"Password123!"}`, user.Email)
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
	rr := httptest.NewRecorder()

	http.HandlerFunc(app.HandleLogin).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	cookies := rr.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "session_token" {
			found = true
			if c.Value != token {
				t.Fatalf("expected existing token %q, got %q", token, c.Value)
			}
		}
	}
	if !found {
		t.Fatal("expected session_token cookie")
	}
}

func TestHandleRegister_MethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/register", nil)
	rr := httptest.NewRecorder()

	http.HandlerFunc(app.HandleRegister).ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestHandleRegister_BadPayload(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader("{bad"))
	rr := httptest.NewRecorder()

	http.HandlerFunc(app.HandleRegister).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestHandleRegister_InvalidName(t *testing.T) {
	body := fmt.Sprintf(`{"first_name":"John123","last_name":"Doe","email":"%s","password":"Password123!","captcha_id":"x","captcha_answer":"1234"}`,
		uniqueEmail("register_invalid_name"))
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body))
	rr := httptest.NewRecorder()

	http.HandlerFunc(app.HandleRegister).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestHandleRegister_InvalidEmail(t *testing.T) {
	body := `{"first_name":"John","last_name":"Doe","email":"invalid-email","password":"Password123!","captcha_id":"x","captcha_answer":"1234"}`
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body))
	rr := httptest.NewRecorder()

	http.HandlerFunc(app.HandleRegister).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestHandleRegister_InvalidPassword(t *testing.T) {
	body := fmt.Sprintf(`{"first_name":"John","last_name":"Doe","email":"%s","password":"short","captcha_id":"x","captcha_answer":"1234"}`,
		uniqueEmail("register_invalid_password"))
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body))
	rr := httptest.NewRecorder()

	http.HandlerFunc(app.HandleRegister).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestHandleRegister_InvalidCaptcha(t *testing.T) {
	body := fmt.Sprintf(`{"first_name":"John","last_name":"Doe","email":"%s","password":"Password123!","captcha_id":"missing","captcha_answer":"1234"}`,
		uniqueEmail("register_invalid_captcha"))
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body))
	rr := httptest.NewRecorder()

	http.HandlerFunc(app.HandleRegister).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestHandleRegister_EmailExists(t *testing.T) {
	existing := &models.User{
		FirstName: "Jane",
		LastName:  "Doe",
		Email:     uniqueEmail("register_exists"),
		Password:  "Password123!",
	}
	app.DB.SeedUser(t, existing)

	captchaID := fmt.Sprintf("captcha_%d", time.Now().UnixNano())
	app.DB.SeedCaptcha(t, captchaID, "7777")

	body := fmt.Sprintf(`{"first_name":"Jane","last_name":"Doe","email":"%s","password":"Password123!","captcha_id":"%s","captcha_answer":"7777"}`,
		existing.Email, captchaID)
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body))
	rr := httptest.NewRecorder()

	http.HandlerFunc(app.HandleRegister).ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, rr.Code)
	}
}

func TestHandleRegister_Success(t *testing.T) {
	captchaID := fmt.Sprintf("captcha_%d", time.Now().UnixNano())
	app.DB.SeedCaptcha(t, captchaID, "4242")

	email := uniqueEmail("register_success")
	body := fmt.Sprintf(`{"first_name":"John","last_name":"Doe","email":"%s","password":"Password123!","captcha_id":"%s","captcha_answer":"4242"}`,
		email, captchaID)
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body))
	rr := httptest.NewRecorder()

	http.HandlerFunc(app.HandleRegister).ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}

	cookies := rr.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "session_token" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected session_token cookie")
	}
}

func TestSessionHandler_Unauthenticated(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/session", nil)
	rr := httptest.NewRecorder()

	http.HandlerFunc(app.SessionHandler).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var body map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if authenticated, ok := body["authenticated"].(bool); !ok || authenticated {
		t.Fatalf("expected authenticated=false, got %v", body["authenticated"])
	}
}

func TestSessionHandler_Authenticated(t *testing.T) {
	user := &models.User{
		FirstName: "Auth",
		LastName:  "User",
		Email:     uniqueEmail("session_handler"),
		Password:  "Password123!",
	}
	userID := app.DB.SeedUser(t, user)

	req := httptest.NewRequest(http.MethodGet, "/session", nil)
	req = req.WithContext(context.WithValue(req.Context(), "userID", int(userID)))
	rr := httptest.NewRecorder()

	http.HandlerFunc(app.SessionHandler).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var body map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if authenticated, ok := body["authenticated"].(bool); !ok || !authenticated {
		t.Fatalf("expected authenticated=true, got %v", body["authenticated"])
	}
}

func TestSessionHandler_UserNotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/session", nil)
	req = req.WithContext(context.WithValue(req.Context(), "userID", 9999999))
	rr := httptest.NewRecorder()

	http.HandlerFunc(app.SessionHandler).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var body map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if authenticated, ok := body["authenticated"].(bool); !ok || authenticated {
		t.Fatalf("expected authenticated=false, got %v", body["authenticated"])
	}
}

func TestHandleLogout_MethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/logout", nil)
	rr := httptest.NewRecorder()

	http.HandlerFunc(app.HandleLogout).ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestHandleUpdateName_MethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/profile/name", nil)
	rr := httptest.NewRecorder()

	http.HandlerFunc(app.HandleUpdateName).ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestHandleUpdateName_Unauthorized(t *testing.T) {
	body := `{"first_name":"New","last_name":"Name"}`
	req := httptest.NewRequest(http.MethodPut, "/profile/name", strings.NewReader(body))
	rr := httptest.NewRecorder()

	http.HandlerFunc(app.HandleUpdateName).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestHandleUpdateName_BadRequest(t *testing.T) {
	user := &models.User{FirstName: "Bad", LastName: "Req", Email: uniqueEmail("update_name_bad"), Password: "Password123!"}
	userID := app.DB.SeedUser(t, user)

	req := httptest.NewRequest(http.MethodPut, "/profile/name", strings.NewReader("{bad"))
	req = req.WithContext(context.WithValue(req.Context(), "userID", int(userID)))
	rr := httptest.NewRecorder()

	http.HandlerFunc(app.HandleUpdateName).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestHandleUpdateName_InvalidName(t *testing.T) {
	user := &models.User{FirstName: "Bad", LastName: "Name", Email: uniqueEmail("update_name_invalid"), Password: "Password123!"}
	userID := app.DB.SeedUser(t, user)

	body := `{"first_name":"New123","last_name":"Name"}`
	req := httptest.NewRequest(http.MethodPut, "/profile/name", strings.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), "userID", int(userID)))
	rr := httptest.NewRecorder()

	http.HandlerFunc(app.HandleUpdateName).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestHandleUpdateName_Success(t *testing.T) {
	user := &models.User{
		FirstName: "Old",
		LastName:  "Name",
		Email:     uniqueEmail("update_name"),
		Password:  "Password123!",
	}
	userID := app.DB.SeedUser(t, user)

	body := `{"first_name":"New","last_name":"Name"}`
	req := httptest.NewRequest(http.MethodPut, "/profile/name", strings.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), "userID", int(userID)))
	rr := httptest.NewRecorder()

	http.HandlerFunc(app.HandleUpdateName).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestHandleUpdatePassword_Success(t *testing.T) {
	user := &models.User{
		FirstName: "Pass",
		LastName:  "Change",
		Email:     uniqueEmail("update_password"),
		Password:  "Password123!",
	}
	userID := app.DB.SeedUser(t, user)

	body := `{"current_password":"Password123!","new_password":"NewPass123!"}`
	req := httptest.NewRequest(http.MethodPut, "/profile/password", strings.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), "userID", int(userID)))
	rr := httptest.NewRecorder()

	http.HandlerFunc(app.HandleUpdatePassword).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestHandleUpdatePassword_MethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/profile/password", nil)
	rr := httptest.NewRecorder()

	http.HandlerFunc(app.HandleUpdatePassword).ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestHandleUpdatePassword_Unauthorized(t *testing.T) {
	body := `{"current_password":"Password123!","new_password":"NewPass123!"}`
	req := httptest.NewRequest(http.MethodPut, "/profile/password", strings.NewReader(body))
	rr := httptest.NewRecorder()

	http.HandlerFunc(app.HandleUpdatePassword).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestHandleUpdatePassword_BadRequest(t *testing.T) {
	user := &models.User{FirstName: "Bad", LastName: "Req", Email: uniqueEmail("update_pass_bad"), Password: "Password123!"}
	userID := app.DB.SeedUser(t, user)

	req := httptest.NewRequest(http.MethodPut, "/profile/password", strings.NewReader("{bad"))
	req = req.WithContext(context.WithValue(req.Context(), "userID", int(userID)))
	rr := httptest.NewRecorder()

	http.HandlerFunc(app.HandleUpdatePassword).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestHandleUpdatePassword_InvalidPassword(t *testing.T) {
	user := &models.User{FirstName: "Invalid", LastName: "Pass", Email: uniqueEmail("update_pass_invalid"), Password: "Password123!"}
	userID := app.DB.SeedUser(t, user)

	body := `{"current_password":"Password123!","new_password":"short"}`
	req := httptest.NewRequest(http.MethodPut, "/profile/password", strings.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), "userID", int(userID)))
	rr := httptest.NewRecorder()

	http.HandlerFunc(app.HandleUpdatePassword).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestHandleUpdatePassword_IncorrectCurrentPassword(t *testing.T) {
	user := &models.User{FirstName: "Wrong", LastName: "Current", Email: uniqueEmail("update_pass_wrong_current"), Password: "Password123!"}
	userID := app.DB.SeedUser(t, user)

	body := `{"current_password":"WrongPassword123!","new_password":"NewPass123!"}`
	req := httptest.NewRequest(http.MethodPut, "/profile/password", strings.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), "userID", int(userID)))
	rr := httptest.NewRecorder()

	http.HandlerFunc(app.HandleUpdatePassword).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestHomeHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	http.HandlerFunc(app.HomeHandler).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK && rr.Code != http.StatusNotFound {
		t.Fatalf("expected status %d or %d, got %d", http.StatusOK, http.StatusNotFound, rr.Code)
	}
}
