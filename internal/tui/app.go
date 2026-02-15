package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/temujinlabs/shinkansen/internal/cache"
	"github.com/temujinlabs/shinkansen/internal/config"
	"github.com/temujinlabs/shinkansen/internal/jira"
)

type view int

const (
	viewIssues view = iota
	viewBoard
	viewDetail
	viewSearch
	viewHelp
)

// Messages
type syncDoneMsg struct{ result cache.SyncResult }
type tickMsg time.Time

type App struct {
	client *jira.Client
	store  *cache.Store
	cfg    *config.Config

	currentView view
	activePanel int // 0 = issues, 1 = board

	issues    IssueList
	board     BoardView
	detail    DetailView
	search    SearchView
	showHelp  bool

	width  int
	height int

	syncStatus string
	lastSync   time.Time
	syncing    bool
}

func NewApp(client *jira.Client, store *cache.Store, cfg *config.Config) *App {
	return &App{
		client:      client,
		store:       store,
		cfg:         cfg,
		currentView: viewIssues,
		issues:      NewIssueList(),
		board:       NewBoardView(),
		detail:      NewDetailView(),
		search:      NewSearchView(),
		syncStatus:  "Loading...",
	}
}

func (a *App) Init() tea.Cmd {
	return tea.Batch(
		a.doSync,
		a.tickCmd(),
	)
}

func (a *App) tickCmd() tea.Cmd {
	interval := 60
	if a.cfg.SyncInterval > 0 {
		interval = a.cfg.SyncInterval
	}
	return tea.Tick(time.Duration(interval)*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (a *App) doSync() tea.Msg {
	result := cache.Sync(a.client, a.store)
	return syncDoneMsg{result: result}
}

func (a *App) loadFromCache() {
	issues, err := a.store.GetAllIssues()
	if err != nil {
		return
	}
	a.issues.SetIssues(issues)
	a.board.SetIssues(issues)
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		return a, nil

	case syncDoneMsg:
		a.syncing = false
		if msg.result.Err != nil {
			a.syncStatus = fmt.Sprintf("Sync failed: %v", msg.result.Err)
			// Load from cache anyway
			a.loadFromCache()
		} else {
			a.syncStatus = fmt.Sprintf("Synced %d issues in %dms", msg.result.ItemsSynced, msg.result.Duration.Milliseconds())
			a.lastSync = time.Now()
			a.loadFromCache()
		}
		return a, nil

	case tickMsg:
		a.syncing = true
		a.syncStatus = "Syncing..."
		return a, tea.Batch(a.doSync, a.tickCmd())

	case tea.KeyMsg:
		// Global keys
		switch msg.String() {
		case "q", "ctrl+c":
			if a.currentView == viewDetail {
				a.currentView = viewIssues
				return a, nil
			}
			if a.currentView == viewSearch {
				a.currentView = viewIssues
				return a, nil
			}
			if a.showHelp {
				a.showHelp = false
				return a, nil
			}
			return a, tea.Quit

		case "?":
			if a.currentView != viewSearch {
				a.showHelp = !a.showHelp
				return a, nil
			}

		case "r":
			if a.currentView != viewSearch && a.currentView != viewDetail {
				a.syncing = true
				a.syncStatus = "Syncing..."
				return a, a.doSync
			}

		case "/":
			if a.currentView != viewDetail {
				a.currentView = viewSearch
				a.search.Reset()
				return a, nil
			}

		case "h":
			if a.currentView == viewIssues || a.currentView == viewBoard {
				if a.activePanel == 1 {
					a.activePanel = 0
					a.currentView = viewIssues
				}
				return a, nil
			}

		case "l":
			if a.currentView == viewIssues || a.currentView == viewBoard {
				if a.activePanel == 0 {
					a.activePanel = 1
					a.currentView = viewBoard
				}
				return a, nil
			}
		}

		if a.showHelp {
			return a, nil
		}
	}

	// Delegate to active view
	var cmd tea.Cmd
	switch a.currentView {
	case viewIssues:
		a.issues, cmd = a.issues.Update(msg, a)
	case viewBoard:
		a.board, cmd = a.board.Update(msg, a)
	case viewDetail:
		a.detail, cmd = a.detail.Update(msg, a)
	case viewSearch:
		a.search, cmd = a.search.Update(msg, a)
	}
	return a, cmd
}

func (a *App) View() string {
	if a.width == 0 {
		return "Loading..."
	}

	if a.showHelp {
		return a.renderHelp()
	}

	header := titleStyle.Render("SHINKANSEN") + "  " +
		statusBarStyle.Render(a.syncStatus)

	var content string
	switch a.currentView {
	case viewDetail:
		content = a.detail.View(a.width, a.height-4)
	case viewSearch:
		content = a.search.View(a.width, a.height-4)
	default:
		// Side-by-side: issues | board
		halfWidth := a.width/2 - 2
		contentHeight := a.height - 4

		leftPanel := a.issues.View(halfWidth, contentHeight, a.activePanel == 0)
		rightPanel := a.board.View(halfWidth, contentHeight, a.activePanel == 1)
		content = lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
	}

	footer := a.renderFooter()

	return lipgloss.JoinVertical(lipgloss.Left, header, content, footer)
}

func (a *App) renderFooter() string {
	keys := "j/k:nav  h/l:panels  enter:detail  m:move  c:comment  n:new  /:search  r:sync  ?:help  q:quit"
	return statusBarStyle.Render(keys)
}

func (a *App) renderHelp() string {
	help := lipgloss.JoinVertical(lipgloss.Left,
		detailHeaderStyle.Render("Shinkansen — Keyboard Shortcuts"),
		"",
		helpKeyStyle.Render("j/k    ")+" "+helpDescStyle.Render("Navigate up/down"),
		helpKeyStyle.Render("h/l    ")+" "+helpDescStyle.Render("Switch between panels"),
		helpKeyStyle.Render("Enter  ")+" "+helpDescStyle.Render("Open issue detail"),
		helpKeyStyle.Render("m      ")+" "+helpDescStyle.Render("Move issue (status transition)"),
		helpKeyStyle.Render("c      ")+" "+helpDescStyle.Render("Add comment"),
		helpKeyStyle.Render("n      ")+" "+helpDescStyle.Render("Create new issue"),
		helpKeyStyle.Render("/      ")+" "+helpDescStyle.Render("Fuzzy search"),
		helpKeyStyle.Render("r      ")+" "+helpDescStyle.Render("Refresh / sync from Jira"),
		helpKeyStyle.Render("?      ")+" "+helpDescStyle.Render("Toggle this help"),
		helpKeyStyle.Render("q      ")+" "+helpDescStyle.Render("Quit"),
		"",
		helpDescStyle.Render("Press any key to close"),
	)
	return lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center, help)
}
