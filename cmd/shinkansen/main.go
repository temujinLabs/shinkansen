package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/temujinlabs/shinkansen/internal/cache"
	"github.com/temujinlabs/shinkansen/internal/config"
	"github.com/temujinlabs/shinkansen/internal/jira"
	"github.com/temujinlabs/shinkansen/internal/tui"
)

var version = "dev"

func main() {
	loginCmd := flag.NewFlagSet("login", flag.ExitOnError)
	oauthFlag := loginCmd.Bool("oauth", false, "Use OAuth 2.0 (3LO) instead of API token")
	clientIDFlag := loginCmd.String("client-id", "", "OAuth client ID (from developer.atlassian.com)")
	clientSecretFlag := loginCmd.String("client-secret", "", "OAuth client secret")

	versionFlag := flag.Bool("version", false, "Print version")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("shinkansen %s\n", version)
		os.Exit(0)
	}

	if len(os.Args) > 1 && os.Args[1] == "login" {
		loginCmd.Parse(os.Args[2:])
		var err error
		if *oauthFlag {
			err = runOAuthLogin(*clientIDFlag, *clientSecretFlag)
		} else {
			err = runLogin()
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Login failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	cfg, err := config.Load()
	if err != nil || cfg.JiraURL == "" {
		fmt.Println("Not configured. Run 'shinkansen login' first.")
		os.Exit(1)
	}

	// Use config-aware client (supports both API token and OAuth)
	client := jira.NewClientFromConfig(cfg)

	store, err := cache.NewStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cache error: %v\n", err)
		os.Exit(1)
	}
	defer store.Close()

	app := tui.NewApp(client, store, cfg)
	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		cmd = exec.Command("open", url)
	}
	cmd.Start()
}

func runLogin() error {
	var jiraURL, email, token string

	fmt.Print("Jira URL (e.g. https://yourorg.atlassian.net): ")
	fmt.Scanln(&jiraURL)

	fmt.Print("Email: ")
	fmt.Scanln(&email)

	tokenURL := "https://id.atlassian.com/manage-profile/security/api-tokens"
	fmt.Printf("Opening %s in your browser...\n", tokenURL)
	openBrowser(tokenURL)
	fmt.Print("API Token (paste from browser): ")
	fmt.Scanln(&token)

	if jiraURL == "" || email == "" || token == "" {
		return fmt.Errorf("all fields are required")
	}

	// Verify credentials
	client := jira.NewClient(jiraURL, email, token)
	user, err := client.GetMyself()
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}
	fmt.Printf("Authenticated as: %s (%s)\n", user.DisplayName, user.EmailAddress)

	// Detect projects
	projects, err := client.GetProjects()
	if err != nil {
		fmt.Printf("Warning: could not fetch projects: %v\n", err)
	} else {
		fmt.Printf("Found %d projects\n", len(projects))
	}

	cfg := &config.Config{
		JiraURL:    jiraURL,
		Email:      email,
		APIToken:   token,
		AccountID:  user.AccountID,
		AuthMethod: "api-token",
	}
	if len(projects) > 0 {
		cfg.DefaultProject = projects[0].Key
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println("Configuration saved. Run 'shinkansen' to start.")
	return nil
}

func runOAuthLogin(clientID, clientSecret string) error {
	if clientID == "" {
		clientID = os.Getenv("SHINKANSEN_OAUTH_CLIENT_ID")
	}
	if clientSecret == "" {
		clientSecret = os.Getenv("SHINKANSEN_OAUTH_SECRET")
	}

	if clientID == "" {
		fmt.Print("OAuth Client ID (from developer.atlassian.com): ")
		fmt.Scanln(&clientID)
	}
	if clientSecret == "" {
		fmt.Print("OAuth Client Secret: ")
		fmt.Scanln(&clientSecret)
	}

	if clientID == "" || clientSecret == "" {
		return fmt.Errorf("client ID and secret are required")
	}

	fmt.Println("Starting OAuth 2.0 authorization flow...")
	fmt.Println("A browser window will open for you to authorize Shinkansen.")

	// Open browser for OAuth authorization
	authURL := fmt.Sprintf("https://auth.atlassian.com/authorize?audience=api.atlassian.com&client_id=%s&scope=%s&redirect_uri=%s&response_type=code&prompt=consent",
		clientID,
		"read%3Ajira-work+write%3Ajira-work+read%3Ajira-user+offline_access",
		"http%3A%2F%2Flocalhost%3A8089%2Fcallback",
	)
	openBrowser(authURL)

	cfg, err := config.OAuthFlow(clientID, clientSecret)
	if err != nil {
		return err
	}

	// Verify by fetching user info
	client := jira.NewClientFromConfig(cfg)
	user, err := client.GetMyself()
	if err != nil {
		return fmt.Errorf("authentication verification failed: %w", err)
	}
	fmt.Printf("Authenticated as: %s (%s)\n", user.DisplayName, user.EmailAddress)
	cfg.AccountID = user.AccountID

	// Detect projects
	projects, err := client.GetProjects()
	if err != nil {
		fmt.Printf("Warning: could not fetch projects: %v\n", err)
	} else {
		fmt.Printf("Found %d projects\n", len(projects))
		if len(projects) > 0 {
			cfg.DefaultProject = projects[0].Key
		}
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println("OAuth configuration saved. Run 'shinkansen' to start.")
	return nil
}
