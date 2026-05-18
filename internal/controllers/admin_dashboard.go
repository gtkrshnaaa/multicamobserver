package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gtkrshnaaa/multicamobserver/internal/models"
)

// CameraNodeStatus represents active state mapping for dashboard rendering
type CameraNodeStatus struct {
	ID          int    `json:"id"`
	NodeID      string `json:"node_id"`
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

	// Extract query parameter messages for feedback popups
	successMsg := r.URL.Query().Get("success")
	errorMsg := r.URL.Query().Get("error")

	c.Render(w, r, "dashboard.html", map[string]interface{}{
		"Cameras": cameraStatuses,
		"Success": successMsg,
		"Error":   errorMsg,
	})
}

// CreateBroadcaster registers a new broadcaster node on request
func (c *BaseController) CreateBroadcaster(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.FormValue("name")
	password := r.FormValue("password")

	if name == "" || password == "" {
		http.Redirect(w, r, "/admin/dashboard?error=Broadcaster+name+and+password+are+required", http.StatusSeeOther)
		return
	}

	_, err := models.CreateBroadcaster(c.DB, name, password)
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
	name := r.FormValue("name")
	password := r.FormValue("password")

	var id int
	_, err := fmt.Sscanf(idStr, "%d", &id)
	if err != nil || name == "" {
		http.Redirect(w, r, "/admin/dashboard?error=Invalid+ID+or+missing+name", http.StatusSeeOther)
		return
	}

	err = models.UpdateBroadcaster(c.DB, id, name, password)
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
