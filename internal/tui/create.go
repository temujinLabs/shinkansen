package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/temujinlabs/shinkansen/internal/cache"
)

// createDoneMsg is sent after a new issue is created successfully.
type createDoneMsg struct {
	issueKey string
}

// createErrMsg is sent when issue creation fails.
type createErrMsg struct {
	err error
}

// createField identifies which field is being edited.
type createField int

const (
	fieldSummary createField = iota
	fieldType
	fieldPriority
	fieldDescription
	fieldCount // sentinel: total number of fields
)

// issueTypes available for creation.
var issueTypes = []string{"Task", "Bug", "Story"}

// issuePriorities available for creation.
var issuePriorities = []string{"Highest", "High", "Medium", "Low", "Lowest"}

// CreateView handles the multi-field issue creation form.
type CreateView struct {
	visible  bool
	field    createField
	summary  string
	typeIdx  int // index into issueTypes
	prioIdx  int // index into issuePriorities
	desc     string
	errMsg   string
}

func NewCreateView() CreateView {
	return CreateView{
		prioIdx: 2, // default to Medium
	}
}

// Show opens the create form with default values.
func (cv *CreateView) Show() {
	cv.visible = true
	cv.field = fieldSummary
	cv.summary = ""
	cv.typeIdx = 0
	cv.prioIdx = 2
	cv.desc = ""
	cv.errMsg = ""
}

// Hide closes the create form.
func (cv *CreateView) Hide() {
	cv.visible = false
}

func (cv CreateView) Update(msg tea.Msg, app *App) (CreateView, tea.Cmd) {
	if !cv.visible {
		return cv, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			cv.Hide()
			app.currentView = viewIssues
			return cv, nil

		case "tab":
			// Move to next field
			cv.field = (cv.field + 1) % fieldCount
			return cv, nil

		case "shift+tab":
			// Move to previous field
			cv.field = (cv.field - 1 + fieldCount) % fieldCount
			return cv, nil

		case "ctrl+s", "ctrl+enter":
			// Submit the form
			if cv.summary == "" {
				cv.errMsg = "Summary is required"
				return cv, nil
			}
			cv.errMsg = ""
			summary := cv.summary
			issueType := issueTypes[cv.typeIdx]
			priority := issuePriorities[cv.prioIdx]
			description := cv.desc
			projectKey := app.cfg.DefaultProject

			cv.Hide()
			app.currentView = viewIssues
			app.flashMsg = "Creating issue..."

			return cv, func() tea.Msg {
				issue, err := app.client.CreateIssueWithDetails(projectKey, summary, issueType, priority, description)
				if err != nil {
					return createErrMsg{err: err}
				}

				// Try to add to active sprint
				if app.cfg.DefaultBoard > 0 {
					sprints, err := app.client.GetSprints(app.cfg.DefaultBoard)
					if err == nil {
						for _, s := range sprints {
							if s.State == "active" {
								app.client.MoveToSprint(s.ID, issue.Key)
								break
							}
						}
					}
				}

				// Sync to refresh the board
				result := cache.Sync(app.client, app.store, app.cfg.DefaultProject)
				if result.Err != nil {
					return createDoneMsg{issueKey: issue.Key}
				}
				return createDoneMsg{issueKey: issue.Key}
			}

		case "enter":
			// In description field, enter adds a newline; otherwise move to next field
			if cv.field == fieldDescription {
				cv.desc += "\n"
				return cv, nil
			}
			// For other fields, treat as tab (next field)
			cv.field = (cv.field + 1) % fieldCount
			return cv, nil

		case "left":
			switch cv.field {
			case fieldType:
				if cv.typeIdx > 0 {
					cv.typeIdx--
				}
			case fieldPriority:
				if cv.prioIdx > 0 {
					cv.prioIdx--
				}
			}
			return cv, nil

		case "right":
			switch cv.field {
			case fieldType:
				if cv.typeIdx < len(issueTypes)-1 {
					cv.typeIdx++
				}
			case fieldPriority:
				if cv.prioIdx < len(issuePriorities)-1 {
					cv.prioIdx++
				}
			}
			return cv, nil

		case "backspace":
			switch cv.field {
			case fieldSummary:
				if len(cv.summary) > 0 {
					cv.summary = cv.summary[:len(cv.summary)-1]
				}
			case fieldDescription:
				if len(cv.desc) > 0 {
					cv.desc = cv.desc[:len(cv.desc)-1]
				}
			}
			return cv, nil

		default:
			ch := msg.String()
			if len(ch) == 1 || ch == " " {
				switch cv.field {
				case fieldSummary:
					cv.summary += ch
				case fieldDescription:
					cv.desc += ch
				}
			}
			return cv, nil
		}
	}
	return cv, nil
}

func (cv CreateView) View(width, height int) string {
	if !cv.visible {
		return ""
	}

	var lines []string
	lines = append(lines, detailHeaderStyle.Render("Create New Issue"))
	lines = append(lines, "")

	// Summary field
	summaryLabel := "  Summary:"
	if cv.field == fieldSummary {
		summaryLabel = searchPromptStyle.Render("> Summary:")
	}
	summaryVal := cv.summary
	if cv.field == fieldSummary {
		summaryVal += "\u2588" // block cursor
	}
	lines = append(lines, summaryLabel+"  "+summaryVal)
	lines = append(lines, "")

	// Type field (selector)
	typeLabel := "  Type:"
	if cv.field == fieldType {
		typeLabel = searchPromptStyle.Render("> Type:")
	}
	var typeParts []string
	for i, t := range issueTypes {
		if i == cv.typeIdx {
			typeParts = append(typeParts, selectedStyle.Render(" "+t+" "))
		} else {
			typeParts = append(typeParts, helpDescStyle.Render(" "+t+" "))
		}
	}
	lines = append(lines, typeLabel+"     "+strings.Join(typeParts, "  "))
	lines = append(lines, "")

	// Priority field (selector)
	prioLabel := "  Priority:"
	if cv.field == fieldPriority {
		prioLabel = searchPromptStyle.Render("> Priority:")
	}
	var prioParts []string
	for i, p := range issuePriorities {
		if i == cv.prioIdx {
			prioParts = append(prioParts, selectedStyle.Render(" "+p+" "))
		} else {
			prioParts = append(prioParts, helpDescStyle.Render(" "+p+" "))
		}
	}
	lines = append(lines, prioLabel+" "+strings.Join(prioParts, "  "))
	lines = append(lines, "")

	// Description field (multiline text)
	descLabel := "  Description:"
	if cv.field == fieldDescription {
		descLabel = searchPromptStyle.Render("> Description:")
	}
	lines = append(lines, descLabel)
	descText := cv.desc
	if cv.field == fieldDescription {
		descText += "\u2588"
	}
	if descText == "" && cv.field != fieldDescription {
		lines = append(lines, helpDescStyle.Render("    (optional)"))
	} else {
		for _, dl := range strings.Split(descText, "\n") {
			lines = append(lines, "    "+dl)
		}
	}

	// Error message
	if cv.errMsg != "" {
		lines = append(lines, "")
		lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("#cc3333")).Bold(true).Render("  "+cv.errMsg))
	}

	lines = append(lines, "")
	lines = append(lines, helpDescStyle.Render("  Tab/Shift+Tab: navigate fields  Left/Right: select option"))
	lines = append(lines, helpDescStyle.Render("  Ctrl+S: create issue  Esc: cancel"))

	content := strings.Join(lines, "\n")
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center,
		panelStyle.Width(min(width-4, 80)).Render(content),
	)
}

