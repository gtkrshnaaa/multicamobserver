package models

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// User represents an Administrator who can view camera feeds
type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

// Broadcaster represents a physical camera node credentials
type Broadcaster struct {
	ID           int       `json:"id"`
	NodeID       string    `json:"node_id"`
	Username     string    `json:"username"`
	Name         string    `json:"name"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

// Viewer represents a connected administrator browser watching a stream
type Viewer struct {
	Conn *websocket.Conn
}

// ActiveStream manages the real-time websocket connection for a single camera node
type ActiveStream struct {
	NodeID       string
	Name         string
	Broadcaster  *websocket.Conn
	Viewers      map[*websocket.Conn]*Viewer
	Mu           sync.RWMutex
	LastActive   time.Time
}

// StreamRegistry keeps track of all active online streams in memory
type StreamRegistry struct {
	Streams map[string]*ActiveStream
	Mu      sync.RWMutex
}

// NewStreamRegistry creates a new in-memory stream registry
func NewStreamRegistry() *StreamRegistry {
	return &StreamRegistry{
		Streams: make(map[string]*ActiveStream),
	}
}
