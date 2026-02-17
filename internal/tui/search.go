package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/temujinlabs/shinkansen/internal/cache"
	"github.com/temujinlabs/shinkansen/internal/jira"
)

type SearchView struct {
	query    string
	results  []jira.Issue
	cursor   int
	creating bool // true when in "new issue" mode
}

func NewSearchView() SearchView {
	return SearchView{}
}

func (sv *SearchView) Reset() {
	sv.query = ""
	sv.results = nil
	sv.cursor = 0
	sv.creating = false
}

func (sv *SearchView) StartCreate() {
	sv.Reset()
	sv.creating = true
}

func (sv SearchView) Update(msg tea.Msg, app *App) (SearchView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			app.currentView = viewIssues
			sv.Reset()
			return sv, nil

		case "enter":
			if sv.creating {
				if sv.query != "" {
					summary := sv.query
					projectKey := app.cfg.DefaultProject
					sv.Reset()
					app.currentView = viewIssues
					return sv, func() tea.Msg {
						app.client.CreateIssue(projectKey, summary, "Task")
						return syncDoneMsg{result: cache.Sync(app.client, app.store, app.cfg.DefaultProject)}
					}
				}
				return sv, nil
			}

			// Select a search result
			if len(sv.results) > 0 && sv.cursor < len(sv.results) {
				issue := sv.results[sv.cursor]
				app.detail.SetIssue(&issue)
				app.currentView = viewDetail
				sv.Reset()
			}
			return sv, nil

		case "down", "ctrl+n":
			if sv.cursor < len(sv.results)-1 {
				sv.cursor++
			}
		case "up", "ctrl+p":
			if sv.cursor > 0 {
				sv.cursor--
			}

		case "backspace":
			if len(sv.query) > 0 {
				sv.query = sv.query[:len(sv.query)-1]
				sv.doSearch(app)
			}

		default:
			if len(msg.String()) == 1 || msg.String() == " " {
				sv.query += msg.String()
				if !sv.creating {
					sv.doSearch(app)
				}
			}
		}
	}
	return sv, nil
}

func (sv *SearchView) doSearch(app *App) {
	if sv.query == "" {
		sv.results = nil
		sv.cursor = 0
		return
	}

	issues, err := app.store.SearchIssues(sv.query)
	if err != nil {
		sv.results = nil
		return
	}
	sv.results = issues
	sv.cursor = 0
}

func (sv SearchView) View(width, height int) string {
	var lines []string

	if sv.creating {
		lines = append(lines, searchPromptStyle.Render("New Issue Title: ")+sv.query+"█")
		lines = append(lines, "")
		lines = append(lines, helpDescStyle.Render("Enter: create  Esc: cancel"))
	} else {
		lines = append(lines, searchPromptStyle.Render("Search: ")+sv.query+"█")
		lines = append(lines, "")

		if len(sv.results) == 0 && sv.query != "" {
			lines = append(lines, helpDescStyle.Render("No results"))
		}

		maxResults := height - 6
		for i, issue := range sv.results {
			if i >= maxResults {
				lines = append(lines, helpDescStyle.Render(fmt.Sprintf("  +%d more", len(sv.results)-maxResults)))
				break
			}

			line := fmt.Sprintf("  %s  %s  [%s]", issue.Key, issue.Fields.Summary, issue.Fields.Status.Name)
			if len(line) > width-4 {
				line = line[:width-7] + "..."
			}
			if i == sv.cursor {
				line = selectedStyle.Width(width - 4).Render(line)
			}
			lines = append(lines, line)
		}

		lines = append(lines, "")
		lines = append(lines, helpDescStyle.Render("Enter: select  Esc: cancel"))
	}

	content := strings.Join(lines, "\n")
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Top,
		panelStyle.Width(width-4).Render(content),
	)
}
