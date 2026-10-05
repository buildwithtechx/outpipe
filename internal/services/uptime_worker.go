package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"outpipe.dev/outpipe/internal/models"
)

type uptimeScheduleRepository interface {
	ClaimDueMonitors(context.Context, time.Time, int) ([]models.UptimeMonitor, error)
}

func (s *UptimeService) RunScheduler(ctx context.Context, reportError func(error)) error {
	repo, ok := s.repo.(uptimeScheduleRepository)
	if !ok {
		return fmt.Errorf("uptime repository does not support scheduling")
	}
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	var pending sync.WaitGroup
	defer pending.Wait()
	slots := make(chan struct{}, 16)
	for {
		available := cap(slots) - len(slots)
		if available > 0 && ctx.Err() == nil {
			monitors, err := repo.ClaimDueMonitors(ctx, time.Now().UTC(), available)
			if err != nil {
				if reportError != nil {
					reportError(fmt.Errorf("schedule uptime probes: %w", err))
				}
			}
			for _, monitor := range monitors {
				slots <- struct{}{}
				pending.Add(1)
				go func(monitor models.UptimeMonitor) {
					defer pending.Done()
					defer func() { <-slots }()
					if _, err := s.Probe(ctx, &monitor); err != nil && reportError != nil {
						reportError(fmt.Errorf("run scheduled uptime probe: %w", err))
					}
				}(monitor)
			}
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}
