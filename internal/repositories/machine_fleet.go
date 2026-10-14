package repositories

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"outpipe.dev/outpipe/internal/models"
)

type MachineFleetRepository interface {
	CreateFleet(context.Context, *models.MachineFleet) error
	FindFleet(context.Context, string, string) (models.MachineFleet, error)
	FindFleetByID(context.Context, string) (models.MachineFleet, error)
	ListFleets(context.Context, string) ([]models.MachineFleet, error)
	DeleteFleet(context.Context, string, string) error
	AttachToken(context.Context, *models.MachineFleetToken) error
	DetachToken(context.Context, string, string, string) error
	ListFleetTokens(context.Context, string, string) ([]models.MachineFleetToken, error)
	FindFleetTokensByMachineToken(context.Context, string) ([]models.MachineFleetToken, error)
	UpsertFleetTunnel(context.Context, *models.MachineFleetTunnel) error
	RemoveFleetTunnel(context.Context, string, string) error
	ListFleetTunnels(context.Context, string, string) ([]models.MachineFleetTunnel, error)
	ListAllFleetTunnels(context.Context) ([]models.MachineFleetTunnel, error)
	CountFleetTunnels(context.Context, string, string) (int64, error)
}

type GormMachineFleetRepository struct{ db *gorm.DB }

func NewMachineFleetRepository(db *gorm.DB) (*GormMachineFleetRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("database is required")
	}
	return &GormMachineFleetRepository{db: db}, nil
}

func (r *GormMachineFleetRepository) CreateFleet(ctx context.Context, fleet *models.MachineFleet) error {
	if fleet == nil {
		return fmt.Errorf("machine fleet is required")
	}
	return wrap(r.db.WithContext(ctx).Create(fleet).Error, "create machine fleet")
}

func (r *GormMachineFleetRepository) FindFleet(ctx context.Context, orgID, fleetID string) (models.MachineFleet, error) {
	var fleet models.MachineFleet
	err := r.db.WithContext(ctx).Where("organization_id = ? AND id = ?", orgID, fleetID).First(&fleet).Error
	if err != nil {
		return models.MachineFleet{}, mapError(err)
	}
	return fleet, nil
}

func (r *GormMachineFleetRepository) FindFleetByID(ctx context.Context, fleetID string) (models.MachineFleet, error) {
	var fleet models.MachineFleet
	err := r.db.WithContext(ctx).Where("id = ?", fleetID).First(&fleet).Error
	if err != nil {
		return models.MachineFleet{}, mapError(err)
	}
	return fleet, nil
}

func (r *GormMachineFleetRepository) ListFleets(ctx context.Context, orgID string) ([]models.MachineFleet, error) {
	var fleets []models.MachineFleet
	err := r.db.WithContext(ctx).Where("organization_id = ?", orgID).Order("created_at desc").Find(&fleets).Error
	return fleets, wrap(err, "list machine fleets")
}

func (r *GormMachineFleetRepository) DeleteFleet(ctx context.Context, orgID, fleetID string) error {
	return wrap(r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("fleet_id = ?", fleetID).Delete(&models.MachineFleetTunnel{}).Error; err != nil {
			return fmt.Errorf("delete fleet tunnels: %w", err)
		}
		if err := tx.Where("organization_id = ? AND fleet_id = ?", orgID, fleetID).Delete(&models.MachineFleetToken{}).Error; err != nil {
			return fmt.Errorf("delete fleet tokens: %w", err)
		}
		if err := tx.Where("organization_id = ? AND id = ?", orgID, fleetID).Delete(&models.MachineFleet{}).Error; err != nil {
			return fmt.Errorf("delete machine fleet: %w", err)
		}
		return nil
	}), "delete machine fleet")
}

func (r *GormMachineFleetRepository) AttachToken(ctx context.Context, attachment *models.MachineFleetToken) error {
	if attachment == nil {
		return fmt.Errorf("fleet token attachment is required")
	}
	return wrap(r.db.WithContext(ctx).Create(attachment).Error, "attach fleet token")
}

func (r *GormMachineFleetRepository) DetachToken(ctx context.Context, orgID, fleetID, machineTokenID string) error {
	return wrap(r.db.WithContext(ctx).Where("organization_id = ? AND fleet_id = ? AND machine_token_id = ?", orgID, fleetID, machineTokenID).Delete(&models.MachineFleetToken{}).Error, "detach fleet token")
}

func (r *GormMachineFleetRepository) ListFleetTokens(ctx context.Context, orgID, fleetID string) ([]models.MachineFleetToken, error) {
	var attachments []models.MachineFleetToken
	err := r.db.WithContext(ctx).Where("organization_id = ? AND fleet_id = ?", orgID, fleetID).Order("created_at desc").Find(&attachments).Error
	return attachments, wrap(err, "list fleet tokens")
}

func (r *GormMachineFleetRepository) FindFleetTokensByMachineToken(ctx context.Context, machineTokenID string) ([]models.MachineFleetToken, error) {
	var attachments []models.MachineFleetToken
	err := r.db.WithContext(ctx).Where("machine_token_id = ?", machineTokenID).Find(&attachments).Error
	return attachments, wrap(err, "find fleet tokens by machine token")
}

func (r *GormMachineFleetRepository) UpsertFleetTunnel(ctx context.Context, desired *models.MachineFleetTunnel) error {
	if desired == nil {
		return fmt.Errorf("fleet tunnel is required")
	}
	return wrap(r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing models.MachineFleetTunnel
		err := tx.Where("fleet_id = ? AND tunnel_id = ?", desired.FleetID, desired.TunnelID).First(&existing).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return tx.Create(desired).Error
			}
			return err
		}
		existing.DesiredStatus = desired.DesiredStatus
		existing.TargetHost = desired.TargetHost
		existing.TargetPort = desired.TargetPort
		existing.Protocol = desired.Protocol
		existing.Name = desired.Name
		return tx.Save(&existing).Error
	}), "upsert fleet tunnel")
}

func (r *GormMachineFleetRepository) RemoveFleetTunnel(ctx context.Context, fleetID, tunnelID string) error {
	return wrap(r.db.WithContext(ctx).Where("fleet_id = ? AND tunnel_id = ?", fleetID, tunnelID).Delete(&models.MachineFleetTunnel{}).Error, "remove fleet tunnel")
}

func (r *GormMachineFleetRepository) ListFleetTunnels(ctx context.Context, orgID, fleetID string) ([]models.MachineFleetTunnel, error) {
	var desired []models.MachineFleetTunnel
	err := r.db.WithContext(ctx).Table("machine_fleet_tunnels").Joins("JOIN machine_fleets ON machine_fleets.id = machine_fleet_tunnels.fleet_id").Where("machine_fleets.organization_id = ? AND machine_fleet_tunnels.fleet_id = ?", orgID, fleetID).Find(&desired).Error
	return desired, wrap(err, "list fleet tunnels")
}

func (r *GormMachineFleetRepository) ListAllFleetTunnels(ctx context.Context) ([]models.MachineFleetTunnel, error) {
	var desired []models.MachineFleetTunnel
	err := r.db.WithContext(ctx).Find(&desired).Error
	return desired, wrap(err, "list all fleet tunnels")
}

func (r *GormMachineFleetRepository) CountFleetTunnels(ctx context.Context, orgID, fleetID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("machine_fleet_tunnels").Joins("JOIN machine_fleets ON machine_fleets.id = machine_fleet_tunnels.fleet_id").Where("machine_fleets.organization_id = ? AND machine_fleet_tunnels.fleet_id = ?", orgID, fleetID).Count(&count).Error
	return count, wrap(err, "count fleet tunnels")
}
