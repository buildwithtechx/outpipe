package config

import (
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestEnvironmentExamplesCoverConfiguration(t *testing.T) {
	examples := []struct {
		path   string
		config any
		unused []string
	}{
		{"cmd/server/.env.example", APIConfig{}, nil},
		{"cmd/tunnel/.env.example", RelayConfig{}, []string{"ACME_ENABLED", "ACME_EMAIL", "ACME_DIRECTORY", "CERTIFICATE_CACHE_DIR"}},
		{"cmd/cron/.env.example", CronConfig{}, []string{"ACME_ENABLED", "ACME_EMAIL", "ACME_DIRECTORY", "CERTIFICATE_CACHE_DIR", "REQUIRE_TLS", "TLS_CERT_FILE", "TLS_KEY_FILE"}},
		{"cmd/check/.env.example", CheckConfig{}, []string{"ACME_ENABLED", "ACME_EMAIL", "ACME_DIRECTORY", "CERTIFICATE_CACHE_DIR", "REQUIRE_TLS", "TLS_CERT_FILE", "TLS_KEY_FILE"}},
		{"cmd/cli/.env.example", CLIConfig{}, nil},
	}
	for _, example := range examples {
		t.Run(example.path, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("..", "..", example.path))
			if err != nil {
				t.Fatal(err)
			}
			keys := map[string]bool{}
			for _, line := range strings.Split(string(data), "\n") {
				key, _, found := strings.Cut(strings.TrimSpace(line), "=")
				if found && !strings.HasPrefix(key, "#") {
					if keys[key] {
						t.Errorf("duplicate variable %s", key)
					}
					keys[key] = true
				}
			}
			var check func(reflect.Type, string)
			check = func(kind reflect.Type, prefix string) {
				for i := 0; i < kind.NumField(); i++ {
					field := kind.Field(i)
					if field.Type.Kind() == reflect.Struct {
						check(field.Type, prefix+field.Tag.Get("envPrefix"))
						continue
					}
					name := field.Tag.Get("env")
					if slices.Contains(example.unused, name) {
						if keys[prefix+name] {
							t.Errorf("unused variable %s", prefix+name)
						}
						continue
					}
					if name != "" && name != "-" && !keys[prefix+name] {
						t.Errorf("missing variable %s", prefix+name)
					}
				}
			}
			check(reflect.TypeOf(example.config), "")
		})
	}
}
