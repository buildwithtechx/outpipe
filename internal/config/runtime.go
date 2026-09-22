package config

import "strings"

func (c AppConfig) ListenAddress() string {
	port := strings.TrimSpace(c.Port)

	if strings.HasPrefix(port, ":") {
		return port
	}

	return ":" + port
}
