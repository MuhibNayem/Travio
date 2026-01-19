package realtime

import (
	"sync"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
)

// ClientChan represents a single client connection that receives SSE data strings
type ClientChan chan string

// Manager handles SSE connections mapped by TripID
type Manager struct {
	clients map[string][]ClientChan // tripID -> []ClientChan
	lock    sync.RWMutex
}

// NewManager creates a new realtime manager
func NewManager() *Manager {
	return &Manager{
		clients: make(map[string][]ClientChan),
	}
}

// Subscribe adds a client to a trip's updates and returns their channel
func (m *Manager) Subscribe(tripID string) ClientChan {
	m.lock.Lock()
	defer m.lock.Unlock()

	// Buffer to prevent blocking the broadcaster if one client is slow
	ch := make(ClientChan, 10)
	m.clients[tripID] = append(m.clients[tripID], ch)

	logger.Info("Client subscribed to realtime updates", "trip_id", tripID)
	return ch
}

// Unsubscribe removes a client and closes their channel
func (m *Manager) Unsubscribe(tripID string, ch ClientChan) {
	m.lock.Lock()
	defer m.lock.Unlock()

	clients := m.clients[tripID]
	for i, client := range clients {
		if client == ch {
			// Remove client from slice (standard filtering)
			m.clients[tripID] = append(clients[:i], clients[i+1:]...)
			close(ch)
			break
		}
	}

	// Clean up map entry if empty
	if len(m.clients[tripID]) == 0 {
		delete(m.clients, tripID)
	}
}

// Broadcast sends a message to all clients watching a trip
func (m *Manager) Broadcast(tripID string, payload string) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	clients, ok := m.clients[tripID]
	if !ok {
		return
	}

	for _, ch := range clients {
		select {
		case ch <- payload:
		default:
			// Drop message if client is slow/blocked to avoid blocking everyone
			// relying on client-side reconnection/refresh for recovery
		}
	}
}
