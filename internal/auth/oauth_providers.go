package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
	"outpipe.dev/outpipe/internal/infra/httpclient"
)

type OAuthConfig struct {
	GoogleClientID     string
	GoogleClientSecret string
	GitHubClientID     string
	GitHubClientSecret string
	HTTPClient         *http.Client
}

type provider struct {
	name     string
	config   *oauth2.Config
	client   *http.Client
	profile  string
	identity func(context.Context, *oauth2.Token) (OAuthProfile, error)
}

func NewOAuthProviders(cfg OAuthConfig) map[string]OAuthProvider {
	providers := make(map[string]OAuthProvider)
	client := cfg.HTTPClient

	if client == nil {
		client = httpclient.New(0)
	}

	if cfg.GoogleClientID != "" && cfg.GoogleClientSecret != "" {
		providers["google"] = &provider{name: "google", config: &oauth2.Config{ClientID: cfg.GoogleClientID, ClientSecret: cfg.GoogleClientSecret, Endpoint: google.Endpoint, Scopes: []string{"openid", "email", "profile"}}, client: client, profile: "https://openidconnect.googleapis.com/v1/userinfo", identity: googleProfile(client)}
	}

	if cfg.GitHubClientID != "" && cfg.GitHubClientSecret != "" {
		providers["github"] = &provider{name: "github", config: &oauth2.Config{ClientID: cfg.GitHubClientID, ClientSecret: cfg.GitHubClientSecret, Endpoint: github.Endpoint, Scopes: []string{"read:user", "user:email"}}, client: client, profile: "https://api.github.com/user", identity: githubProfile(client)}
	}

	return providers
}

func (p *provider) Name() string { return p.name }

func (p *provider) AuthorizeURL(state, redirectURI, codeChallenge string) string {
	config := *p.config
	config.RedirectURL = redirectURI
	return config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("code_challenge", codeChallenge), oauth2.SetAuthURLParam("code_challenge_method", "S256"))
}

func (p *provider) Exchange(ctx context.Context, code, redirectURI, verifier string) (OAuthProfile, error) {
	config := *p.config
	config.RedirectURL = redirectURI
	requestContext := context.WithValue(ctx, oauth2.HTTPClient, p.client)
	token, err := config.Exchange(requestContext, code, oauth2.SetAuthURLParam("code_verifier", verifier))

	if err != nil {
		return OAuthProfile{}, fmt.Errorf("exchange %s oauth code: %w", p.name, err)
	}

	profile, err := p.identity(requestContext, token)

	if err != nil {
		return OAuthProfile{}, err
	}

	profile.Provider = p.name
	profile.AccessToken = token.AccessToken
	profile.RefreshToken = token.RefreshToken

	if !token.Expiry.IsZero() {
		profile.TokenExpiresAt = &token.Expiry
	}

	return profile, nil
}

func googleProfile(client *http.Client) func(context.Context, *oauth2.Token) (OAuthProfile, error) {
	return func(ctx context.Context, token *oauth2.Token) (OAuthProfile, error) {
		var body struct {
			Subject string `json:"sub"`
			Email   string `json:"email"`
			Name    string `json:"name"`
			Valid   bool   `json:"email_verified"`
		}
		if err := fetchProfile(ctx, client, "https://openidconnect.googleapis.com/v1/userinfo", token, &body); err != nil {
			return OAuthProfile{}, err
		}

		return OAuthProfile{Subject: body.Subject, Email: strings.ToLower(body.Email), Name: body.Name, EmailVerified: body.Valid}, nil
	}
}

func githubProfile(client *http.Client) func(context.Context, *oauth2.Token) (OAuthProfile, error) {
	return func(ctx context.Context, token *oauth2.Token) (OAuthProfile, error) {
		var user struct {
			ID    int    `json:"id"`
			Email string `json:"email"`
			Name  string `json:"name"`
			Login string `json:"login"`
		}
		if err := fetchProfile(ctx, client, "https://api.github.com/user", token, &user); err != nil {
			return OAuthProfile{}, err
		}

		email := user.Email
		verified := email != ""

		if email == "" {
			var emails []struct {
				Email    string `json:"email"`
				Primary  bool   `json:"primary"`
				Verified bool   `json:"verified"`
			}
			if err := fetchProfile(ctx, client, "https://api.github.com/user/emails", token, &emails); err != nil {
				return OAuthProfile{}, err
			}

			for _, candidate := range emails {

				if candidate.Primary && candidate.Verified {
					email, verified = candidate.Email, true
					break
				}
			}
		}

		name := user.Name

		if name == "" {
			name = user.Login
		}

		return OAuthProfile{Subject: fmt.Sprint(user.ID), Email: strings.ToLower(email), Name: name, EmailVerified: verified}, nil
	}
}

func fetchProfile(ctx context.Context, client *http.Client, endpoint string, token *oauth2.Token, target any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)

	if err != nil {
		return fmt.Errorf("create oauth profile request: %w", err)
	}

	token.SetAuthHeader(request)
	request.Header.Set("Accept", "application/json")
	response, err := client.Do(request)

	if err != nil {
		return fmt.Errorf("fetch oauth profile: %w", err)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("oauth profile returned status %d", response.StatusCode)
	}

	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		return fmt.Errorf("decode oauth profile: %w", err)
	}

	return nil
}
