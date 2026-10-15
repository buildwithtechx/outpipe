package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"outpipe.dev/outpipe/internal/config"
	"outpipe.dev/outpipe/pkg/client"
)

type DomainDTO struct {
	ID                 string `json:"id"`
	OrganizationID     string `json:"organizationId"`
	Hostname           string `json:"hostname"`
	Status             string `json:"status"`
	VerificationMethod string `json:"verificationMethod"`
	CertificateStatus  string `json:"certificateStatus"`
}

type DNSProviderDTO struct {
	ID       string `json:"id"`
	Provider string `json:"provider"`
	Label    string `json:"label"`
	LastFour string `json:"lastFour"`
}

func newDomainsCommand(cfg config.CLIConfig) *cobra.Command {
	command := &cobra.Command{Use: "domains", Short: "manage custom domains", Args: cobra.NoArgs}
	command.AddCommand(
		&cobra.Command{Use: "add HOSTNAME", Short: "register a custom domain", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
			return runDomainAction(cmd, cfg, "add", args)
		}},
		&cobra.Command{Use: "add-bundle APEX", Short: "register an apex plus wildcard bundle", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
			return runDomainAction(cmd, cfg, "add-bundle", args)
		}},
		&cobra.Command{Use: "list", Short: "list custom domains", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, args []string) error {
			return runDomainAction(cmd, cfg, "list", args)
		}},
		&cobra.Command{Use: "verify DOMAIN_ID", Short: "verify a custom domain", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
			return runDomainAction(cmd, cfg, "verify", args)
		}},
		&cobra.Command{Use: "rm DOMAIN_ID", Short: "delete a custom domain", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
			return runDomainAction(cmd, cfg, "rm", args)
		}},
	)
	command.PersistentFlags().Bool("json", false, "output in JSON format")
	command.PersistentFlags().String("organization", "", "organization ID")
	command.PersistentFlags().String("method", "dns", "verification method (dns, http, dns-automation, dns-delegation)")
	command.PersistentFlags().String("provider", "", "dns provider for automated verification")
	command.PersistentFlags().Bool("delegated", false, "use CNAME delegation (add-bundle only)")
	command.PersistentFlags().String("tunnel-id", "", "tunnel to attach the domain to")
	command.PersistentFlags().String("token", "", "verification token")
	return command
}

func runDomainAction(cmd *cobra.Command, cfg config.CLIConfig, action string, args []string) error {
	jsonOutput, err := cmd.Flags().GetBool("json")
	if err != nil {
		return err
	}
	organizationID, err := cmd.Flags().GetString("organization")
	if err != nil {
		return err
	}
	apiClient, err := client.New(client.Config{BaseURL: cfg.APIURL, APIKey: cfg.APIKey})
	if err != nil {
		return fmt.Errorf("initialize client: %w", err)
	}
	switch action {
	case "list":
		if organizationID == "" {
			return fmt.Errorf("organization is required")
		}
		var domains []DomainDTO
		if err := apiClient.Do(cmd.Context(), http.MethodGet, "/api/v1/organizations/"+organizationID+"/domains", nil, &domains); err != nil {
			return fmt.Errorf("list domains: %w", err)
		}
		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(domains)
		}
		for _, domain := range domains {
			fmt.Printf("%s\t%s\t%s\t%s\t%s\n", domain.ID, domain.Hostname, domain.VerificationMethod, domain.Status, domain.CertificateStatus)
		}
	case "add":
		if organizationID == "" {
			return fmt.Errorf("organization is required")
		}
		method, err := cmd.Flags().GetString("method")
		if err != nil {
			return err
		}
		tunnelID, err := cmd.Flags().GetString("tunnel-id")
		if err != nil {
			return err
		}
		provider, err := cmd.Flags().GetString("provider")
		if err != nil {
			return err
		}
		payload := map[string]any{"hostname": args[0], "verificationMethod": method, "dnsProvider": provider}
		if tunnelID != "" {
			payload["tunnelId"] = tunnelID
		}
		var result struct {
			Domain            DomainDTO `json:"domain"`
			VerificationToken string    `json:"verificationToken"`
		}
		if err := apiClient.Do(cmd.Context(), http.MethodPost, "/api/v1/organizations/"+organizationID+"/domains", payload, &result); err != nil {
			return fmt.Errorf("register domain: %w", err)
		}
		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(result)
		}
		fmt.Printf("domain %s registered (%s)\nverification token (shown once): %s\n", result.Domain.Hostname, result.Domain.VerificationMethod, result.VerificationToken)
	case "add-bundle":
		if organizationID == "" {
			return fmt.Errorf("organization is required")
		}
		provider, err := cmd.Flags().GetString("provider")
		if err != nil {
			return err
		}
		tunnelID, err := cmd.Flags().GetString("tunnel-id")
		if err != nil {
			return err
		}
		delegated, err := cmd.Flags().GetBool("delegated")
		if err != nil {
			return err
		}
		payload := map[string]any{"hostname": args[0], "wildcard": true, "delegated": delegated, "dnsProvider": provider}
		if tunnelID != "" {
			payload["tunnelId"] = tunnelID
		}
		var bundle []DomainDTO
		if err := apiClient.Do(cmd.Context(), http.MethodPost, "/api/v1/organizations/"+organizationID+"/domains/bundle", payload, &bundle); err != nil {
			return fmt.Errorf("register domain bundle: %w", err)
		}
		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(bundle)
		}
		for _, domain := range bundle {
			fmt.Printf("domain %s registered (%s)\n", domain.Hostname, domain.VerificationMethod)
		}
	case "verify":
		token, err := cmd.Flags().GetString("token")
		if err != nil {
			return err
		}
		if err := apiClient.Do(cmd.Context(), http.MethodPost, "/api/v1/domains/"+args[0]+"/verify", map[string]any{"token": token}, nil); err != nil {
			return fmt.Errorf("verify domain: %w", err)
		}
		fmt.Printf("domain %s verified\n", args[0])
	case "rm":
		if err := apiClient.Do(cmd.Context(), http.MethodDelete, "/api/v1/domains/"+args[0], nil, nil); err != nil {
			return fmt.Errorf("delete domain: %w", err)
		}
		fmt.Printf("domain %s deleted\n", args[0])
	}
	return nil
}

func newDNSProvidersCommand(cfg config.CLIConfig) *cobra.Command {
	command := &cobra.Command{Use: "dns-providers", Short: "manage DNS provider credentials for automated verification", Args: cobra.NoArgs}
	command.AddCommand(
		&cobra.Command{Use: "set", Short: "connect a DNS provider", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, args []string) error {
			return runDNSProviderAction(cmd, cfg, "set", args)
		}},
		&cobra.Command{Use: "list", Short: "list connected DNS providers", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, args []string) error {
			return runDNSProviderAction(cmd, cfg, "list", args)
		}},
		&cobra.Command{Use: "rm CREDENTIAL_ID", Short: "disconnect a DNS provider", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
			return runDNSProviderAction(cmd, cfg, "rm", args)
		}},
	)
	command.PersistentFlags().Bool("json", false, "output in JSON format")
	command.PersistentFlags().String("organization", "", "organization ID")
	command.PersistentFlags().String("provider", "cloudflare", "dns provider (cloudflare, route53)")
	command.PersistentFlags().String("token", "", "provider API token (or OUTPIPE_DNS_PROVIDER_TOKEN)")
	command.PersistentFlags().String("access-key-id", "", "route53 access key id")
	command.PersistentFlags().String("secret-access-key", "", "route53 secret access key (or OUTPIPE_DNS_PROVIDER_SECRET)")
	command.PersistentFlags().String("session-token", "", "route53 session token")
	command.PersistentFlags().String("label", "", "credential label")
	return command
}

func runDNSProviderAction(cmd *cobra.Command, cfg config.CLIConfig, action string, args []string) error {
	jsonOutput, err := cmd.Flags().GetBool("json")
	if err != nil {
		return err
	}
	organizationID, err := cmd.Flags().GetString("organization")
	if err != nil {
		return err
	}
	if organizationID == "" {
		return fmt.Errorf("organization is required")
	}
	apiClient, err := client.New(client.Config{BaseURL: cfg.APIURL, APIKey: cfg.APIKey})
	if err != nil {
		return fmt.Errorf("initialize client: %w", err)
	}
	switch action {
	case "list":
		var credentials []DNSProviderDTO
		if err := apiClient.Do(cmd.Context(), http.MethodGet, "/api/v1/organizations/"+organizationID+"/dns-providers", nil, &credentials); err != nil {
			return fmt.Errorf("list dns providers: %w", err)
		}
		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(credentials)
		}
		for _, credential := range credentials {
			fmt.Printf("%s\t%s\t%s\t...%s\n", credential.ID, credential.Provider, credential.Label, credential.LastFour)
		}
	case "set":
		provider, err := cmd.Flags().GetString("provider")
		if err != nil {
			return err
		}
		token, err := cmd.Flags().GetString("token")
		if err != nil {
			return err
		}
		if token == "" {
			token = cliEnvValue("OUTPIPE_DNS_PROVIDER_TOKEN")
		}
		if provider == "route53" && token == "" {
			accessKey, err := cmd.Flags().GetString("access-key-id")
			if err != nil {
				return err
			}
			secret, err := cmd.Flags().GetString("secret-access-key")
			if err != nil {
				return err
			}
			if secret == "" {
				secret = cliEnvValue("OUTPIPE_DNS_PROVIDER_SECRET")
			}
			session, err := cmd.Flags().GetString("session-token")
			if err != nil {
				return err
			}
			if accessKey != "" && secret != "" {
				encoded, err := json.Marshal(map[string]string{"access_key_id": accessKey, "secret_access_key": secret, "session_token": session})
				if err != nil {
					return err
				}
				token = string(encoded)
			}
		}
		if token == "" {
			return fmt.Errorf("provider credential is required via --token, route53 key flags, or OUTPIPE_DNS_PROVIDER_TOKEN")
		}
		label, err := cmd.Flags().GetString("label")
		if err != nil {
			return err
		}
		var credential DNSProviderDTO
		if err := apiClient.Do(cmd.Context(), http.MethodPost, "/api/v1/organizations/"+organizationID+"/dns-providers", map[string]any{"provider": provider, "token": token, "label": label}, &credential); err != nil {
			return fmt.Errorf("connect dns provider: %w", err)
		}
		fmt.Printf("dns provider %s connected (%s)\n", credential.Provider, credential.ID)
	case "rm":
		if err := apiClient.Do(cmd.Context(), http.MethodDelete, "/api/v1/organizations/"+organizationID+"/dns-providers/"+args[0], nil, nil); err != nil {
			return fmt.Errorf("disconnect dns provider: %w", err)
		}
		fmt.Printf("dns provider credential %s deleted\n", args[0])
	}
	return nil
}
