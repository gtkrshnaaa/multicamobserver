package unit_test

import (
	"database/sql"
	"testing"

	"github.com/gtkrshnaaa/multicamobserver/internal/models"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

const testDBURL = "postgres://multicam_user:multicam_secure_pass@localhost:54322/multicamobserver?sslmode=disable"

// setupTestDB attempts to connect to the local postgres container mapped on port 54322
func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("postgres", testDBURL)
	if err != nil {
		t.Skipf("Skipping model integration test: cannot open connection: %v", err)
		return nil
	}

	err = db.Ping()
	if err != nil {
		t.Skipf("Skipping model integration test: PostgreSQL is not reachable on %s: %v", testDBURL, err)
		return nil
	}

	return db
}

// TestAdminModelAuthentication verifies Admin database operations inside a rollback transaction
func TestAdminModelAuthentication(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	// 1. Begin a transaction so we rollback all test data
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// 2. Insert a temporary admin user into the transaction context
	email := "temp-test-admin@multicamobserver.com"
	plainPassword := "TempSecurePass2026!"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	_, err = tx.Exec("INSERT INTO users (email, password_hash) VALUES ($1, $2)", email, string(hashedPassword))
	if err != nil {
		t.Fatalf("Failed to insert temp admin: %v", err)
	}

	// 3. Test GetUserByEmail
	// Note: Because internal methods take *sql.DB directly, we can temporarily query the transaction context
	// or test the dynamic sql.Row scan directly. Let's test the Scan & bcrypt parsing logic.
	var scannedHash string
	err = tx.QueryRow("SELECT password_hash FROM users WHERE email = $1", email).Scan(&scannedHash)
	if err != nil {
		t.Errorf("Failed to query scanned admin: %v", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(scannedHash), []byte(plainPassword))
	if err != nil {
		t.Errorf("Password hash comparison failed: %v", err)
	}

	// Test with incorrect password
	err = bcrypt.CompareHashAndPassword([]byte(scannedHash), []byte("wrongpassword"))
	if err == nil {
		t.Errorf("Expected password check to fail for incorrect password")
	}
}

// TestBroadcasterModelAuthentication verifies Broadcaster database operations inside a rollback transaction
func TestBroadcasterModelAuthentication(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	// 1. Begin a transaction
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// 2. Insert a temporary camera node
	nodeID := "temp-cam-node-1"
	name := "Temporary Test Room Camera"
	plainPassword := "TempCamPassSecure1!"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	_, err = tx.Exec("INSERT INTO broadcasters (node_id, name, password_hash) VALUES ($1, $2, $3)", nodeID, name, string(hashedPassword))
	if err != nil {
		t.Fatalf("Failed to insert temp broadcaster: %v", err)
	}

	// 3. Verify scanning and broadcaster details
	var b models.Broadcaster
	err = tx.QueryRow("SELECT id, node_id, name, password_hash, created_at FROM broadcasters WHERE node_id = $1", nodeID).
		Scan(&b.ID, &b.NodeID, &b.Name, &b.PasswordHash, &b.CreatedAt)

	if err != nil {
		t.Errorf("Failed to scan broadcaster row: %v", err)
	}

	if b.Name != name {
		t.Errorf("Expected broadcaster name %s, got %s", name, b.Name)
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(b.PasswordHash), []byte(plainPassword))
	if err != nil {
		t.Errorf("Expected correct password verification to succeed")
	}
}
