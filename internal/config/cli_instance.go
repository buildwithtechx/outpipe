package config

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const defaultCLIAPIURL = "https://api.outpipe.dev"

const legacyCLIConfigFilename = "config.json"

func CanonicalAPIOrigin(value string) (string, error) {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return "", fmt.Errorf("OUTPIPE_API_URL must be a valid API origin")
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" {
		return "", fmt.Errorf("OUTPIPE_API_URL must be a valid API origin")
	}
	hostname := strings.ToLower(strings.TrimSuffix(parsed.Hostname(), "."))
	address := strings.Trim(strings.ToLower(hostname), "[]")
	loopback := hostname == "localhost" || address == "::1" || isLoopbackIPv4(address)
	if parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || (parsed.Path != "" && parsed.Path != "/") ||
		(parsed.Scheme != "https" && !(parsed.Scheme == "http" && loopback)) {
		return "", fmt.Errorf("OUTPIPE_API_URL must be an HTTPS origin (or loopback HTTP), without credentials, a path, query or fragment")
	}
	host := hostname
	if port := parsed.Port(); port != "" && !isDefaultPort(parsed.Scheme, port) {
		host = net.JoinHostPort(hostname, port)
	}
	return parsed.Scheme + "://" + host, nil
}

func isLoopbackIPv4(address string) bool {
	parsed := net.ParseIP(address)
	return parsed != nil && parsed.IsLoopback() && strings.Contains(address, ".")
}

func isDefaultPort(scheme, port string) bool {
	return (scheme == "https" && port == "443") || (scheme == "http" && port == "80")
}

func CLIConfigFilename(apiURL string) string {
	origin, err := CanonicalAPIOrigin(apiURL)
	if err != nil {
		return legacyCLIConfigFilename
	}
	if origin == defaultCLIAPIURL {
		return legacyCLIConfigFilename
	}
	sum := sha256.Sum256([]byte(origin))
	return "config.instance." + hex.EncodeToString(sum[:]) + ".json"
}

func defaultCLIConfigPath(apiURL string) (string, error) {
	directory, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("locate user configuration directory: %w", err)
	}
	return filepath.Join(directory, "outpipe", CLIConfigFilename(apiURL)), nil
}
