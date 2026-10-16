package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"gorm.io/gorm"
	"outpipe.dev/outpipe/internal/config"
	"outpipe.dev/outpipe/internal/infra/certificates"
	"outpipe.dev/outpipe/internal/infra/mail"
	"outpipe.dev/outpipe/internal/infra/storage"
	"outpipe.dev/outpipe/internal/repositories"
	"outpipe.dev/outpipe/internal/services"
	"outpipe.dev/outpipe/internal/workers"
)

func buildDomainCertificateJob(db *gorm.DB, cfg config.CronConfig) *workers.DomainCertificateJob {
	domainRepository, err := repositories.NewDomainRepository(db)
	if err != nil {
		log.Fatal(err)
	}
	dnsCredentialRepository, err := repositories.NewDNSProviderRepository(db)
	if err != nil {
		log.Fatal(err)
	}
	domainService, err := services.NewDomainService(domainRepository)
	if err != nil {
		log.Fatal(err)
	}
	dnsCredentialService, err := services.NewDNSCredentialService(dnsCredentialRepository)
	if err != nil {
		log.Fatal(err)
	}
	domainService.SetDNSCredentials(dnsCredentialService)

	if cfg.App.ACMEEnabled && cfg.App.ACMEEmail != "" {
		dnsIssuer, err := certificates.NewDNS01Issuer(certificates.ACMEConfig{Email: cfg.App.ACMEEmail, Directory: cfg.App.ACMEDirectory, CacheDir: cfg.App.CertificateCache})
		if err != nil {
			log.Fatal(err)
		}
		domainService.SetDNS01Issuer(dnsIssuer)
		autocertIssuer, err := certificates.NewACMEIssuer(certificates.ACMEConfig{Email: cfg.App.ACMEEmail, Directory: cfg.App.ACMEDirectory, CacheDir: cfg.App.CertificateCache, AllowedHost: func(host string) bool {
			return strings.Contains(host, ".") && !strings.Contains(host, "..") && !strings.Contains(host, "*")
		}})
		if err != nil {
			log.Fatal(err)
		}
		domainService.SetCertificateIssuer(autocertIssuer)
	}

	if strings.TrimSpace(cfg.App.ChallengeZone) != "" && strings.TrimSpace(cfg.App.ChallengeDNSCredential) != "" {
		challengeSolver, err := services.SolverFor(cfg.App.ChallengeDNSProvider, cfg.App.ChallengeDNSCredential)
		if err != nil {
			log.Fatal(err)
		}
		domainService.SetChallengeSolver(cfg.App.ChallengeZone, challengeSolver)
	}

	if cfg.Auth.EncryptionKey != "" {
		secretBox, err := storage.NewSecretBox([]byte(cfg.Auth.EncryptionKey))
		if err != nil {
			log.Fatal(err)
		}
		domainService.SetDNSProtector(secretBox)
	} else {
		log.Print("warning: OUTPIPE_AUTH_ENCRYPTION_KEY is empty; automated domain renewals will fail")
	}

	job, err := workers.NewDomainCertificateJob(domainService, 30*24*time.Hour)
	if err != nil {
		log.Fatal(err)
	}
	return job
}

func buildProductJobs(
	db *gorm.DB,
	cfg config.CronConfig,
	organizations repositories.OrganizationRepository,
	emailRepository repositories.EmailRepository,
	webhookRepository repositories.WebhookRepository,
	telemetryExportRepository *repositories.GormTelemetryExportRepository,
	alerts *services.AlertService,
) (*workers.MachineFleetJob, *workers.SubscriberNotificationJob, *workers.AlertBridgeJob, *workers.PipelineHealthJob) {
	secretRepository, err := repositories.NewSecretRepository(db)
	if err != nil {
		log.Fatal(err)
	}
	secretService, err := services.NewSecretService(secretRepository, cfg.Auth.EncryptionKey)
	if err != nil {
		log.Fatal(err)
	}
	tunnelRepository, err := repositories.NewTunnelRepository(db)
	if err != nil {
		log.Fatal(err)
	}
	tunnelService, err := services.NewTunnelService(tunnelRepository)
	if err != nil {
		log.Fatal(err)
	}
	fleetRepository, err := repositories.NewMachineFleetRepository(db)
	if err != nil {
		log.Fatal(err)
	}
	fleetService, err := services.NewMachineFleetService(fleetRepository, secretService, tunnelService)
	if err != nil {
		log.Fatal(err)
	}
	fleetJob, err := workers.NewMachineFleetJob(fleetService)
	if err != nil {
		log.Fatal(err)
	}

	uptimeRepository, err := repositories.NewUptimeRepository(db)
	if err != nil {
		log.Fatal(err)
	}
	uptimeService, err := services.NewUptimeService(uptimeRepository)
	if err != nil {
		log.Fatal(err)
	}
	subscriberMailer, err := mail.NewOutboxMailer(emailRepository, func(_ context.Context, _ string) (string, error) {
		return "", fmt.Errorf("direct email enqueue only; recipient resolution is unavailable in cron")
	}, cfg.App.DashboardURL, cfg.Mail.Support)
	if err != nil {
		log.Fatal(err)
	}
	uptimeService.SetSubscriberMailer(subscriberMailer)
	uptimeService.SetSubscriberSigningKey(cfg.Auth.EncryptionKey)
	subscriberBaseURL := strings.TrimSpace(cfg.App.StatusBaseURL)
	if subscriberBaseURL == "" {
		subscriberBaseURL = strings.TrimRight(cfg.App.DashboardURL, "/")
	}
	uptimeService.SetSubscriberBaseURL(subscriberBaseURL)
	subscriberJob, err := workers.NewSubscriberNotificationJob(uptimeService)
	if err != nil {
		log.Fatal(err)
	}

	alertRepository, err := repositories.NewAlertRepository(db)
	if err != nil {
		log.Fatal(err)
	}
	alertBridge, err := services.NewAlertBridgeService(alertRepository, uptimeRepository, uptimeService)
	if err != nil {
		log.Fatal(err)
	}
	alertBridgeJob, err := workers.NewAlertBridgeJob(alertBridge)
	if err != nil {
		log.Fatal(err)
	}

	pipelineHealth, err := services.NewPipelineHealthService(webhookRepository, telemetryExportRepository, uptimeRepository)
	if err != nil {
		log.Fatal(err)
	}
	orgLister, err := workers.NewOrganizationListAdapter(organizations)
	if err != nil {
		log.Fatal(err)
	}
	pipelineJob, err := workers.NewPipelineHealthJob(pipelineHealth, orgLister, alerts)
	if err != nil {
		log.Fatal(err)
	}

	return fleetJob, subscriberJob, alertBridgeJob, pipelineJob
}

func mustWrapJob(job workers.Job, retryConfig workers.RetryConfig, dlh workers.DeadLetterHandler, tracker *workers.StatusTracker) *workers.RetryableJob {
	wrapped, err := workers.NewRetryableJob(job, retryConfig, dlh, tracker)
	if err != nil {
		log.Fatal(err)
	}
	return wrapped
}
