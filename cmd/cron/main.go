package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"outpipe.dev/outpipe/internal/config"
	"outpipe.dev/outpipe/internal/infra/locks"
	"outpipe.dev/outpipe/internal/infra/mail"
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
	if err := postgres.Migrate(db); err != nil {
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
	telemetryQuota, err := repositories.NewTelemetryQuotaRepository(db)
	if err != nil {
		log.Fatal(err)
	}
	audit, err := repositories.NewAuditRepository(db)
	if err != nil {
		log.Fatal(err)
	}
	emailRepository, err := repositories.NewEmailRepository(db)
	if err != nil {
		log.Fatal(err)
	}
	webhookRepository, err := repositories.NewWebhookRepository(db)
	if err != nil {
		log.Fatal(err)
	}

	domainCertificateJob := buildDomainCertificateJob(db, cfg)

	telemetryExportRepository := repositories.NewGormTelemetryExportRepository(db)
	telemetryExportService := services.NewObservabilityService(repositories.NewGormObservabilityRepository(db))
	telemetryExportService.SetExportQueue(telemetryExportRepository)
	telemetryExporter, err := services.NewHTTPAnalyticsExporter(cfg.Analytics)
	if err != nil {
		log.Fatal(err)
	}
	telemetryExportService.SetTelemetryExporter(telemetryExporter)
	telemetryExportJob, err := workers.NewTelemetryExportJob(telemetryExportService)
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
	telemetryJob, err := workers.NewTelemetryReceiptJob(telemetryQuota)
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
	webhookService, err := services.NewWebhookService(webhookRepository)
	if err != nil {
		log.Fatal(err)
	}
	webhookJob, err := workers.NewWebhookJob(webhookService)
	if err != nil {
		log.Fatal(err)
	}

	fleetJob, subscriberJob, alertBridgeJob, pipelineJob := buildProductJobs(db, cfg, organizations, emailRepository, webhookRepository, telemetryExportRepository, alerts)

	var emailJob *workers.EmailJob
	if cfg.Mail.ZeptoAPIKey != "" {
		zepto, mailErr := mail.NewZeptoClient(mail.Config{URL: cfg.Mail.ZeptoURL, APIKey: cfg.Mail.ZeptoAPIKey, FromAddress: cfg.Mail.FromAddress}, nil)
		if mailErr != nil {
			log.Fatal(mailErr)
		}
		emailJob, err = workers.NewEmailJob(emailRepository, mail.NewSender(zepto), nil)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Print("warning: OUTPIPE_ZEPTO_API_KEY is empty; queued emails will remain pending")
	}

	var backupJob *workers.BackupJob
	if cfg.Backup.Directory != "" {
		backupJob, err = workers.NewBackupJob(workers.BackupConfig{Directory: cfg.Backup.Directory, DatabaseURL: cfg.Database.URL, PgDumpPath: cfg.Backup.PgDumpPath, PgRestorePath: cfg.Backup.PgRestorePath, Keep: cfg.Backup.Keep})
		if err != nil {
			log.Fatal(err)
		}
	}

	tracker := workers.NewStatusTracker()
	retryConfig := workers.DefaultRetryConfig()
	dlh, err := workers.NewAlertingDeadLetterHandler(nil, alerts)
	if err != nil {
		log.Fatal(err)
	}

	wrappedRetention := mustWrapJob(retentionJob, retryConfig, dlh, tracker)
	wrappedBilling := mustWrapJob(billingJob, retryConfig, dlh, tracker)

	jobs := []workers.Job{
		mustWrapJob(cleanup, retryConfig, dlh, tracker),
		mustWrapJob(usageJob, retryConfig, dlh, tracker),
		wrappedRetention,
		wrappedBilling,
		mustWrapJob(reconciliation, retryConfig, dlh, tracker),
		mustWrapJob(webhookJob, retryConfig, dlh, tracker),
		mustWrapJob(telemetryJob, retryConfig, dlh, tracker),
		mustWrapJob(subscriberJob, retryConfig, dlh, tracker),
		mustWrapJob(domainCertificateJob, retryConfig, dlh, tracker),
		mustWrapJob(pipelineJob, retryConfig, dlh, tracker),
		mustWrapJob(alertBridgeJob, retryConfig, dlh, tracker),
	}
	if cfg.Analytics.TinybirdURL != "" || cfg.Analytics.ClickHouseURL != "" {
		jobs = append(jobs, mustWrapJob(telemetryExportJob, retryConfig, dlh, tracker))
	} else {
		log.Print("warning: analytics destination URLs are empty; telemetry export job disabled")
	}
	if fleetJob != nil {
		jobs = append(jobs, mustWrapJob(fleetJob, retryConfig, dlh, tracker))
	}
	if emailJob != nil {
		jobs = append(jobs, mustWrapJob(emailJob, retryConfig, dlh, tracker))
	}
	if backupJob != nil {
		jobs = append(jobs, mustWrapJob(backupJob, retryConfig, dlh, tracker))
	}

	if startupLease, err := locks.Acquire(ctx, redisClient.Raw(), "outpipe:cron:startup-subscription-maintenance", 5*time.Minute); err == nil {
		_ = wrappedRetention.Run(ctx)
		_ = wrappedBilling.Run(ctx)
		_ = startupLease.Release(context.Background())
	}

	runner, err := workers.NewRunner(jobs, time.Hour, nil)
	if err != nil {
		log.Fatal(err)
	}
	if err := runner.RunOnce(ctx); err != nil {
		log.Fatal(err)
	}

	exporter := telemetry.NewMetricsExporter()
	for _, status := range tracker.Statuses() {
		exporter.SetWorkerStatus(status.Name, string(status.State))
		log.Printf("job status: name=%s state=%s runs=%d failures=%d lastError=%s", status.Name, status.State, status.RunCount, status.FailureCount, status.LastError)
	}

	if cfg.Metrics || os.Getenv("OUTPIPE_CRON_METRICS") == "1" {
		fmt.Println(exporter.ExportPrometheus())
	}
}
