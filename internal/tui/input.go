package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/temujinlabs/shinkansen/internal/cache"
	"github.com/temujinlabs/shinkansen/internal/jira"
)

// TransitionPicker shows available status transitions for an issue.
type TransitionPicker struct {
	issueKey    string
	transitions []jira.Transition
	cursor      int
	visible     bool
}

func (tp *TransitionPicker) Show(issueKey string, transitions []jira.Transition) {
	tp.issueKey = issueKey
	tp.transitions = transitions
	tp.cursor = 0
	tp.visible = true
}

func (tp *TransitionPicker) Hide() {
	tp.visible = false
}

func (tp TransitionPicker) Update(msg tea.Msg, app *App) (TransitionPicker, tea.Cmd) {
	if !tp.visible {
		return tp, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			tp.Hide()
		case "j", "down":
			if tp.cursor < len(tp.transitions)-1 {
				tp.cursor++
			}
		case "k", "up":
			if tp.cursor > 0 {
				tp.cursor--
			}
		case "enter":
			if tp.cursor < len(tp.transitions) {
				t := tp.transitions[tp.cursor]
				key := tp.issueKey
				tp.Hide()
				return tp, func() tea.Msg {
					app.client.TransitionIssue(key, t.ID)
					return syncDoneMsg{result: cache.Sync(app.client, app.store)}
				}
			}
		}
	}
	return tp, nil
}

func (tp TransitionPicker) View(width, height int) string {
	if !tp.visible || len(tp.transitions) == 0 {
		return ""
	}

	title := searchPromptStyle.Render(fmt.Sprintf("Move %s to:", tp.issueKey))
	var options []string
	for i, t := range tp.transitions {
		line := fmt.Sprintf("  %s → %s", t.Name, t.To.Name)
		if i == tp.cursor {
			line = selectedStyle.Render(line)
		}
		options = append(options, line)
	}

	content := title + "\n\n"
	for _, o := range options {
		content += o + "\n"
	}
	content += "\n" + helpDescStyle.Render("Enter: select  Esc: cancel")

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center,
		panelStyle.Render(content),
	)
}
