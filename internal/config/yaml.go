package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func resolveConfigPath(rel string) string {
	if rel == "" {
		return ""
	}
	if fi, err := os.Stat(rel); err == nil && !fi.IsDir() {
		return rel
	}
	up2 := filepath.Join("..", "..", rel)
	if fi, err := os.Stat(up2); err == nil && !fi.IsDir() {
		return up2
	}
	up1 := filepath.Join("..", rel)
	if fi, err := os.Stat(up1); err == nil && !fi.IsDir() {
		return up1
	}
	return ""
}

func environmentCandidates(serviceOrPath string) []string {
	envName := strings.ToLower(strings.TrimSpace(os.Getenv("OUTPIPE_ENV")))
	if envName == "" {
		envName = strings.ToLower(strings.TrimSpace(os.Getenv("ENV")))
	}
	if envName == "" {
		envName = "development"
	}

	if strings.HasSuffix(serviceOrPath, ".yaml") || strings.HasSuffix(serviceOrPath, ".yml") {
		nameWithoutExt := strings.TrimSuffix(serviceOrPath, filepath.Ext(serviceOrPath))
		ext := filepath.Ext(serviceOrPath)
		return []string{
			fmt.Sprintf("%s.%s%s", nameWithoutExt, envName, ext),
			serviceOrPath,
		}
	}

	service := strings.TrimPrefix(serviceOrPath, "configs/")
	return []string{
		fmt.Sprintf("configs/%s.%s.yaml", service, envName),
		fmt.Sprintf("configs/%s.%s.yml", service, envName),
	}
}

func loadYAMLConfig[T any](cfg *T, defaultPaths ...string) error {
	if len(defaultPaths) == 0 {
		return nil
	}

	explicitPath := os.Getenv("OUTPIPE_CONFIG_FILE")
	if explicitPath == "" {
		if p := os.Getenv("OUTPIPE_CONFIG_PATH"); strings.HasSuffix(p, ".yaml") || strings.HasSuffix(p, ".yml") {
			explicitPath = p
		}
	}
	if explicitPath != "" {
		contents, err := os.ReadFile(explicitPath)
		if err != nil {
			return fmt.Errorf("read config file %q: %w", explicitPath, err)
		}
		if err := yaml.Unmarshal(contents, cfg); err != nil {
			return fmt.Errorf("decode config file %q: %w", explicitPath, err)
		}
		return nil
	}

	for _, basePath := range defaultPaths {
		for _, candidate := range environmentCandidates(basePath) {
			if resolved := resolveConfigPath(candidate); resolved != "" {
				contents, err := os.ReadFile(resolved)
				if err != nil {
					return fmt.Errorf("read config file %q: %w", resolved, err)
				}
				if err := yaml.Unmarshal(contents, cfg); err != nil {
					return fmt.Errorf("decode config file %q: %w", resolved, err)
				}
				return nil
			}
		}
	}

	return nil
}
