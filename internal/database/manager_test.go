package database

import (
	"testing"

	"github.com/deadjoe/benchphant/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestManager(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	storage := NewMemoryStorage()

	// Create encryption key for testing
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	manager, err := NewManager(storage, key, logger)
	require.NoError(t, err)
	require.NotNil(t, manager)

	// Test AddConnection
	t.Run("AddConnection", func(t *testing.T) {
		conn := &models.DBConnection{
			Name:        "test_db",
			Type:        models.MySQL,
			Host:        "localhost",
			Port:        3306,
			Username:    "root",
			Password:    "password",
			Database:    "test",
			MaxIdleConn: 10,
			MaxOpenConn: 100,
		}

		err := manager.AddConnection(conn)
		require.NoError(t, err)
		assert.NotZero(t, conn.ID)
		assert.NotEmpty(t, conn.GetEncryptedPassword())
		assert.Empty(t, conn.Password) // Password should be cleared
	})

	// Test GetConnection
	t.Run("GetConnection", func(t *testing.T) {
		conn := &models.DBConnection{
			Name:        "test_db2",
			Type:        models.PostgreSQL,
			Host:        "localhost",
			Port:        5432,
			Username:    "postgres",
			Password:    "secret",
			Database:    "test",
			MaxIdleConn: 10,
			MaxOpenConn: 100,
		}

		err := manager.AddConnection(conn)
		require.NoError(t, err)

		retrieved, err := manager.GetConnection(conn.ID)
		require.NoError(t, err)
		assert.Equal(t, conn.Name, retrieved.Name)
		assert.Equal(t, "secret", retrieved.Password) // Password should be decrypted
	})

	// Test ListConnections
	t.Run("ListConnections", func(t *testing.T) {
		connections, err := manager.ListConnections()
		assert.NoError(t, err)
		assert.Len(t, connections, 2) // We added two connections in previous tests
	})

	// Test UpdateConnection
	t.Run("UpdateConnection", func(t *testing.T) {
		conn := &models.DBConnection{
			Name:        "test_db3",
			Type:        models.MySQL,
			Host:        "localhost",
			Port:        3306,
			Username:    "root",
			Password:    "oldpass",
			Database:    "test",
			MaxIdleConn: 10,
			MaxOpenConn: 100,
		}

		err := manager.AddConnection(conn)
		require.NoError(t, err)

		oldEncrypted := conn.GetEncryptedPassword()

		// Update connection with same password
		conn.Name = "updated_name"
		err = manager.UpdateConnection(conn)
		require.NoError(t, err)

		// Verify update with same password
		updated, err := manager.GetConnection(conn.ID)
		require.NoError(t, err)
		assert.Equal(t, "updated_name", updated.Name)
		assert.Equal(t, "oldpass", updated.Password)
		assert.Equal(t, oldEncrypted, updated.GetEncryptedPassword()) // Should keep the same encrypted password

		// Update connection with new password
		conn.Password = "newpass"
		err = manager.UpdateConnection(conn)
		require.NoError(t, err)

		// Verify update with new password
		updated, err = manager.GetConnection(conn.ID)
		require.NoError(t, err)
		assert.Equal(t, "updated_name", updated.Name)
		assert.Equal(t, "newpass", updated.Password)
		assert.NotEqual(t, oldEncrypted, updated.GetEncryptedPassword()) // Should have different encrypted password
	})

	// Test DeleteConnection
	t.Run("DeleteConnection", func(t *testing.T) {
		conn := &models.DBConnection{
			Name:        "test_db4",
			Type:        models.MySQL,
			Host:        "localhost",
			Port:        3306,
			Username:    "root",
			Password:    "password",
			Database:    "test",
			MaxIdleConn: 10,
			MaxOpenConn: 100,
		}

		err := manager.AddConnection(conn)
		require.NoError(t, err)

		err = manager.DeleteConnection(conn.ID)
		require.NoError(t, err)

		_, err = manager.GetConnection(conn.ID)
		assert.Error(t, err)
	})

	// Test GetPool
	t.Run("GetPool", func(t *testing.T) {
		conn := &models.DBConnection{
			Name:        "test_db5",
			Type:        models.MySQL,
			Host:        "invalid-host", // Use invalid host to ensure connection fails
			Port:        3306,
			Username:    "root",
			Password:    "password",
			Database:    "test",
			MaxIdleConn: 10,
			MaxOpenConn: 100,
		}

		err := manager.AddConnection(conn)
		require.NoError(t, err)

		// Get pool first time - should return pool but fail on first use
		pool1, err := manager.GetPool(conn.ID)
		require.NoError(t, err)
		require.NotNil(t, pool1)

		// Try to use the pool - should fail
		db := pool1.GetDB()
		err = db.Ping()
		require.Error(t, err)

		// Update connection
		conn.Port = 3307
		err = manager.UpdateConnection(conn)
		assert.NoError(t, err)

		// Get pool after update - should return new pool but fail on first use
		pool2, err := manager.GetPool(conn.ID)
		require.NoError(t, err)
		require.NotNil(t, pool2)

		// Try to use the new pool - should fail
		db = pool2.GetDB()
		err = db.Ping()
		require.Error(t, err)
	})

	// Test TestConnection
	t.Run("TestConnection", func(t *testing.T) {
		conn := &models.DBConnection{
			Name:        "test_db6",
			Type:        models.MySQL,
			Host:        "invalid-host",
			Port:        3306,
			Username:    "root",
			Password:    "password",
			Database:    "test",
			MaxIdleConn: 10,
			MaxOpenConn: 100,
		}

		// Should fail because host is invalid
		err := manager.TestConnection(conn)
		assert.Error(t, err)
	})

	// Test ListConnections
	t.Run("ListConnections", func(t *testing.T) {
		connections, err := manager.ListConnections()
		require.NoError(t, err)
		assert.NotEmpty(t, connections)

		// Verify passwords are decrypted
		for _, conn := range connections {
			if conn.GetEncryptedPassword() != "" {
				assert.NotEmpty(t, conn.Password)
			}
		}
	})
}
