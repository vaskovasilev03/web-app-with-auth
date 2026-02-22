package server_tests

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"web-app/internal/models"
	"web-app/pkg/server"
)

func uniqueEmail(prefix string) string {
	return fmt.Sprintf("%s_%d@test.com", prefix, time.Now().UnixNano())
}

func TestNewApp(t *testing.T) {
	a := server.NewApp(app.DB)
	if a == nil {
		t.Fatal("expected app instance, got nil")
	}
	if a.DB != app.DB {
		t.Fatal("expected NewApp to keep provided DB reference")
	}
}

func TestRequireAuth_UnauthenticatedRedirects(t *testing.T) {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler := app.RequireAuth(next)
	req := httptest.NewRequest(http.MethodGet, "/profile", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected status %d, got %d", http.StatusSeeOther, rr.Code)
	}
	if nextCalled {
		t.Fatal("next handler should not be called for unauthenticated request")
	}
	if location := rr.Header().Get("Location"); location != "/login" {
		t.Fatalf("expected redirect to /login, got %q", location)
	}
}

func TestRequireAuth_AuthenticatedCallsNext(t *testing.T) {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusNoContent)
	})

	handler := app.RequireAuth(next)
	req := httptest.NewRequest(http.MethodGet, "/profile", nil)
	req = req.WithContext(context.WithValue(req.Context(), "userID", 1))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if !nextCalled {
		t.Fatal("next handler was not called")
	}
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rr.Code)
	}
}

func TestRedirectIfAuthenticated_AuthenticatedRedirects(t *testing.T) {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler := app.RedirectIfAuthenticated(next)
	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	req = req.WithContext(context.WithValue(req.Context(), "userID", 1))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected status %d, got %d", http.StatusSeeOther, rr.Code)
	}
	if nextCalled {
		t.Fatal("next handler should not be called for authenticated request")
	}
	if location := rr.Header().Get("Location"); location != "/" {
		t.Fatalf("expected redirect to /, got %q", location)
	}
}

func TestRedirectIfAuthenticated_UnauthenticatedCallsNext(t *testing.T) {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusCreated)
	})

	handler := app.RedirectIfAuthenticated(next)
	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if !nextCalled {
		t.Fatal("next handler was not called")
	}
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}
}

func TestSessionLoader_NoCookie(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := r.Context().Value("userID").(int); ok {
			t.Fatal("did not expect userID in context")
		}
		w.WriteHeader(http.StatusOK)
	})

	handler := app.SessionLoader(next)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestSessionLoader_InvalidCookie(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := r.Context().Value("userID").(int); ok {
			t.Fatal("did not expect userID in context")
		}
		w.WriteHeader(http.StatusOK)
	})

	handler := app.SessionLoader(next)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: "missing-token"})
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestSessionLoader_ValidCookieSetsContext(t *testing.T) {
	user := &models.User{
		FirstName: "Token",
		LastName:  "User",
		Email:     uniqueEmail("session_loader"),
		Password:  "Password123!",
	}
	userID := app.DB.SeedUser(t, user)

	loginBody := fmt.Sprintf(`{"email":"%s", "password":"Password123!"}`, user.Email)
	loginReq := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(loginBody))
	loginRR := httptest.NewRecorder()
	http.HandlerFunc(app.HandleLogin).ServeHTTP(loginRR, loginReq)
	if loginRR.Code != http.StatusOK {
		t.Fatalf("expected login status %d, got %d", http.StatusOK, loginRR.Code)
	}

	var sessionToken string
	for _, cookie := range loginRR.Result().Cookies() {
		if cookie.Name == "session_token" {
			sessionToken = cookie.Value
			break
		}
	}
	if sessionToken == "" {
		t.Fatal("expected session_token cookie from login")
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxUserID, ok := r.Context().Value("userID").(int)
		if !ok {
			t.Fatal("expected userID in context")
		}
		if ctxUserID != int(userID) {
			t.Fatalf("expected userID %d, got %d", userID, ctxUserID)
		}
		w.WriteHeader(http.StatusOK)
	})

	handler := app.SessionLoader(next)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: sessionToken})
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}
