package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/temujinlabs/shinkansen/internal/jira"
)

type BoardView struct {
	columns    []boardColumn
	colCursor  int
	rowCursor  int
	rowOffset  int
	maxVisible int
}

type boardColumn struct {
	name   string
	issues []jira.Issue
}

func NewBoardView() BoardView {
	return BoardView{
		columns: []boardColumn{
			{name: "To Do"},
			{name: "In Progress"},
			{name: "Done"},
		},
	}
}

func (bv *BoardView) SetIssues(issues []jira.Issue) {
	// Reset columns
	for i := range bv.columns {
		bv.columns[i].issues = nil
	}

	for _, issue := range issues {
		status := strings.ToLower(issue.Fields.Status.Name)
		switch {
		case strings.Contains(status, "done") || strings.Contains(status, "closed") || strings.Contains(status, "resolved"):
			bv.columns[2].issues = append(bv.columns[2].issues, issue)
		case strings.Contains(status, "progress") || strings.Contains(status, "review"):
			bv.columns[1].issues = append(bv.columns[1].issues, issue)
		default:
			bv.columns[0].issues = append(bv.columns[0].issues, issue)
		}
	}
}

func (bv *BoardView) SelectedIssue() *jira.Issue {
	if bv.colCursor >= len(bv.columns) {
		return nil
	}
	col := bv.columns[bv.colCursor]
	if bv.rowCursor >= len(col.issues) {
		return nil
	}
	return &col.issues[bv.rowCursor]
}

func (bv BoardView) Update(msg tea.Msg, app *App) (BoardView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "down":
			col := bv.columns[bv.colCursor]
			if bv.rowCursor < len(col.issues)-1 {
				bv.rowCursor++
				if bv.maxVisible > 0 && bv.rowCursor-bv.rowOffset >= bv.maxVisible {
					bv.rowOffset++
				}
			}
		case "up":
			if bv.rowCursor > 0 {
				bv.rowCursor--
				if bv.rowCursor < bv.rowOffset {
					bv.rowOffset--
				}
			}
		case "tab":
			bv.colCursor = (bv.colCursor + 1) % len(bv.columns)
			bv.rowCursor = 0
			bv.rowOffset = 0
		case "shift+tab":
			bv.colCursor--
			if bv.colCursor < 0 {
				bv.colCursor = len(bv.columns) - 1
			}
			bv.rowCursor = 0
			bv.rowOffset = 0
		case "enter":
			if issue := bv.SelectedIssue(); issue != nil {
				app.detail.SetIssue(issue)
				app.currentView = viewDetail
			}
		case "m":
			if issue := bv.SelectedIssue(); issue != nil {
				return bv, app.showTransitions(issue.Key)
			}
		}
	}
	return bv, nil
}

func (bv BoardView) View(width, height int, active bool) string {
	title := panelTitleStyle.Render("Sprint Board")

	colWidth := (width - 6) / len(bv.columns)
	var cols []string

	maxRows := height - 6
	if maxRows < 1 {
		maxRows = 1
	}
	bv.maxVisible = maxRows

	for ci, col := range bv.columns {
		header := lipgloss.NewStyle().
			Bold(true).
			Foreground(colorAccent).
			Width(colWidth).
			Align(lipgloss.Center).
			Render(fmt.Sprintf("%s (%d)", col.name, len(col.issues)))

		var rows []string
		rows = append(rows, header)
		rows = append(rows, strings.Repeat("─", colWidth))

		// Apply scroll offset only to the active column
		startIdx := 0
		if ci == bv.colCursor {
			startIdx = bv.rowOffset
		}

		visible := 0
		for ri := startIdx; ri < len(col.issues) && visible < maxRows; ri++ {
			issue := col.issues[ri]
			line := fmt.Sprintf("  %s", issue.Key)
			if len(issue.Fields.Summary) > colWidth-14 {
				line += " " + issue.Fields.Summary[:colWidth-17] + "..."
			} else {
				line += " " + issue.Fields.Summary
			}

			if ci == bv.colCursor && ri == bv.rowCursor {
				line = selectedStyle.Width(colWidth).Render(line)
			}
			rows = append(rows, line)
			visible++
		}

		remaining := len(col.issues) - (startIdx + visible)
		if remaining > 0 {
			rows = append(rows, helpDescStyle.Render(fmt.Sprintf("  +%d more ↓", remaining)))
		}
		if startIdx > 0 {
			// Insert scroll-up indicator after the separator
			rows = append(rows[:2], append([]string{helpDescStyle.Render(fmt.Sprintf("  ↑ %d above", startIdx))}, rows[2:]...)...)
		}

		cols = append(cols, strings.Join(rows, "\n"))
	}

	content := lipgloss.JoinHorizontal(lipgloss.Top, cols...)
	body := lipgloss.JoinVertical(lipgloss.Left, title, content)

	style := panelStyle
	if active {
		style = activePanelStyle
	}
	return style.Width(width).Height(height).Render(body)
}
