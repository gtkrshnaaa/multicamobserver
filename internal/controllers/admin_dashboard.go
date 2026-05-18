package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gtkrshnaaa/multicamobserver/internal/middleware"
	"github.com/gtkrshnaaa/multicamobserver/internal/models"
)

// CameraNodeStatus represents active state mapping for dashboard rendering
type CameraNodeStatus struct {
	ID          int    `json:"id"`
	NodeID      string `json:"node_id"`
	Username    string `json:"username"`
	Name        string `json:"name"`
	IsOnline    bool   `json:"is_online"`
	ViewerCount int    `json:"viewer_count"`
}

// ShowDashboard serves the main multicam grid view interface for the administrator
func (c *BaseController) ShowDashboard(w http.ResponseWriter, r *http.Request) {
	// Retrieve all registered camera nodes from database
	broadcasters, err := models.GetAllBroadcasters(c.DB)
	if err != nil {
		http.Error(w, "Failed to load camera list: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var cameraStatuses []CameraNodeStatus

	c.StreamRegistry.Mu.RLock()
	for _, b := range broadcasters {
		status := CameraNodeStatus{
			ID:          b.ID,
			NodeID:      b.NodeID,
			Username:    b.Username,
			Name:        b.Name,
			IsOnline:    false,
			ViewerCount: 0,
		}

		// Match against in-memory registry to verify active connection
		if active, online := c.StreamRegistry.Streams[b.NodeID]; online {
			active.Mu.RLock()
			status.IsOnline = (active.Broadcaster != nil)
			status.ViewerCount = len(active.Viewers)
			active.Mu.RUnlock()
		}

		cameraStatuses = append(cameraStatuses, status)
	}
	c.StreamRegistry.Mu.RUnlock()

	// Retrieve current logged-in admin user to pre-fill admin edit form
	claims := middleware.GetUserClaims(r)
	var adminUser *models.User
	if claims != nil {
		adminUser, _ = models.GetUserByEmail(c.DB, claims.Subject)
	}

	// Extract query parameter messages for feedback popups
	successMsg := r.URL.Query().Get("success")
	errorMsg := r.URL.Query().Get("error")

	c.Render(w, r, "dashboard.html", map[string]interface{}{
		"Cameras":   cameraStatuses,
		"Success":   successMsg,
		"Error":     errorMsg,
		"AdminUser": adminUser,
	})
}

// CreateBroadcaster registers a new broadcaster node on request
func (c *BaseController) CreateBroadcaster(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.FormValue("username")
	name := r.FormValue("name")
	password := r.FormValue("password")

	if username == "" || name == "" || password == "" {
		http.Redirect(w, r, "/admin/dashboard?error=Broadcaster+username,+name,+and+password+are+required", http.StatusSeeOther)
		return
	}

	_, err := models.CreateBroadcaster(c.DB, username, name, password)
	if err != nil {
		http.Redirect(w, r, "/admin/dashboard?error=Failed+to+create:+"+err.Error(), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/dashboard?success=Broadcaster+created+successfully", http.StatusSeeOther)
}

// UpdateBroadcaster modifies credentials for an existing broadcaster node
func (c *BaseController) UpdateBroadcaster(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.FormValue("id")
	username := r.FormValue("username")
	name := r.FormValue("name")
	password := r.FormValue("password")

	var id int
	_, err := fmt.Sscanf(idStr, "%d", &id)
	if err != nil || username == "" || name == "" {
		http.Redirect(w, r, "/admin/dashboard?error=Invalid+ID,+missing+username,+or+missing+name", http.StatusSeeOther)
		return
	}

	err = models.UpdateBroadcaster(c.DB, id, username, name, password)
	if err != nil {
		http.Redirect(w, r, "/admin/dashboard?error=Failed+to+update:+"+err.Error(), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/dashboard?success=Broadcaster+updated+successfully", http.StatusSeeOther)
}

// DeleteBroadcaster removes a broadcaster node from database
func (c *BaseController) DeleteBroadcaster(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.FormValue("id")
	var id int
	_, err := fmt.Sscanf(idStr, "%d", &id)
	if err != nil {
		http.Redirect(w, r, "/admin/dashboard?error=Invalid+ID", http.StatusSeeOther)
		return
	}

	err = models.DeleteBroadcaster(c.DB, id)
	if err != nil {
		http.Redirect(w, r, "/admin/dashboard?error=Failed+to+delete:+"+err.Error(), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/dashboard?success=Broadcaster+deleted+successfully", http.StatusSeeOther)
}

// PurgeBroadcasters permanently deletes all broadcaster camera nodes from the system
func (c *BaseController) PurgeBroadcasters(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := models.PurgeAllBroadcasters(c.DB)
	if err != nil {
		http.Redirect(w, r, "/admin/dashboard?error=Failed+to+purge+camera+nodes:+"+err.Error(), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/dashboard?success=All+broadcaster+camera+nodes+successfully+purged", http.StatusSeeOther)
}

// UpdateAdminCredentials allows the logged-in administrator to modify their own email/username and password
func (c *BaseController) UpdateAdminCredentials(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.FormValue("id")
	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")

	var id int
	_, err := fmt.Sscanf(idStr, "%d", &id)
	if err != nil || username == "" || email == "" {
		http.Redirect(w, r, "/admin/dashboard?error=Invalid+ID,+missing+username,+or+missing+email", http.StatusSeeOther)
		return
	}

	err = models.UpdateUser(c.DB, id, username, email, password)
	if err != nil {
		http.Redirect(w, r, "/admin/dashboard?error=Failed+to+update+admin+credentials:+"+err.Error(), http.StatusSeeOther)
		return
	}

	// Regenerate JWT token to maintain valid active session for the updated admin email
	tokenString, err := middleware.GenerateJWT(email, "admin", c.JWTSecret)
	if err == nil {
		http.SetCookie(w, &http.Cookie{
			Name:     "auth_token",
			Value:    tokenString,
			Path:     "/",
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
			Secure:   false, // Set to true in production over HTTPS
			SameSite: http.SameSiteLaxMode,
		})
	}

	http.Redirect(w, r, "/admin/dashboard?success=Administrator+credentials+updated+successfully", http.StatusSeeOther)
}

// ShowHealth returns a JSON healthcheck status representing service availability
func (c *BaseController) ShowHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":   "healthy",
		"database": "online",
	}

	// Verify database connection is alive
	if err := c.DB.Ping(); err != nil {
		health["status"] = "unhealthy"
		health["database"] = "offline: " + err.Error()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(health)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(health)
}
