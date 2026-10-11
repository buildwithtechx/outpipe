package http

import (
	"fmt"
	"gorm.io/gorm"
	"outpipe.dev/outpipe/internal/config"
	"outpipe.dev/outpipe/internal/repositories"
	"outpipe.dev/outpipe/internal/services"
)

func productServices(db *gorm.DB, cfg config.APIConfig) (*services.SecretService, *services.ShareService, *services.UptimeService, *services.ObservabilityService, *services.AIAgentService, error) {
	secretRepo, err := repositories.NewSecretRepository(db)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	secretService, err := services.NewSecretService(secretRepo, cfg.Auth.EncryptionKey)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	shareService, err := services.NewShareService(secretRepo)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	uptimeRepo, err := repositories.NewUptimeRepository(db)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	uptimeService, err := services.NewUptimeService(uptimeRepo)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	obsRepo := repositories.NewGormObservabilityRepository(db)
	obsService := services.NewObservabilityService(obsRepo)
	exporter, err := services.NewHTTPAnalyticsExporter(cfg.Analytics)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("configure telemetry analytics: %w", err)
	}
	obsService.SetTelemetryExporter(exporter)

	aiRepo, err := repositories.NewAIAgentRepository(db)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	aiService, err := services.NewAIAgentService(aiRepo, obsService)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	return secretService, shareService, uptimeService, obsService, aiService, nil
}
