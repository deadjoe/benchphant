package database

import (
	"fmt"
	"sync"
	"time"

	"github.com/deadjoe/benchphant/internal/models"
)

// MemoryStorage implements Storage interface using in-memory storage
type MemoryStorage struct {
	connections map[int64]*models.DBConnection
	nextID      int64
	mu          sync.RWMutex
}

// NewMemoryStorage creates a new memory storage
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		connections: make(map[int64]*models.DBConnection),
		nextID:      1,
	}
}

// AddConnection adds a new database connection
func (s *MemoryStorage) AddConnection(conn *models.DBConnection) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Set ID and store connection
	conn.ID = s.nextID
	s.nextID++
	s.connections[conn.ID] = conn

	return nil
}

// SaveConnection saves a database connection
func (s *MemoryStorage) SaveConnection(conn *models.DBConnection) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if conn.ID == 0 {
		// This is a new connection
		conn.ID = s.nextID
		s.nextID++
	}

	s.connections[conn.ID] = conn
	return nil
}

// UpdateConnection updates an existing database connection
func (s *MemoryStorage) UpdateConnection(conn *models.DBConnection) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.connections[conn.ID]; !exists {
		return fmt.Errorf("connection not found: %d", conn.ID)
	}

	s.connections[conn.ID] = conn
	return nil
}

// UpdateLastUsed updates the last used timestamp for a connection
func (s *MemoryStorage) UpdateLastUsed(id int64, lastUsed time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	conn, exists := s.connections[id]
	if !exists {
		return fmt.Errorf("connection not found: %d", id)
	}

	conn.LastUsedAt = lastUsed
	return nil
}

// DeleteConnection deletes a database connection
func (s *MemoryStorage) DeleteConnection(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.connections[id]; !exists {
		return fmt.Errorf("connection not found: %d", id)
	}

	delete(s.connections, id)
	return nil
}

// GetConnection retrieves a database connection by ID
func (s *MemoryStorage) GetConnection(id int64) (*models.DBConnection, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	conn, exists := s.connections[id]
	if !exists {
		return nil, fmt.Errorf("connection not found: %d", id)
	}

	return conn, nil
}

// ListConnections returns all database connections
func (s *MemoryStorage) ListConnections() ([]*models.DBConnection, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	connections := make([]*models.DBConnection, 0, len(s.connections))
	for _, conn := range s.connections {
		connections = append(connections, conn)
	}

	return connections, nil
}

// Close implements Storage.Close
func (s *MemoryStorage) Close() error {
	return nil
}
