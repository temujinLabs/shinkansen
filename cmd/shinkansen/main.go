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
	versionFlag := flag.Bool("version", false, "Print version")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("shinkansen %s\n", version)
		os.Exit(0)
	}

	if len(os.Args) > 1 && os.Args[1] == "login" {
		loginCmd.Parse(os.Args[2:])
		if err := runLogin(); err != nil {
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

	client := jira.NewClient(cfg.JiraURL, cfg.Email, cfg.APIToken)

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
		JiraURL:   jiraURL,
		Email:     email,
		APIToken:  token,
		AccountID: user.AccountID,
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
