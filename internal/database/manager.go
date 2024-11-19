package database

import (
	"fmt"
	"sync"
	"time"

	"github.com/deadjoe/benchphant/internal/crypto"
	"github.com/deadjoe/benchphant/internal/models"
	"go.uber.org/zap"
)

// Manager handles database connections and their lifecycle
type Manager struct {
	storage   Storage
	pools     map[int64]*ConnectionPool
	encryptor *crypto.Encryptor
	logger    *zap.Logger
	mu        sync.RWMutex
}

// NewManager creates a new database manager
func NewManager(storage Storage, encryptionKey []byte, logger *zap.Logger) (*Manager, error) {
	encryptor, err := crypto.NewEncryptor(encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create encryptor: %w", err)
	}

	return &Manager{
		storage:   storage,
		pools:     make(map[int64]*ConnectionPool),
		encryptor: encryptor,
		logger:    logger,
	}, nil
}

// AddConnection adds a new database connection
func (m *Manager) AddConnection(conn *models.DBConnection) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Set timestamps
	now := time.Now()
	conn.CreatedAt = now
	conn.UpdatedAt = now
	conn.LastUsedAt = now

	// Encrypt password before storage
	if conn.Password != "" {
		encrypted, err := m.encryptor.Encrypt(conn.Password)
		if err != nil {
			return fmt.Errorf("failed to encrypt password: %w", err)
		}
		conn.SetEncryptedPassword(encrypted)
	}

	// Save to storage
	if err := m.storage.SaveConnection(conn); err != nil {
		return fmt.Errorf("failed to save connection: %w", err)
	}

	m.logger.Info("Added new database connection",
		zap.Int64("id", conn.ID),
		zap.String("name", conn.Name),
		zap.Stringer("type", conn.Type))

	return nil
}

// UpdateConnection updates an existing database connection
func (m *Manager) UpdateConnection(conn *models.DBConnection) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get existing connection to preserve encrypted password if not changed
	existing, err := m.storage.GetConnection(conn.ID)
	if err != nil {
		return fmt.Errorf("failed to get existing connection: %w", err)
	}

	// Update timestamp
	conn.UpdatedAt = time.Now()

	// If password is empty, keep the existing encrypted password
	if conn.Password == "" {
		conn.SetEncryptedPassword(existing.GetEncryptedPassword())
	} else {
		// Encrypt new password
		encrypted, err := m.encryptor.Encrypt(conn.Password)
		if err != nil {
			return fmt.Errorf("failed to encrypt password: %w", err)
		}
		conn.SetEncryptedPassword(encrypted)
	}

	// Update in storage
	if err := m.storage.UpdateConnection(conn); err != nil {
		return fmt.Errorf("failed to update connection: %w", err)
	}

	// Close and remove existing pool if exists
	if pool, exists := m.pools[conn.ID]; exists {
		pool.Close()
		delete(m.pools, conn.ID)
	}

	m.logger.Info("Updated database connection",
		zap.Int64("id", conn.ID),
		zap.String("name", conn.Name))

	return nil
}

// GetConnection retrieves a database connection by ID
func (m *Manager) GetConnection(id int64) (*models.DBConnection, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	conn, err := m.storage.GetConnection(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	// Decrypt password if encrypted
	if encrypted := conn.GetEncryptedPassword(); encrypted != "" {
		password, err := m.encryptor.Decrypt(encrypted)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt password: %w", err)
		}
		conn.Password = password
	}

	return conn, nil
}

// GetPool gets or creates a connection pool for the specified connection
func (m *Manager) GetPool(id int64) (*ConnectionPool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if pool already exists
	if pool, exists := m.pools[id]; exists {
		return pool, nil
	}

	// Get connection configuration
	conn, err := m.storage.GetConnection(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	// Decrypt password if encrypted
	if encrypted := conn.GetEncryptedPassword(); encrypted != "" {
		password, err := m.encryptor.Decrypt(encrypted)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt password: %w", err)
		}
		conn.Password = password
	}

	// Create new pool
	pool, err := NewConnectionPool(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Store pool for reuse
	m.pools[id] = pool

	m.logger.Info("Created new connection pool",
		zap.Int64("id", id),
		zap.String("name", conn.Name))

	return pool, nil
}

// DeleteConnection deletes a database connection
func (m *Manager) DeleteConnection(id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Close and remove pool if exists
	if pool, exists := m.pools[id]; exists {
		pool.Close()
		delete(m.pools, id)
	}

	// Delete from storage
	if err := m.storage.DeleteConnection(id); err != nil {
		return fmt.Errorf("failed to delete connection: %w", err)
	}

	m.logger.Info("Deleted database connection", zap.Int64("id", id))
	return nil
}

// ListConnections lists all database connections
func (m *Manager) ListConnections() ([]*models.DBConnection, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	connections, err := m.storage.ListConnections()
	if err != nil {
		return nil, fmt.Errorf("failed to list connections: %w", err)
	}

	// Decrypt passwords for all connections
	for _, conn := range connections {
		if encrypted := conn.GetEncryptedPassword(); encrypted != "" {
			password, err := m.encryptor.Decrypt(encrypted)
			if err != nil {
				return nil, fmt.Errorf("failed to decrypt password for connection %d: %w", conn.ID, err)
			}
			conn.Password = password
		}
	}

	return connections, nil
}

// TestConnection tests if a database connection is valid
func (m *Manager) TestConnection(conn *models.DBConnection) error {
	pool, err := NewConnectionPool(conn)
	if err != nil {
		return fmt.Errorf("failed to create test pool: %w", err)
	}
	defer pool.Close()

	// Try to get a connection from the pool
	db, err := pool.Get()
	if err != nil {
		return fmt.Errorf("failed to get connection from pool: %w", err)
	}
	defer db.Close()

	// Try to ping the database
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return nil
}

// Close closes all connection pools and the storage
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Close all pools
	for _, pool := range m.pools {
		pool.Close()
	}
	m.pools = make(map[int64]*ConnectionPool)

	// Close storage
	return m.storage.Close()
}
