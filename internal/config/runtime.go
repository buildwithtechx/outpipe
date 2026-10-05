package config

import (
	"fmt"
	"strings"
)

type CLIConfigVersionError struct {
	Version int `json:"version"`
}

func (e *CLIConfigVersionError) Error() string {
	return fmt.Sprintf("cli config version %d is newer than supported version %d", e.Version, CurrentCLIConfigVersion)
}

func (c AppConfig) ListenAddress() string {
	port := strings.TrimSpace(c.Port)

	if strings.HasPrefix(port, ":") {
		return port
	}

	return ":" + port
}
