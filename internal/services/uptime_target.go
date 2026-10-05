package services

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"strings"

	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/validation"
)

func (s *UptimeService) probeTarget(ctx context.Context, monitor *models.UptimeMonitor) (int, error) {
	switch monitor.Protocol {
	case "icmp":
		return 0, probeICMP(ctx, monitor.URL)
	case "tcp":
		host, port, err := net.SplitHostPort(monitor.URL)
		if err != nil {
			return 0, fmt.Errorf("parse TCP target: %w", err)
		}
		ip, err := resolveProbeIP(ctx, host)
		if err != nil {
			return 0, err
		}
		connection, err := (&net.Dialer{}).DialContext(ctx, "tcp", net.JoinHostPort(ip.String(), port))
		if err != nil {
			return 0, fmt.Errorf("connect TCP target: %w", err)
		}
		if err := connection.Close(); err != nil {
			return 0, fmt.Errorf("close TCP connection: %w", err)
		}
		return 0, nil
	case "http", "https":
		target := monitor.URL
		if !strings.Contains(target, "://") {
			target = monitor.Protocol + "://" + target
		}
		if err := validation.ValidateWebhookURL(target); err != nil {
			return 0, fmt.Errorf("validate probe target: %w", err)
		}
		request, err := http.NewRequestWithContext(ctx, monitor.Method, target, nil)
		if err != nil {
			return 0, fmt.Errorf("create probe request: %w", err)
		}
		request.Header.Set("User-Agent", "Outpipe-Uptime-Bot/1.0")
		response, err := s.httpClient.Do(request)
		if err != nil {
			return 0, fmt.Errorf("request probe target: %w", err)
		}
		defer response.Body.Close()
		if monitor.ExpectedStatusCode != 0 {
			if response.StatusCode != monitor.ExpectedStatusCode {
				return response.StatusCode, fmt.Errorf("expected HTTP %d, received %d", monitor.ExpectedStatusCode, response.StatusCode)
			}
		} else if response.StatusCode < 200 || response.StatusCode >= 400 {
			return response.StatusCode, fmt.Errorf("HTTP status %d", response.StatusCode)
		}
		if monitor.BodyRegex != "" {
			pattern, err := regexp.Compile(monitor.BodyRegex)
			if err != nil {
				return response.StatusCode, fmt.Errorf("compile body assertion: %w", err)
			}
			body, err := io.ReadAll(io.LimitReader(response.Body, 1024*1024+1))
			if err != nil {
				return response.StatusCode, fmt.Errorf("read assertion body: %w", err)
			}
			if len(body) > 1024*1024 {
				return response.StatusCode, fmt.Errorf("assertion body exceeds 1 MiB")
			}
			if !pattern.Match(body) {
				return response.StatusCode, fmt.Errorf("response body assertion failed")
			}
		}
		return response.StatusCode, nil
	default:
		return 0, fmt.Errorf("unsupported monitor protocol")
	}
}
