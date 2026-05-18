package feature_test

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gtkrshnaaa/multicamobserver/internal/controllers"
	"github.com/gtkrshnaaa/multicamobserver/internal/models"
)

// setupTestController instantiates a BaseController with parsed real templates for controller tests
func setupTestController(t *testing.T) *controllers.BaseController {
	// Parse real templates from the project relative to the tests/Feature folder
	tmpl, err := template.ParseGlob("../../ui/html/*.html")
	if err != nil {
		t.Fatalf("Failed to parse real templates for feature tests: %v", err)
	}

	registry := models.NewStreamRegistry()
	return controllers.NewBaseController(nil, testSecret, tmpl, registry)
}

// TestShowLogin validates that GET /login serves the SSR login page perfectly with HTTP 200 OK
func TestShowLogin(t *testing.T) {
	ctrl := setupTestController(t)

	req := httptest.NewRequest("GET", "/login", nil)
	rr := httptest.NewRecorder()

	ctrl.ShowLogin(rr, req)

	// Validate status is 200 OK
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %d", rr.Code)
	}

	// Validate content type is HTML
	contentType := rr.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Errorf("Expected text/html, got %s", contentType)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "login-card") && !strings.Contains(body, "MulticamObserver") {
		t.Errorf("Expected body to contain login page markers")
	}
}

// TestHandleLogout checks HTTP cookie clearing and proper redirection
func TestHandleLogout(t *testing.T) {
	ctrl := setupTestController(t)

	req := httptest.NewRequest("GET", "/logout", nil)
	rr := httptest.NewRecorder()

	ctrl.HandleLogout(rr, req)

	// Expect redirect status 303
	if rr.Code != http.StatusSeeOther {
		t.Errorf("Expected status 303 See Other, got %d", rr.Code)
	}

	location := rr.Header().Get("Location")
	if location != "/login" {
		t.Errorf("Expected redirect to /login, got %s", location)
	}

	// Check that the auth_token cookie was expired in the past
	setCookie := rr.Header().Get("Set-Cookie")
	if !strings.Contains(setCookie, "auth_token=") || (!strings.Contains(setCookie, "Max-Age=0") && !strings.Contains(setCookie, "Max-Age=-1")) {
		t.Errorf("Expected set-cookie to expire auth_token, got %s", setCookie)
	}
}

// TestHandleLoginValidation verifies invalid role form validation renders login with error payload
func TestHandleLoginValidation(t *testing.T) {
	ctrl := setupTestController(t)

	// Submit an invalid role to trigger dynamic template feedback
	formData := url.Values{}
	formData.Set("role", "invalid-role")
	formData.Set("username", "somebody@multicamobserver.com")
	formData.Set("password", "anypass")

	req := httptest.NewRequest("POST", "/login", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	ctrl.HandleLogin(rr, req)

	// Since it's validation failure, it should render login.html (HTTP 200) instead of redirecting
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200 for error rendering, got %d", rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "Invalid role selected") {
		t.Errorf("Expected error message 'Invalid role selected' to be rendered, got: %s", body)
	}
}
