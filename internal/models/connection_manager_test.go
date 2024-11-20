package models

import (
	"testing"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestConnectionManager(t *testing.T) {
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
				manager, err := NewConnectionManager(tt.conn)
				if tt.wantErr {
					assert.Error(t, err)
					assert.Nil(t, manager)
				} else {
					assert.NoError(t, err)
					assert.NotNil(t, manager)
					assert.NotEmpty(t, manager.connections)
					assert.NotNil(t, manager.available)
				}
			})
		}
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
		}

		manager, err := NewConnectionManager(conn)
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
		}

		manager, err := NewConnectionManager(conn)
		assert.NoError(t, err)

		// Note: This will fail since we're not actually connecting to a database
		assert.Error(t, manager.TestConnection())
	})
}
