// SPDX-License-Identifier: GPL-3.0-or-later
// SPDX-FileCopyrightText: 2026 The MetaBoard authors

package ui

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gotunix.net/metaboard/internal/git"
	"gotunix.net/metaboard/internal/models"
	"gotunix.net/metaboard/internal/store"
)

func (m *dashboardModel) moveCursor(dir int) {
	if len(m.lines) == 0 {
		return
	}
	start := m.cursor
	for {
		m.cursor += dir
		if m.cursor < 0 {
			m.cursor = len(m.lines) - 1
		} else if m.cursor >= len(m.lines) {
			m.cursor = 0
		}
		if m.lines[m.cursor].isSelectable {
			break
		}
		if m.cursor == start {
			break
		}
	}
}

func (m *dashboardModel) refresh() {
	milestones, err := store.ListMilestones()
	if err != nil {
		m.lines = []dashboardLine{{text: "Error listing milestones: " + err.Error(), isSelectable: false}}
		return
	}
	store.SortMilestones(milestones)
	tasks, _ := store.ListTasks()
	pullRequests, _ := store.ListPullRequests()

	taskMap := make(map[string]models.Task)
	for _, t := range tasks {
		taskMap[t.ID] = t
	}
	prMap := make(map[string]models.PullRequest)
	for _, pr := range pullRequests {
		prMap[pr.ID] = pr
	}

	isMatch := func(status string) bool {
		s := strings.ToUpper(status)
		switch strings.ToLower(m.filter) {
		case "active":
			return s != models.StatusCompleted && s != models.StatusCancelled && s != models.StatusMerged && s != models.StatusClosed && s != models.StatusRejected
		case "closed":
			return s == models.StatusCompleted || s == models.StatusMerged || s == models.StatusClosed || s == models.StatusRejected
		case "cancelled":
			return s == models.StatusCancelled
		default:
			return true
		}
	}

	isExpanded := func(slug string) bool {
		val, ok := m.expanded[slug]
		if !ok {
			return true // default to expanded
		}
		return val
	}

	var filteredMilestones []models.Milestone
	if m.mode == "milestone" && m.target != "" {
		for _, mVal := range milestones {
			if mVal.ID == m.target || mVal.Slug == m.target || (len(m.target) >= 4 && strings.HasPrefix(mVal.ID, m.target)) {
				filteredMilestones = append(filteredMilestones, mVal)
				break
			}
		}
	} else {
		for _, mVal := range milestones {
			if isMatch(mVal.Status) {
				filteredMilestones = append(filteredMilestones, mVal)
			}
		}
	}

	var flatRows []FlatRow

	for _, mVal := range filteredMilestones {
		// Calculate Milestone stats
		totalTasks := 0
		doneTasks := 0
		for _, tID := range mVal.Tasks {
			if t, ok := taskMap[tID]; ok {
				totalTasks++
				if strings.ToUpper(t.Status) == "COMPLETED" {
					doneTasks++
				}
			}
		}

		var progress float64
		if totalTasks > 0 {
			progress = float64(doneTasks) / float64(totalTasks)
		}
		progressText := fmt.Sprintf("%d/%d", doneTasks, totalTasks)

		mExpanded := isExpanded(mVal.Slug)

		flatRows = append(flatRows, FlatRow{
			Type:         TypeMilestone,
			Title:        mVal.Title,
			ItemType:     "milestone",
			Slug:         mVal.Slug,
			Status:       mVal.Status,
			Progress:     progress,
			ProgressText: progressText,
			Depth:        0,
			Expanded:     mExpanded,
		})

		if mExpanded {
			// Add standalone milestone tasks
			var activeMilestoneTasks []models.Task
			for _, tID := range mVal.Tasks {
				if t, ok := taskMap[tID]; ok {
					if isMatch(t.Status) {
						activeMilestoneTasks = append(activeMilestoneTasks, t)
					}
				}
			}
			store.SortTasks(activeMilestoneTasks)

			for _, t := range activeMilestoneTasks {
				flatRows = append(flatRows, FlatRow{
					Type:         TypeTask,
					Title:        t.Title,
					ItemType:     "task",
					Slug:         t.Slug,
					Status:       t.Status,
					Progress:     0,
					ProgressText: "",
					Depth:        1,
					Expanded:     false,
				})
			}
		}
	}

	flatRows = append(flatRows, FlatRow{Depth: -1})

	// Backlog
	claimedTasks := make(map[string]bool)
	for _, mVal := range milestones {
		for _, tID := range mVal.Tasks {
			claimedTasks[tID] = true
		}
	}

	var unclaimedTasks []models.Task
	for _, t := range tasks {
		if !claimedTasks[t.ID] && isMatch(t.Status) {
			unclaimedTasks = append(unclaimedTasks, t)
		}
	}

	if len(unclaimedTasks) > 0 {
		unclaimedExpanded := isExpanded("unclaimed-items")
		flatRows = append(flatRows, FlatRow{
			Type:         TypeMilestone,
			Title:        "Backlog",
			ItemType:     "milestone",
			Slug:         "unclaimed-items",
			Status:       "BACKLOG",
			Progress:     0,
			ProgressText: "",
			Depth:        0,
			Expanded:     unclaimedExpanded,
		})

		if unclaimedExpanded {
			store.SortTasks(unclaimedTasks)
			for _, t := range unclaimedTasks {
				flatRows = append(flatRows, FlatRow{
					Type:         TypeTask,
					Title:        t.Title,
					ItemType:     "task",
					Slug:         t.Slug,
					Status:       t.Status,
					Progress:     0,
					ProgressText: "",
					Depth:        1,
					Expanded:     false,
				})
			}
		}
	}

	flatRows = append(flatRows, FlatRow{Depth: -1})

	// Pull Requests
	var activePRs []models.PullRequest
	for _, pr := range pullRequests {
		if isMatch(pr.Status) {
			activePRs = append(activePRs, pr)
		}
	}
	if len(activePRs) > 0 {
		prExpanded := isExpanded("pull-requests")
		flatRows = append(flatRows, FlatRow{
			Type:         TypeMilestone,
			Title:        "Active Pull Requests",
			ItemType:     "milestone",
			Slug:         "pull-requests",
			Status:       "OPEN",
			Progress:     0,
			ProgressText: "",
			Depth:        0,
			Expanded:     prExpanded,
		})

		if prExpanded {
			for _, pr := range activePRs {
				prExpandedRow := isExpanded(pr.Slug)
				flatRows = append(flatRows, FlatRow{
					Type:         TypePR,
					Title:        fmt.Sprintf("%s [%s]", pr.Description, pr.Slug),
					ItemType:     "pr",
					Slug:         pr.Slug,
					Status:       pr.Status,
					Progress:     0,
					ProgressText: "",
					Depth:        1,
					Expanded:     prExpandedRow,
				})

				if prExpandedRow {
					for _, tID := range pr.Tasks {
						if t, ok := taskMap[tID]; ok && isMatch(t.Status) {
							flatRows = append(flatRows, FlatRow{
								Type:         TypeTask,
								Title:        t.Title,
								ItemType:     "task",
								Slug:         t.Slug,
								Status:       t.Status,
								Progress:     0,
								ProgressText: "",
								Depth:        2,
								Expanded:     false,
							})
						}
					}
				}
			}
		}
	}

	// Now build column widths based on targetWidth (4 columns: Name, Type, Status, Progress)
	totalWidth := m.width
	if totalWidth <= 0 {
		totalWidth = GetTerminalWidth()
	}
	targetWidth := int(float64(totalWidth) * 0.95)
	if targetWidth < 80 {
		targetWidth = 80
	}
	contentWidth := targetWidth - 20

	widths := []int{28, 10, 14, 18}
	baseColsWidth := 28 + 10 + 14 + 18

	if contentWidth > baseColsWidth {
		extra := contentWidth - baseColsWidth
		titleExtra := int(float64(extra) * 0.6)
		progressExtra := extra - titleExtra
		widths[0] += titleExtra
		widths[3] += progressExtra
	} else if contentWidth < baseColsWidth {
		shrink := baseColsWidth - contentWidth
		if widths[3]-shrink >= 10 {
			widths[3] -= shrink
		} else {
			shrink -= (widths[3] - 10)
			widths[3] = 10
			if widths[0]-shrink >= 15 {
				widths[0] -= shrink
			} else {
				widths[0] = 15
			}
		}
	}

	m.widths = widths
	m.flatRows = flatRows

	// Clear and reconstruct m.lines
	m.lines = nil
	if len(flatRows) == 0 {
		msg := "No items found."
		m.lines = append(m.lines, dashboardLine{
			text:         lipgloss.NewStyle().Foreground(Gray).Render(msg),
			isSelectable: false,
		})
		return
	}

	for _, r := range flatRows {
		m.lines = append(m.lines, dashboardLine{
			isSelectable: r.Depth >= 0,
			itemType:     r.ItemType,
			slug:         r.Slug,
		})
	}

	if m.cursor >= len(m.lines) {
		m.cursor = len(m.lines) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	if len(m.lines) > 0 && !m.lines[m.cursor].isSelectable {
		m.moveCursor(1)
	}
	m.updateViewContent()
}

type refreshMsg struct{}

type statusClearMsg struct{}

type fetchResultMsg struct{ err error }

type pushResultMsg struct{ err error }

// statusCmd clears the status bar after 3 seconds.
func statusCmd() tea.Cmd {
	return tea.Tick(3*time.Second, func(time.Time) tea.Msg {
		return statusClearMsg{}
	})
}


func (m *dashboardModel) Init() tea.Cmd {
	return nil
}

func (m *dashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	model, cmd := m.updateInternal(msg)
	if dm, ok := model.(*dashboardModel); ok {
		dm.clampScroll()
	}
	return model, cmd
}

func (m *dashboardModel) updateInternal(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle resize unconditionally so all states are responsive.
	if wm, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = wm.Width
		m.height = wm.Height
		m.refresh()
	}

	// Handle async result messages unconditionally BEFORE any state-gated
	// blocks. These messages arrive as callbacks from background operations
	// (editor exec, git fetch/push) and must never be swallowed by a
	// state handler's early return.
	if _, ok := msg.(editorFinishedMsg); ok {
		if m.pendingPlanPath != "" {
			// Use paths captured at editor-launch time — never re-compute
			// from store/filesystem after the editor exits.
			absPath := m.pendingPlanPath
			slug := m.pendingPlanSlug
			root := m.pendingPlanRoot
			m.pendingPlanPath = ""
			m.pendingPlanSlug = ""
			m.pendingPlanRoot = ""
			if root != "" && git.IsGitRepo(root) {
				err := git.Commit(root, []string{absPath}, fmt.Sprintf("boards: plan created for task %s", slug))
				if err != nil {
					m.statusMsg = "Error: Failed to commit plan: " + err.Error()
					m.statusIsError = true
				} else {
					m.statusMsg = "OK — Plan saved and committed"
					m.statusIsError = false
				}
				m.refresh()
				return m, statusCmd()
			}
		}
m.refresh()
		return m, nil
	}

	isFormState := m.state == StateCreateMilestone || m.state == StateEditMilestone ||
		m.state == StateCreateTask || m.state == StateEditTask ||
		m.state == StateCreatePR || m.state == StateEditPR

	// 1. Check if we are in form state
	if isFormState {
		// Handle keypresses manually to match form-demo!
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "esc", "ctrl+c":
				switch m.state {
				case StateEditMilestone:
					m.state = StateViewMilestone
				case StateEditTask:
					m.state = StateViewTask
				case StateEditPR:
					m.state = StateViewPR
				default:
					m.state = StateDashboard
				}
				m.refresh()
				return m, nil
			case "ctrl+d", "pgdown":
				m.formScroll += 3
				return m, nil
			case "ctrl+u", "pgup":
				m.formScroll -= 3
				if m.formScroll < 0 {
					m.formScroll = 0
				}
				return m, nil
			case "tab":
				m.formFocusIndex++
				if m.formFocusIndex >= m.formTotalFields {
					m.formFocusIndex = 0
				}
				m.updateFormFieldsFocus()
				m.formScrollToFocus()
				return m, nil
			case "shift+tab":
				m.formFocusIndex--
				if m.formFocusIndex < 0 {
					m.formFocusIndex = m.formTotalFields - 1
				}
				m.updateFormFieldsFocus()
				m.formScrollToFocus()
				return m, nil
			case "up":
				// On a select field, up moves selection; otherwise navigate focus
				it, _, _ := m.getActiveFieldInfo()
				if it == "select" || it == "priority-select" || it == "type-select" || it == "changelog-select" || it == "milestone-select" || it == "task-select" || it == "depends-select" {
					m.handleSelectMove(-1)
				} else {
					m.formFocusIndex--
					if m.formFocusIndex < 0 {
						m.formFocusIndex = m.formTotalFields - 1
					}
					m.updateFormFieldsFocus()
					m.formScrollToFocus()
				}
				return m, nil
			case "down":
				// On a select field, down moves selection; otherwise navigate focus
				it, _, _ := m.getActiveFieldInfo()
				if it == "select" || it == "priority-select" || it == "type-select" || it == "changelog-select" || it == "milestone-select" || it == "task-select" || it == "depends-select" {
					m.handleSelectMove(1)
				} else {
					m.formFocusIndex++
					if m.formFocusIndex >= m.formTotalFields {
						m.formFocusIndex = 0
					}
					m.updateFormFieldsFocus()
					m.formScrollToFocus()
				}
				return m, nil
			case "left", "h":
				// Only move select if currently on a select-type field
				it, _, _ := m.getActiveFieldInfo()
				if it == "select" || it == "priority-select" || it == "type-select" || it == "changelog-select" || it == "milestone-select" || it == "task-select" {
					m.handleSelectMove(-1)
					return m, nil
				}
			case "right", "l":
				it, _, _ := m.getActiveFieldInfo()
				if it == "select" || it == "priority-select" || it == "type-select" || it == "changelog-select" || it == "milestone-select" || it == "task-select" {
					m.handleSelectMove(1)
					return m, nil
				}
			case "space", " ":
				it, _, _ := m.getActiveFieldInfo()
				if it == "task-select" {
					m.toggleTaskSelection()
					return m, nil
				}
				if it == "depends-select" {
					m.toggleDependsOnSelection()
					return m, nil
				}
			case "enter":
				it, _, _ := m.getActiveFieldInfo()
				if it == "task-select" {
					m.toggleTaskSelection()
					return m, nil
				}
				if it == "depends-select" {
					m.toggleDependsOnSelection()
					return m, nil
				}
				// If we are on the Submit button, save!
				if m.formFocusIndex == m.formTotalFields-1 {
					cmd := m.submitForm()
					return m, cmd
				}
				// Otherwise, advance focus to next field
				m.formFocusIndex++
				if m.formFocusIndex >= m.formTotalFields {
					m.formFocusIndex = 0
				}
				m.updateFormFieldsFocus()
				m.formScrollToFocus()
				return m, nil
			}
		}

		// Forward key message to currently focused text input or textarea
		it, _, _ := m.getActiveFieldInfo()
		if it != "select" && it != "priority-select" && it != "type-select" && it != "changelog-select" && it != "milestone-select" && it != "task-select" && it != "depends-select" && it != "button" {
			modelPtr, isTextInput := m.getActiveFieldModel()
			if isTextInput {
				ti := modelPtr.(*textinput.Model)
				var cmd tea.Cmd
				*ti, cmd = ti.Update(msg)
				return m, cmd
			} else {
				ta := modelPtr.(*textarea.Model)
				var cmd tea.Cmd
				*ta, cmd = ta.Update(msg)
				return m, cmd
			}
		}
		return m, nil
	}

	// 1.5 View State Keyboard Handler
	if m.state == StateViewMilestone || m.state == StateViewTask || m.state == StateViewPR || m.state == StateViewVersion || m.state == StateViewChangelog {
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "esc", "q":
				m.state = StateDashboard
				m.refresh()
				return m, nil
			case "e":
				switch m.state {
				case StateViewMilestone:
					m.state = StateEditMilestone
					_ = m.initEditMilestoneForm(m.viewSlug)
				case StateViewTask:
					m.state = StateEditTask
					_ = m.initEditTaskForm(m.viewSlug)
				case StateViewPR:
					m.state = StateEditPR
					_ = m.initEditPRForm(m.viewSlug)
				}
				return m, nil
			case "p":
				if m.state == StateViewTask {
					planPath, err := store.EnsureTaskPlan(m.viewSlug)
					if err == nil {
						// Resolve plan path and data root to absolute paths NOW,
						// before handing off to the editor. Storing them on the
						// model means editorFinishedMsg uses the same paths
						// without any re-computation that could differ after
						// the editor process exits.
						absPlanPath, absErr := filepath.Abs(planPath)
						if absErr != nil {
							absPlanPath = planPath
						}
						m.pendingPlanPath = absPlanPath
						m.pendingPlanSlug = m.viewSlug
						// Capture the data root as an absolute path too.
						if root, errR := store.GetDataRoot(); errR == nil {
							if absRoot, errA := filepath.Abs(root); errA == nil {
								m.pendingPlanRoot = absRoot
							} else {
								m.pendingPlanRoot = root
							}
							// Stage the file immediately so it is tracked before
							// the editor opens — handles brand-new files and
							// any that were created by a previous (buggy) session.
							if git.IsGitRepo(m.pendingPlanRoot) {
								_ = git.Stage(m.pendingPlanRoot, []string{absPlanPath})
							}
						}
						editor := os.Getenv("EDITOR")
						if editor == "" {
							editor = "vim"
						}
						c := exec.Command(editor, absPlanPath)
						return m, tea.Exec(execCmdWrapper{c}, func(err error) tea.Msg {
							return editorFinishedMsg{err}
						})
					}
				}
				return m, nil

			case "ctrl+d", "pgdown", "j", "down":
				m.previewScroll += 3
				return m, nil
			case "ctrl+u", "pgup", "k", "up":
				m.previewScroll -= 3
				return m, nil
			case "d", "delete", "x":
				if m.state == StateViewVersion || m.state == StateViewChangelog {
					return m, nil
				}
				m.deleteViewSlug = m.viewSlug
				m.deleteViewType = ""
				switch m.state {
				case StateViewMilestone: m.deleteViewType = "milestone"
				case StateViewTask: m.deleteViewType = "task"
				case StateViewPR: m.deleteViewType = "pr"
				}
				if m.deleteViewType != "" {
					m.deleteErrorMsg = ""
					m.state = StateConfirmDelete
				}
				return m, nil
			case "u":
				if m.state == StateViewChangelog {
					err := GenerateChangelog(".")
					if err != nil {
						m.changelogMessage = "Error: " + err.Error()
					} else {
						m.changelogMessage = "OK — Changelog updated successfully!"
					}
					m.refresh()
					return m, nil
				}
				if m.state == StateViewVersion || m.state == StateViewMilestone {
					return m, nil
				}
				m.unlinkViewSlug = m.viewSlug
				m.unlinkViewType = ""
				switch m.state {
				case StateViewTask: m.unlinkViewType = "task"
				case StateViewPR: m.unlinkViewType = "pr"
				}
				if m.unlinkViewType != "" {
					m.unlinkErrorMsg = ""
					m.state = StateConfirmUnlink
				}
				return m, nil
			case "r":
				m.refresh()
				return m, nil
			}
		}
		return m, nil
	}

	if m.state == StatePromptInit {
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "y", "Y":
				// Proceed to location selection step
				m.initError = ""
				m.state = StatePromptInitLocation
				return m, nil
			case "n", "N", "esc", "q":
				return m, tea.Quit
			}
		}
		return m, nil
	}

	if m.state == StatePromptInitLocation {
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "1":
				m.initLocation = "."
			case "2":
				m.initLocation = "boards"
			case "esc", "q":
				return m, tea.Quit
			default:
				return m, nil
			}
			if err := store.Initialize(m.initLocation); err != nil {
				m.initError = err.Error()
				m.state = StatePromptInit
			} else {
				if git.IsGitRepo(m.initLocation) {
					_ = git.Commit(m.initLocation, []string{"."}, "boards: initialized")
				}
				m.state = StateDashboard
				m.refresh()
			}
		}
		return m, nil
	}

	if m.state == StatePromptPush {
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "enter":
				branch := strings.TrimSpace(m.pushBranchField.Value())
				root, _ := store.GetDataRoot()
				m.state = StateDashboard
				return m, func() tea.Msg {
					err := git.Push(root, branch)
					return pushResultMsg{err: err}
				}
			case "esc", "q":
				m.state = StateDashboard
				return m, nil
			default:
				var cmd tea.Cmd
				m.pushBranchField, cmd = m.pushBranchField.Update(msg)
				return m, cmd
			}
		}
		return m, nil
	}

	if m.state == StateConfirmDelete {
		if msg, ok := msg.(tea.KeyMsg); ok {
			if m.deleteErrorMsg != "" {
				m.deleteErrorMsg = ""
				m.state = StateDashboard
				m.refresh()
				return m, nil
			}
			switch msg.String() {
			case "y", "Y":
				m.executeDelete()
				return m, statusCmd()
			case "n", "N", "esc", "q":
				m.deleteErrorMsg = ""
				m.state = StateDashboard
				m.refresh()
				return m, nil
			}
		}
		return m, nil
	}

	if m.state == StateConfirmUnlink {
		if msg, ok := msg.(tea.KeyMsg); ok {
			if m.unlinkErrorMsg != "" {
				m.unlinkErrorMsg = ""
				m.state = StateDashboard
				m.refresh()
				return m, nil
			}
			switch msg.String() {
			case "y", "Y":
				m.executeUnlink()
				return m, statusCmd()
			case "n", "N", "esc", "q":
				m.unlinkErrorMsg = ""
				m.state = StateDashboard
				m.refresh()
				return m, nil
			}
		}
		return m, nil
	}

	// 2. Normal dashboard keyboard handler
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "up", "k":
			m.moveCursor(-1)
			m.previewScroll = 0
		case "down", "j":
			m.moveCursor(1)
			m.previewScroll = 0
		case "space", " ":
			if m.cursor >= 0 && m.cursor < len(m.lines) {
				line := m.lines[m.cursor]
				if line.itemType == "milestone" || line.itemType == "pr" {
					cur, ok := m.expanded[line.slug]
					if !ok {
						cur = true
					}
					m.expanded[line.slug] = !cur
					m.refresh()
					return m, nil
				}
			}
		case "d", "delete", "x":
			if m.cursor >= 0 && m.cursor < len(m.lines) {
				line := m.lines[m.cursor]
				if line.isSelectable {
					m.deleteViewSlug = line.slug
					m.deleteViewType = line.itemType
					m.deleteErrorMsg = ""
					m.state = StateConfirmDelete
					return m, nil
				}
			}
		case "u":
			if m.cursor >= 0 && m.cursor < len(m.lines) {
				line := m.lines[m.cursor]
				if line.isSelectable && line.itemType != "milestone" {
					m.unlinkViewSlug = line.slug
					m.unlinkViewType = line.itemType
					m.unlinkErrorMsg = ""
					m.state = StateConfirmUnlink
					return m, nil
				}
			}
		case "v":
			m.state = StateViewVersion
			m.previewScroll = 0
			m.refresh()
			return m, nil
		case "c":
			m.state = StateViewChangelog
			m.previewScroll = 0
			m.changelogMessage = ""
			m.refresh()
			return m, nil
		case "f":
			root, _ := store.GetDataRoot()
			if git.IsGitRepo(root) {
				return m, func() tea.Msg {
					err := git.Fetch(root)
					return fetchResultMsg{err: err}
				}
			} else {
				m.statusMsg = "Warning: Metadata directory is not git-tracked"
				m.statusIsError = true
				return m, statusCmd()
			}
		case "g":
			root, _ := store.GetDataRoot()
			if git.IsGitRepo(root) {
				branch, _ := git.CurrentBranch(root)
			m.pushBranchField.SetValue(branch)
			m.pushBranchField.CursorEnd()
				m.pushError = ""
				m.state = StatePromptPush
			} else {
				m.statusMsg = "Warning: Metadata directory is not git-tracked"
				m.statusIsError = true
				return m, statusCmd()
			}
			return m, nil
		case "ctrl+d", "pgdown":
			m.previewScroll += 5
		case "ctrl+u", "pgup":
			m.previewScroll -= 5
		case "m":
			m.state = StateCreateMilestone
			m.initCreateMilestoneForm()
			return m, nil
		case "t":
			m.state = StateCreateTask
			m.initCreateTaskForm()
			return m, nil
		case "p":
			m.state = StateCreatePR
			m.initCreatePRForm()
			return m, nil
		case "e":
			if m.cursor >= 0 && m.cursor < len(m.lines) {
				line := m.lines[m.cursor]
				if line.isSelectable {
					m.viewSlug = line.slug
					switch line.itemType {
					case "milestone":
						m.state = StateEditMilestone
						_ = m.initEditMilestoneForm(line.slug)
						return m, nil
					case "task":
						m.state = StateEditTask
						_ = m.initEditTaskForm(line.slug)
						return m, nil
					case "pr":
						m.state = StateEditPR
						_ = m.initEditPRForm(line.slug)
						return m, nil
					}
				}
			}
		case "enter":
			if m.cursor >= 0 && m.cursor < len(m.lines) {
				line := m.lines[m.cursor]
				if line.isSelectable {
					m.viewSlug = line.slug
					switch line.itemType {
					case "milestone":
						m.state = StateViewMilestone
					case "task":
						m.state = StateViewTask
					case "pr":
						m.state = StateViewPR
					}
					m.previewScroll = 0
					m.refresh()
					return m, nil
				}
			}
		}
	case fetchResultMsg:
		if msg.err != nil {
			m.statusMsg = "Error: Fetch failed: " + msg.err.Error()
			m.statusIsError = true
		} else {
			m.statusMsg = "OK — Fetched from origin"
			m.statusIsError = false
		}
		return m, statusCmd()
	case pushResultMsg:
		if msg.err != nil {
			if errors.Is(msg.err, git.ErrAlreadyUpToDate) {
				m.statusMsg = "OK — Already up to date"
				m.statusIsError = false
			} else {
				m.statusMsg = "Error: Push failed: " + msg.err.Error()
				m.statusIsError = true
			}
		} else {
			m.statusMsg = "OK — Pushed successfully"
			m.statusIsError = false
		}
		return m, statusCmd()
	case refreshMsg:
		m.refresh()
	case statusClearMsg:
		m.statusMsg = ""
		m.statusIsError = false
	}
	return m, nil
}


func (m *dashboardModel) executeDelete() {
	var err error
	var path string
	var msg string

	switch m.deleteViewType {
	case "milestone":
		mItem, _ := store.GetMilestone(m.deleteViewSlug)
		path, _ = store.GetMilestonePath(mItem.ID)
		err = store.DeleteMilestone(m.deleteViewSlug)
		msg = fmt.Sprintf("delete milestone [%s]", m.deleteViewSlug)
	case "task":
		tItem, _ := store.GetTask(m.deleteViewSlug)
		path, _ = store.GetTaskPath(tItem.ID)
		err = store.DeleteTask(m.deleteViewSlug)
		msg = fmt.Sprintf("delete task [%s]", m.deleteViewSlug)
	case "pr":
		pItem, _ := store.GetPullRequest(m.deleteViewSlug)
		path, _ = store.GetPullRequestPath(pItem.ID)
		err = store.DeletePR(m.deleteViewSlug)
		msg = fmt.Sprintf("delete pull request [%s]", m.deleteViewSlug)
	}

	if err != nil {
		m.deleteErrorMsg = err.Error()
		m.state = StateConfirmDelete
		return
	}

	root, _ := store.GetDataRoot()
	if git.IsGitRepo(root) {
		if err := git.Commit(root, []string{path}, msg); err != nil {
			m.statusMsg = "Error: Git commit failed: " + err.Error()
			m.statusIsError = true
		} else {
			m.statusMsg = "OK — Deleted successfully"
			m.statusIsError = false
		}
	} else {
		m.statusMsg = "OK — Deleted successfully"
		m.statusIsError = false
	}

	m.deleteErrorMsg = ""
	m.state = StateDashboard
	m.refresh()
}

func (m *dashboardModel) executeUnlink() {
	err := store.UnlinkEntity(m.unlinkViewSlug)

	if err != nil {
		m.unlinkErrorMsg = err.Error()
		m.state = StateConfirmUnlink
		return
	}

	root, _ := store.GetDataRoot()
	if git.IsGitRepo(root) {
		msg := fmt.Sprintf("boards: unlink %s [%s]", m.unlinkViewType, m.unlinkViewSlug)
		if err := git.Commit(root, []string{"."}, msg); err != nil {
			m.statusMsg = "Error: Git commit failed: " + err.Error()
			m.statusIsError = true
		} else {
			m.statusMsg = "OK — Unlinked successfully"
			m.statusIsError = false
		}
	} else {
		m.statusMsg = "OK — Unlinked successfully"
		m.statusIsError = false
	}

	m.unlinkErrorMsg = ""
	m.state = StateDashboard
	m.refresh()
}


func (m *dashboardModel) clampScroll() {
	isFormState := m.state == StateCreateMilestone || m.state == StateEditMilestone ||
		m.state == StateCreateTask || m.state == StateEditTask ||
		m.state == StateCreatePR || m.state == StateEditPR

	if isFormState {
		max := m.maxFormScroll()
		if m.formScroll > max {
			m.formScroll = max
		}
		if m.formScroll < 0 {
			m.formScroll = 0
		}
	} else if m.state == StateViewMilestone || m.state == StateViewTask || m.state == StateViewPR || m.state == StateViewVersion || m.state == StateViewChangelog {
		max := m.maxPreviewScroll()
		if m.previewScroll > max {
			m.previewScroll = max
		}
		if m.previewScroll < 0 {
			m.previewScroll = 0
		}
	}
}

func (m *dashboardModel) maxFormScroll() int {
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

	formSections := m.renderFormView(totalWidth - 6)

	var fieldSpacer string
	if totalHeight < 24 {
		fieldSpacer = "\n"
	} else {
		fieldSpacer = "\n\n"
	}

	formView := strings.Join(formSections, fieldSpacer)
	formLines := strings.Split(formView, "\n")
	targetLines := totalHeight - 2

	if len(formLines) > targetLines {
		return len(formLines) - targetLines
	}
	return 0
}

func (m *dashboardModel) maxPreviewScroll() int {
	termHeight := m.height
	if termHeight <= 0 {
		termHeight = 24
	}
	totalHeight := int(float64(termHeight) * 0.95)
	if totalHeight < 12 {
		totalHeight = 12
	}
	targetLines := totalHeight - 2
	lines := strings.Split(m.viewContent, "\n")
	if len(lines) > targetLines {
		return len(lines) - targetLines
	}
	return 0
}

