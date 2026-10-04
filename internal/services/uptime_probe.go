package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"outpipe.dev/outpipe/internal/models"
)

func (s *UptimeService) Probe(ctx context.Context, monitor *models.UptimeMonitor) (models.UptimeCheck, error) {
	start := time.Now()
	var statusCode int
	var success bool
	var errorMsg string

	timeout := time.Duration(monitor.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	switch monitor.Protocol {
	case "tcp":
		conn, err := net.DialTimeout("tcp", monitor.URL, timeout)
		if err != nil {
			success = false
			errorMsg = err.Error()
		} else {
			_ = conn.Close()
			success = true
			statusCode = 200
		}
	default:
		reqURL := monitor.URL
		if !strings.HasPrefix(reqURL, "http://") && !strings.HasPrefix(reqURL, "https://") {
			reqURL = "https://" + reqURL
		}
		req, err := http.NewRequestWithContext(ctx, monitor.Method, reqURL, nil)
		if err != nil {
			success = false
			errorMsg = err.Error()
		} else {
			req.Header.Set("User-Agent", "Outpipe-Uptime-Bot/1.0")
			resp, err := s.httpClient.Do(req)
			if err != nil {
				success = false
				errorMsg = err.Error()
			} else {
				_ = resp.Body.Close()
				statusCode = resp.StatusCode
				success = statusCode >= 200 && statusCode < 400
				if !success {
					errorMsg = fmt.Sprintf("HTTP status %d", statusCode)
				}
			}
		}
	}

	latency := time.Since(start).Milliseconds()
	checkID := make([]byte, 16)
	_, _ = rand.Read(checkID)

	now := time.Now().UTC()
	check := models.UptimeCheck{
		ID:           hex.EncodeToString(checkID),
		MonitorID:    monitor.ID,
		StatusCode:   statusCode,
		LatencyMs:    latency,
		Success:      success,
		ErrorMessage: errorMsg,
		CheckedAt:    now,
	}

	_ = s.repo.RecordCheck(ctx, &check)

	monitor.LastCheckAt = &now
	monitor.LatencyMs = latency
	if success {
		monitor.Status = models.MonitorStatusUp
	} else {
		monitor.Status = models.MonitorStatusDown
	}

	recent, err := s.repo.GetRecentChecks(ctx, monitor.ID, 30)
	if err == nil && len(recent) > 0 {
		successCount := 0
		for _, c := range recent {
			if c.Success {
				successCount++
			}
		}
		monitor.UptimeRatio = float64(successCount) / float64(len(recent)) * 100.0
	}
	monitor.UpdatedAt = now
	_ = s.repo.UpdateMonitor(ctx, monitor)

	return check, nil
}
