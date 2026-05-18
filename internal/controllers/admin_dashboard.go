package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/gtkrshnaaa/multicamobserver/internal/models"
)

// CameraNodeStatus represents active state mapping for dashboard rendering
type CameraNodeStatus struct {
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

	c.Render(w, r, "dashboard.html", map[string]interface{}{
		"Cameras": cameraStatuses,
	})
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
