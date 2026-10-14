package certificates

import (
	"path/filepath"
	"strings"
)

func HostCertificatePaths(cacheDir, hostname string) (string, string) {
	normalized := strings.ToLower(strings.TrimSpace(strings.TrimSuffix(hostname, ".")))
	base := strings.ReplaceAll(strings.TrimPrefix(normalized, "*."), ".", "_")
	if strings.HasPrefix(normalized, "*.") {
		base = "wildcard_" + base
	}
	return filepath.Join(cacheDir, base+".crt"), filepath.Join(cacheDir, base+".key")
}
