package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"outpipe.dev/outpipe/internal/config"
	"outpipe.dev/outpipe/internal/infra/locks"
	"outpipe.dev/outpipe/internal/infra/postgres"
	"outpipe.dev/outpipe/internal/infra/redis"
	"outpipe.dev/outpipe/internal/infra/telemetry"
	"outpipe.dev/outpipe/internal/repositories"
	"outpipe.dev/outpipe/internal/services"
	"outpipe.dev/outpipe/internal/workers"
)

var version = "dev"

func main() {
	cfg, err := config.LoadCron()
	if err != nil {
		log.Fatal(err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	db, err := postgres.Open(ctx, postgres.Config{DSN: cfg.Database.URL, MaxOpenConns: cfg.Database.MaxConns, MaxIdleConns: cfg.Database.MaxConns, ConnMaxLifetime: cfg.Database.MaxLifetime})
	if err != nil {
		log.Fatal(err)
	}
	redisClient, err := redis.Open(ctx, redis.Config{Host: cfg.Redis.Host, Port: cfg.Redis.Port, Password: cfg.Redis.Password, DB: cfg.Redis.DB})
	if err != nil {
		log.Fatal(err)
	}
	defer redisClient.Close()
	lease, err := locks.Acquire(ctx, redisClient.Raw(), "outpipe:cron", 10*time.Minute)
	if err != nil {
		log.Fatal(err)
	}
	defer lease.Release(context.Background())

	reporter := telemetry.NewSlog(nil)
	alerts, err := services.NewAlertService(reporter)
	if err != nil {
		log.Fatal(err)
	}

	sessions, err := repositories.NewSessionRepository(db)
	if err != nil {
		log.Fatal(err)
	}
	keys, err := repositories.NewAPIKeyRepository(db)
	if err != nil {
		log.Fatal(err)
	}
	deviceLogins, err := repositories.NewDeviceLoginRepository(db)
	if err != nil {
		log.Fatal(err)
	}
	organizations, err := repositories.NewOrganizationRepository(db)
	if err != nil {
		log.Fatal(err)
	}
	billing, err := repositories.NewBillingRepository(db)
	if err != nil {
		log.Fatal(err)
	}
	usageRepository, err := repositories.NewUsageRepository(db)
	if err != nil {
		log.Fatal(err)
	}
	audit, err := repositories.NewAuditRepository(db)
	if err != nil {
		log.Fatal(err)
	}
	usageService, err := services.NewUsageService(usageRepository)
	if err != nil {
		log.Fatal(err)
	}
	aggregation, err := services.NewUsageAggregationService(organizations, usageService)
	if err != nil {
		log.Fatal(err)
	}
	aggregation.SetAlerts(alerts)
	aggregation.SetBilling(billing)

	retention, err := services.NewRetentionService(organizations, billing, usageRepository, audit)
	if err != nil {
		log.Fatal(err)
	}
	billingReconciler, err := services.NewSubscriptionReconcilerService(billing, alerts)
	if err != nil {
		log.Fatal(err)
	}

	cleanup, err := workers.NewCleanupJob(sessions, keys, deviceLogins)
	if err != nil {
		log.Fatal(err)
	}
	usageJob, err := workers.NewUsageJob(aggregation)
	if err != nil {
		log.Fatal(err)
	}
	retentionJob, err := workers.NewRetentionJob(retention)
	if err != nil {
		log.Fatal(err)
	}
	billingJob, err := workers.NewBillingJob(billingReconciler)
	if err != nil {
		log.Fatal(err)
	}

	operations, err := redis.NewOperations(redisClient)
	if err != nil {
		log.Fatal(err)
	}
	operationalService, err := services.NewOperationsService(organizations, operations)
	if err != nil {
		log.Fatal(err)
	}
	operationalService.SetAlerts(alerts)

	reconciliation, err := workers.NewReconciliationJob(operationalService)
	if err != nil {
		log.Fatal(err)
	}

	tracker := workers.NewStatusTracker()
	retryConfig := workers.DefaultRetryConfig()
	dlh := workers.NewSlogDeadLetterHandler(nil)

	wrappedCleanup, err := workers.NewRetryableJob(cleanup, retryConfig, dlh, tracker)
	if err != nil {
		log.Fatal(err)
	}
	wrappedUsage, err := workers.NewRetryableJob(usageJob, retryConfig, dlh, tracker)
	if err != nil {
		log.Fatal(err)
	}
	wrappedRetention, err := workers.NewRetryableJob(retentionJob, retryConfig, dlh, tracker)
	if err != nil {
		log.Fatal(err)
	}
	wrappedBilling, err := workers.NewRetryableJob(billingJob, retryConfig, dlh, tracker)
	if err != nil {
		log.Fatal(err)
	}
	wrappedReconciliation, err := workers.NewRetryableJob(reconciliation, retryConfig, dlh, tracker)
	if err != nil {
		log.Fatal(err)
	}

	if startupLease, err := locks.Acquire(ctx, redisClient.Raw(), "outpipe:cron:startup-subscription-maintenance", 5*time.Minute); err == nil {
		_ = wrappedRetention.Run(ctx)
		_ = wrappedBilling.Run(ctx)
		_ = startupLease.Release(context.Background())
	}

	runner, err := workers.NewRunner([]workers.Job{wrappedCleanup, wrappedUsage, wrappedRetention, wrappedBilling, wrappedReconciliation}, time.Hour, nil)
	if err != nil {
		log.Fatal(err)
	}

	if err := runner.RunOnce(ctx); err != nil {
		log.Fatal(err)
	}
}
