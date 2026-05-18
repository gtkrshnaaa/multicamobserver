package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gtkrshnaaa/multicamobserver/internal/controllers"
	"github.com/gtkrshnaaa/multicamobserver/internal/middleware"
	"github.com/gtkrshnaaa/multicamobserver/internal/models"
	
	_ "github.com/lib/pq"
)

func main() {
	log.Println("──────────────────────────────────────────────────")
	log.Println(" 🚀 Starting MulticamObserver Monolithic Signaling Hub")
	log.Println("──────────────────────────────────────────────────")

	// 1. Parse configuration from environment variables
	port := os.Getenv("PORT")
	if port == "" {
		port = "51177" // Standard required port
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Local development default URL if DB env is not provided
		dbURL = "postgres://novel_user:novel_secure_pass@localhost:5432/ceritadariayah?sslmode=disable"
	}

	sessionSecret := os.Getenv("SESSION_SECRET")
	if sessionSecret == "" {
		sessionSecret = "multicamobserver-super-secret-session-key-2026"
	}

	// 2. Establish PostgreSQL database connection with retry backoff
	var db *sql.DB
	var err error
	maxRetries := 15

	log.Printf("Connecting to database: %s", redactDBURL(dbURL))
	for i := 1; i <= maxRetries; i++ {
		db, err = sql.Open("postgres", dbURL)
		if err == nil {
			err = db.Ping()
		}

		if err == nil {
			log.Println("✅ Successfully connected to PostgreSQL database.")
			break
		}

		log.Printf("  [%d/%d] Waiting for database connection... (%v)", i, maxRetries, err)
		if db != nil {
			db.Close()
		}
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatalf("❌ Fatal Error: Could not connect to database after maximum retries: %v", err)
	}
	defer db.Close()

	// 3. Automatically bootstrap and run database schema
	bootstrapDatabase(db)

	// 4. Parse Server-Side HTML templates
	tmplPath := filepath.Join("ui", "html", "*.html")
	tmpl, err := template.ParseGlob(tmplPath)
	if err != nil {
		log.Fatalf("❌ Fatal Error: Failed to parse templates: %v", err)
	}
	log.Println("✅ Server-Side Rendered (SSR) HTML templates loaded successfully.")

	// 5. Initialize In-Memory Stream Registry
	streamRegistry := models.NewStreamRegistry()
	log.Println("✅ High-Concurrency In-Memory Stream Registry initialized.")

	// 6. Initialize Base Controller
	ctrl := controllers.NewBaseController(db, []byte(sessionSecret), tmpl, streamRegistry)

	// 7. Configure Route Multiplexing & Middleware Authorization
	mux := http.NewServeMux()

	// Static Assets Server
	staticFS := http.FileServer(http.Dir("ui/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", staticFS))

	// Unified Authentication Routes
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			ctrl.ShowLogin(w, r)
		} else if r.Method == http.MethodPost {
			ctrl.HandleLogin(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/logout", ctrl.HandleLogout)

	// Protected Admin Dashboard Route
	mux.HandleFunc("/admin/dashboard", middleware.RequireAuth("admin", ctrl.JWTSecret, ctrl.ShowDashboard))
	mux.HandleFunc("/admin/broadcaster/create", middleware.RequireAuth("admin", ctrl.JWTSecret, ctrl.CreateBroadcaster))
	mux.HandleFunc("/admin/broadcaster/update", middleware.RequireAuth("admin", ctrl.JWTSecret, ctrl.UpdateBroadcaster))
	mux.HandleFunc("/admin/broadcaster/delete", middleware.RequireAuth("admin", ctrl.JWTSecret, ctrl.DeleteBroadcaster))
	mux.HandleFunc("/admin/broadcaster/purge", middleware.RequireAuth("admin", ctrl.JWTSecret, ctrl.PurgeBroadcasters))
	mux.HandleFunc("/admin/credentials/update", middleware.RequireAuth("admin", ctrl.JWTSecret, ctrl.UpdateAdminCredentials))

	// Protected Camera Node Route
	mux.HandleFunc("/broadcaster/camera", middleware.RequireAuth("broadcaster", ctrl.JWTSecret, ctrl.ShowBroadcaster))

	// WebSocket Media Hub Endpoint (Multiplexes viewer vs broadcaster inside)
	mux.HandleFunc("/ws", ctrl.HandleWebSocket)

	// API Health check endpoint
	mux.HandleFunc("/health", ctrl.ShowHealth)

	// Root Redirect Handler
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})

	// 8. Start HTTP / WSS Server
	serverAddr := ":" + port
	log.Println("══════════════════════════════════════════════")
	log.Printf(" 🌐 APPLICATION RUNNING SUCCESSFULLY")
	log.Printf(" 👉 Access Portal: http://localhost:%s", port)
	log.Println("══════════════════════════════════════════════")

	if err := http.ListenAndServe(serverAddr, mux); err != nil {
		log.Fatalf("❌ Fatal Error: Server stopped: %v", err)
	}
}

// bootstrapDatabase runs database schema and seeder commands to initialize tables
func bootstrapDatabase(db *sql.DB) {
	// 1. Run Schema Migrations (DDL)
	schemaFile := filepath.Join("database", "migrations", "0001_create_tables.sql")
	schemaBytes, err := os.ReadFile(schemaFile)
	if err != nil {
		log.Printf("⚠️  Warning: database migration file not found, skipping DDL migration: %v", err)
		return
	}

	_, err = db.Exec(string(schemaBytes))
	if err != nil {
		log.Fatalf("❌ Fatal Error: Failed to bootstrap database DDL schema: %v", err)
	}
	log.Println("✅ Successfully initialized database DDL schema tables.")

	// 2. Run Seeder Initializer (DML)
	seederFile := filepath.Join("database", "seeders", "0001_seed_credentials.sql")
	seederBytes, err := os.ReadFile(seederFile)
	if err != nil {
		log.Printf("⚠️  Warning: database seeder file not found, skipping DML seed: %v", err)
		return
	}

	_, err = db.Exec(string(seederBytes))
	if err != nil {
		log.Fatalf("❌ Fatal Error: Failed to seed database initial records: %v", err)
	}
	log.Println("✅ Successfully seeded database initial credentials.")
}

// redactDBURL removes passwords from DB connection logs for security
func redactDBURL(url string) string {
	// A simple helper to prevent logging clear-text passwords
	return "postgres://***:***@host/dbname"
}
