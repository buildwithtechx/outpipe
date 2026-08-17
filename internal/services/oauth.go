package services

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log/slog"
	"strings"

	"outpipe.dev/outpipe/internal/auth"
	"outpipe.dev/outpipe/internal/models"
)

type OAuthService struct {
	auth       *AuthService
	providers  map[string]auth.OAuthProvider
	stateStore auth.OAuthStateStore
	welcome    WelcomeMailer
}

type WelcomeMailer interface {
	SendWelcome(context.Context, string, string) error
}

func NewOAuthService(authService *AuthService, providers map[string]auth.OAuthProvider, stateStore auth.OAuthStateStore) (*OAuthService, error) {

	if authService == nil || len(providers) == 0 || stateStore == nil {
		return nil, fmt.Errorf("auth service, oauth providers, and state store are required")
	}

	return &OAuthService{auth: authService, providers: providers, stateStore: stateStore}, nil
}

func (s *OAuthService) SetWelcomeMailer(welcome WelcomeMailer) { s.welcome = welcome }

func (s *OAuthService) Start(ctx context.Context, providerName, redirectURI, returnPath string) (string, error) {
	provider, ok := s.providers[strings.ToLower(strings.TrimSpace(providerName))]

	if !ok {
		return "", fmt.Errorf("oauth provider %q is unavailable", providerName)
	}

	state, err := auth.NewToken("", 32)

	if err != nil {
		return "", err
	}

	verifier, err := auth.NewToken("", 32)

	if err != nil {
		return "", err
	}

	digest := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(digest[:])

	if err := s.stateStore.Save(ctx, state, auth.OAuthState{Provider: provider.Name(), RedirectURI: redirectURI, ReturnPath: returnPath, Verifier: verifier}); err != nil {
		return "", fmt.Errorf("save oauth state: %w", err)
	}

	return provider.AuthorizeURL(state, redirectURI, challenge), nil
}

func (s *OAuthService) Callback(ctx context.Context, state, code, userAgent, ipAddress string) (string, models.Session, string, error) {
	stateValue, err := s.stateStore.Take(ctx, state)

	if err != nil {
		return "", models.Session{}, "", fmt.Errorf("consume oauth state: %w", err)
	}

	provider, ok := s.providers[stateValue.Provider]

	if !ok {
		return "", models.Session{}, "", fmt.Errorf("oauth provider %q is unavailable", stateValue.Provider)
	}

	profile, err := provider.Exchange(ctx, code, stateValue.RedirectURI, stateValue.Verifier)

	if err != nil {
		return "", models.Session{}, "", err
	}

	user, created, err := s.auth.FindOrCreateOAuthUser(ctx, profile)

	if err != nil {
		return "", models.Session{}, "", err
	}

	if created && s.welcome != nil {

		if err := s.welcome.SendWelcome(ctx, user.Email, user.Name); err != nil {
			slog.Default().Warn("welcome email delivery failed", "email", user.Email, "error", err)
		}
	}

	raw, session, err := s.auth.CreateSession(ctx, user.ID, userAgent, ipAddress)

	if err != nil {
		return "", models.Session{}, "", err
	}

	return raw, session, stateValue.ReturnPath, nil
}
