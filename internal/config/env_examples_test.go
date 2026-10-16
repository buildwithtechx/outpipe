package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestEnvironmentExamplesAndYAMLCoverConfiguration(t *testing.T) {
	allowedEnvOnlyKeys := map[string]bool{
		"OUTPIPE_ENV":                      true,
		"OUTPIPE_DATABASE_URL":             true,
		"OUTPIPE_REDIS_PASSWORD":           true,
		"OUTPIPE_INTERNAL_API_SECRET":      true,
		"OUTPIPE_AUTH_ENCRYPTION_KEY":      true,
		"OUTPIPE_CHALLENGE_DNS_CREDENTIAL": true,
		"OUTPIPE_GITHUB_CLIENT_ID":         true,
		"OUTPIPE_GITHUB_CLIENT_SECRET":     true,
		"OUTPIPE_GOOGLE_CLIENT_ID":         true,
		"OUTPIPE_GOOGLE_CLIENT_SECRET":     true,
		"OUTPIPE_ZEPTO_API_KEY":            true,
		"OUTPIPE_POLAR_ACCESS_TOKEN":       true,
		"OUTPIPE_POLAR_WEBHOOK_SECRET":     true,
		"OUTPIPE_PAYSTACK_SECRET_KEY":      true,
		"OUTPIPE_BILLING_WEBHOOK_SECRET":   true,
		"OUTPIPE_TINYBIRD_TOKEN":           true,
		"OUTPIPE_CLICKHOUSE_PASSWORD":      true,
		"OPENAI_API_KEY":                   true,
		"XAI_API_KEY":                      true,
		"ANTHROPIC_API_KEY":                true,
		"TINYBIRD_INGEST_TOKEN":            true,
		"TINYBIRD_TUNNEL_INGEST_TOKEN":     true,
		"TINYBIRD_QUERY_TOKEN":             true,
		"OUTPIPE_API_KEY":                  true,
		"OUTPIPE_TOKEN":                    true,
		"OUTPIPE_AGENT_TOKEN":              true,
		"OUTPIPE_PASSWORD":                 true,
	}

	examples := []struct {
		envPath  string
		yamlPath string
		config   any
		unused   []string
	}{
		{".env.example", "configs/outpipe.development.yaml", APIConfig{}, nil},
		{"cmd/server/.env.example", "configs/outpipe.development.yaml", APIConfig{}, nil},
		{"cmd/tunnel/.env.example", "configs/relay.development.yaml", RelayConfig{}, []string{"CHALLENGE_DNS_CREDENTIAL"}},
		{"cmd/cron/.env.example", "configs/cron.development.yaml", CronConfig{}, nil},
		{"cmd/check/.env.example", "configs/check.development.yaml", CheckConfig{}, []string{"CHALLENGE_DNS_CREDENTIAL"}},
		{"cmd/cli/.env.example", "", CLIConfig{}, nil},
	}

	for _, example := range examples {
		t.Run(example.envPath, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("..", "..", example.envPath))
			if err != nil {
				t.Fatal(err)
			}
			envKeys := map[string]bool{}
			for _, line := range strings.Split(string(data), "\n") {
				key, _, found := strings.Cut(strings.TrimSpace(line), "=")
				if found && !strings.HasPrefix(key, "#") {
					if envKeys[key] {
						t.Errorf("duplicate variable %s", key)
					}
					if !allowedEnvOnlyKeys[key] {
						t.Errorf("non-sensitive variable %s should be in configs/*.yaml, not %s", key, example.envPath)
					}
					envKeys[key] = true
				}
			}

			var yamlMap map[string]any
			if example.yamlPath != "" {
				yamlData, err := os.ReadFile(filepath.Join("..", "..", example.yamlPath))
				if err != nil {
					t.Fatal(err)
				}
				if err := yaml.Unmarshal(yamlData, &yamlMap); err != nil {
					t.Fatal(err)
				}
			}

			var check func(reflect.Type, string, map[string]any)
			check = func(kind reflect.Type, prefix string, section map[string]any) {
				for i := 0; i < kind.NumField(); i++ {
					field := kind.Field(i)
					yamlKey := strings.Split(field.Tag.Get("yaml"), ",")[0]
					if field.Type.Kind() == reflect.Struct {
						var subSection map[string]any
						if section != nil {
							if m, ok := section[yamlKey].(map[string]any); ok {
								subSection = m
							}
						}
						check(field.Type, prefix+field.Tag.Get("envPrefix"), subSection)
						continue
					}
					envName := field.Tag.Get("env")
					if envName == "" || envName == "-" {
						continue
					}
					for _, unused := range example.unused {
						if envName == unused {
							goto nextField
						}
					}
					{
						fullEnv := prefix + envName
						isSensitive := field.Tag.Get("json") == "-" || allowedEnvOnlyKeys[fullEnv]
						if isSensitive && !envKeys[fullEnv] {
							t.Errorf("missing sensitive variable %s in %s", fullEnv, example.envPath)
						}
					}
				nextField:
				}
			}
			check(reflect.TypeOf(example.config), "", yamlMap)
		})
	}
}
