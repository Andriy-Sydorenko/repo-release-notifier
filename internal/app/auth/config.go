package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/Andriy-Sydorenko/repo-release-notifier/internal/shared/config"
)

const minJWTSecretLen = 32

type OAuthProvider struct {
	ClientID     string
	ClientSecret string
}

func (p OAuthProvider) Enabled() bool { return p.ClientID != "" && p.ClientSecret != "" }

type Config struct {
	JWTSecret           string
	SessionTTL          time.Duration
	Secure              bool
	Google              OAuthProvider
	GitHub              OAuthProvider
	RedirectBase        string
	AccessRequestEmail  string
	EnforceTokenVersion bool
}

func LoadConfig() Config {
	return Config{
		JWTSecret:           config.GetEnvOrDefault("JWT_SECRET", ""),
		SessionTTL:          config.GetEnvDuration("SESSION_TTL", 24*time.Hour),
		Secure:              config.GetEnvOrDefault("COOKIE_SECURE", "true") == "true",
		Google:              OAuthProvider{config.GetEnvOrDefault("GOOGLE_CLIENT_ID", ""), config.GetEnvOrDefault("GOOGLE_CLIENT_SECRET", "")},
		GitHub:              OAuthProvider{config.GetEnvOrDefault("GITHUB_OAUTH_CLIENT_ID", ""), config.GetEnvOrDefault("GITHUB_OAUTH_CLIENT_SECRET", "")},
		RedirectBase:        config.GetEnvOrDefault("OAUTH_REDIRECT_BASE", config.GetEnvOrDefault("BASE_URL", "")),
		AccessRequestEmail:  config.GetEnvOrDefault("ACCESS_REQUEST_EMAIL", ""),
		EnforceTokenVersion: config.GetEnvOrDefault("ENFORCE_TOKEN_VERSION", "false") == "true",
	}
}

func (c *Config) Validate() error {
	var errs []error
	if len(c.JWTSecret) < minJWTSecretLen {
		errs = append(errs, fmt.Errorf("JWT_SECRET must be at least %d chars", minJWTSecretLen))
	}
	if c.SessionTTL <= 0 {
		errs = append(errs, errors.New("SESSION_TTL must be positive"))
	}
	return errors.Join(errs...)
}
