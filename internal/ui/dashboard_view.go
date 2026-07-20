// SPDX-License-Identifier: GPL-3.0-or-later
// SPDX-FileCopyrightText: 2026 The MetaBoard authors

package ui

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gotunix.net/metaboard/internal/version"
	"gotunix.net/metaboard/internal/store"
)

func (m *dashboardModel) View() string {
	if len(m.lines) == 0 && m.state == StateDashboard {
		return "No items to display. Press q to quit."
	}

	// Capture outer terminal dimensions
	termWidth := m.width
	if termWidth <= 0 {
		termWidth = GetTerminalWidth()
	}
	termHeight := m.height
	if termHeight <= 0 {
		termHeight = 24
	}

	// Calculate target dimensions at 85%
	totalWidth := int(float64(termWidth) * 0.95)
	if totalWidth < 80 {
		totalWidth = 80
	}
	totalHeight := int(float64(termHeight) * 0.95)
	if totalHeight < 12 {
		totalHeight = 12
	}

	isFormState := m.state == StateCreateMilestone || m.state == StateEditMilestone ||
		m.state == StateCreateTask || m.state == StateEditTask ||
		m.state == StateCreatePR || m.state == StateEditPR

	if m.state == StatePromptInit {
		titleBanner := TitleStyle.Copy().Width(totalWidth).Render(" INITIALIZE METABOARD ")

		var bodyText string
		if m.initError != "" {
			bodyText = lipgloss.NewStyle().Foreground(catMochaRed).Bold(true).Render("Initialization Error: " + m.initError) +
				"\n\nPress 'y' to try again, or 'n' / 'esc' / 'q' to exit."
		} else {
			bodyText = "No Metaboard database structure was found in this directory.\n\n" +
				"Would you like to initialize a new Metaboard project here?\n\n" +
				"Press " +
				lipgloss.NewStyle().Foreground(catMochaGreen).Bold(true).Render("'y'") +
				" to continue, or " +
				lipgloss.NewStyle().Foreground(catMochaRed).Bold(true).Render("'n' / 'esc' / 'q'") +
				" to exit."
		}
		bodyStyled := lipgloss.NewStyle().Padding(2, 4).Render(bodyText)

		view := lipgloss.JoinVertical(lipgloss.Left, titleBanner, bodyStyled)

		return lipgloss.Place(
			termWidth, termHeight,
			lipgloss.Center, lipgloss.Center,
			lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(catMochaMauve).
				Width(totalWidth).
				Height(totalHeight).
				Render(view),
		)
	}

	if m.state == StatePromptInitLocation {
		titleBanner := TitleStyle.Copy().Width(totalWidth).Render(" CHOOSE INIT LOCATION ")

		keyStyle := lipgloss.NewStyle().Foreground(catMochaMauve).Bold(true)
		pathStyle := lipgloss.NewStyle().Foreground(catMochaPeach)
		hintStyle := lipgloss.NewStyle().Foreground(catMochaSubtext)

		bodyText := "Where should Metaboard create the project directories?\n\n" +
			keyStyle.Render("[1]") + "  " + pathStyle.Render("./") + "            " + hintStyle.Render("(current directory)") + "\n" +
			keyStyle.Render("[2]") + "  " + pathStyle.Render("./boards/") + "      " + hintStyle.Render("(boards subdirectory)") + "\n\n" +
			"Press " + keyStyle.Render("1") + " or " + keyStyle.Render("2") + " to select, or " +
			lipgloss.NewStyle().Foreground(catMochaRed).Bold(true).Render("'esc' / 'q'") +
			" to exit."
		bodyStyled := lipgloss.NewStyle().Padding(2, 4).Render(bodyText)

		view := lipgloss.JoinVertical(lipgloss.Left, titleBanner, bodyStyled)

		return lipgloss.Place(
			termWidth, termHeight,
			lipgloss.Center, lipgloss.Center,
			lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(catMochaMauve).
				Width(totalWidth).
				Height(totalHeight).
				Render(view),
		)
	}

	if m.state == StatePromptPush {
		titleBanner := TitleStyle.Copy().Width(totalWidth).Render(" PUSH TO REMOTE ")

		keyStyle := lipgloss.NewStyle().Foreground(catMochaMauve).Bold(true)

		var bodyText string
		if m.pushError != "" {
			bodyText = lipgloss.NewStyle().Foreground(catMochaRed).Bold(true).Render("Push Error: " + m.pushError) +
				"\n\nPress " + keyStyle.Render("'esc'" ) + " to return."
		} else {
			bodyText = "Push local commits to a remote branch.\n\n" +
				"Branch: " + m.pushBranchField.View() + "\n\n" +
				"Type to change the branch name, then press " + keyStyle.Render("Enter") + " to push,\n" +
				"or " + lipgloss.NewStyle().Foreground(catMochaRed).Bold(true).Render("'esc' / 'q'") + " to cancel."
		}
		bodyStyled := lipgloss.NewStyle().Padding(2, 4).Render(bodyText)

		view := lipgloss.JoinVertical(lipgloss.Left, titleBanner, bodyStyled)

		return lipgloss.Place(
			termWidth, termHeight,
			lipgloss.Center, lipgloss.Center,
			lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(catMochaMauve).
				Width(totalWidth).
				Height(totalHeight).
				Render(view),
		)
	}

	if m.state == StateConfirmDelete {
		titleBanner := TitleStyle.Copy().Width(totalWidth).Render(" CONFIRM DELETION ")
		
		var bodyText string
		if m.deleteErrorMsg != "" {
			bodyText = lipgloss.NewStyle().Foreground(catMochaRed).Bold(true).Render("Error: " + m.deleteErrorMsg) + "\n\nPress any key to cancel."
		} else {
			bodyText = fmt.Sprintf("Are you sure you want to delete this %s ([%s])?\n\nPress 'y' to confirm, or 'n' / 'esc' to cancel.", m.deleteViewType, m.deleteViewSlug)
		}
		bodyStyled := lipgloss.NewStyle().Padding(2, 4).Render(bodyText)

		view := lipgloss.JoinVertical(lipgloss.Left, titleBanner, bodyStyled)

		return lipgloss.Place(
			termWidth, termHeight,
			lipgloss.Center, lipgloss.Center,
			lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(catMochaMauve).
				Width(totalWidth).
				Height(totalHeight).
				Render(view),
		)
	}

	if m.state == StateConfirmUnlink {
		titleBanner := TitleStyle.Copy().Width(totalWidth).Render(" CONFIRM UNLINK ")
		
		var bodyText string
		if m.unlinkErrorMsg != "" {
			bodyText = lipgloss.NewStyle().Foreground(catMochaRed).Bold(true).Render("Error: " + m.unlinkErrorMsg) + "\n\nPress any key to cancel."
		} else {
			bodyText = fmt.Sprintf("Are you sure you want to unlink this %s ([%s]) from its parent?\n\nPress 'y' to confirm, or 'n' / 'esc' to cancel.", m.unlinkViewType, m.unlinkViewSlug)
		}
		bodyStyled := lipgloss.NewStyle().Padding(2, 4).Render(bodyText)

		view := lipgloss.JoinVertical(lipgloss.Left, titleBanner, bodyStyled)

		return lipgloss.Place(
			termWidth, termHeight,
			lipgloss.Center, lipgloss.Center,
			lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(catMochaMauve).
				Width(totalWidth).
				Height(totalHeight).
				Render(view),
		)
	}

	// Render inline forms if currently active
	if isFormState {
		var titleText string
		switch m.state {
		case StateCreateMilestone:
			titleText = " CREATE MILESTONE "
		case StateEditMilestone:
			titleText = fmt.Sprintf(" EDIT MILESTONE: %s ", m.formSlug)
		case StateCreateTask:
			titleText = " CREATE TASK "
		case StateEditTask:
			titleText = fmt.Sprintf(" EDIT TASK: %s ", m.formSlug)
		case StateCreatePR:
			titleText = " CREATE PULL REQUEST "
		case StateEditPR:
			titleText = fmt.Sprintf(" EDIT PULL REQUEST: %s ", m.formSlug)
		}

		titleBanner := TitleStyle.Copy().Width(totalWidth).Render(titleText)

		// Render the custom fields matching samples-golang/form!
		formSections := m.renderFormView(totalWidth - 6)

		var fieldSpacer string
		if totalHeight < 24 {
			fieldSpacer = "\n"
		} else {
			fieldSpacer = "\n\n"
		}

		formView := strings.Join(formSections, fieldSpacer)
		formLines := strings.Split(formView, "\n")
		for i := 0; i < len(formLines); i++ {
			formLines[i] = "  " + formLines[i]
		}

		targetLines := totalHeight - 2
		clampedFormScroll := m.formScroll
		if len(formLines) > targetLines {
			maxScroll := len(formLines) - targetLines
			if clampedFormScroll > maxScroll {
				clampedFormScroll = maxScroll
			}
			if clampedFormScroll < 0 {
				clampedFormScroll = 0
			}
			formLines = formLines[clampedFormScroll : clampedFormScroll+targetLines]
		} else if len(formLines) < targetLines {
			for len(formLines) < targetLines {
				formLines = append(formLines, "")
			}
		}

		// Scroll indicator
		scrollHint := ""
		if m.formScroll > 0 {
			scrollHint = " ↑ scroll ↑"
		}

		helpText := scrollHint + " Tab/Shift+Tab: navigate • Enter/Space: select • PgDn/PgUp: scroll • Esc/Ctrl+C: Cancel "
		helpLegendStyled := TitleStyle.Copy().Width(totalWidth).Align(lipgloss.Center).Render(helpText)

		finalSections := []string{titleBanner}
		finalSections = append(finalSections, formLines...)
		finalSections = append(finalSections, helpLegendStyled)

		windowStyle := lipgloss.NewStyle().
			Padding(0, 0).
			Width(totalWidth).
			Height(totalHeight).
			Background(catMochaBase)

		renderedWindow := windowStyle.Render(strings.Join(finalSections, "\n"))
		return lipgloss.Place(termWidth, termHeight, lipgloss.Center, lipgloss.Center, renderedWindow, lipgloss.WithWhitespaceBackground(catMochaBase))
	}

	// Render inline view detail pages if active
	if m.state == StateViewMilestone || m.state == StateViewTask || m.state == StateViewPR || m.state == StateViewVersion || m.state == StateViewChangelog {
		content := m.viewContent

		// Calculate available height inside padding
		lines := strings.Split(content, "\n")
		for i := 5; i < len(lines); i++ {
			if len(lines[i]) > 0 {
				lines[i] = "  " + lines[i]
			}
		}
		targetLines := totalHeight - 2

		// Help legend with scroll position indicator
		helpText := " Esc/q: Return • j/k/PgDn/PgUp: Scroll • e: Edit Item "
		if m.state == StateViewVersion {
			helpText = " Esc/q: Return • j/k/PgDn/PgUp: Scroll "
		} else if m.state == StateViewChangelog {
			helpText = " Esc/q: Return • j/k/PgDn/PgUp: Scroll • u: Update/Regenerate "
		} else if m.state == StateViewTask {
			helpText = " Esc/q: Return • j/k/PgDn/PgUp: Scroll • e: Edit Task • p: Edit Plan "
		}
		clampedPreviewScroll := m.previewScroll
		if len(lines) > targetLines {
			maxScroll := len(lines) - targetLines
			pct := int(float64(clampedPreviewScroll) / float64(maxScroll) * 100)
			if pct < 0 {
				pct = 0
			} else if pct > 100 {
				pct = 100
			}
			helpText += fmt.Sprintf("  │  Scroll: %d%%", pct)
		}
		helpTextStyled := TitleStyle.Copy().Width(totalWidth).Align(lipgloss.Center).Render(helpText)

		if len(lines) > targetLines {
			maxScroll := len(lines) - targetLines
			if clampedPreviewScroll < 0 {
				clampedPreviewScroll = 0
			}
			if clampedPreviewScroll > maxScroll {
				clampedPreviewScroll = maxScroll
			}
			lines = lines[clampedPreviewScroll : clampedPreviewScroll+targetLines]
		} else if len(lines) < targetLines {
			for len(lines) < targetLines {
				lines = append(lines, "")
			}
		} else if len(lines) > targetLines {
			lines = lines[:targetLines]
		}
		
		finalLines := []string{lines[0]}
		finalLines = append(finalLines, lines[1:]...)
		finalLines = append(finalLines, helpTextStyled)

		windowStyle := lipgloss.NewStyle().
			Padding(0, 0).
			Width(totalWidth).
			Height(totalHeight).
			Background(catMochaBase)

		renderedWindow := windowStyle.Render(strings.Join(finalLines, "\n"))
		return lipgloss.Place(termWidth, termHeight, lipgloss.Center, lipgloss.Center, renderedWindow, lipgloss.WithWhitespaceBackground(catMochaBase))
	}

	leftWidth := totalWidth
	leftHeight := totalHeight

	var sections []string

	// 1. Title Banner
	cursorIndicator := ""
	if len(m.lines) > 0 {
		cursorIndicator = fmt.Sprintf(" (%d/%d)", m.cursor+1, len(m.lines))
	}
	sections = append(sections, lipgloss.NewStyle().Width(leftWidth).Align(lipgloss.Center).Background(catMochaMauve).Foreground(catMochaBase).Bold(true).Render(" METABOARD "+version.AppVersion+cursorIndicator+" "))
	// 2. Table Headers
	if len(m.widths) == 4 {
		headerRow := lipgloss.NewStyle().Width(leftWidth).Background(catMochaMauve).Render(m.renderHeader(m.widths, leftWidth))
		sections = append(sections, headerRow)
	}

	// 3. Dynamic scrolling viewport of rows
	maxVisibleRows := leftHeight - 10
	if maxVisibleRows < 4 {
		maxVisibleRows = 4
	}

	// Scroll control
	if m.cursor >= m.viewStart+maxVisibleRows {
		m.viewStart = m.cursor - maxVisibleRows + 1
	} else if m.cursor < m.viewStart {
		m.viewStart = m.cursor
	}

	if m.viewStart < 0 {
		m.viewStart = 0
	}
	if m.viewStart > len(m.lines)-maxVisibleRows {
		m.viewStart = len(m.lines) - maxVisibleRows
	}
	if m.viewStart < 0 {
		m.viewStart = 0
	}

	start := m.viewStart
	end := start + maxVisibleRows
	if end > len(m.lines) {
		end = len(m.lines)
	}

	for i := start; i < end; i++ {
		if i >= len(m.flatRows) {
			break
		}
		r := m.flatRows[i]
		isCursor := i == m.cursor

		if r.Depth < 0 {
			sections = append(sections, "")
			continue
		}

		var rowStyle lipgloss.Style
		switch r.Type {
		case TypeMilestone:
			rowStyle = milestoneRowStyle
		case TypeTask:
			rowStyle = taskRowStyle
		case TypePR:
			rowStyle = storyRowStyle
		}

		statusCell := formatStatusCell(r.Status, m.widths[2], isCursor)

		var progressCell string
		if r.Type == TypeMilestone {
			progressCell = renderProgressBar(r.Progress, r.ProgressText, m.widths[3])
		} else {
			progressCell = r.ProgressText
		}

		typeCell := formatTypeCell(r.ItemType, m.widths[1], isCursor)

		lineText := renderRow(getRowTitle(r.Depth, r.Expanded, r.Title), typeCell, statusCell, progressCell, m.widths, isCursor, rowStyle)
		sections = append(sections, lineText)
	}

	actualRenderedRows := end - start
	if actualRenderedRows < maxVisibleRows {
		for i := 0; i < maxVisibleRows-actualRenderedRows; i++ {
			sections = append(sections, "")
		}
	}

	// 4. Help legend (centered inside with purple banner styling)
	helpText := " ↑/↓/j/k: navigate • space: expand/collapse • enter/e: edit • t/s/m/p: create • c: changelog • f: fetch • g: push • v: version • q: quit "
	helpTextStyled := lipgloss.NewStyle().Width(leftWidth).Align(lipgloss.Center).Background(catMochaMauve).Foreground(catMochaBase).Bold(true).Render(helpText)
	sections = append(sections, helpTextStyled)

	// 5. Status bar (transient, shown below help when non-empty)
	if m.statusMsg != "" {
		statusStyle := lipgloss.NewStyle().Width(leftWidth).Align(lipgloss.Center).Background(catMochaMauve).Foreground(catMochaBase).Bold(true)
		sections = append(sections, statusStyle.Render(" "+m.statusMsg+" "))
	}

	leftContent := strings.Join(sections, "\n")
	leftStyle := lipgloss.NewStyle().
		Padding(0, 0).
		Width(leftWidth).
		Height(leftHeight).
		Background(catMochaBase)

	leftPanel := leftStyle.Render(leftContent)

	return lipgloss.Place(termWidth, termHeight, lipgloss.Center, lipgloss.Center, leftPanel, lipgloss.WithWhitespaceBackground(catMochaBase))
}

func (m *dashboardModel) renderHeader(widths []int, totalWidth int) string {
	cells := []string{}
	headers := []string{"BOARD ITEM", "TYPE", "STATUS", "PROGRESS"}

	for i, col := range headers {
		w := widths[i]
		var styled string
		if len(col) > w {
			styled = lipgloss.NewStyle().Foreground(catMochaBase).Background(catMochaMauve).Bold(true).Render(col[:w-3] + "...")
		} else {
			styled = lipgloss.NewStyle().Foreground(catMochaBase).Background(catMochaMauve).Bold(true).Render(col) + lipgloss.NewStyle().Background(catMochaMauve).Render(strings.Repeat(" ", w-len(col)))
		}
		cells = append(cells, styled)
	}

	prefix := lipgloss.NewStyle().Background(catMochaMauve).Render("  ")
	separator := lipgloss.NewStyle().Foreground(catMochaBase).Background(catMochaMauve).Render(" │ ")
	middle := prefix + strings.Join(cells, separator)

	var parts []string
	for _, w := range widths {
		parts = append(parts, strings.Repeat("─", w))
	}
	borderStyle := lipgloss.NewStyle().Foreground(catMochaBase).Background(catMochaMauve)
	
	topBorder := borderStyle.Render("──" + strings.Join(parts, "─┬─"))
	botBorder := borderStyle.Render("──" + strings.Join(parts, "─┼─"))
	
	// Pad the right side to match totalWidth exactly
	currentWidth := lipgloss.Width(middle)
	if currentWidth < totalWidth {
		padding := strings.Repeat(" ", totalWidth-currentWidth)
		paddingLine := strings.Repeat("─", totalWidth-currentWidth)
		topBorder += borderStyle.Render(paddingLine)
		middle += lipgloss.NewStyle().Background(catMochaMauve).Render(padding)
		botBorder += borderStyle.Render(paddingLine)
	}

	return topBorder + "\n" + middle + "\n" + botBorder
}

func CenterText(text string, width int) string {
	textWidth := lipgloss.Width(text)
	if textWidth >= width {
		return text
	}
	padding := (width - textWidth) / 2
	return strings.Repeat(" ", padding) + text
}

func RenderDashboard(mode, target, filter string) error {
	p := tea.NewProgram(newDashboardModel(mode, target, filter), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func getRowTitle(depth int, expanded bool, title string) string {
	indent := strings.Repeat("  ", depth)
	var icon string
	if depth < 2 {
		if expanded {
			icon = "▼ "
		} else {
			icon = "▶ "
		}
	} else {
		icon = "▪ "
	}
	return indent + icon + title
}

func formatTypeCell(itemType string, width int, isCursor bool) string {
	s := strings.ToUpper(itemType)

	var badge string
	switch s {
	case "MILESTONE":
		badge = lipgloss.NewStyle().Foreground(catMochaBase).Background(catMochaMauve).Bold(true).Width(width).Align(lipgloss.Center).Render(s)
	case "TASK":
		badge = lipgloss.NewStyle().Foreground(catMochaBase).Background(catMochaText).Bold(true).Width(width).Align(lipgloss.Center).Render(s)
	case "PR", "PULL REQUEST":
		badge = lipgloss.NewStyle().Foreground(catMochaBase).Background(catMochaPeach).Bold(true).Width(width).Align(lipgloss.Center).Render(s)
	default:
		badge = lipgloss.NewStyle().Foreground(catMochaBase).Background(catMochaOverlay).Bold(true).Width(width).Align(lipgloss.Center).Render(s)
	}

	vWidth := lipgloss.Width(badge)
	spaces := width - vWidth
	if spaces < 0 {
		spaces = 0
	}

	if isCursor {
		return badge + lipgloss.NewStyle().Background(catMochaOverlay).Render(strings.Repeat(" ", spaces))
	}
	return badge + lipgloss.NewStyle().Background(catMochaBase).Render(strings.Repeat(" ", spaces))
}

func formatStatusCell(status string, width int, isCursor bool) string {
	badge := renderStatusBadge(status)
	vWidth := lipgloss.Width(badge)

	spaces := width - vWidth
	if spaces < 0 {
		spaces = 0
	}

	if isCursor {
		return badge + lipgloss.NewStyle().Background(catMochaOverlay).Render(strings.Repeat(" ", spaces))
	}
	return badge + lipgloss.NewStyle().Background(catMochaBase).Render(strings.Repeat(" ", spaces))
}

func renderStatusBadge(status string) string {
	s := strings.ToUpper(status)
	badgeWidth := 13

	switch s {
	case "COMPLETED", "DONE":
		return statusDoneStyle.Copy().Padding(0, 0).Width(badgeWidth).Align(lipgloss.Center).Render(s)
	case "MERGED":
		return statusMergedStyle.Copy().Padding(0, 0).Width(badgeWidth).Align(lipgloss.Center).Render(s)
	case "CLOSED":
		return statusDoneStyle.Copy().Padding(0, 0).Width(badgeWidth).Align(lipgloss.Center).Render(s)
	case "ACTIVE", "IN-PROGRESS", "OPEN", "DOING":
		return statusProgressStyle.Copy().Padding(0, 0).Width(badgeWidth).Align(lipgloss.Center).Render(s)
	case "BACKLOG", "DRAFT", "TODO", "CANCELLED", "REJECTED":
		if s == "CANCELLED" || s == "REJECTED" {
			return lipgloss.NewStyle().Foreground(catMochaBase).Background(catMochaRed).Bold(true).Width(badgeWidth).Align(lipgloss.Center).Render(s)
		}
		return statusTodoStyle.Copy().Padding(0, 0).Width(badgeWidth).Align(lipgloss.Center).Render(s)
	default:
		return statusTodoStyle.Copy().Padding(0, 0).Width(badgeWidth).Align(lipgloss.Center).Render(s)
	}
}

func renderProgressBar(pct float64, text string, width int) string {
	if pct < 0 {
		pct = 0
	} else if pct > 1 {
		pct = 1
	}

	textWidth := 6
	barWidth := width - textWidth - 2 // 1 space between bar and text, 1 leading space
	if barWidth < 3 {
		barWidth = 3
	}

	filled := int(float64(barWidth) * pct)
	empty := barWidth - filled

	bar := "[" + strings.Repeat("█", filled) + strings.Repeat("░", empty) + "]"
	result := fmt.Sprintf("%s %*s", bar, textWidth, text)

	// Pad to exactly width so the column header aligns
	resultWidth := lipgloss.Width(result)
	if resultWidth < width {
		result += strings.Repeat(" ", width-resultWidth)
	}
	return result
}

func renderRow(title, typeStr, statusCell, progress string, widths []int, isCursor bool, rowStyle lipgloss.Style) string {
	formatCol := func(txt string, w int, style lipgloss.Style) string {
		var val string
		vWidth := lipgloss.Width(txt)
		if vWidth > w {
			runes := []rune(txt)
			var curWidth int
			var safeRunes []rune
			for _, r := range runes {
				rw := lipgloss.Width(string(r))
				if curWidth+rw > w-3 {
					break
				}
				curWidth += rw
				safeRunes = append(safeRunes, r)
			}
			val = string(safeRunes) + "..."
			valWidth := lipgloss.Width(val)
			if valWidth < w {
				val += strings.Repeat(" ", w-valWidth)
			}
		} else {
			val = txt + strings.Repeat(" ", w-vWidth)
		}

		if isCursor {
			return style.Copy().Background(catMochaOverlay).Foreground(catMochaText).Render(val)
		}
		return style.Copy().Background(catMochaBase).Render(val)
	}

	c0 := formatCol(title, widths[0], rowStyle)
	
	c1 := typeStr
	if lipgloss.Width(typeStr) < widths[1] {
		c1 = formatCol(typeStr, widths[1], rowStyle)
	}

	c2 := statusCell
	// If it's a raw string (header row), format it like a normal col
	if lipgloss.Width(statusCell) < widths[2] {
		c2 = formatCol(statusCell, widths[2], rowStyle)
	}
	c3 := formatCol(progress, widths[3], rowStyle)

	var prefix string
	if isCursor {
		prefix = lipgloss.NewStyle().Background(catMochaOverlay).Foreground(catMochaGreen).Bold(true).Render("▸ ")
	} else {
		prefix = lipgloss.NewStyle().Background(catMochaBase).Render("  ")
	}

	var sepStyle lipgloss.Style
	if isCursor {
		sepStyle = lipgloss.NewStyle().Foreground(catMochaOverlay).Background(catMochaOverlay)
	} else {
		sepStyle = lipgloss.NewStyle().Foreground(catMochaOverlay).Background(catMochaBase)
	}
	separator := sepStyle.Render(" │ ")

	return prefix + c0 + separator + c1 + separator + c2 + separator + c3
}

type execCmdWrapper struct {
	*exec.Cmd
}

func (w execCmdWrapper) SetStdin(r io.Reader) {
	w.Cmd.Stdin = r
}

func (w execCmdWrapper) SetStdout(wr io.Writer) {
	w.Cmd.Stdout = wr
}

func (w execCmdWrapper) SetStderr(wr io.Writer) {
	w.Cmd.Stderr = wr
}

type editorFinishedMsg struct{ err error }

func (m *dashboardModel) viewMilestoneContent(slug string, width, height int) string {
	mil, err := store.GetMilestone(slug)
	if err != nil {
		return "Error loading milestone: " + err.Error()
	}

	col1Width := (width - 3) / 2
	col2Width := width - col1Width - 3
	if col1Width < 30 {
		col1Width = 30
	}
	if col2Width < 30 {
		col2Width = 30
	}

	var sb strings.Builder

	// Title Banner
	sb.WriteString(TitleStyle.Copy().Width(width).Render(fmt.Sprintf("MILESTONE DETAILS: %s", strings.ToUpper(mil.Title))) + "\n")
	dividerText := strings.Repeat("─", col1Width+1) + "┬" + strings.Repeat("─", width-col1Width-2)
	sb.WriteString(lipgloss.NewStyle().Foreground(catMochaBase).Background(TableBg).Render(dividerText) + "\n")

	// Info Grid
	labelStyle := lipgloss.NewStyle().Foreground(LabelFg).Bold(true).Background(TableBg)
	gridValStyle := lipgloss.NewStyle().Foreground(ValFg).Background(TableBg)
	valStyle := lipgloss.NewStyle().Foreground(catMochaText)
	bgStyle := lipgloss.NewStyle().Background(TableBg)
	sepStyle := lipgloss.NewStyle().Foreground(catMochaBase).Background(TableBg)

	formatField := func(label, val string, colW int) string {
		lWidth := lipgloss.Width(label)
		vWidth := colW - lWidth - 1
		if vWidth < 0 {
			vWidth = 0
		}
		if lipgloss.Width(val) > vWidth {
			val = val[:vWidth]
		}
		return labelStyle.Render(label) + bgStyle.Render(" ") + gridValStyle.Render(val) + bgStyle.Render(strings.Repeat(" ", vWidth-lipgloss.Width(val)))
	}

	row1 := formatField("Status:  ", mil.Status, col1Width) + sepStyle.Render(" │ ") + formatField("Version: ", fmt.Sprintf("v%d", mil.Version), col2Width)
	row2 := formatField("ID:      ", mil.ID, col1Width) + sepStyle.Render(" │ ") + formatField("Created: ", mil.CreatedAt, col2Width)
	row3 := formatField("Updated: ", mil.UpdatedAt, col1Width) + sepStyle.Render(" │ ") + formatField("Done At: ", mil.CompletedAt, col2Width)

	sb.WriteString(row1 + "\n" + row2 + "\n" + row3 + "\n\n")

	// Description
	sb.WriteString(lipgloss.NewStyle().Foreground(catMochaMauve).Bold(true).Render("Description:") + "\n")
	desc := mil.Description
	if desc == "" {
		desc = "(No description provided)"
	}
	descLines := strings.Split(desc, "\n")
	for i, dl := range descLines {
		if i > 3 {
			sb.WriteString(lipgloss.NewStyle().Foreground(catMochaOverlay).Render("  ... (truncated)") + "\n")
			break
		}
		sb.WriteString("  " + valStyle.Render(dl) + "\n")
	}
	sb.WriteString("\n")

	return sb.String()
}

func (m *dashboardModel) viewTaskContent(slug string, width, height int) string {
	t, err := store.GetTask(slug)
	if err != nil {
		return "Error loading task: " + err.Error()
	}

	milestones, _ := store.ListMilestones()
	parentMilestone := "(None)"
	for _, mVal := range milestones {
		for _, tID := range mVal.Tasks {
			if tID == t.ID || tID == t.Slug {
				parentMilestone = mVal.Title
				break
			}
		}
	}

	col1Width := (width - 3) / 2
	col2Width := width - col1Width - 3
	if col1Width < 30 {
		col1Width = 30
	}
	if col2Width < 30 {
		col2Width = 30
	}

	var sb strings.Builder

	// Title Banner
	sb.WriteString(TitleStyle.Copy().Width(width).Render(fmt.Sprintf("TASK DETAILS: %s", strings.ToUpper(t.Title))) + "\n")
	dividerText := strings.Repeat("─", col1Width+1) + "┬" + strings.Repeat("─", width-col1Width-2)
	sb.WriteString(lipgloss.NewStyle().Foreground(catMochaBase).Background(TableBg).Render(dividerText) + "\n")

	// Info Grid
	labelStyle := lipgloss.NewStyle().Foreground(LabelFg).Bold(true).Background(TableBg)
	gridValStyle := lipgloss.NewStyle().Foreground(ValFg).Background(TableBg)
	valStyle := lipgloss.NewStyle().Foreground(catMochaText)
	bgStyle := lipgloss.NewStyle().Background(TableBg)
	sepStyle := lipgloss.NewStyle().Foreground(catMochaBase).Background(TableBg)

	formatField := func(label, val string, colW int) string {
		lWidth := lipgloss.Width(label)
		vWidth := colW - lWidth - 1
		if vWidth < 0 {
			vWidth = 0
		}
		if lipgloss.Width(val) > vWidth {
			val = val[:vWidth]
		}
		return labelStyle.Render(label) + bgStyle.Render(" ") + gridValStyle.Render(val) + bgStyle.Render(strings.Repeat(" ", vWidth-lipgloss.Width(val)))
	}

	row1 := formatField("Status:   ", t.Status, col1Width) + sepStyle.Render(" │ ") + formatField("Priority: ", t.Priority, col2Width)
	row2 := formatField("Assigned: ", t.AssignedTo, col1Width) + sepStyle.Render(" │ ") + formatField("Milestone:", parentMilestone, col2Width)
	row3 := formatField("Created:  ", t.CreatedAt, col1Width) + sepStyle.Render(" │ ") + formatField("Done At:  ", t.CompletedAt, col2Width)

	sb.WriteString(row1 + "\n" + row2 + "\n" + row3 + "\n\n")

	// Description
	sb.WriteString(lipgloss.NewStyle().Foreground(catMochaMauve).Bold(true).Render("Description:") + "\n")
	desc := t.Description
	if desc == "" {
		desc = "(No description provided)"
	}
	descLines := strings.Split(desc, "\n")
	for i, dl := range descLines {
		if i > 3 {
			sb.WriteString(lipgloss.NewStyle().Foreground(catMochaOverlay).Render("  ... (truncated)") + "\n")
			break
		}
		sb.WriteString("  " + valStyle.Render(dl) + "\n")
	}
	sb.WriteString("\n")

	// Tags & PRs
	tagsStr := strings.Join(t.Tags, ", ")
	if tagsStr == "" {
		tagsStr = "(None)"
	}
	prsStr := strings.Join(t.PullRequests, ", ")
	if prsStr == "" {
		prsStr = "(None)"
	}

	dependsStr := strings.Join(t.DependsOn, ", ")
	if dependsStr == "" {
		dependsStr = "(None)"
	}

	sb.WriteString(lipgloss.NewStyle().Foreground(catMochaMauve).Bold(true).Render("Metadata:") + "\n")
	sb.WriteString(fmt.Sprintf("  • Tags:          %s\n", valStyle.Render(tagsStr)))
	sb.WriteString(fmt.Sprintf("  • Depends On:    %s\n", valStyle.Render(dependsStr)))
	sb.WriteString(fmt.Sprintf("  • Pull Requests: %s\n", valStyle.Render(prsStr)))
	sb.WriteString("\n")

	planPath, err := store.GetTaskPlanPath(t.ID)
	if err == nil {
		if b, err := os.ReadFile(planPath); err == nil && len(b) > 0 {
			sb.WriteString(lipgloss.NewStyle().Foreground(catMochaMauve).Bold(true).Render("Additional Details:") + "\n")
			planLines := strings.Split(string(b), "\n")
			for _, pl := range planLines {
				sb.WriteString("  " + valStyle.Render(pl) + "\n")
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func (m *dashboardModel) viewPRContent(slug string, width, height int) string {
	pr, err := store.GetPullRequest(slug)
	if err != nil {
		return "Error loading pull request: " + err.Error()
	}

	col1Width := (width - 3) / 2
	col2Width := width - col1Width - 3
	if col1Width < 30 {
		col1Width = 30
	}
	if col2Width < 30 {
		col2Width = 30
	}

	var sb strings.Builder

	// Title Banner
	sb.WriteString(TitleStyle.Copy().Width(width).Render(fmt.Sprintf("PULL REQUEST DETAILS: %s", strings.ToUpper(pr.Description))) + "\n")
	dividerText := strings.Repeat("─", col1Width+1) + "┬" + strings.Repeat("─", width-col1Width-2)
	sb.WriteString(lipgloss.NewStyle().Foreground(catMochaBase).Background(TableBg).Render(dividerText) + "\n")

	// Info Grid
	labelStyle := lipgloss.NewStyle().Foreground(LabelFg).Bold(true).Background(TableBg)
	gridValStyle := lipgloss.NewStyle().Foreground(ValFg).Background(TableBg)
	valStyle := lipgloss.NewStyle().Foreground(catMochaText)
	bgStyle := lipgloss.NewStyle().Background(TableBg)
	sepStyle := lipgloss.NewStyle().Foreground(catMochaBase).Background(TableBg)

	formatField := func(label, val string, colW int) string {
		lWidth := lipgloss.Width(label)
		vWidth := colW - lWidth - 1
		if vWidth < 0 {
			vWidth = 0
		}
		if lipgloss.Width(val) > vWidth {
			val = val[:vWidth]
		}
		return labelStyle.Render(label) + bgStyle.Render(" ") + gridValStyle.Render(val) + bgStyle.Render(strings.Repeat(" ", vWidth-lipgloss.Width(val)))
	}

	row1 := formatField("Status:   ", pr.Status, col1Width) + sepStyle.Render(" │ ") + formatField("ID:       ", pr.ID, col2Width)
	row2 := formatField("Head:     ", pr.HeadBranch, col1Width) + sepStyle.Render(" │ ") + formatField("Base:     ", pr.BaseBranch, col2Width)
	row3 := formatField("Created:  ", pr.CreatedAt, col1Width) + sepStyle.Render(" │ ") + formatField("Updated:  ", pr.UpdatedAt, col2Width)

	sb.WriteString(row1 + "\n" + row2 + "\n" + row3 + "\n\n")

	// Description/Body
	sb.WriteString(lipgloss.NewStyle().Foreground(catMochaMauve).Bold(true).Render("Source Repositories:") + "\n")
	sb.WriteString(fmt.Sprintf("  • Source:      %s\n", valStyle.Render(pr.SourceRepo)))
	sb.WriteString(fmt.Sprintf("  • Destination: %s\n", valStyle.Render(pr.DestRepo)))

	return sb.String()
}

func (m *dashboardModel) viewVersionContent(width, height int) string {
	return GetVersionString(width)
}

func (m *dashboardModel) viewChangelogContent(width, height int) string {
	var sb strings.Builder
	sb.WriteString(TitleStyle.Copy().Width(width).Render("CHANGELOG") + "\n")

	if m.changelogMessage != "" {
		msgStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2"))
		if strings.HasPrefix(m.changelogMessage, "Error:") {
			msgStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("1"))
		}
		sb.WriteString("  " + msgStyle.Render(m.changelogMessage) + "\n\n")
	}

	content, err := os.ReadFile("CHANGELOG.md")
	if err != nil {
		sb.WriteString("\n  No CHANGELOG.md found in the current directory.\n")
		sb.WriteString("  Press 'u' to generate/update the changelog from your tasks.\n")
		return sb.String()
	}

	sb.WriteString(string(content))
	return sb.String()
}


func (m *dashboardModel) updateViewContent() {
	if m.state != StateViewMilestone && m.state != StateViewTask && m.state != StateViewPR && m.state != StateViewVersion && m.state != StateViewChangelog {
		return
	}

	termWidth := m.width
	if termWidth <= 0 {
		termWidth = GetTerminalWidth()
	}
	termHeight := m.height
	if termHeight <= 0 {
		termHeight = 24
	}

	totalWidth := int(float64(termWidth) * 0.95)
	if totalWidth < 80 {
		totalWidth = 80
	}
	totalHeight := int(float64(termHeight) * 0.95)
	if totalHeight < 12 {
		totalHeight = 12
	}

	var content string
	switch m.state {
	case StateViewMilestone:
		content = m.viewMilestoneContent(m.viewSlug, totalWidth, totalHeight)
	case StateViewTask:
		content = m.viewTaskContent(m.viewSlug, totalWidth, totalHeight)
	case StateViewPR:
		content = m.viewPRContent(m.viewSlug, totalWidth, totalHeight)
	case StateViewVersion:
		content = m.viewVersionContent(totalWidth, totalHeight)
	case StateViewChangelog:
		content = m.viewChangelogContent(totalWidth, totalHeight)
	}
	m.viewContent = content
}

