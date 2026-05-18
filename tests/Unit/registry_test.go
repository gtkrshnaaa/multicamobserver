package unit_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/gtkrshnaaa/multicamobserver/internal/models"
)

// TestRegistryBasicOperations validates standard CRUD-like operations of the registry
func TestRegistryBasicOperations(t *testing.T) {
	registry := models.NewStreamRegistry()

	nodeID := "cam-test-1"
	broadcasterName := "Test Camera 1"

	// 1. Initially, node should be inactive
	registry.Mu.RLock()
	_, found := registry.Streams[nodeID]
	registry.Mu.RUnlock()
	if found {
		t.Errorf("Expected node %s to be inactive initially", nodeID)
	}

	// 2. Register active stream
	registry.Mu.Lock()
	registry.Streams[nodeID] = &models.ActiveStream{
		NodeID:      nodeID,
		Name:        broadcasterName,
		Viewers:     make(map[*websocket.Conn]*models.Viewer),
		LastActive:  time.Now(),
	}
	registry.Mu.Unlock()

	// 3. Node should now be active
	registry.Mu.RLock()
	stream, found := registry.Streams[nodeID]
	registry.Mu.RUnlock()
	if !found {
		t.Errorf("Expected node %s to be active after registration", nodeID)
	}

	// 4. Check properties
	if stream.Name != broadcasterName {
		t.Errorf("Expected broadcaster name %s, got %s", broadcasterName, stream.Name)
	}

	// 5. Deregister/Remove stream
	registry.Mu.Lock()
	delete(registry.Streams, nodeID)
	registry.Mu.Unlock()

	// 6. Node should be inactive again
	registry.Mu.RLock()
	_, found = registry.Streams[nodeID]
	registry.Mu.RUnlock()
	if found {
		t.Errorf("Expected node %s to be inactive after deregistration", nodeID)
	}
}

// TestRegistryConcurrencySafety performs extreme parallel reading/writing of streams to verify sync.RWMutex thread safety
func TestRegistryConcurrencySafety(t *testing.T) {
	registry := models.NewStreamRegistry()
	var wg sync.WaitGroup

	numGoroutines := 100
	wg.Add(numGoroutines)

	// Spin up 100 goroutines doing concurrent reads/writes
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			nodeID := fmt.Sprintf("cam-concur-%d", id)

			// Concurrently write
			registry.Mu.Lock()
			registry.Streams[nodeID] = &models.ActiveStream{
				NodeID:      nodeID,
				Name:        fmt.Sprintf("Camera %d", id),
				LastActive:  time.Now(),
			}
			registry.Mu.Unlock()

			// Concurrently read
			registry.Mu.RLock()
			_ = registry.Streams[nodeID]
			registry.Mu.RUnlock()

			// Concurrently delete
			if id%2 == 0 {
				registry.Mu.Lock()
				delete(registry.Streams, nodeID)
				registry.Mu.Unlock()
			}
		}(i)
	}

	wg.Wait()
}
