package services_test

import (
	"context"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/services"
	"testing"
)

func TestReplayRejectsUnsafeTargets(t *testing.T) {
	svc := services.NewObservabilityService(nil)
	for _, target := range []string{"http://127.0.0.1", "http://[::1]", "http://169.254.169.254/latest/meta-data", "http://10.1.2.3", "http://192.168.1.1", "http://user:pass@example.com", "file:///etc/passwd", "http://[::ffff:127.0.0.1]", "http://224.0.0.1", "http://localhost"} {
		t.Run(target, func(t *testing.T) {
			if _, err := svc.ReplayRequest(context.Background(), models.ReplayRequestInput{URL: target}); err == nil {
				t.Fatal("unsafe replay target accepted")
			}
		})
	}
}
