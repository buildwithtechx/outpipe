package workers

import (
	"context"
	"fmt"
	"time"
)

type DomainCertificateRenewer interface {
	RenewDueCertificates(context.Context, time.Duration) (int, error)
}

type DomainCertificateJob struct {
	renewer DomainCertificateRenewer
	within  time.Duration
}

func NewDomainCertificateJob(renewer DomainCertificateRenewer, within time.Duration) (*DomainCertificateJob, error) {
	if renewer == nil {
		return nil, fmt.Errorf("domain certificate renewer is required")
	}
	if within <= 0 {
		within = 30 * 24 * time.Hour
	}
	return &DomainCertificateJob{renewer: renewer, within: within}, nil
}

func (j *DomainCertificateJob) Name() string {
	return "domain-certificates"
}

func (j *DomainCertificateJob) Run(ctx context.Context) error {
	if _, err := j.renewer.RenewDueCertificates(ctx, j.within); err != nil {
		return fmt.Errorf("renew domain certificates: %w", err)
	}
	return nil
}
