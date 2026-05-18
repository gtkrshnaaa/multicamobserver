package feature_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gtkrshnaaa/multicamobserver/internal/middleware"
)

var testSecret = []byte("super-secret-key-for-testing-purposes-2026")

// dummyHandler is a mock endpoint that we protect using the middleware
func dummyHandler(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	if claims == nil {
		http.Error(w, "No claims found", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Access Granted: " + claims.Subject))
}

// TestGenerateJWT verifies JWT creation does not return errors and parses properly
func TestGenerateJWT(t *testing.T) {
	token, err := middleware.GenerateJWT("admin@multicamobserver.com", "admin", testSecret)
	if err != nil {
		t.Fatalf("Failed to generate JWT: %v", err)
	}
	if len(token) == 0 {
		t.Errorf("Expected token to be non-empty")
	}
}

// TestRequireAuthMissingCookie checks redirect to /login when no cookie is present
func TestRequireAuthMissingCookie(t *testing.T) {
	handler := middleware.RequireAuth("admin", testSecret, dummyHandler)

	req := httptest.NewRequest("GET", "/admin/dashboard", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Expect redirect status code 303
	if rr.Code != http.StatusSeeOther {
		t.Errorf("Expected status code 303, got %d", rr.Code)
	}

	location := rr.Header().Get("Location")
	if location != "/login" {
		t.Errorf("Expected redirect to /login, got %s", location)
	}
}

// TestRequireAuthInvalidToken checks clean rollback and redirect when token is corrupted
func TestRequireAuthInvalidToken(t *testing.T) {
	handler := middleware.RequireAuth("admin", testSecret, dummyHandler)

	req := httptest.NewRequest("GET", "/admin/dashboard", nil)
	// Inject invalid cookie value
	req.AddCookie(&http.Cookie{
		Name:  "auth_token",
		Value: "invalid-garbage-token-string",
	})
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Expect redirect status code 303
	if rr.Code != http.StatusSeeOther {
		t.Errorf("Expected status code 303, got %d", rr.Code)
	}

	// Verify that the invalid cookie was cleared (Expires in past)
	cookieHeader := rr.Header().Get("Set-Cookie")
	if cookieHeader == "" {
		t.Errorf("Expected Set-Cookie header to clear invalid session")
	}
}

// TestRequireAuthIncorrectRole verifies StatusForbidden (403) when role is mismatched
func TestRequireAuthIncorrectRole(t *testing.T) {
	handler := middleware.RequireAuth("admin", testSecret, dummyHandler)

	// Generate a token for a broadcaster, but the handler requires "admin"
	token, err := middleware.GenerateJWT("cam-workspace", "broadcaster", testSecret)
	if err != nil {
		t.Fatalf("Failed to generate JWT: %v", err)
	}

	req := httptest.NewRequest("GET", "/admin/dashboard", nil)
	req.AddCookie(&http.Cookie{
		Name:  "auth_token",
		Value: token,
	})
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Expect status 403 Forbidden
	if rr.Code != http.StatusForbidden {
		t.Errorf("Expected status 403 Forbidden, got %d", rr.Code)
	}
}

// TestRequireAuthSuccess validates complete path authorization and claims extraction
func TestRequireAuthSuccess(t *testing.T) {
	handler := middleware.RequireAuth("admin", testSecret, dummyHandler)

	adminEmail := "admin@multicamobserver.com"
	token, err := middleware.GenerateJWT(adminEmail, "admin", testSecret)
	if err != nil {
		t.Fatalf("Failed to generate JWT: %v", err)
	}

	req := httptest.NewRequest("GET", "/admin/dashboard", nil)
	req.AddCookie(&http.Cookie{
		Name:  "auth_token",
		Value: token,
	})
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Expect status 200 OK
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %d", rr.Code)
	}

	body := rr.Body.String()
	expectedBody := "Access Granted: " + adminEmail
	if body != expectedBody {
		t.Errorf("Expected body %s, got %s", expectedBody, body)
	}
}
