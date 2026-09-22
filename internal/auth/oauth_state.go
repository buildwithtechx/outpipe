package auth

import "context"

type OAuthState struct {
	Provider    string `json:"provider"`
	RedirectURI string `json:"redirectUri"`
	ReturnPath  string `json:"returnPath"`
	Verifier    string `json:"verifier"`
}

type OAuthStateStore interface {
	Save(context.Context, string, OAuthState) error
	Take(context.Context, string) (OAuthState, error)
}
