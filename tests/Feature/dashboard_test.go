package feature_test

import (
	"database/sql"
	"encoding/json"
	"fmt"
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

	_, err = db.Exec("INSERT INTO broadcasters (node_id, username, name, password_hash) VALUES ($1, $2, $3, $4)", nodeID, "feature_test_cam", name, string(hashedPassword))
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

// TestCreateBroadcaster verifies successful creation of a new broadcaster node
func TestCreateBroadcaster(t *testing.T) {
	ctrl, db := setupTestControllerWithDB(t)
	if ctrl == nil {
		return
	}
	defer db.Close()

	formData := "username=yard_camera&name=Yard+Camera&password=YardSecurePassword2026!"
	req := httptest.NewRequest("POST", "/admin/broadcaster/create", strings.NewReader(formData))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	ctrl.CreateBroadcaster(rr, req)

	// Should redirect on success
	if rr.Code != http.StatusSeeOther {
		t.Errorf("Expected redirect status 303, got %d", rr.Code)
	}

	loc := rr.Header().Get("Location")
	if !strings.Contains(loc, "success=") {
		t.Errorf("Expected success query param in redirect URL, got %s", loc)
	}

	// Verify insertion in database
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM broadcasters WHERE username = 'yard_camera'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query database count: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected exactly 1 broadcaster with name 'Yard Camera' to be created, got %d", count)
	}

	// Clean up
	_, _ = db.Exec("DELETE FROM broadcasters WHERE username = 'yard_camera'")
}

// TestUpdateBroadcaster checks dynamic editing of name and password parameters
func TestUpdateBroadcaster(t *testing.T) {
	ctrl, db := setupTestControllerWithDB(t)
	if ctrl == nil {
		return
	}
	defer db.Close()

	// Insert test camera
	var id int
	err := db.QueryRow("INSERT INTO broadcasters (username, name, password_hash) VALUES ('old_lobby_cam', 'Old Lobby Cam', 'dummyhash') RETURNING id").Scan(&id)
	if err != nil {
		t.Fatalf("Failed to insert mock camera: %v", err)
	}
	defer func() {
		_, _ = db.Exec("DELETE FROM broadcasters WHERE id = $1", id)
	}()

	formData := fmt.Sprintf("id=%d&username=new_lobby_cam&name=New+Lobby+Cam&password=LobbyPassUpdated1!", id)
	req := httptest.NewRequest("POST", "/admin/broadcaster/update", strings.NewReader(formData))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	ctrl.UpdateBroadcaster(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("Expected status 303, got %d", rr.Code)
	}

	// Verify update in DB
	var name string
	err = db.QueryRow("SELECT name FROM broadcasters WHERE id = $1", id).Scan(&name)
	if err != nil {
		t.Fatalf("Failed to query broadcaster name: %v", err)
	}
	if name != "New Lobby Cam" {
		t.Errorf("Expected updated name 'New Lobby Cam', got: %s", name)
	}
}

// TestDeleteBroadcaster confirms deletion of camera nodes from database
func TestDeleteBroadcaster(t *testing.T) {
	ctrl, db := setupTestControllerWithDB(t)
	if ctrl == nil {
		return
	}
	defer db.Close()

	// Insert test camera to delete
	var id int
	err := db.QueryRow("INSERT INTO broadcasters (username, name, password_hash) VALUES ('to_delete_cam', 'To Delete Cam', 'dummy') RETURNING id").Scan(&id)
	if err != nil {
		t.Fatalf("Failed to insert mock camera: %v", err)
	}

	formData := fmt.Sprintf("id=%d", id)
	req := httptest.NewRequest("POST", "/admin/broadcaster/delete", strings.NewReader(formData))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	ctrl.DeleteBroadcaster(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("Expected status 303, got %d", rr.Code)
	}

	// Verify deleted from DB
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM broadcasters WHERE id = $1", id).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query count: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected broadcaster to be deleted, but still found in DB")
	}
}

// TestPurgeBroadcasters verifies deleting all camera nodes at once
func TestPurgeBroadcasters(t *testing.T) {
	ctrl, db := setupTestControllerWithDB(t)
	if ctrl == nil {
		return
	}
	defer db.Close()

	// Insert two camera nodes
	_, _ = db.Exec("INSERT INTO broadcasters (username, name, password_hash) VALUES ('mock_cam_a', 'Mock Cam A', 'dummy'), ('mock_cam_b', 'Mock Cam B', 'dummy')")

	req := httptest.NewRequest("POST", "/admin/broadcaster/purge", nil)
	rr := httptest.NewRecorder()

	ctrl.PurgeBroadcasters(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("Expected status 303, got %d", rr.Code)
	}

	// Verify database is completely empty of broadcaster records
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM broadcasters").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query total count: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected broadcasters table to be purged, but found %d items", count)
	}
}

// TestUpdateAdminCredentials verifies administrator can update their email/username and password
func TestUpdateAdminCredentials(t *testing.T) {
	ctrl, db := setupTestControllerWithDB(t)
	if ctrl == nil {
		return
	}
	defer db.Close()

	// Insert mock administrator user
	var id int
	err := db.QueryRow("INSERT INTO users (username, email, password_hash) VALUES ('mock-admin', 'mock-admin@test.com', 'dummy') RETURNING id").Scan(&id)
	if err != nil {
		t.Fatalf("Failed to insert mock admin: %v", err)
	}
	defer func() {
		_, _ = db.Exec("DELETE FROM users WHERE id = $1", id)
	}()

	formData := fmt.Sprintf("id=%d&username=updated-mock-admin&email=updated-mock-admin@test.com&password=UpdatedMockPass2026!", id)
	req := httptest.NewRequest("POST", "/admin/credentials/update", strings.NewReader(formData))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	ctrl.UpdateAdminCredentials(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("Expected status 303, got %d", rr.Code)
	}

	// Verify credentials updated in DB
	var email string
	err = db.QueryRow("SELECT email FROM users WHERE id = $1", id).Scan(&email)
	if err != nil {
		t.Fatalf("Failed to query email: %v", err)
	}
	if email != "updated-mock-admin@test.com" {
		t.Errorf("Expected updated email 'updated-mock-admin@test.com', got: %s", email)
	}
}
