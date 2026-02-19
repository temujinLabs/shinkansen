package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	JiraURL        string `json:"jira_url"`
	Email          string `json:"email"`
	APIToken       string `json:"api_token"`
	AccountID      string `json:"account_id,omitempty"`
	DefaultProject string `json:"default_project,omitempty"`
	DefaultBoard   int    `json:"default_board,omitempty"`
	SyncInterval   int    `json:"sync_interval,omitempty"` // seconds, default 60

	// OAuth 2.0 (3LO) fields
	AuthMethod    string `json:"auth_method,omitempty"`     // "api-token" or "oauth"
	OAuthClientID string `json:"oauth_client_id,omitempty"` // from developer.atlassian.com
	OAuthSecret   string `json:"oauth_secret,omitempty"`
	AccessToken   string `json:"access_token,omitempty"`
	RefreshToken  string `json:"refresh_token,omitempty"`
	CloudID       string `json:"cloud_id,omitempty"`
	TokenExpiry   string `json:"token_expiry,omitempty"` // RFC3339
}

// IsOAuth returns true if the config uses OAuth authentication.
func (c *Config) IsOAuth() bool {
	return c.AuthMethod == "oauth" && c.AccessToken != ""
}

// OAuthBaseURL returns the Atlassian API base URL for OAuth access.
func (c *Config) OAuthBaseURL() string {
	if c.CloudID != "" {
		return "https://api.atlassian.com/ex/jira/" + c.CloudID
	}
	return c.JiraURL
}

// BrowseURL returns the Jira web URL for an issue key.
func (c *Config) BrowseURL(issueKey string) string {
	return c.JiraURL + "/browse/" + issueKey
}

func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "shinkansen"), nil
}

func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{SyncInterval: 60}, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if cfg.SyncInterval == 0 {
		cfg.SyncInterval = 60
	}
	return &cfg, nil
}

func Save(cfg *Config) error {
	dir, err := configDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	path, err := configPath()
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
