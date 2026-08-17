package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"outpipe.dev/outpipe/internal/auth"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
)

type AuthService struct {
	users      repositories.UserRepository
	identities repositories.OAuthIdentityRepository
	sessions   repositories.SessionRepository
	admins     platformAdminAuthorizer
	now        func() time.Time
	sessionTTL time.Duration
	protector  SecretProtector
}

type SecretProtector interface{ Seal(string) (string, error) }

type platformAdminAuthorizer interface {
	IsPlatformAdmin(context.Context, string) (bool, error)
}

func (s *AuthService) SetSecretProtector(protector SecretProtector) { s.protector = protector }

func (s *AuthService) IsPlatformAdmin(ctx context.Context, userID string) (bool, error) {
	return s.admins.IsPlatformAdmin(ctx, userID)
}

func (s *AuthService) EnsureUserActive(ctx context.Context, userID string) error {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("find authenticated user: %w", err)
	}
	if user.Status == models.UserStatusDisabled || user.DeletedAt != nil {
		return fmt.Errorf("user account is disabled")
	}
	return nil
}

func (s *AuthService) CurrentUser(ctx context.Context, userID string) (models.User, error) {
	if strings.TrimSpace(userID) == "" {
		return models.User{}, fmt.Errorf("user id is required")
	}
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return models.User{}, fmt.Errorf("find current user: %w", err)
	}
	if user.Status == models.UserStatusDisabled || user.DeletedAt != nil {
		return models.User{}, fmt.Errorf("user account is disabled")
	}
	return user, nil
}

func NewAuthService(users repositories.UserRepository, identities repositories.OAuthIdentityRepository, sessions repositories.SessionRepository, admins platformAdminAuthorizer, sessionTTL time.Duration) (*AuthService, error) {
	if users == nil || identities == nil || sessions == nil || admins == nil {
		return nil, fmt.Errorf("auth repositories are required")
	}
	if sessionTTL <= 0 {
		return nil, fmt.Errorf("session ttl must be positive")
	}
	return &AuthService{users: users, identities: identities, sessions: sessions, admins: admins, now: time.Now, sessionTTL: sessionTTL}, nil
}

func (s *AuthService) CreateSession(ctx context.Context, userID, userAgent, ipAddress string) (string, models.Session, error) {
	if userID == "" {
		return "", models.Session{}, fmt.Errorf("user id is required")
	}
	now := s.now()
	raw, err := auth.NewToken("cds", 32)
	if err != nil {
		return "", models.Session{}, err
	}
	session := models.Session{UserID: userID, TokenHash: auth.HashToken(raw), UserAgent: userAgent, IPAddress: ipAddress, ExpiresAt: now.Add(s.sessionTTL), LastSeenAt: &now}
	if err := s.sessions.Create(ctx, &session); err != nil {
		return "", models.Session{}, fmt.Errorf("create auth session: %w", err)
	}
	return raw, session, nil
}

func (s *AuthService) AuthenticateSession(ctx context.Context, raw string) (models.Session, error) {
	if strings.TrimSpace(raw) == "" {
		return models.Session{}, fmt.Errorf("session token is required")
	}
	now := s.now()
	session, err := s.sessions.FindActive(ctx, auth.HashToken(raw), now)
	if err != nil {
		return models.Session{}, fmt.Errorf("find auth session: %w", err)
	}
	user, err := s.users.FindByID(ctx, session.UserID)
	if err != nil {
		return models.Session{}, fmt.Errorf("find session user: %w", err)
	}
	if user.Status == models.UserStatusDisabled {
		return models.Session{}, fmt.Errorf("user account is disabled")
	}
	if err := s.sessions.Touch(ctx, session.ID, now); err != nil {
		return models.Session{}, fmt.Errorf("touch auth session: %w", err)
	}
	session.LastSeenAt = &now
	return session, nil
}

func (s *AuthService) RevokeSession(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("session id is required")
	}
	if err := s.sessions.Revoke(ctx, sessionID, s.now()); err != nil {
		return fmt.Errorf("revoke auth session: %w", err)
	}
	return nil
}

func (s *AuthService) FindOrCreateOAuthUser(ctx context.Context, profile auth.OAuthProfile) (models.User, bool, error) {
	if err := auth.ValidateOAuthProfile(profile); err != nil {
		return models.User{}, false, err
	}
	identity, err := s.identities.Find(ctx, profile.Provider, profile.Subject)
	if err == nil {
		user, findErr := s.users.FindByID(ctx, identity.UserID)
		if findErr != nil {
			return models.User{}, false, fmt.Errorf("find oauth user: %w", findErr)
		}
		return user, false, nil
	}
	if err != repositories.ErrNotFound {
		return models.User{}, false, fmt.Errorf("find oauth identity: %w", err)
	}
	user, err := s.users.FindByEmail(ctx, strings.ToLower(strings.TrimSpace(profile.Email)))
	created := false
	if err == repositories.ErrNotFound {
		name := strings.TrimSpace(profile.Name)
		if name == "" {
			name = strings.TrimSpace(strings.Split(profile.Email, "@")[0])
		}
		user = models.User{Email: strings.ToLower(strings.TrimSpace(profile.Email)), Name: name, Status: models.UserStatusActive}
		if profile.EmailVerified {
			now := s.now()
			user.EmailVerifiedAt = &now
		}
		if err := s.users.Create(ctx, &user); err != nil {
			return models.User{}, false, fmt.Errorf("create oauth user: %w", err)
		}
		created = true
	} else if err != nil {
		return models.User{}, false, fmt.Errorf("find oauth email: %w", err)
	}
	accessToken, refreshToken := profile.AccessToken, profile.RefreshToken
	if s.protector != nil {
		if accessToken != "" {
			accessToken, err = s.protector.Seal(accessToken)
			if err != nil {
				return models.User{}, false, fmt.Errorf("encrypt oauth access token: %w", err)
			}
		}
		if refreshToken != "" {
			refreshToken, err = s.protector.Seal(refreshToken)
			if err != nil {
				return models.User{}, false, fmt.Errorf("encrypt oauth refresh token: %w", err)
			}
		}
	}
	identity = models.OAuthIdentity{UserID: user.ID, Provider: profile.Provider, Subject: profile.Subject, Email: profile.Email, AccessToken: accessToken, RefreshToken: refreshToken, TokenExpiresAt: profile.TokenExpiresAt}
	if err := s.identities.Save(ctx, &identity); err != nil {
		return models.User{}, false, fmt.Errorf("save oauth identity: %w", err)
	}
	if err := s.users.UpdateLastLogin(ctx, user.ID, s.now()); err != nil {
		return models.User{}, false, fmt.Errorf("update oauth login: %w", err)
	}
	return user, created, nil
}
