package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"outpipe.dev/outpipe/internal/config"
	"outpipe.dev/outpipe/internal/infra/redis"
)

var version = "dev"

type CheckResult struct {
	Check     string `json:"check"`
	Status    string `json:"status"`
	Error     string `json:"error,omitempty"`
	LatencyMS int64  `json:"latencyMs"`
}

func main() {
	cfg, err := config.LoadCheck()

	if err != nil {
		log.Fatal(err)
	}

	flags := flag.NewFlagSet("check", flag.ExitOnError)
	jsonOutput := flags.Bool("json", false, "print results as JSON for monitoring systems")
	apiURL := flags.String("api-url", cfg.Service.InternalAPIURL, "internal API URL to probe")
	relayURL := flags.String("relay-url", cfg.RelayURL, "relay URL to probe (http/https/ws/wss); empty to skip")
	databaseURL := flags.String("database-url", cfg.Database.URL, "database connection URL to probe; empty to skip")
	redisHost := flags.String("redis-host", cfg.Redis.Host, "redis host to probe; empty to skip")
	timeout := flags.Duration("timeout", 5*time.Second, "overall probe timeout")
	_ = flags.Parse(os.Args[1:])

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	var checks []func() CheckResult

	if strings.TrimSpace(*apiURL) != "" {
		checks = append(checks, func() CheckResult { return checkAPI(ctx, *apiURL, cfg.Service.InternalAPISecret) })
	}

	if strings.TrimSpace(*relayURL) != "" {
		checks = append(checks, func() CheckResult { return checkRelay(ctx, *relayURL) })
	}

	if strings.TrimSpace(*databaseURL) != "" {
		checks = append(checks, func() CheckResult { return checkDatabase(ctx, *databaseURL) })
	}

	if strings.TrimSpace(*redisHost) != "" {
		checks = append(checks, func() CheckResult {
			return checkRedis(ctx, redis.Config{Host: *redisHost, Port: cfg.Redis.Port, Password: cfg.Redis.Password, DB: cfg.Redis.DB})
		})
	}

	if len(checks) == 0 {
		log.Fatal("no checks configured")
	}

	results := make([]CheckResult, len(checks))

	if len(checks) == 1 {
		results[0] = checks[0]()
	} else {
		var wait sync.WaitGroup
		wait.Add(len(checks))

		for index, run := range checks {
			index, run := index, run
			go func() {
				defer wait.Done()
				results[index] = run()
			}()
		}

		wait.Wait()
	}

	if *jsonOutput {
		encoded, err := json.MarshalIndent(struct {
			Version string        `json:"version"`
			Checked string        `json:"checkedAt"`
			Results []CheckResult `json:"results"`
		}{Version: version, Checked: time.Now().UTC().Format(time.RFC3339), Results: results}, "", "  ")

		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(string(encoded))
	} else {

		for _, result := range results {
			fmt.Printf("%s: %s\n", result.Check, result.Status)

			if result.Error != "" {
				fmt.Printf("  %s\n", result.Error)
			}
		}
	}

	for _, result := range results {

		if result.Status != "ok" {
			os.Exit(1)
		}
	}
}

func checkAPI(ctx context.Context, apiURL string, secret string) CheckResult {
	started := time.Now()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(apiURL, "/")+"/readyz", nil)

	if err != nil {
		return fail("api", started, fmt.Errorf("create api readiness request: %w", err))
	}

	if secret != "" {
		request.Header.Set("X-Internal-Secret", secret)
	}

	response, err := http.DefaultClient.Do(request)

	if err != nil {
		return fail("api", started, fmt.Errorf("api readiness check: %w", err))
	}

	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fail("api", started, fmt.Errorf("api readiness returned status %d", response.StatusCode))
	}

	return pass("api", started)
}

func checkRelay(ctx context.Context, relayURL string) CheckResult {
	started := time.Now()
	parsed, err := url.Parse(relayURL)

	if err != nil {
		return fail("relay", started, fmt.Errorf("invalid relay url: %w", err))
	}

	switch parsed.Scheme {
	case "ws":
		parsed.Scheme = "http"
	case "wss":
		parsed.Scheme = "https"
	case "http", "https":
	default:
		return fail("relay", started, fmt.Errorf("unsupported relay url scheme %q", parsed.Scheme))
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String()+"/healthz", nil)

	if err != nil {
		return fail("relay", started, fmt.Errorf("create relay health request: %w", err))
	}

	response, err := http.DefaultClient.Do(request)

	if err != nil {
		return fail("relay", started, fmt.Errorf("relay health check: %w", err))
	}

	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fail("relay", started, fmt.Errorf("relay health returned status %d", response.StatusCode))
	}

	return pass("relay", started)
}

func checkDatabase(ctx context.Context, databaseURL string) CheckResult {
	started := time.Now()
	database, err := sql.Open("pgx", databaseURL)

	if err != nil {
		return fail("database", started, fmt.Errorf("open database: %w", err))
	}

	defer database.Close()

	if err := database.PingContext(ctx); err != nil {
		return fail("database", started, fmt.Errorf("database ping: %w", err))
	}

	return pass("database", started)
}

func checkRedis(ctx context.Context, redisConfig redis.Config) CheckResult {
	started := time.Now()
	client, err := redis.Open(ctx, redisConfig)

	if err != nil {
		return fail("redis", started, fmt.Errorf("connect redis: %w", err))
	}

	defer client.Close()

	if err := client.Raw().Ping(ctx).Err(); err != nil {
		return fail("redis", started, fmt.Errorf("redis ping: %w", err))
	}

	return pass("redis", started)
}

func pass(check string, started time.Time) CheckResult {
	return CheckResult{Check: check, Status: "ok", LatencyMS: time.Since(started).Milliseconds()}
}

func fail(check string, started time.Time, err error) CheckResult {
	return CheckResult{Check: check, Status: "failed", Error: err.Error(), LatencyMS: time.Since(started).Milliseconds()}
}
