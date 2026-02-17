package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/temujinlabs/shinkansen/internal/jira"
)

type IssueList struct {
	issues   []jira.Issue
	cursor   int
	offset   int
	maxVisible int
}

func NewIssueList() IssueList {
	return IssueList{}
}

func (il *IssueList) SetIssues(issues []jira.Issue) {
	il.issues = issues
	if il.cursor >= len(issues) {
		il.cursor = max(0, len(issues)-1)
	}
}

func (il *IssueList) SelectedIssue() *jira.Issue {
	if len(il.issues) == 0 || il.cursor >= len(il.issues) {
		return nil
	}
	return &il.issues[il.cursor]
}

func (il IssueList) Update(msg tea.Msg, app *App) (IssueList, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "down":
			if il.cursor < len(il.issues)-1 {
				il.cursor++
				if il.cursor-il.offset >= il.maxVisible {
					il.offset++
				}
			}
		case "up":
			if il.cursor > 0 {
				il.cursor--
				if il.cursor < il.offset {
					il.offset--
				}
			}
		case "enter":
			if issue := il.SelectedIssue(); issue != nil {
				app.detail.SetIssue(issue)
				app.currentView = viewDetail
			}
		case "m":
			if issue := il.SelectedIssue(); issue != nil {
				return il, app.showTransitions(issue.Key)
			}
		case "c":
			if issue := il.SelectedIssue(); issue != nil {
				app.currentView = viewDetail
				app.detail.SetIssue(issue)
				app.detail.StartComment()
			}
		case "n":
			app.currentView = viewSearch
			app.search.StartCreate()
		}
	}
	return il, nil
}

func (il IssueList) View(width, height int, active bool) string {
	il.maxVisible = height - 4

	title := panelTitleStyle.Render(fmt.Sprintf("My Issues (%d)", len(il.issues)))

	var rows []string
	end := min(il.offset+il.maxVisible, len(il.issues))
	for i := il.offset; i < end; i++ {
		issue := il.issues[i]
		key := issueKeyStyle.Render(issue.Key)
		summary := issue.Fields.Summary
		if len(summary) > width-30 {
			summary = summary[:width-33] + "..."
		}
		status := issueStatusStyle.Render(issue.Fields.Status.Name)

		line := fmt.Sprintf("%s %s %s", key, issueSummaryStyle.Render(summary), status)
		if i == il.cursor {
			line = selectedStyle.Width(width - 4).Render(
				fmt.Sprintf("%-12s %s %14s", issue.Key, summary, issue.Fields.Status.Name),
			)
		}
		rows = append(rows, line)
	}

	content := strings.Join(rows, "\n")
	if len(il.issues) == 0 {
		content = helpDescStyle.Render("No issues found")
	}

	body := lipgloss.JoinVertical(lipgloss.Left, title, content)

	style := panelStyle
	if active {
		style = activePanelStyle
	}
	return style.Width(width).Height(height).Render(body)
}

func (app *App) showTransitions(issueKey string) tea.Cmd {
	return func() tea.Msg {
		transitions, err := app.client.GetTransitions(issueKey)
		if err != nil {
			return nil
		}
		app.store.UpsertTransitions(issueKey, transitions)
		return transitionsMsg{issueKey: issueKey, transitions: transitions}
	}
}

type transitionsMsg struct {
	issueKey    string
	transitions []jira.Transition
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
