package tui

import (
	"fmt"
	"os/exec"
	"runtime"
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
type assignDoneMsg struct{ issueKey string }
type logWorkDoneMsg struct{ issueKey string }
type bulkMoveDoneMsg struct{ count int }
type statusMsg string

type App struct {
	client *jira.Client
	store  *cache.Store
	cfg    *config.Config

	currentView view
	activePanel int // 0 = issues, 1 = board

	issues   IssueList
	board    BoardView
	detail   DetailView
	search   SearchView
	picker   TransitionPicker
	showHelp bool

	// Selections for bulk operations
	selections map[string]bool

	width  int
	height int

	syncStatus string
	lastSync   time.Time
	syncing    bool
	flashMsg   string // Temporary status message
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
		selections:  make(map[string]bool),
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
	result := cache.Sync(a.client, a.store, a.cfg.DefaultProject)
	return syncDoneMsg{result: result}
}

func (a *App) loadFromCache() {
	issues, err := a.store.GetAllIssues()
	if err != nil {
		return
	}
	a.issues.SetIssues(issues)
	a.board.SetIssues(issues)

	// Refresh the detail view if it's showing an issue
	if a.detail.issue != nil {
		for i := range issues {
			if issues[i].Key == a.detail.issue.Key {
				a.detail.issue = &issues[i]
				a.detail.commentSent = false
				break
			}
		}
	}
}

// isInputMode returns true when a text input is active and global keys should be suppressed.
func (a *App) isInputMode() bool {
	if a.currentView == viewSearch {
		return true
	}
	if a.currentView == viewDetail && (a.detail.commenting || a.detail.logging) {
		return true
	}
	return false
}

// selectionCount returns the number of selected issues.
func (a *App) selectionCount() int {
	count := 0
	for _, v := range a.selections {
		if v {
			count++
		}
	}
	return count
}

// toggleSelection toggles an issue's selection state.
func (a *App) toggleSelection(key string) {
	if a.selections[key] {
		delete(a.selections, key)
	} else {
		a.selections[key] = true
	}
}

// clearSelections removes all selections.
func (a *App) clearSelections() {
	a.selections = make(map[string]bool)
}

// openBrowser opens a URL in the default browser.
func openBrowser(url string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Start()
	case "linux":
		return exec.Command("xdg-open", url).Start()
	default:
		return exec.Command("open", url).Start()
	}
}

// doBulkTransition moves all selected issues to a given transition.
func (a *App) doBulkTransition(transitionID string) tea.Cmd {
	keys := make([]string, 0, len(a.selections))
	for k, v := range a.selections {
		if v {
			keys = append(keys, k)
		}
	}
	return func() tea.Msg {
		for _, key := range keys {
			a.client.TransitionIssue(key, transitionID)
		}
		return bulkMoveDoneMsg{count: len(keys)}
	}
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
			a.loadFromCache()
		} else {
			a.syncStatus = fmt.Sprintf("Synced %d issues in %dms", msg.result.ItemsSynced, msg.result.Duration.Milliseconds())
			a.lastSync = time.Now()
			a.loadFromCache()
		}
		return a, nil

	case transitionsMsg:
		a.picker.Show(msg.issueKey, msg.transitions)
		return a, nil

	case assignDoneMsg:
		a.flashMsg = fmt.Sprintf("Assigned %s to you", msg.issueKey)
		a.syncing = true
		return a, a.doSync

	case logWorkDoneMsg:
		a.flashMsg = fmt.Sprintf("Time logged on %s", msg.issueKey)
		return a, nil

	case bulkMoveDoneMsg:
		a.flashMsg = fmt.Sprintf("Moved %d issues", msg.count)
		a.clearSelections()
		a.syncing = true
		return a, a.doSync

	case statusMsg:
		a.flashMsg = string(msg)
		return a, nil

	case tickMsg:
		a.syncing = true
		a.syncStatus = "Syncing..."
		a.flashMsg = ""
		return a, tea.Batch(a.doSync, a.tickCmd())

	case tea.KeyMsg:
		// Transition picker captures all input when visible
		if a.picker.visible {
			var cmd tea.Cmd
			a.picker, cmd = a.picker.Update(msg, a)
			return a, cmd
		}

		// In input mode (search, commenting, logging), only ctrl+c quits
		if a.isInputMode() {
			if msg.String() == "ctrl+c" {
				return a, tea.Quit
			}
			var cmd tea.Cmd
			switch a.currentView {
			case viewSearch:
				a.search, cmd = a.search.Update(msg, a)
			case viewDetail:
				a.detail, cmd = a.detail.Update(msg, a)
			}
			return a, cmd
		}

		// Global keys (only active when NOT in input mode)
		switch msg.String() {
		case "q", "ctrl+c":
			if a.currentView == viewDetail {
				a.currentView = viewIssues
				return a, nil
			}
			if a.showHelp {
				a.showHelp = false
				return a, nil
			}
			return a, tea.Quit

		case "?":
			a.showHelp = !a.showHelp
			return a, nil

		case "r":
			if a.currentView != viewDetail {
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

		case "o":
			// Open in browser
			var issueKey string
			switch a.currentView {
			case viewIssues:
				if i := a.issues.SelectedIssue(); i != nil {
					issueKey = i.Key
				}
			case viewBoard:
				if i := a.board.SelectedIssue(); i != nil {
					issueKey = i.Key
				}
			case viewDetail:
				if a.detail.issue != nil {
					issueKey = a.detail.issue.Key
				}
			}
			if issueKey != "" {
				openBrowser(a.cfg.BrowseURL(issueKey))
				a.flashMsg = fmt.Sprintf("Opened %s in browser", issueKey)
			}
			return a, nil

		case "a":
			// Assign to self
			if a.cfg.AccountID == "" {
				a.flashMsg = "Run 'shinkansen login' to enable assign"
				return a, nil
			}
			var issueKey string
			switch a.currentView {
			case viewIssues:
				if i := a.issues.SelectedIssue(); i != nil {
					issueKey = i.Key
				}
			case viewBoard:
				if i := a.board.SelectedIssue(); i != nil {
					issueKey = i.Key
				}
			case viewDetail:
				if a.detail.issue != nil {
					issueKey = a.detail.issue.Key
				}
			}
			if issueKey != "" {
				key := issueKey
				accountID := a.cfg.AccountID
				a.flashMsg = fmt.Sprintf("Assigning %s...", key)
				return a, func() tea.Msg {
					a.client.AssignIssue(key, accountID)
					return assignDoneMsg{issueKey: key}
				}
			}
			return a, nil

		case " ":
			// Toggle selection for bulk operations
			if a.currentView == viewIssues {
				if i := a.issues.SelectedIssue(); i != nil {
					a.toggleSelection(i.Key)
				}
				return a, nil
			}
			if a.currentView == viewBoard {
				if i := a.board.SelectedIssue(); i != nil {
					a.toggleSelection(i.Key)
				}
				return a, nil
			}

		case "escape", "esc":
			// Clear selections if any
			if a.selectionCount() > 0 {
				a.clearSelections()
				return a, nil
			}

		case "left":
			if a.currentView == viewBoard {
				if a.board.colCursor > 0 {
					a.board.colCursor--
					a.board.rowCursor = 0
					a.board.rowOffset = 0
				} else {
					a.activePanel = 0
					a.currentView = viewIssues
				}
				return a, nil
			}
			if a.currentView == viewIssues {
				return a, nil
			}

		case "right":
			if a.currentView == viewIssues {
				a.activePanel = 1
				a.currentView = viewBoard
				return a, nil
			}
			if a.currentView == viewBoard {
				if a.board.colCursor < len(a.board.columns)-1 {
					a.board.colCursor++
					a.board.rowCursor = 0
					a.board.rowOffset = 0
				}
				return a, nil
			}
		}

		if a.showHelp {
			a.showHelp = false
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

	// Transition picker overlay
	if a.picker.visible {
		return a.picker.View(a.width, a.height)
	}

	if a.showHelp {
		return a.renderHelp()
	}

	// Build header with hints and status
	selCount := a.selectionCount()
	var hints string
	if selCount > 0 {
		hints = helpKeyStyle.Render(fmt.Sprintf("[%d selected]", selCount)) + "  " +
			helpDescStyle.Render("m:move all  space:toggle  esc:clear")
	} else {
		hints = helpDescStyle.Render("enter:open  o:browser  a:assign  m:move  t:log  ?:help")
	}

	status := a.syncStatus
	if a.flashMsg != "" {
		status = a.flashMsg
	}

	header := titleStyle.Render("SHINKANSEN") + "  " + hints + "  " +
		statusBarStyle.Render(status)

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

		leftPanel := a.issues.View(halfWidth, contentHeight, a.activePanel == 0, a.selections)
		rightPanel := a.board.View(halfWidth, contentHeight, a.activePanel == 1, a.selections)
		content = lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
	}

	return lipgloss.JoinVertical(lipgloss.Left, header, content)
}

func (a *App) renderHelp() string {
	help := lipgloss.JoinVertical(lipgloss.Left,
		detailHeaderStyle.Render("Shinkansen -- Keyboard Shortcuts"),
		"",
		helpKeyStyle.Render("↑/↓      ")+" "+helpDescStyle.Render("Navigate up/down"),
		helpKeyStyle.Render("←/→      ")+" "+helpDescStyle.Render("Switch between panels & columns"),
		helpKeyStyle.Render("Enter    ")+" "+helpDescStyle.Render("Open issue detail"),
		helpKeyStyle.Render("o        ")+" "+helpDescStyle.Render("Open issue in browser"),
		helpKeyStyle.Render("a        ")+" "+helpDescStyle.Render("Assign issue to yourself"),
		helpKeyStyle.Render("m        ")+" "+helpDescStyle.Render("Move issue (status transition)"),
		helpKeyStyle.Render("c        ")+" "+helpDescStyle.Render("Add comment"),
		helpKeyStyle.Render("t        ")+" "+helpDescStyle.Render("Log time (e.g. 2h, 30m)"),
		helpKeyStyle.Render("n        ")+" "+helpDescStyle.Render("Create new issue"),
		helpKeyStyle.Render("Space    ")+" "+helpDescStyle.Render("Select/deselect issue (bulk ops)"),
		helpKeyStyle.Render("/        ")+" "+helpDescStyle.Render("Fuzzy search"),
		helpKeyStyle.Render("r        ")+" "+helpDescStyle.Render("Refresh / sync from Jira"),
		helpKeyStyle.Render("?        ")+" "+helpDescStyle.Render("Toggle this help"),
		helpKeyStyle.Render("q        ")+" "+helpDescStyle.Render("Quit"),
		"",
		helpDescStyle.Render("Press any key to close"),
	)
	return lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center, help)
}
