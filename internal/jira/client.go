package jira

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/temujinlabs/shinkansen/internal/config"
)

type Client struct {
	cfg        *config.Config
	baseURL    string
	email      string
	token      string
	httpClient *http.Client
}

func NewClient(baseURL, email, token string) *Client {
	return &Client{
		baseURL: baseURL,
		email:   email,
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewClientFromConfig creates a client that uses the config for auth,
// supporting both API token and OAuth authentication.
func NewClientFromConfig(cfg *config.Config) *Client {
	baseURL := cfg.JiraURL
	if cfg.IsOAuth() {
		baseURL = cfg.OAuthBaseURL()
	}
	return &Client{
		cfg:     cfg,
		baseURL: baseURL,
		email:   cfg.Email,
		token:   cfg.APIToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) do(method, path string, body interface{}) ([]byte, error) {
	// Auto-refresh OAuth token if expired
	if c.cfg != nil && c.cfg.IsOAuth() && c.cfg.TokenExpired() {
		if err := config.RefreshAccessToken(c.cfg); err != nil {
			return nil, fmt.Errorf("token refresh: %w", err)
		}
		c.baseURL = c.cfg.OAuthBaseURL()
	}

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	url := c.baseURL + path
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	if c.cfg != nil {
		req.Header.Set("Authorization", c.cfg.AuthHeader())
	} else {
		req.Header.Set("Authorization", config.BasicAuthHeader(c.email, c.token))
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func (c *Client) GetMyself() (*User, error) {
	data, err := c.do("GET", "/rest/api/3/myself", nil)
	if err != nil {
		return nil, err
	}
	var user User
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, fmt.Errorf("parse user: %w", err)
	}
	return &user, nil
}

func (c *Client) GetProjects() ([]Project, error) {
	data, err := c.do("GET", "/rest/api/3/project", nil)
	if err != nil {
		return nil, err
	}
	var projects []Project
	if err := json.Unmarshal(data, &projects); err != nil {
		return nil, fmt.Errorf("parse projects: %w", err)
	}
	return projects, nil
}

func (c *Client) GetBoards() ([]Board, error) {
	data, err := c.do("GET", "/rest/agile/1.0/board", nil)
	if err != nil {
		return nil, err
	}
	var resp BoardsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parse boards: %w", err)
	}
	return resp.Values, nil
}
