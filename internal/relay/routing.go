package relay

import (
	"strings"

	"outpipe.dev/outpipe/pkg/protocol"
)

func bearerToken(value string) string {
	value = strings.TrimSpace(value)
	parts := strings.SplitN(value, " ", 2)

	if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
		return strings.TrimSpace(parts[1])
	}

	return ""
}

func publicURL(open protocol.OpenTunnel, tunnelID, domain string) string {

	if open.CustomDomain != "" {
		return "https://" + strings.TrimSuffix(open.CustomDomain, ".")
	}

	name := open.Subdomain

	if name == "" {
		name = strings.Split(tunnelID, "-")[0]
	}

	return "https://" + name + "." + strings.TrimSuffix(domain, ".")
}
