package feature_test

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gtkrshnaaa/multicamobserver/internal/controllers"
	"github.com/gtkrshnaaa/multicamobserver/internal/models"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

const testDBURL = "postgres://multicam_user:multicam_secure_pass@localhost:54322/multicamobserver?sslmode=disable"

// setupTestControllerWithDB instantiates BaseController connected to the live PostgreSQL test database
func setupTestControllerWithDB(t *testing.T) (*controllers.BaseController, *sql.DB) {
	tmpl, err := template.ParseGlob("../../ui/html/*.html")
	if err != nil {
		t.Fatalf("Failed to parse templates: %v", err)
	}

	db, err := sql.Open("postgres", testDBURL)
	if err != nil {
		t.Skipf("Skipping dashboard feature test: cannot connect to db: %v", err)
		return nil, nil
	}

	err = db.Ping()
	if err != nil {
		t.Skipf("Skipping dashboard feature test: DB not reachable on %s: %v", testDBURL, err)
		return nil, nil
	}

	registry := models.NewStreamRegistry()
	return controllers.NewBaseController(db, testSecret, tmpl, registry), db
}

// TestShowDashboard validates administrator SSR dashboard rendering containing database elements
func TestShowDashboard(t *testing.T) {
	ctrl, db := setupTestControllerWithDB(t)
	if ctrl == nil {
		return
	}
	defer db.Close()

	req := httptest.NewRequest("GET", "/admin/dashboard", nil)
	rr := httptest.NewRecorder()

	ctrl.ShowDashboard(rr, req)

	// Validate status 200 OK
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %d", rr.Code)
	}

	contentType := rr.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Errorf("Expected HTML response content, got %s", contentType)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "cam-grid") || !strings.Contains(body, "Central Monitoring Console") {
		t.Errorf("Expected body to contain surveillance grid elements")
	}
}

// TestShowHealth verifies JSON structured format and database connection check reporting
func TestShowHealth(t *testing.T) {
	ctrl, db := setupTestControllerWithDB(t)
	if ctrl == nil {
		return
	}
	defer db.Close()

	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()

	ctrl.ShowHealth(rr, req)

	// Validate status 200 OK
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %d", rr.Code)
	}

	contentType := rr.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("Expected JSON response, got %s", contentType)
	}

	var payload map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &payload)
	if err != nil {
		t.Fatalf("Failed to parse JSON body: %v", err)
	}

	if payload["status"] != "healthy" {
		t.Errorf("Expected status to be healthy, got: %v", payload["status"])
	}
	if payload["database"] != "online" {
		t.Errorf("Expected database to be online, got: %v", payload["database"])
	}
}

// TestShowBroadcasterMissingNodeID asserts Bad Request (400) when accessing camera node without ID query
func TestShowBroadcasterMissingNodeID(t *testing.T) {
	ctrl, db := setupTestControllerWithDB(t)
	if ctrl == nil {
		return
	}
	defer db.Close()

	req := httptest.NewRequest("GET", "/broadcaster/camera", nil)
	rr := httptest.NewRecorder()

	ctrl.ShowBroadcaster(rr, req)

	// Expect 400 Bad Request
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 Bad Request, got %d", rr.Code)
	}
}

// TestShowBroadcasterValidNodeID checks correct broadcaster camera node interface render
func TestShowBroadcasterValidNodeID(t *testing.T) {
	ctrl, db := setupTestControllerWithDB(t)
	if ctrl == nil {
		return
	}
	defer db.Close()

	nodeID := "test-feature-cam-node"
	name := "Feature Test Camera Office"
	plainPassword := "SecretCameraPass1!"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to generate password hash: %v", err)
	}

	_, err = db.Exec("INSERT INTO broadcasters (node_id, name, password_hash) VALUES ($1, $2, $3)", nodeID, name, string(hashedPassword))
	if err != nil {
		t.Fatalf("Failed to insert temp broadcaster: %v", err)
	}
	defer func() {
		_, _ = db.Exec("DELETE FROM broadcasters WHERE node_id = $1", nodeID)
	}()

	req := httptest.NewRequest("GET", "/broadcaster/camera?node_id="+nodeID, nil)
	rr := httptest.NewRecorder()

	ctrl.ShowBroadcaster(rr, req)

	// Validate status 200 OK
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %d", rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, name) || !strings.Contains(body, "broadcaster-layout") {
		t.Errorf("Expected rendered page to contain camera name and streaming controls")
	}
}
