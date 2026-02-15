package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/temujinlabs/shinkansen/internal/cache"
	"github.com/temujinlabs/shinkansen/internal/jira"
)

type DetailView struct {
	issue      *jira.Issue
	scrollY    int
	commenting bool
	commentBuf string
}

func NewDetailView() DetailView {
	return DetailView{}
}

func (dv *DetailView) SetIssue(issue *jira.Issue) {
	dv.issue = issue
	dv.scrollY = 0
	dv.commenting = false
	dv.commentBuf = ""
}

func (dv *DetailView) StartComment() {
	dv.commenting = true
	dv.commentBuf = ""
}

func (dv DetailView) Update(msg tea.Msg, app *App) (DetailView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if dv.commenting {
			switch msg.String() {
			case "enter":
				if dv.commentBuf != "" && dv.issue != nil {
					comment := dv.commentBuf
					key := dv.issue.Key
					dv.commenting = false
					dv.commentBuf = ""
					return dv, func() tea.Msg {
						app.client.AddComment(key, comment)
						return syncDoneMsg{result: cache.Sync(app.client, app.store)}
					}
				}
				dv.commenting = false
			case "esc":
				dv.commenting = false
				dv.commentBuf = ""
			case "backspace":
				if len(dv.commentBuf) > 0 {
					dv.commentBuf = dv.commentBuf[:len(dv.commentBuf)-1]
				}
			default:
				if len(msg.String()) == 1 || msg.String() == " " {
					dv.commentBuf += msg.String()
				}
			}
			return dv, nil
		}

		switch msg.String() {
		case "q", "esc":
			app.currentView = viewIssues
		case "j", "down":
			dv.scrollY++
		case "k", "up":
			if dv.scrollY > 0 {
				dv.scrollY--
			}
		case "c":
			dv.StartComment()
		case "m":
			if dv.issue != nil {
				return dv, app.showTransitions(dv.issue.Key)
			}
		}
	}
	return dv, nil
}

func (dv DetailView) View(width, height int) string {
	if dv.issue == nil {
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center,
			helpDescStyle.Render("No issue selected"))
	}

	i := dv.issue
	var lines []string

	lines = append(lines, detailHeaderStyle.Render(fmt.Sprintf("%s: %s", i.Key, i.Fields.Summary)))
	lines = append(lines, "")

	lines = append(lines, detailLabelStyle.Render("Status:")+" "+detailValueStyle.Render(i.Fields.Status.Name))
	lines = append(lines, detailLabelStyle.Render("Priority:")+" "+detailValueStyle.Render(i.Fields.Priority.Name))
	lines = append(lines, detailLabelStyle.Render("Type:")+" "+detailValueStyle.Render(i.Fields.IssueType.Name))
	lines = append(lines, detailLabelStyle.Render("Assignee:")+" "+detailValueStyle.Render(i.AssigneeName()))

	if i.Fields.Reporter != nil {
		lines = append(lines, detailLabelStyle.Render("Reporter:")+" "+detailValueStyle.Render(i.Fields.Reporter.DisplayName))
	}
	lines = append(lines, detailLabelStyle.Render("Project:")+" "+detailValueStyle.Render(i.Fields.Project.Name))
	lines = append(lines, detailLabelStyle.Render("Updated:")+" "+detailValueStyle.Render(i.Fields.Updated))
	lines = append(lines, "")

	if i.Fields.Description != "" {
		lines = append(lines, detailLabelStyle.Render("Description:"))
		// Wrap long descriptions
		descLines := strings.Split(i.Fields.Description, "\n")
		for _, dl := range descLines {
			if len(dl) > width-6 {
				for len(dl) > width-6 {
					lines = append(lines, "  "+dl[:width-6])
					dl = dl[width-6:]
				}
				if dl != "" {
					lines = append(lines, "  "+dl)
				}
			} else {
				lines = append(lines, "  "+dl)
			}
		}
		lines = append(lines, "")
	}

	// Comments
	if i.Fields.Comment != nil && len(i.Fields.Comment.Comments) > 0 {
		lines = append(lines, detailLabelStyle.Render(fmt.Sprintf("Comments (%d):", len(i.Fields.Comment.Comments))))
		for _, c := range i.Fields.Comment.Comments {
			lines = append(lines, "")
			lines = append(lines, fmt.Sprintf("  %s — %s", helpKeyStyle.Render(c.Author.DisplayName), helpDescStyle.Render(c.Created)))
			lines = append(lines, "  "+c.Body)
		}
	}

	// Comment input
	if dv.commenting {
		lines = append(lines, "")
		lines = append(lines, searchPromptStyle.Render("Add comment: ")+dv.commentBuf+"█")
	}

	// Apply scroll
	if dv.scrollY > 0 && dv.scrollY < len(lines) {
		lines = lines[dv.scrollY:]
	}
	if len(lines) > height-2 {
		lines = lines[:height-2]
	}

	content := strings.Join(lines, "\n")

	footer := statusBarStyle.Render("esc:back  j/k:scroll  c:comment  m:move")
	return lipgloss.JoinVertical(lipgloss.Left,
		panelStyle.Width(width-2).Render(content),
		footer,
	)
}
