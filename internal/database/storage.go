package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/deadjoe/benchphant/internal/models"
	_ "github.com/mattn/go-sqlite3"
)

// Storage defines the interface for database connection storage
type Storage interface {
	// SaveConnection saves a new database connection
	SaveConnection(conn *models.DBConnection) error
	// UpdateConnection updates an existing database connection
	UpdateConnection(conn *models.DBConnection) error
	// DeleteConnection deletes a database connection
	DeleteConnection(id int64) error
	// GetConnection gets a database connection by ID
	GetConnection(id int64) (*models.DBConnection, error)
	// ListConnections lists all database connections
	ListConnections() ([]*models.DBConnection, error)
	// UpdateLastUsed updates the last used timestamp
	UpdateLastUsed(id int64, lastUsed time.Time) error
	// Close closes the storage
	Close() error
}

// SQLiteStorage implements Storage interface using SQLite
type SQLiteStorage struct {
	db *sql.DB
}

// NewSQLiteStorage creates a new SQLite storage
func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	storage := &SQLiteStorage{db: db}
	if err := storage.initialize(); err != nil {
		db.Close()
		return nil, err
	}

	return storage, nil
}

// initialize creates necessary tables if they don't exist
func (s *SQLiteStorage) initialize() error {
	query := `
	CREATE TABLE IF NOT EXISTS connections (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		type TEXT NOT NULL,
		host TEXT NOT NULL,
		port INTEGER NOT NULL,
		username TEXT NOT NULL,
		password TEXT NOT NULL,
		database TEXT NOT NULL,
		options TEXT,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		last_used_at DATETIME NOT NULL,
		is_cluster BOOLEAN NOT NULL DEFAULT 0,
		router_host TEXT,
		router_port INTEGER,
		max_idle_conn INTEGER NOT NULL DEFAULT 10,
		max_open_conn INTEGER NOT NULL DEFAULT 100
	)`

	_, err := s.db.Exec(query)
	return err
}

// SaveConnection implements Storage.SaveConnection
func (s *SQLiteStorage) SaveConnection(conn *models.DBConnection) error {
	query := `
	INSERT INTO connections (
		name, type, host, port, username, password, database, options,
		created_at, updated_at, last_used_at, is_cluster, router_host,
		router_port, max_idle_conn, max_open_conn
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := s.db.Exec(query,
		conn.Name, conn.Type, conn.Host, conn.Port, conn.Username, conn.Password,
		conn.Database, conn.Options, conn.CreatedAt, conn.UpdatedAt, conn.LastUsedAt,
		conn.IsCluster, conn.RouterHost, conn.RouterPort, conn.MaxIdleConn, conn.MaxOpenConn)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	conn.ID = id
	return nil
}

// UpdateConnection implements Storage.UpdateConnection
func (s *SQLiteStorage) UpdateConnection(conn *models.DBConnection) error {
	query := `
	UPDATE connections SET
		name = ?, type = ?, host = ?, port = ?, username = ?, password = ?,
		database = ?, options = ?, updated_at = ?, is_cluster = ?, router_host = ?,
		router_port = ?, max_idle_conn = ?, max_open_conn = ?
	WHERE id = ?`

	result, err := s.db.Exec(query,
		conn.Name, conn.Type, conn.Host, conn.Port, conn.Username, conn.Password,
		conn.Database, conn.Options, conn.UpdatedAt, conn.IsCluster, conn.RouterHost,
		conn.RouterPort, conn.MaxIdleConn, conn.MaxOpenConn, conn.ID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("connection not found: %d", conn.ID)
	}

	return nil
}

// DeleteConnection implements Storage.DeleteConnection
func (s *SQLiteStorage) DeleteConnection(id int64) error {
	result, err := s.db.Exec("DELETE FROM connections WHERE id = ?", id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("connection not found: %d", id)
	}

	return nil
}

// GetConnection implements Storage.GetConnection
func (s *SQLiteStorage) GetConnection(id int64) (*models.DBConnection, error) {
	conn := &models.DBConnection{}
	err := s.db.QueryRow(`
		SELECT id, name, type, host, port, username, password, database, options,
			created_at, updated_at, last_used_at, is_cluster, router_host,
			router_port, max_idle_conn, max_open_conn
		FROM connections WHERE id = ?`, id).Scan(
		&conn.ID, &conn.Name, &conn.Type, &conn.Host, &conn.Port, &conn.Username,
		&conn.Password, &conn.Database, &conn.Options, &conn.CreatedAt, &conn.UpdatedAt,
		&conn.LastUsedAt, &conn.IsCluster, &conn.RouterHost, &conn.RouterPort,
		&conn.MaxIdleConn, &conn.MaxOpenConn)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("connection not found: %d", id)
	}
	return conn, err
}

// ListConnections implements Storage.ListConnections
func (s *SQLiteStorage) ListConnections() ([]*models.DBConnection, error) {
	rows, err := s.db.Query(`
		SELECT id, name, type, host, port, username, password, database, options,
			created_at, updated_at, last_used_at, is_cluster, router_host,
			router_port, max_idle_conn, max_open_conn
		FROM connections ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var connections []*models.DBConnection
	for rows.Next() {
		conn := &models.DBConnection{}
		err := rows.Scan(
			&conn.ID, &conn.Name, &conn.Type, &conn.Host, &conn.Port, &conn.Username,
			&conn.Password, &conn.Database, &conn.Options, &conn.CreatedAt, &conn.UpdatedAt,
			&conn.LastUsedAt, &conn.IsCluster, &conn.RouterHost, &conn.RouterPort,
			&conn.MaxIdleConn, &conn.MaxOpenConn)
		if err != nil {
			return nil, err
		}
		connections = append(connections, conn)
	}

	return connections, rows.Err()
}

// UpdateLastUsed implements Storage.UpdateLastUsed
func (s *SQLiteStorage) UpdateLastUsed(id int64, lastUsed time.Time) error {
	_, err := s.db.Exec("UPDATE connections SET last_used_at = ? WHERE id = ?", lastUsed, id)
	return err
}

// Close implements Storage.Close
func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}
