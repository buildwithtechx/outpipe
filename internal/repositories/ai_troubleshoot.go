package repositories

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"outpipe.dev/outpipe/internal/models"
)

type AIAgentRepository interface {
	AutoMigrate() error
	CreateThread(ctx context.Context, thread *models.AIAgentThread) error
	GetThread(ctx context.Context, orgID, threadID string) (*models.AIAgentThread, error)
	ListThreads(ctx context.Context, orgID, userID string) ([]models.AIAgentThread, error)
	DeleteThread(ctx context.Context, orgID, threadID string) error
	CreateMessage(ctx context.Context, msg *models.AIAgentMessage) error
	UpdateMessage(ctx context.Context, msg *models.AIAgentMessage) error
}

type GormAIAgentRepository struct {
	db *gorm.DB
}

func NewAIAgentRepository(db *gorm.DB) (*GormAIAgentRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection is required")
	}
	repo := &GormAIAgentRepository{db: db}
	if err := repo.AutoMigrate(); err != nil {
		return nil, fmt.Errorf("auto-migrate ai agent tables: %w", err)
	}
	return repo, nil
}

func (r *GormAIAgentRepository) AutoMigrate() error {
	if err := r.db.AutoMigrate(&models.AIAgentThread{}, &models.AIAgentMessage{}); err != nil {
		return fmt.Errorf("migrate ai agent models: %w", err)
	}
	return nil
}

func (r *GormAIAgentRepository) CreateThread(ctx context.Context, thread *models.AIAgentThread) error {
	if err := r.db.WithContext(ctx).Create(thread).Error; err != nil {
		return fmt.Errorf("create ai agent thread: %w", err)
	}
	return nil
}

func (r *GormAIAgentRepository) GetThread(ctx context.Context, orgID, threadID string) (*models.AIAgentThread, error) {
	var thread models.AIAgentThread
	if err := r.db.WithContext(ctx).Preload("Messages", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at asc")
	}).First(&thread, "id = ? AND organization_id = ?", threadID, orgID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("ai agent thread %q not found: %w", threadID, err)
		}
		return nil, fmt.Errorf("find ai agent thread %q: %w", threadID, err)
	}
	return &thread, nil
}

func (r *GormAIAgentRepository) ListThreads(ctx context.Context, orgID, userID string) ([]models.AIAgentThread, error) {
	var threads []models.AIAgentThread
	query := r.db.WithContext(ctx).Where("organization_id = ?", orgID)
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if err := query.Order("updated_at desc").Limit(50).Find(&threads).Error; err != nil {
		return nil, fmt.Errorf("list ai agent threads: %w", err)
	}
	return threads, nil
}

func (r *GormAIAgentRepository) DeleteThread(ctx context.Context, orgID, threadID string) error {
	result := r.db.WithContext(ctx).Where("id = ? AND organization_id = ?", threadID, orgID).Delete(&models.AIAgentThread{})
	if result.Error != nil {
		return fmt.Errorf("delete ai agent thread: %w", result.Error)
	}
	return nil
}

func (r *GormAIAgentRepository) CreateMessage(ctx context.Context, msg *models.AIAgentMessage) error {
	if err := r.db.WithContext(ctx).Create(msg).Error; err != nil {
		return fmt.Errorf("create ai agent message: %w", err)
	}
	return nil
}

func (r *GormAIAgentRepository) UpdateMessage(ctx context.Context, msg *models.AIAgentMessage) error {
	if err := r.db.WithContext(ctx).Save(msg).Error; err != nil {
		return fmt.Errorf("update ai agent message: %w", err)
	}
	return nil
}
