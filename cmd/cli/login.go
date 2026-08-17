package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"outpipe.dev/outpipe/internal/config"
	"outpipe.dev/outpipe/pkg/client"
)

func runLogin(cfg config.CLIConfig) {
	flags := flag.NewFlagSet("login", flag.ExitOnError)
	token := flags.String("agent-token", "", "agent token issued by the dashboard")
	code := flags.String("code", "", "existing browser device code")
	apiKey := flags.String("api-key", cfg.APIKey, "API key for management commands")
	_ = flags.Parse(os.Args[2:])
	if *token != "" {
		saveLogin(cfg, *token, *apiKey)
		return
	}
	apiClient, err := client.New(client.Config{BaseURL: cfg.APIURL})
	if err != nil {
		log.Fatal(err)
	}
	deviceCode := *code
	if deviceCode == "" {
		var started struct {
			Code string `json:"code"`
		}
		if err := apiClient.Do(context.Background(), http.MethodPost, "/api/v1/auth/device/start", nil, &started); err != nil {
			log.Fatal(err)
		}
		deviceCode = started.Code
		fmt.Printf("Authorize this device in the dashboard using code %s\n", deviceCode)
	}
	for {
		var result struct {
			Status string `json:"status"`
			Token  string `json:"token"`
		}
		if err := apiClient.Do(context.Background(), http.MethodGet, "/api/v1/auth/device/poll?code="+deviceCode, nil, &result); err != nil {
			log.Fatal(err)
		}
		if result.Token != "" {
			saveLogin(cfg, result.Token, *apiKey)
			return
		}
		time.Sleep(2 * time.Second)
	}
}

func saveLogin(cfg config.CLIConfig, token, apiKey string) {
	cfg.AgentToken = token
	cfg.APIKey = apiKey
	if err := config.SaveCLI(cfg); err != nil {
		log.Fatalf("save credentials: %v", err)
	}
	fmt.Println("credentials saved")
}
