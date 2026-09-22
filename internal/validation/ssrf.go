package validation

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"
)

func IsPrivateOrLoopbackIP(ip net.IP) bool {
	if ip == nil {
		return true
	}

	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
		return true
	}

	for _, cidr := range []string{
		"0.0.0.0/8", "10.0.0.0/8", "100.64.0.0/10", "127.0.0.0/8", "169.254.0.0/16", "172.16.0.0/12", "192.0.0.0/24", "192.0.2.0/24", "192.168.0.0/16", "198.18.0.0/15", "198.51.100.0/24", "203.0.113.0/24", "::1/128", "fc00::/7", "fe80::/10", "2001:db8::/32",
	} {
		_, block, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}

		if block.Contains(ip) {
			return true
		}
	}

	return false
}

func ValidateSafeTarget(targetHost string) error {
	host := strings.TrimSpace(targetHost)

	if host == "" {
		return fmt.Errorf("target host is empty")
	}

	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}

	ips, err := net.LookupIP(host)

	if err != nil {
		parsed := net.ParseIP(host)

		if parsed != nil {
			ips = []net.IP{parsed}
		} else {
			return fmt.Errorf("resolve target host %s: %w", host, err)
		}
	}

	if slices.ContainsFunc(ips, IsPrivateOrLoopbackIP) {
		return fmt.Errorf("target host %s resolves to private/loopback IP address", host)
	}

	return nil
}

// ValidateWebhookURL validates the parts of a webhook URL that do not depend
// on a DNS lookup. The safe HTTP client enforces the network policy again at
// connection time so DNS changes cannot bypass it.
func ValidateWebhookURL(rawURL string) error {
	if err := ValidateURLSyntax(rawURL); err != nil {
		return err
	}
	parsed, _ := url.ParseRequestURI(strings.TrimSpace(rawURL))

	if ip := net.ParseIP(parsed.Hostname()); ip != nil && IsPrivateOrLoopbackIP(ip) {
		return fmt.Errorf("webhook url resolves to a private/loopback IP address")
	}

	return nil
}

func ValidateURLSyntax(rawURL string) error {
	parsed, err := url.ParseRequestURI(strings.TrimSpace(rawURL))

	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return fmt.Errorf("webhook url must be a valid http or https url")
	}

	if parsed.User != nil {
		return fmt.Errorf("webhook url must not include user information")
	}

	if parsed.Hostname() == "" {
		return fmt.Errorf("webhook url host is required")
	}

	if port := parsed.Port(); port != "" {
		if _, err := strconv.ParseUint(port, 10, 16); err != nil {
			return fmt.Errorf("webhook url port is invalid")
		}
	}

	return nil
}

// NewSafeHTTPClient creates an HTTP client that cannot use an ambient proxy,
// follow redirects, or connect to private and loopback addresses. The IP is
// resolved and validated immediately before dialing, then the dial uses that
// validated IP instead of resolving the hostname a second time.
func NewSafeHTTPClient(timeout time.Duration) *http.Client {
	dialer := &net.Dialer{Timeout: timeout, KeepAlive: 30 * time.Second}
	transport := &http.Transport{
		Proxy:                 nil,
		DialContext:           safeDialContext(dialer),
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          32,
		MaxIdleConnsPerHost:   8,
		IdleConnTimeout:       30 * time.Second,
		TLSHandshakeTimeout:   timeout,
		ResponseHeaderTimeout: timeout,
		ExpectContinueTimeout: time.Second,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func safeDialContext(dialer *net.Dialer) func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)

		if err != nil {
			return nil, fmt.Errorf("split outbound address: %w", err)
		}

		ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)

		if err != nil {
			return nil, fmt.Errorf("resolve outbound host: %w", err)
		}

		for _, ip := range ips {
			if IsPrivateOrLoopbackIP(ip) {
				continue
			}

			conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))

			if err == nil {
				return conn, nil
			}
		}

		return nil, fmt.Errorf("outbound host resolves only to private/loopback or unreachable addresses")
	}
}
