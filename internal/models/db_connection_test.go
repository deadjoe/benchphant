package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDBConnection(t *testing.T) {
	t.Run("DatabaseType", func(t *testing.T) {
		assert.Equal(t, "mysql", string(MySQL))
		assert.Equal(t, "postgresql", string(PostgreSQL))
	})

	t.Run("ValidateConnection", func(t *testing.T) {
		tests := []struct {
			name    string
			conn    *DBConnection
			wantErr bool
		}{
			{
				name: "ValidMySQLConnection",
				conn: &DBConnection{
					Name:        "test_mysql",
					Type:        MySQL,
					Host:        "localhost",
					Port:        3306,
					Username:    "root",
					Password:    "password",
					Database:    "test",
					Options:     map[string]string{"charset": "utf8mb4"},
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
					LastUsedAt:  time.Now(),
					MaxIdleConn: 10,
					MaxOpenConn: 100,
				},
				wantErr: false,
			},
			{
				name: "ValidPostgreSQLConnection",
				conn: &DBConnection{
					Name:        "test_postgres",
					Type:        PostgreSQL,
					Host:        "localhost",
					Port:        5432,
					Username:    "postgres",
					Password:    "password",
					Database:    "test",
					Options:     map[string]string{"sslmode": "disable"},
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
					LastUsedAt:  time.Now(),
					MaxIdleConn: 10,
					MaxOpenConn: 100,
				},
				wantErr: false,
			},
			{
				name: "EmptyName",
				conn: &DBConnection{
					Name:        "",
					Type:        MySQL,
					Host:        "localhost",
					Port:        3306,
					Username:    "root",
					Password:    "password",
					Database:    "test",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
					LastUsedAt:  time.Now(),
					MaxIdleConn: 10,
					MaxOpenConn: 100,
				},
				wantErr: true,
			},
			{
				name: "InvalidType",
				conn: &DBConnection{
					Name:        "test",
					Type:        "invalid",
					Host:        "localhost",
					Port:        3306,
					Username:    "root",
					Password:    "password",
					Database:    "test",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
					LastUsedAt:  time.Now(),
					MaxIdleConn: 10,
					MaxOpenConn: 100,
				},
				wantErr: true,
			},
			{
				name: "EmptyHost",
				conn: &DBConnection{
					Name:        "test",
					Type:        MySQL,
					Host:        "",
					Port:        3306,
					Username:    "root",
					Password:    "password",
					Database:    "test",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
					LastUsedAt:  time.Now(),
					MaxIdleConn: 10,
					MaxOpenConn: 100,
				},
				wantErr: true,
			},
			{
				name: "InvalidPort",
				conn: &DBConnection{
					Name:        "test",
					Type:        MySQL,
					Host:        "localhost",
					Port:        0,
					Username:    "root",
					Password:    "password",
					Database:    "test",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
					LastUsedAt:  time.Now(),
					MaxIdleConn: 10,
					MaxOpenConn: 100,
				},
				wantErr: true,
			},
			{
				name: "EmptyUsername",
				conn: &DBConnection{
					Name:        "test",
					Type:        MySQL,
					Host:        "localhost",
					Port:        3306,
					Username:    "",
					Password:    "password",
					Database:    "test",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
					LastUsedAt:  time.Now(),
					MaxIdleConn: 10,
					MaxOpenConn: 100,
				},
				wantErr: true,
			},
			{
				name: "EmptyDatabase",
				conn: &DBConnection{
					Name:        "test",
					Type:        MySQL,
					Host:        "localhost",
					Port:        3306,
					Username:    "root",
					Password:    "password",
					Database:    "",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
					LastUsedAt:  time.Now(),
					MaxIdleConn: 10,
					MaxOpenConn: 100,
				},
				wantErr: true,
			},
			{
				name: "InvalidMaxIdleConn",
				conn: &DBConnection{
					Name:        "test",
					Type:        MySQL,
					Host:        "localhost",
					Port:        3306,
					Username:    "root",
					Password:    "password",
					Database:    "test",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
					LastUsedAt:  time.Now(),
					MaxIdleConn: -1,
					MaxOpenConn: 100,
				},
				wantErr: true,
			},
			{
				name: "InvalidMaxOpenConn",
				conn: &DBConnection{
					Name:        "test",
					Type:        MySQL,
					Host:        "localhost",
					Port:        3306,
					Username:    "root",
					Password:    "password",
					Database:    "test",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
					LastUsedAt:  time.Now(),
					MaxIdleConn: 10,
					MaxOpenConn: -1,
				},
				wantErr: true,
			},
			{
				name: "ValidClusterConnection",
				conn: &DBConnection{
					Name:        "test_cluster",
					Type:        MySQL,
					Host:        "localhost",
					Port:        3306,
					Username:    "root",
					Password:    "password",
					Database:    "test",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
					LastUsedAt:  time.Now(),
					MaxIdleConn: 10,
					MaxOpenConn: 100,
					IsCluster:   true,
					RouterHost:  "router.example.com",
					RouterPort:  6446,
				},
				wantErr: false,
			},
			{
				name: "InvalidClusterNoRouterHost",
				conn: &DBConnection{
					Name:        "test_cluster",
					Type:        MySQL,
					Host:        "localhost",
					Port:        3306,
					Username:    "root",
					Password:    "password",
					Database:    "test",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
					LastUsedAt:  time.Now(),
					MaxIdleConn: 10,
					MaxOpenConn: 100,
					IsCluster:   true,
					RouterHost:  "",
					RouterPort:  6446,
				},
				wantErr: true,
			},
			{
				name: "InvalidClusterInvalidRouterPort",
				conn: &DBConnection{
					Name:        "test_cluster",
					Type:        MySQL,
					Host:        "localhost",
					Port:        3306,
					Username:    "root",
					Password:    "password",
					Database:    "test",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
					LastUsedAt:  time.Now(),
					MaxIdleConn: 10,
					MaxOpenConn: 100,
					IsCluster:   true,
					RouterHost:  "router.example.com",
					RouterPort:  0,
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := tt.conn.Validate()
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})

	t.Run("EncryptedPassword", func(t *testing.T) {
		conn := &DBConnection{
			Password: "test_password",
		}

		// Test setting encrypted password
		conn.SetEncryptedPassword("encrypted_password")
		assert.Equal(t, "encrypted_password", conn.GetEncryptedPassword())
		assert.Empty(t, conn.Password)

		// Test getting encrypted password
		assert.Equal(t, "encrypted_password", conn.GetEncryptedPassword())
	})

	t.Run("DSN", func(t *testing.T) {
		conn := &DBConnection{
			Name:        "test_mysql",
			Type:        MySQL,
			Host:        "localhost",
			Port:        3306,
			Username:    "root",
			Password:    "password",
			Database:    "test",
			Options:     map[string]string{"charset": "utf8mb4"},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			LastUsedAt:  time.Now(),
			MaxIdleConn: 10,
			MaxOpenConn: 100,
		}

		dsn := conn.DSN()
		assert.Contains(t, dsn, "root:password@tcp(localhost:3306)/test")
		assert.Contains(t, dsn, "charset=utf8mb4")
	})
}
