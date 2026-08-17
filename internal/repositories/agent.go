package repositories

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"outpipe.dev/outpipe/internal/models"
)

type AgentRepository interface {
	Create(context.Context, *models.Agent) error
	FindByID(context.Context, string) (models.Agent, error)
	FindByOrganization(context.Context, string) ([]models.Agent, error)
	FindByTokenHash(context.Context, string) (models.Agent, error)
	Update(context.Context, *models.Agent) error
	Touch(context.Context, string, time.Time, string, string, string) error
	Revoke(context.Context, string, time.Time) error
}

type GormAgentRepository struct{ db *gorm.DB }

func NewAgentRepository(db *gorm.DB) (*GormAgentRepository, error) {

	if db == nil {
		return nil, fmt.Errorf("database is required")
	}

	return &GormAgentRepository{db: db}, nil
}

func (r *GormAgentRepository) Create(ctx context.Context, agent *models.Agent) error {

	if agent == nil {
		return fmt.Errorf("agent is required")
	}

	return wrap(r.db.WithContext(ctx).Create(agent).Error, "create agent")
}

func (r *GormAgentRepository) FindByID(ctx context.Context, id string) (models.Agent, error) {
	var agent models.Agent

	if err := r.db.WithContext(ctx).First(&agent, "id = ?", id).Error; err != nil {
		return models.Agent{}, mapError(err)
	}

	return agent, nil
}

func (r *GormAgentRepository) FindByOrganization(ctx context.Context, organizationID string) ([]models.Agent, error) {
	var agents []models.Agent

	if err := r.db.WithContext(ctx).Where("organization_id = ?", organizationID).Order("created_at ASC").Find(&agents).Error; err != nil {
		return nil, fmt.Errorf("list agents: %w", err)
	}

	return agents, nil
}

func (r *GormAgentRepository) FindByTokenHash(ctx context.Context, tokenHash string) (models.Agent, error) {
	var agent models.Agent

	if err := r.db.WithContext(ctx).Where("token_hash = ?", tokenHash).First(&agent).Error; err != nil {
		return models.Agent{}, mapError(err)
	}

	return agent, nil
}

func (r *GormAgentRepository) Update(ctx context.Context, agent *models.Agent) error {

	if agent == nil {
		return fmt.Errorf("agent is required")
	}

	return wrap(r.db.WithContext(ctx).Save(agent).Error, "update agent")
}

func (r *GormAgentRepository) Touch(ctx context.Context, id string, at time.Time, version, hostname, platform string) error {
	result := r.db.WithContext(ctx).Model(&models.Agent{}).Where("id = ? AND revoked_at IS NULL", id).Updates(map[string]any{"status": models.AgentStatusOnline, "last_seen_at": at, "connected_at": at, "version": version, "hostname": hostname, "platform": platform})

	if result.Error != nil {
		return fmt.Errorf("touch agent: %w", result.Error)
	}

	if result.RowsAffected != 1 {
		return ErrNotFound
	}

	return nil
}

func (r *GormAgentRepository) Revoke(ctx context.Context, id string, at time.Time) error {
	result := r.db.WithContext(ctx).Model(&models.Agent{}).Where("id = ? AND revoked_at IS NULL", id).Updates(map[string]any{"status": models.AgentStatusRevoked, "revoked_at": at})

	if result.Error != nil {
		return fmt.Errorf("revoke agent: %w", result.Error)
	}

	if result.RowsAffected != 1 {
		return ErrNotFound
	}

	return nil
}
