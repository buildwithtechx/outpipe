package http

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

func databaseReady(db *gorm.DB) func(context.Context) error {
	return func(ctx context.Context) error {
		sqlDB, err := db.DB()
		if err != nil {
			return fmt.Errorf("get database connection: %w", err)
		}
		if err := sqlDB.PingContext(ctx); err != nil {
			return fmt.Errorf("ping database: %w", err)
		}
		return nil
	}
}
