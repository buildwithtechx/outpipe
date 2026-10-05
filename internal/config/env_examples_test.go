package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestEnvironmentExamplesCoverConfiguration(t *testing.T) {
	examples := []struct {
		path   string
		config any
	}{
		{"cmd/server/.env.example", APIConfig{}},
		{"cmd/tunnel/.env.example", RelayConfig{}},
		{"cmd/cron/.env.example", CronConfig{}},
		{"cmd/check/.env.example", CheckConfig{}},
		{"cmd/cli/.env.example", CLIConfig{}},
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
					if name != "" && name != "-" && !keys[prefix+name] {
						t.Errorf("missing variable %s", prefix+name)
					}
				}
			}
			check(reflect.TypeOf(example.config), "")
		})
	}
}
