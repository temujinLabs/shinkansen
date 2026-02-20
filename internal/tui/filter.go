package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/temujinlabs/shinkansen/internal/cache"
)

// FilterView handles the JQL filter input with history support.
type FilterView struct {
	visible     bool
	query       string
	history     []string // recent JQL filters from cache
	historyCur  int      // cursor position in history list (-1 = input field)
	browsingHis bool     // true when navigating history with up/down
	store       *cache.Store
	errMsg      string
}

func NewFilterView(store *cache.Store) FilterView {
	return FilterView{
		store:      store,
		historyCur: -1,
	}
}

// Show opens the filter input and loads history from cache.
func (fv *FilterView) Show() {
	fv.visible = true
	fv.query = ""
	fv.historyCur = -1
	fv.browsingHis = false
	fv.errMsg = ""

	// Load filter history
	if fv.store != nil {
		filters, err := fv.store.GetJQLFilters()
		if err == nil {
			fv.history = filters
		}
	}
}

// Hide closes the filter view.
func (fv *FilterView) Hide() {
	fv.visible = false
}

func (fv FilterView) Update(msg tea.Msg, app *App) (FilterView, tea.Cmd) {
	if !fv.visible {
		return fv, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			fv.Hide()
			app.currentView = viewIssues
			// Reload from cache (undo any filter)
			app.loadFromCache()
			return fv, nil

		case "enter":
			jql := fv.query
			if fv.browsingHis && fv.historyCur >= 0 && fv.historyCur < len(fv.history) {
				jql = fv.history[fv.historyCur]
			}
			if jql == "" {
				return fv, nil
			}

			// Save to history
			if fv.store != nil {
				fv.store.SaveJQLFilter(jql)
			}

			fv.Hide()
			app.currentView = viewIssues
			app.flashMsg = "Filtering..."

			query := jql
			return fv, func() tea.Msg {
				issues, err := app.client.SearchAll(query)
				if err != nil {
					return statusMsg(fmt.Sprintf("Filter failed: %v", err))
				}
				return filterAppliedMsg{issues: issues}
			}

		case "up":
			// Browse history
			if len(fv.history) > 0 {
				fv.browsingHis = true
				if fv.historyCur < len(fv.history)-1 {
					fv.historyCur++
				}
			}
			return fv, nil

		case "down":
			if fv.browsingHis {
				if fv.historyCur > 0 {
					fv.historyCur--
				} else {
					fv.browsingHis = false
					fv.historyCur = -1
				}
			}
			return fv, nil

		case "backspace":
			if !fv.browsingHis && len(fv.query) > 0 {
				fv.query = fv.query[:len(fv.query)-1]
			}
			return fv, nil

		default:
			ch := msg.String()
			if len(ch) == 1 || ch == " " {
				fv.browsingHis = false
				fv.historyCur = -1
				fv.query += ch
			}
			return fv, nil
		}
	}
	return fv, nil
}

func (fv FilterView) View(width, height int) string {
	if !fv.visible {
		return ""
	}

	var lines []string
	lines = append(lines, detailHeaderStyle.Render("JQL Filter"))
	lines = append(lines, "")

	// Input field
	prompt := searchPromptStyle.Render("JQL: ")
	inputVal := fv.query
	if !fv.browsingHis {
		inputVal += "\u2588"
	}
	lines = append(lines, prompt+inputVal)
	lines = append(lines, "")

	// History section
	if len(fv.history) > 0 {
		lines = append(lines, helpDescStyle.Render("  Recent filters (Up/Down to browse):"))
		for i, h := range fv.history {
			display := h
			if len(display) > width-12 {
				display = display[:width-15] + "..."
			}
			line := fmt.Sprintf("    %s", display)
			if fv.browsingHis && i == fv.historyCur {
				line = selectedStyle.Width(width - 8).Render(line)
			}
			lines = append(lines, line)
		}
	}

	// Error message
	if fv.errMsg != "" {
		lines = append(lines, "")
		lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("#cc3333")).Bold(true).Render("  "+fv.errMsg))
	}

	lines = append(lines, "")
	lines = append(lines, helpDescStyle.Render("  Enter: apply filter  Up/Down: browse history  Esc: cancel (reload all)"))

	content := strings.Join(lines, "\n")
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center,
		panelStyle.Width(min(width-4, 90)).Render(content),
	)
}
