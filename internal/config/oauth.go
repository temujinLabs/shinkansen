package config

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	atlassianAuthURL  = "https://auth.atlassian.com/authorize"
	atlassianTokenURL = "https://auth.atlassian.com/oauth/token"
	resourcesURL      = "https://api.atlassian.com/oauth/token/accessible-resources"
	callbackPort      = "8089"
	redirectURI       = "http://localhost:8089/callback"
	scopes            = "read:jira-work write:jira-work read:jira-user offline_access"
)

// OAuthTokenResponse is the response from the Atlassian token endpoint.
type OAuthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"` // seconds
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
}

// CloudResource represents an accessible Jira site.
type CloudResource struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

// OAuthFlow runs the Jira Cloud OAuth 2.0 (3LO) authorization flow.
// It starts a local HTTP server, opens the browser for authorization,
// exchanges the code for tokens, and returns the updated config.
func OAuthFlow(clientID, clientSecret string) (*Config, error) {
	state, err := randomState()
	if err != nil {
		return nil, fmt.Errorf("generate state: %w", err)
	}

	// Channel to receive the authorization code
	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	// Start local HTTP server for callback
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			errCh <- fmt.Errorf("invalid state parameter")
			http.Error(w, "Invalid state", http.StatusBadRequest)
			return
		}
		if errMsg := r.URL.Query().Get("error"); errMsg != "" {
			errCh <- fmt.Errorf("authorization denied: %s", errMsg)
			fmt.Fprintf(w, "<html><body><h2>Authorization denied</h2><p>%s</p><p>You can close this tab.</p></body></html>", errMsg)
			return
		}
		code := r.URL.Query().Get("code")
		if code == "" {
			errCh <- fmt.Errorf("no authorization code received")
			http.Error(w, "No code", http.StatusBadRequest)
			return
		}
		codeCh <- code
		fmt.Fprint(w, `<html><body style="font-family:system-ui;text-align:center;padding:60px;">
			<h2 style="color:#f0c232;">&#10003; Authorized!</h2>
			<p>You can close this tab and return to Shinkansen.</p>
		</body></html>`)
	})

	listener, err := net.Listen("tcp", ":"+callbackPort)
	if err != nil {
		return nil, fmt.Errorf("start callback server: %w", err)
	}
	server := &http.Server{Handler: mux}
	go server.Serve(listener)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()

	// Build authorization URL
	authURL := fmt.Sprintf("%s?audience=api.atlassian.com&client_id=%s&scope=%s&redirect_uri=%s&state=%s&response_type=code&prompt=consent",
		atlassianAuthURL,
		url.QueryEscape(clientID),
		url.QueryEscape(scopes),
		url.QueryEscape(redirectURI),
		url.QueryEscape(state),
	)

	fmt.Println("\nOpening browser for Jira authorization...")
	fmt.Printf("If the browser doesn't open, visit:\n%s\n\n", authURL)

	// Wait for code or error
	var code string
	select {
	case code = <-codeCh:
	case err := <-errCh:
		return nil, err
	case <-time.After(5 * time.Minute):
		return nil, fmt.Errorf("authorization timed out (5 minutes)")
	}

	fmt.Println("Authorization received. Exchanging for tokens...")

	// Exchange code for tokens
	tokenResp, err := exchangeCode(clientID, clientSecret, code)
	if err != nil {
		return nil, fmt.Errorf("token exchange: %w", err)
	}

	// Get accessible resources (cloud ID)
	resources, err := getAccessibleResources(tokenResp.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("get resources: %w", err)
	}

	if len(resources) == 0 {
		return nil, fmt.Errorf("no accessible Jira sites found")
	}

	// Use first resource (most users have one site)
	resource := resources[0]
	if len(resources) > 1 {
		fmt.Printf("Found %d Jira sites. Using: %s (%s)\n", len(resources), resource.Name, resource.URL)
	}

	expiry := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	cfg := &Config{
		JiraURL:       resource.URL,
		AuthMethod:    "oauth",
		OAuthClientID: clientID,
		OAuthSecret:   clientSecret,
		AccessToken:   tokenResp.AccessToken,
		RefreshToken:  tokenResp.RefreshToken,
		CloudID:       resource.ID,
		TokenExpiry:   expiry.Format(time.RFC3339),
		SyncInterval:  60,
	}

	return cfg, nil
}

// RefreshAccessToken refreshes an expired OAuth access token.
func RefreshAccessToken(cfg *Config) error {
	if cfg.RefreshToken == "" {
		return fmt.Errorf("no refresh token available")
	}

	form := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {cfg.OAuthClientID},
		"client_secret": {cfg.OAuthSecret},
		"refresh_token": {cfg.RefreshToken},
	}

	resp, err := http.Post(atlassianTokenURL, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("refresh request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("refresh failed (%d): %s", resp.StatusCode, string(body))
	}

	var tokenResp OAuthTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return fmt.Errorf("parse refresh response: %w", err)
	}

	cfg.AccessToken = tokenResp.AccessToken
	if tokenResp.RefreshToken != "" {
		cfg.RefreshToken = tokenResp.RefreshToken
	}
	cfg.TokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second).Format(time.RFC3339)

	return Save(cfg)
}

func exchangeCode(clientID, clientSecret, code string) (*OAuthTokenResponse, error) {
	form := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"code":          {code},
		"redirect_uri":  {redirectURI},
	}

	resp, err := http.Post(atlassianTokenURL, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("token exchange failed (%d): %s", resp.StatusCode, string(body))
	}

	var tokenResp OAuthTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, err
	}
	return &tokenResp, nil
}

func getAccessibleResources(accessToken string) ([]CloudResource, error) {
	req, err := http.NewRequest("GET", resourcesURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("resources request failed (%d): %s", resp.StatusCode, string(body))
	}

	var resources []CloudResource
	if err := json.Unmarshal(body, &resources); err != nil {
		return nil, err
	}
	return resources, nil
}

func randomState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
