package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func Open(ctx context.Context, cfg Config) (*gorm.DB, error) {

	if cfg.DSN == "" {
		return nil, fmt.Errorf("postgres dsn is required")
	}

	db, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{})

	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	sqlDB, err := db.DB()

	if err != nil {
		return nil, fmt.Errorf("get postgres pool: %w", err)
	}

	configurePool(sqlDB, cfg)

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return db, nil
}

func configurePool(db *sql.DB, cfg Config) {

	if cfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	}

	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}

	if cfg.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}
}
