package auth

import (
	"context"
	"fmt"
	"time"
)

type OAuthProfile struct {
	Provider       string
	Subject        string
	Email          string
	Name           string
	EmailVerified  bool
	AccessToken    string
	RefreshToken   string
	TokenExpiresAt *time.Time
}

type OAuthProvider interface {
	Name() string
	AuthorizeURL(state, redirectURI, codeChallenge string) string
	Exchange(context.Context, string, string, string) (OAuthProfile, error)
}

func ValidateOAuthProfile(profile OAuthProfile) error {

	if profile.Provider == "" {
		return fmt.Errorf("oauth provider is required")
	}

	if profile.Subject == "" {
		return fmt.Errorf("oauth subject is required")
	}

	if profile.Email == "" {
		return fmt.Errorf("oauth email is required")
	}

	if !profile.EmailVerified {
		return fmt.Errorf("oauth email is not verified")
	}

	return nil
}
