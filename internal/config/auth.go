package config

import (
	"encoding/base64"
	"fmt"
	"time"
)

// BasicAuthHeader returns the Authorization header value for Jira API token auth.
func BasicAuthHeader(email, token string) string {
	credentials := fmt.Sprintf("%s:%s", email, token)
	encoded := base64.StdEncoding.EncodeToString([]byte(credentials))
	return fmt.Sprintf("Basic %s", encoded)
}

// BearerAuthHeader returns the Authorization header for OAuth Bearer tokens.
func BearerAuthHeader(accessToken string) string {
	return fmt.Sprintf("Bearer %s", accessToken)
}

// AuthHeader returns the appropriate Authorization header based on config.
func (c *Config) AuthHeader() string {
	if c.IsOAuth() {
		return BearerAuthHeader(c.AccessToken)
	}
	return BasicAuthHeader(c.Email, c.APIToken)
}

// TokenExpired returns true if the OAuth token has expired.
func (c *Config) TokenExpired() bool {
	if c.TokenExpiry == "" {
		return true
	}
	t, err := time.Parse(time.RFC3339, c.TokenExpiry)
	if err != nil {
		return true
	}
	return time.Now().After(t.Add(-60 * time.Second)) // 60s buffer
}
