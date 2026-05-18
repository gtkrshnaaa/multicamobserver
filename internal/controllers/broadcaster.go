package controllers

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/gtkrshnaaa/multicamobserver/internal/models"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  8192,
	WriteBufferSize: 8192,
	CheckOrigin: func(r *http.Request) bool {
		return true // Permit cross-origin requests for monitoring nodes
	},
}

// ShowBroadcaster serves the physical camera broadcaster streaming client page
func (c *BaseController) ShowBroadcaster(w http.ResponseWriter, r *http.Request) {
	nodeID := r.URL.Query().Get("node_id")
	if nodeID == "" {
		http.Error(w, "Missing node_id query parameter", http.StatusBadRequest)
		return
	}

	// Fetch camera configuration to display name on page
	b, err := models.GetBroadcasterByNodeID(c.DB, nodeID)
	if err != nil {
		http.Error(w, "Unauthorized camera node", http.StatusUnauthorized)
		return
	}

	c.Render(w, r, "broadcaster.html", map[string]interface{}{
		"NodeID": b.NodeID,
		"Name":   b.Name,
	})
}

// HandleWebSocket upgrades connection and routes frames between Broadcaster and Viewers
func (c *BaseController) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	nodeID := r.URL.Query().Get("node_id")
	role := r.URL.Query().Get("role") // "broadcaster" or "viewer"

	if nodeID == "" || (role != "broadcaster" && role != "viewer") {
		http.Error(w, "Invalid query parameters", http.StatusBadRequest)
		return
	}

	// Upgrade the HTTP server connection to the WebSocket protocol
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	if role == "broadcaster" {
		c.handleBroadcasterWS(conn, nodeID)
	} else {
		c.handleViewerWS(conn, nodeID)
	}
}

// handleBroadcasterWS manages camera node frames feeding into the system
func (c *BaseController) handleBroadcasterWS(conn *websocket.Conn, nodeID string) {
	defer conn.Close()

	log.Printf("🎥 Camera Broadcaster online: %s", nodeID)

	// Fetch broadcaster name for register description
	name := "Camera Node"
	if b, err := models.GetBroadcasterByNodeID(c.DB, nodeID); err == nil {
		name = b.Name
	}

	c.StreamRegistry.Mu.Lock()
	stream, exists := c.StreamRegistry.Streams[nodeID]
	if !exists {
		stream = &models.ActiveStream{
			NodeID:      nodeID,
			Name:        name,
			Viewers:     make(map[*websocket.Conn]*models.Viewer),
			LastActive:  time.Now(),
		}
		c.StreamRegistry.Streams[nodeID] = stream
	}
	
	// Close any prior active broadcaster connection for this ID to avoid overlaps
	if stream.Broadcaster != nil {
		stream.Broadcaster.Close()
	}
	stream.Broadcaster = conn
	stream.LastActive = time.Now()
	c.StreamRegistry.Mu.Unlock()

	// Clean up broadcaster upon exit
	defer func() {
		c.StreamRegistry.Mu.Lock()
		if stream.Broadcaster == conn {
			stream.Broadcaster = nil
		}
		// If there are no viewers and broadcaster went offline, clean up registry entry
		if len(stream.Viewers) == 0 && stream.Broadcaster == nil {
			delete(c.StreamRegistry.Streams, nodeID)
		}
		c.StreamRegistry.Mu.Unlock()
		log.Printf("❌ Camera Broadcaster offline: %s", nodeID)
	}()

	// Incoming frames loop
	for {
		messageType, payload, err := conn.ReadMessage()
		if err != nil {
			break // Connection closed or errored
		}

		// Update activity timestamp
		stream.Mu.Lock()
		stream.LastActive = time.Now()
		stream.Mu.Unlock()

		// Broadcast binary frames (or standard messages) directly to all active viewers
		stream.Mu.RLock()
		for _, viewer := range stream.Viewers {
			go func(vConn *websocket.Conn) {
				// Deliver payload to viewer connection asynchronously
				_ = vConn.WriteMessage(messageType, payload)
			}(viewer.Conn)
		}
		stream.Mu.RUnlock()
	}
}

// handleViewerWS manages admin dashboard client socket connections
func (c *BaseController) handleViewerWS(conn *websocket.Conn, nodeID string) {
	defer conn.Close()

	log.Printf("👤 Viewer connected to stream: %s", nodeID)

	c.StreamRegistry.Mu.Lock()
	stream, exists := c.StreamRegistry.Streams[nodeID]
	if !exists {
		// Create placeholder stream if camera is not online yet
		stream = &models.ActiveStream{
			NodeID:     nodeID,
			Name:       nodeID,
			Viewers:    make(map[*websocket.Conn]*models.Viewer),
			LastActive: time.Now(),
		}
		c.StreamRegistry.Streams[nodeID] = stream
	}

	viewerObj := &models.Viewer{Conn: conn}
	stream.Mu.Lock()
	stream.Viewers[conn] = viewerObj
	stream.Mu.Unlock()
	c.StreamRegistry.Mu.Unlock()

	// Clean up viewer upon disconnect
	defer func() {
		c.StreamRegistry.Mu.Lock()
		stream.Mu.Lock()
		delete(stream.Viewers, conn)
		stream.Mu.Unlock()

		// Cleanup stream entry entirely if completely empty
		if len(stream.Viewers) == 0 && stream.Broadcaster == nil {
			delete(c.StreamRegistry.Streams, nodeID)
		}
		c.StreamRegistry.Mu.Unlock()
		log.Printf("👤 Viewer disconnected from stream: %s", nodeID)
	}()

	// Keep viewer connection alive by continuously reading dummy pings or waiting for close
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
