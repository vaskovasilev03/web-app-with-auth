package server_tests

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"web-app/internal/models"
)

func TestHandleLogout(t *testing.T) {
	req, err := http.NewRequest("POST", "/logout", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.HandleLogout)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	cookies := rr.Result().Cookies()
	foundCookie := false
	for _, cookie := range cookies {
		if cookie.Name == "session_token" {
			foundCookie = true
			if cookie.MaxAge != -1 {
				t.Errorf("session_token cookie MaxAge is not -1: got %v", cookie.MaxAge)
			}
		}
	}

	if !foundCookie {
		t.Error("session_token cookie not found in response")
	}
}

func TestHandleLogin_Success(t *testing.T) {
	user := &models.User{
		FirstName: "Test",
		LastName:  "User",
		Email:     "login_success@test.com",
		Password:  "Password123!",
	}
	app.DB.SeedUser(t, user)

	body := `{"email":"login_success@test.com", "password":"Password123!"}`
	req, err := http.NewRequest("POST", "/login", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.HandleLogin)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	cookies := rr.Result().Cookies()
	foundCookie := false
	for _, cookie := range cookies {
		if cookie.Name == "session_token" {
			foundCookie = true
		}
	}
	if !foundCookie {
		t.Error("session_token cookie not found in response")
	}
}
