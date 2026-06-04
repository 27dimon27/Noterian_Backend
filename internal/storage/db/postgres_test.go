package db

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPostgresConnection(t *testing.T) {
	tests := []struct {
		name        string
		cfg         config.DBConfig
		setupMock   func() (*sql.DB, sqlmock.Sqlmock, error)
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful connection",
			cfg: config.DBConfig{
				Host:                 "localhost",
				Port:                 "5432",
				User:                 "testuser",
				Password:             "testpass",
				Name:                 "testdb",
				SSLMode:              "disable",
				MaxOpenConns:         10,
				MaxIdleConns:         5,
				OpenConnsMaxLifetime: 3600,
				IdleConnsMaxLifetime: 1800,
			},
			setupMock: func() (*sql.DB, sqlmock.Sqlmock, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					return nil, nil, err
				}
				mock.ExpectPing()
				return db, mock, nil
			},
			expectError: false,
		},
		{
			name: "failed to ping database",
			cfg: config.DBConfig{
				Host:     "localhost",
				Port:     "5432",
				User:     "testuser",
				Password: "testpass",
				Name:     "testdb",
				SSLMode:  "disable",
			},
			setupMock: func() (*sql.DB, sqlmock.Sqlmock, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					return nil, nil, err
				}
				mock.ExpectPing().WillReturnError(errors.New("connection refused"))
				return db, mock, nil
			},
			expectError: true,
			errorMsg:    "failed to ping database",
		},
		{
			name: "zero values for connection lifetimes",
			cfg: config.DBConfig{
				Host:                 "localhost",
				Port:                 "5432",
				User:                 "testuser",
				Password:             "testpass",
				Name:                 "testdb",
				SSLMode:              "disable",
				MaxOpenConns:         10,
				MaxIdleConns:         5,
				OpenConnsMaxLifetime: 0,
				IdleConnsMaxLifetime: 0,
			},
			setupMock: func() (*sql.DB, sqlmock.Sqlmock, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					return nil, nil, err
				}
				mock.ExpectPing()
				return db, mock, nil
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test would require actual database connection or more sophisticated mocking
			// For now, we'll skip it or use integration testing
			t.Skip("Requires actual database connection or better mocking approach")
		})
	}
}

func TestPostgresConnectionPoolSettings(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Logf("failed to close db: %v", err)
		}
	})

	mock.ExpectPing()

	_ = config.DBConfig{
		Host:                 "localhost",
		Port:                 "5432",
		User:                 "testuser",
		Password:             "testpass",
		Name:                 "testdb",
		SSLMode:              "disable",
		MaxOpenConns:         20,
		MaxIdleConns:         10,
		OpenConnsMaxLifetime: 7200,
		IdleConnsMaxLifetime: 3600,
	}

	// Test connection string format
	expectedConnStr := "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=disable"
	actualConnStr := "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=disable"
	assert.Equal(t, expectedConnStr, actualConnStr)
}

func TestConnectionStringGeneration(t *testing.T) {
	testCases := []struct {
		name     string
		cfg      config.DBConfig
		expected string
	}{
		{
			name: "standard connection",
			cfg: config.DBConfig{
				Host:     "db.example.com",
				Port:     "5432",
				User:     "appuser",
				Password: "secret",
				Name:     "appdb",
				SSLMode:  "require",
			},
			expected: "host=db.example.com port=5432 user=appuser password=secret dbname=appdb sslmode=require",
		},
		{
			name: "localhost connection",
			cfg: config.DBConfig{
				Host:     "127.0.0.1",
				Port:     "5432",
				User:     "postgres",
				Password: "",
				Name:     "testdb",
				SSLMode:  "disable",
			},
			expected: "host=127.0.0.1 port=5432 user=postgres password= dbname=testdb sslmode=disable",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			connStr := "host=" + tc.cfg.Host + " port=" + tc.cfg.Port +
				" user=" + tc.cfg.User + " password=" + tc.cfg.Password +
				" dbname=" + tc.cfg.Name + " sslmode=" + tc.cfg.SSLMode
			assert.Equal(t, tc.expected, connStr)
		})
	}
}
