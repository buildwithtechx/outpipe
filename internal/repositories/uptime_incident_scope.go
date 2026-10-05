package repositories

import (
	"context"
	"fmt"
	"outpipe.dev/outpipe/internal/models"
)

func (r *GormUptimeRepository) IncidentOrganization(ctx context.Context, id string) (string, error) {
	var incident models.UptimeIncident
	if err := r.db.WithContext(ctx).Select("organization_id").Where("id = ?", id).First(&incident).Error; err != nil {
		return "", fmt.Errorf("find incident organization: %w", mapError(err))
	}
	return incident.OrganizationID, nil
}
