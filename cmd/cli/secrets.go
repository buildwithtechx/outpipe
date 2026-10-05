package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"outpipe.dev/outpipe/internal/config"
)

func newSecretsCommand(cfg config.CLIConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secrets",
		Short: "manage and inject Outpipe secrets into processes",
	}

	cmd.AddCommand(newSecretsRunCommand(cfg))
	return cmd
}

func newSecretsRunCommand(cfg config.CLIConfig) *cobra.Command {
	var project string
	var environment string

	cmd := &cobra.Command{
		Use:   "run -- <command> [args...]",
		Short: "run a command with secrets injected into the environment",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			token := os.Getenv("OUTPIPE_TOKEN")
			if token == "" {
				return fmt.Errorf("set OUTPIPE_TOKEN to a secrets machine token with secrets:read scope")
			}

			apiURL := cfg.APIURL
			if apiURL == "" {
				apiURL = "http://localhost:8080"
			}

			reqURL := fmt.Sprintf("%s/api/v1/cli/secrets", strings.TrimRight(apiURL, "/"))
			queryParams := url.Values{}
			if project != "" {
				queryParams.Set("project", project)
			}
			if environment != "" {
				queryParams.Set("environment", environment)
			}
			if len(queryParams) > 0 {
				reqURL += "?" + queryParams.Encode()
			}

			req, err := http.NewRequestWithContext(cmd.Context(), http.MethodGet, reqURL, nil)
			if err != nil {
				return fmt.Errorf("create request: %w", err)
			}
			req.Header.Set("Authorization", "Bearer "+token)

			client := &http.Client{Timeout: 15 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
			resp, err := client.Do(req)
			if err != nil {
				return fmt.Errorf("fetch secrets: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("fetch secrets failed with status %d", resp.StatusCode)
			}

			var secrets map[string]string
			if err := json.NewDecoder(resp.Body).Decode(&secrets); err != nil {
				return fmt.Errorf("decode secrets: %w", err)
			}

			childEnv := os.Environ()
			for k, v := range secrets {
				childEnv = append(childEnv, fmt.Sprintf("%s=%s", k, v))
			}

			childCmd := exec.CommandContext(cmd.Context(), args[0], args[1:]...)
			childCmd.Env = childEnv
			childCmd.Stdin = os.Stdin
			childCmd.Stdout = os.Stdout
			childCmd.Stderr = os.Stderr

			return childCmd.Run()
		},
	}

	cmd.Flags().StringVarP(&project, "project", "p", "", "project slug or ID")
	cmd.Flags().StringVarP(&environment, "environment", "e", "", "environment slug or ID")
	return cmd
}
