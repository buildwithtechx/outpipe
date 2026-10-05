package repositories

import (
	"context"
	"fmt"
	"outpipe.dev/outpipe/internal/models"
)

func (r *GormTunnelRepository) SetCapture(ctx context.Context, id string, enabled bool) error {
	result := r.db.WithContext(ctx).Model(&models.Tunnel{}).Where("id = ? AND status <> ? AND revoked_at IS NULL", id, models.TunnelStatusRevoked).Update("capture_enabled", enabled)
	if result.Error != nil {
		return fmt.Errorf("persist capture setting: %w", result.Error)
	}
	if result.RowsAffected != 1 {
		return ErrNotFound
	}
	return nil
}
