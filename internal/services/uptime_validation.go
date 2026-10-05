package services

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"strings"

	"outpipe.dev/outpipe/internal/validation"
)

func validateMonitorInput(ctx context.Context, input CreateMonitorInput) error {
	if input.TimeoutSeconds < 1 || input.TimeoutSeconds > 60 || input.IntervalSeconds < 10 || input.IntervalSeconds > 86400 {
		return fmt.Errorf("timeout must be 1-60 seconds and interval 10-86400 seconds")
	}
	if input.MaxLatencyMs < 0 || input.MaxLatencyMs > 60000 || len(input.BodyRegex) > 1024 || (input.ExpectedStatusCode != 0 && (input.ExpectedStatusCode < 100 || input.ExpectedStatusCode > 599)) {
		return fmt.Errorf("invalid monitor assertion")
	}
	if input.BodyRegex != "" {
		if _, err := regexp.Compile(input.BodyRegex); err != nil {
			return fmt.Errorf("compile body assertion: %w", err)
		}
	}
	switch input.Protocol {
	case "http", "https":
		target := input.URL
		if !strings.Contains(target, "://") {
			target = input.Protocol + "://" + target
		}
		if err := validation.ValidateWebhookURL(target); err != nil {
			return fmt.Errorf("validate monitor URL: %w", err)
		}
	case "tcp", "icmp":
		if input.ExpectedStatusCode != 0 || input.BodyRegex != "" {
			return fmt.Errorf("HTTP assertions require an HTTP monitor")
		}
		host := input.URL
		if input.Protocol == "tcp" {
			var port string
			var err error
			host, port, err = net.SplitHostPort(input.URL)
			if err != nil {
				return fmt.Errorf("TCP target must be host:port: %w", err)
			}
			if value, err := net.LookupPort("tcp", port); err != nil || value < 1 {
				return fmt.Errorf("invalid TCP port")
			}
		}
		if _, err := resolveProbeIP(ctx, host); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported monitor protocol")
	}
	return nil
}

func resolveProbeIP(ctx context.Context, host string) (net.IP, error) {
	ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
	if err != nil {
		return nil, fmt.Errorf("resolve probe target: %w", err)
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("probe target has no addresses")
	}
	for _, ip := range ips {
		if validation.IsPrivateOrLoopbackIP(ip) || !ip.IsGlobalUnicast() {
			return nil, fmt.Errorf("probe target is not a public address")
		}
	}
	return ips[0], nil
}
