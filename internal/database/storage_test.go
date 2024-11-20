package database

import (
	"os"
	"testing"
	"time"

	"github.com/deadjoe/benchphant/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLiteStorage(t *testing.T) {
	// Create a temporary database file
	tmpfile, err := os.CreateTemp("", "test_storage_*.db")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	// Create storage
	storage, err := NewSQLiteStorage(tmpfile.Name())
	require.NoError(t, err)
	defer storage.Close()

	// Test SaveConnection
	t.Run("SaveConnection", func(t *testing.T) {
		conn := &models.DBConnection{
			Name:        "test_db",
			Type:        "mysql",
			Host:        "localhost",
			Port:        3306,
			Username:    "root",
			Password:    "password",
			Database:    "test",
			Options:     map[string]string{"charset": "utf8mb4"},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			LastUsedAt:  time.Now(),
			IsCluster:   false,
			MaxIdleConn: 10,
			MaxOpenConn: 100,
		}

		err := storage.SaveConnection(conn)
		assert.NoError(t, err)
		assert.Greater(t, conn.ID, int64(0))
	})

	// Test GetConnection
	t.Run("GetConnection", func(t *testing.T) {
		conn := &models.DBConnection{
			Name:        "test_db2",
			Type:        "mysql",
			Host:        "localhost",
			Port:        3306,
			Username:    "root",
			Password:    "password",
			Database:    "test",
			Options:     map[string]string{"charset": "utf8mb4"},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			LastUsedAt:  time.Now(),
			IsCluster:   false,
			MaxIdleConn: 10,
			MaxOpenConn: 100,
		}

		err := storage.SaveConnection(conn)
		require.NoError(t, err)

		retrieved, err := storage.GetConnection(conn.ID)
		assert.NoError(t, err)
		assert.Equal(t, conn.Name, retrieved.Name)
		assert.Equal(t, conn.Type, retrieved.Type)
		assert.Equal(t, conn.Host, retrieved.Host)
		assert.Equal(t, conn.Port, retrieved.Port)
	})

	// Test ListConnections
	t.Run("ListConnections", func(t *testing.T) {
		connections, err := storage.ListConnections()
		assert.NoError(t, err)
		assert.Len(t, connections, 2) // We added two connections in previous tests
	})

	// Test UpdateConnection
	t.Run("UpdateConnection", func(t *testing.T) {
		conn := &models.DBConnection{
			Name:        "test_db3",
			Type:        "mysql",
			Host:        "localhost",
			Port:        3306,
			Username:    "root",
			Password:    "password",
			Database:    "test",
			Options:     map[string]string{"charset": "utf8mb4"},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			LastUsedAt:  time.Now(),
			IsCluster:   false,
			MaxIdleConn: 10,
			MaxOpenConn: 100,
		}

		err := storage.SaveConnection(conn)
		require.NoError(t, err)

		conn.Name = "updated_name"
		conn.Port = 3307
		err = storage.UpdateConnection(conn)
		assert.NoError(t, err)

		retrieved, err := storage.GetConnection(conn.ID)
		assert.NoError(t, err)
		assert.Equal(t, "updated_name", retrieved.Name)
		assert.Equal(t, 3307, retrieved.Port)
	})

	// Test DeleteConnection
	t.Run("DeleteConnection", func(t *testing.T) {
		conn := &models.DBConnection{
			Name:        "test_db4",
			Type:        "mysql",
			Host:        "localhost",
			Port:        3306,
			Username:    "root",
			Password:    "password",
			Database:    "test",
			Options:     map[string]string{"charset": "utf8mb4"},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			LastUsedAt:  time.Now(),
			IsCluster:   false,
			MaxIdleConn: 10,
			MaxOpenConn: 100,
		}

		err := storage.SaveConnection(conn)
		require.NoError(t, err)

		err = storage.DeleteConnection(conn.ID)
		assert.NoError(t, err)

		_, err = storage.GetConnection(conn.ID)
		assert.Error(t, err)
	})

	// Test UpdateLastUsed
	t.Run("UpdateLastUsed", func(t *testing.T) {
		conn := &models.DBConnection{
			Name:        "test_db5",
			Type:        "mysql",
			Host:        "localhost",
			Port:        3306,
			Username:    "root",
			Password:    "password",
			Database:    "test",
			Options:     map[string]string{"charset": "utf8mb4"},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			LastUsedAt:  time.Now(),
			IsCluster:   false,
			MaxIdleConn: 10,
			MaxOpenConn: 100,
		}

		err := storage.SaveConnection(conn)
		require.NoError(t, err)

		newLastUsed := time.Now().Add(time.Hour)
		err = storage.UpdateLastUsed(conn.ID, newLastUsed)
		assert.NoError(t, err)

		retrieved, err := storage.GetConnection(conn.ID)
		assert.NoError(t, err)
		assert.WithinDuration(t, newLastUsed, retrieved.LastUsedAt, time.Second)
	})
}
