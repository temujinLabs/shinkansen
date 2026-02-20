package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/temujinlabs/shinkansen/internal/cache"
	"github.com/temujinlabs/shinkansen/internal/jira"
)

// ProjectPicker shows a list of available projects for switching.
type ProjectPicker struct {
	projects []jira.Project
	cursor   int
	visible  bool
	loading  bool
}

func NewProjectPicker() ProjectPicker {
	return ProjectPicker{}
}

// Show opens the project picker in loading state.
func (pp *ProjectPicker) Show() {
	pp.visible = true
	pp.loading = true
	pp.cursor = 0
}

// Hide closes the project picker.
func (pp *ProjectPicker) Hide() {
	pp.visible = false
	pp.loading = false
}

// SetProjects populates the project list after fetching.
func (pp *ProjectPicker) SetProjects(projects []jira.Project) {
	pp.projects = projects
	pp.loading = false
	pp.cursor = 0
}

func (pp ProjectPicker) Update(msg tea.Msg, app *App) (ProjectPicker, tea.Cmd) {
	if !pp.visible {
		return pp, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			pp.Hide()
			return pp, nil

		case "down":
			if pp.cursor < len(pp.projects)-1 {
				pp.cursor++
			}
			return pp, nil

		case "up":
			if pp.cursor > 0 {
				pp.cursor--
			}
			return pp, nil

		case "enter":
			if len(pp.projects) > 0 && pp.cursor < len(pp.projects) {
				project := pp.projects[pp.cursor]
				pp.Hide()

				// If same project, do nothing
				if project.Key == app.cfg.DefaultProject {
					return pp, nil
				}

				newProjectKey := project.Key
				return pp, func() tea.Msg {
					// Clear cache for old project and sync new one
					result := cache.Sync(app.client, app.store, newProjectKey)
					if result.Err != nil {
						return statusMsg(fmt.Sprintf("Switch failed: %v", result.Err))
					}
					return projectSwitchedMsg{projectKey: newProjectKey}
				}
			}
			return pp, nil
		}
	}
	return pp, nil
}

func (pp ProjectPicker) View(width, height int) string {
	if !pp.visible {
		return ""
	}

	var lines []string
	lines = append(lines, detailHeaderStyle.Render("Switch Project"))
	lines = append(lines, "")

	if pp.loading {
		lines = append(lines, helpDescStyle.Render("  Loading projects..."))
	} else if len(pp.projects) == 0 {
		lines = append(lines, helpDescStyle.Render("  No projects found"))
	} else {
		for i, p := range pp.projects {
			indicator := "  "
			line := fmt.Sprintf("%s%s  %s", indicator, helpKeyStyle.Render(p.Key), p.Name)
			if len(line) > width-8 {
				line = line[:width-11] + "..."
			}
			if i == pp.cursor {
				line = selectedStyle.Width(width - 8).Render(
					fmt.Sprintf("  %-10s %s", p.Key, p.Name),
				)
			}
			lines = append(lines, line)
		}
	}

	lines = append(lines, "")
	lines = append(lines, helpDescStyle.Render("  Enter: switch  Esc: cancel"))

	content := strings.Join(lines, "\n")
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center,
		panelStyle.Width(min(width-4, 60)).Render(content),
	)
}
