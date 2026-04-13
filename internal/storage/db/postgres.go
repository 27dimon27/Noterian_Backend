package db

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	_ "github.com/lib/pq"
)

func NewPostgresConnection(cfg config.DBConfig) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		closeErr := db.Close()
		return nil, errors.Join(
			fmt.Errorf("failed to ping database: %w", err),
			fmt.Errorf("failed to close connection: %v", closeErr),
		)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)

	return db, nil
}
