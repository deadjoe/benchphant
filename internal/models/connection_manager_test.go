package models

import (
	"testing"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"database/sql"
)

func TestConnectionManager(t *testing.T) {
	// Create a mock database connection
	db := &sql.DB{}

	t.Run("NewConnectionManager", func(t *testing.T) {
		tests := []struct {
			name    string
			conn    *DBConnection
			wantErr bool
		}{
			{
				name: "ValidConnection",
				conn: &DBConnection{
					Name:     "test",
					Type:     MySQL,
					Host:     "localhost",
					Port:     3306,
					Username: "root",
					Password: "password",
					Database: "test",
					Options:  map[string]string{"charset": "utf8mb4"},
				},
				wantErr: false,
			},
			{
				name:    "NilConnection",
				conn:    nil,
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				manager, err := NewConnectionManager(db)
				assert.NoError(t, err)
				if tt.conn != nil {
					err := manager.AddConnection(tt.conn)
					if tt.wantErr {
						assert.Error(t, err)
					} else {
						assert.NoError(t, err)
						assert.NotNil(t, manager)
						assert.NotEmpty(t, manager.connections)
						assert.NotNil(t, manager.available)
					}
				}
			})
		}
	})

	t.Run("GetConnection", func(t *testing.T) {
		manager, err := NewConnectionManager(db)
		assert.NoError(t, err)

		conn := &DBConnection{
			Name:     "test",
			Type:     MySQL,
			Host:     "localhost",
			Port:     3306,
			Username: "root",
			Password: "password",
			Database: "test",
			Options:  map[string]string{"charset": "utf8mb4"},
		}

		err = manager.AddConnection(conn)
		assert.NoError(t, err)

		got, err := manager.GetConnection(conn.ID)
		assert.NoError(t, err)
		assert.Equal(t, conn, got)
	})

	t.Run("GetConnectionNotFound", func(t *testing.T) {
		manager, err := NewConnectionManager(db)
		assert.NoError(t, err)

		got, err := manager.GetConnection(999)
		assert.Error(t, err)
		assert.Nil(t, got)
	})

	t.Run("GetPut", func(t *testing.T) {
		conn := &DBConnection{
			Name:     "test",
			Type:     MySQL,
			Host:     "localhost",
			Port:     3306,
			Username: "root",
			Password: "password",
			Database: "test",
			Options:  map[string]string{"charset": "utf8mb4"},
		}

		manager, err := NewConnectionManager(db)
		assert.NoError(t, err)

		err = manager.AddConnection(conn)
		assert.NoError(t, err)

		// Test Get
		got, err := manager.Get()
		assert.NoError(t, err)
		assert.Equal(t, conn, got)
		assert.Empty(t, manager.connections)

		// Test Put
		manager.Put(got)
		assert.NotEmpty(t, manager.connections)
		assert.Contains(t, manager.connections, conn)
	})

	t.Run("TestConnection", func(t *testing.T) {
		conn := &DBConnection{
			Name:     "test",
			Type:     MySQL,
			Host:     "localhost",
			Port:     3306,
			Username: "root",
			Password: "password",
			Database: "test",
			Options:  map[string]string{"charset": "utf8mb4"},
		}

		manager, err := NewConnectionManager(db)
		assert.NoError(t, err)

		err = manager.AddConnection(conn)
		assert.NoError(t, err)

		// Note: This will fail since we're not actually connecting to a database
		assert.Error(t, manager.TestConnection())
	})
}
