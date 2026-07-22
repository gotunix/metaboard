// SPDX-License-Identifier: GPL-3.0-or-later
// SPDX-FileCopyrightText: 2026 The MetaBoard authors

package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gotunix.net/metaboard/internal/git"
	"gotunix.net/metaboard/internal/store"
)

func (m *dashboardModel) initCreateMilestoneForm() {
	m.formScroll = 0
	m.formTitleField = createFormTextInput("Milestone Title")
	m.formTitleField.Focus()
	m.formDescField = createFormTextArea("Milestone Description...")

	m.formStatusList = []string{"BACKLOG", "ACTIVE", "COMPLETED", "CANCELLED"}
	m.formStatusSel = 0

	m.resizeFormFields()
	m.formFocusIndex = 0
	m.formTotalFields = 4
}

func (m *dashboardModel) initEditMilestoneForm(slug string) error {
	m.formScroll = 0
	mil, err := store.GetMilestone(slug)
	if err != nil {
		return err
	}

	m.editingID = mil.ID
	m.formSlug = mil.Slug

	m.formTitleField = createFormTextInput("")
	m.formTitleField.SetValue(mil.Title)
	m.formTitleField.Focus()

	m.formSlugField = createFormTextInput("")
	m.formSlugField.SetValue(mil.Slug)

	m.formStatusList = []string{"BACKLOG", "ACTIVE", "COMPLETED", "CANCELLED"}
	m.formStatusSel = 0
	for i, st := range m.formStatusList {
		if st == mil.Status {
			m.formStatusSel = i
			break
		}
	}

	m.formDescField = createFormTextArea("")
	m.formDescField.SetValue(mil.Description)

	m.formTasksField = createFormTextInput("")
	m.formTasksField.SetValue(strings.Join(mil.Tasks, ", "))

	m.resizeFormFields()
	m.formFocusIndex = 0
	m.formTotalFields = 6
	return nil
}

func (m *dashboardModel) saveCreatedMilestone() tea.Cmd {
	finalSlug, err := store.CreateMilestone(strings.TrimSpace(m.formTitle), "", strings.TrimSpace(m.formDesc))
	if err == nil {
		if m.formStatus != "BACKLOG" {
			_ = store.UpdateMilestoneStatus(finalSlug, m.formStatus)
		}
		mil, errM := store.GetMilestone(finalSlug)
		if errM == nil {
			path, _ := store.GetMilestonePath(mil.ID)
			root, _ := store.GetDataRoot()
			if git.IsGitRepo(root) {
				_ = git.Commit(root, []string{path}, fmt.Sprintf("boards: create milestone [%s] - %s", mil.Slug, mil.Title))
			}
		}
		m.statusMsg = "OK — Milestone created"
		m.statusIsError = false
	} else {
		m.statusMsg = "Error: Failed to create milestone: " + err.Error()
		m.statusIsError = true
	}
	m.state = StateDashboard
	m.refresh()
	return statusCmd()
}

func (m *dashboardModel) saveEditedMilestone() tea.Cmd {
	mil, err := store.GetMilestone(m.editingID)
	if err == nil {
		mil.Title = strings.TrimSpace(m.formTitle)
		mil.Slug = strings.TrimSpace(m.formSlug)
		mil.Status = m.formStatus
		mil.Description = strings.TrimSpace(m.formDesc)

		mil.Tasks = []string{}
		for _, id := range strings.Split(m.formTasks, ",") {
			id = strings.TrimSpace(id)
			if id != "" {
				mil.Tasks = append(mil.Tasks, id)
			}
		}

		if mil.Status == "COMPLETED" {
			if mil.CompletedAt == "" {
				mil.CompletedAt = time.Now().Format(time.RFC3339Nano)
			}
		} else {
			mil.CompletedAt = ""
		}

		_ = store.SaveMilestone(*mil)

		path, _ := store.GetMilestonePath(mil.ID)
		root, _ := store.GetDataRoot()
		if git.IsGitRepo(root) {
			_ = git.Commit(root, []string{path}, fmt.Sprintf("boards: update milestone [%s] - %s", mil.Slug, mil.Title))
		}
		m.statusMsg = "OK — Milestone updated"
		m.statusIsError = false
	} else {
		m.statusMsg = "Error: Failed to update milestone"
		m.statusIsError = true
	}
	m.state = StateDashboard
	m.refresh()
	return statusCmd()
}

func (m *dashboardModel) initCreateTaskForm() {
	m.formScroll = 0
	m.formTitleField = createFormTextInput("Task Title")
	m.formTitleField.Focus()
	m.formDescField = createFormTextArea("Task Description...")

	m.formStatusList = []string{"BACKLOG", "ACTIVE", "IN-PROGRESS", "COMPLETED", "CANCELLED"}
	m.formStatusSel = 0

	m.formPriorityList = []string{"LOW", "NORMAL", "HIGH", "CRITICAL"}
	m.formPrioritySel = 1 // default NORMAL

	m.formTypeList = []string{"FEATURE", "BUG", "CHORE", "R&D", "MAINTENANCE", "INFRA"}
	m.formTypeSel = 0

	m.formAssignedField = createFormTextInput("Unassigned")
	m.formAssignedField.SetValue("Unassigned")

	m.formTagsField = createFormTextInput("docs, backend")
	m.formTagsField.SetValue("")

	m.formChangelogList = []string{"YES", "NO"}
	m.formChangelogSel = 1 // default NO (or YES/NO select)

	tasksList, _ := store.ListTasks()
	m.formDependsOnOptions = []string{}
	m.formDependsOnSlugs = []string{}
	m.formDependsOnSel = []int{}
	m.formDependsOnCursor = 0
	for _, t := range tasksList {
		label := t.Title
		if label == "" {
			label = t.Slug
		}
		m.formDependsOnOptions = append(m.formDependsOnOptions, label)
		m.formDependsOnSlugs = append(m.formDependsOnSlugs, t.Slug)
	}

	milestones, _ := store.ListMilestones()
	m.formMilestones = []string{"None"}
	m.formMilestonesSlugs = []string{""}
	m.formMilestonesSel = 0
	for _, mil := range milestones {
		mStatus := strings.ToUpper(mil.Status)
		if mStatus == "ACTIVE" || mStatus == "BACKLOG" || mStatus == "" {
			m.formMilestones = append(m.formMilestones, mil.Title)
			m.formMilestonesSlugs = append(m.formMilestonesSlugs, mil.Slug)
		}
	}

	m.resizeFormFields()
	m.formFocusIndex = 0
	m.formTotalFields = 11
}

func (m *dashboardModel) saveCreatedTask() tea.Cmd {
	var tags []string
	for _, tag := range strings.Split(m.formTags, ",") {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			tags = append(tags, tag)
		}
	}

	finalSlug, err := store.CreateTask(
		strings.TrimSpace(m.formTitle),
		"",
		m.formPriority,
		m.formType,
		strings.TrimSpace(m.formAssignedTo),
		strings.TrimSpace(m.formDesc),
	)
	if err == nil {
		t, errT := store.GetTask(finalSlug)
		if errT == nil {
			t.Status = m.formStatus
			t.Priority = m.formPriority
			t.Type = m.formType
			t.AssignedTo = strings.TrimSpace(m.formAssignedTo)
			t.Tags = tags
			t.DependsOn = []string{}
		for _, sel := range m.formDependsOnSel {
			if sel >= 0 && sel < len(m.formDependsOnSlugs) {
				t.DependsOn = append(t.DependsOn, m.formDependsOnSlugs[sel])
			}
		}
			t.Changelog = m.formChangelog

			if t.Status == "COMPLETED" {
				t.CompletedAt = time.Now().Format(time.RFC3339Nano)
			}

			_ = store.SaveTask(*t)

			if m.formMilestoneSlug != "" {
				_ = store.LinkEntities(finalSlug, m.formMilestoneSlug)
			}

			paths := []string{}
			path, _ := store.GetTaskPath(t.ID)
			paths = append(paths, path)
			if m.formMilestoneSlug != "" {
				if mil, errM := store.GetMilestone(m.formMilestoneSlug); errM == nil {
					mPath, _ := store.GetMilestonePath(mil.ID)
					paths = append(paths, mPath)
				}
			}

			root, _ := store.GetDataRoot()
			if git.IsGitRepo(root) {
				_ = git.Commit(root, paths, fmt.Sprintf("boards: create task [%s] - %s", t.Slug, t.Title))
			}
		}
		m.statusMsg = "OK — Task created"
		m.statusIsError = false
	} else {
		m.statusMsg = "Error: Failed to create task: " + err.Error()
		m.statusIsError = true
	}
	m.state = StateDashboard
	m.refresh()
	return statusCmd()
}

func (m *dashboardModel) initEditTaskForm(slug string) error {
	m.formScroll = 0
	t, err := store.GetTask(slug)
	if err != nil {
		return err
	}

	m.editingID = t.ID
	m.formSlug = t.Slug

	m.formTitleField = createFormTextInput("")
	m.formTitleField.SetValue(t.Title)
	m.formTitleField.Focus()

	m.formSlugField = createFormTextInput("")
	m.formSlugField.SetValue(t.Slug)

	m.formStatusList = []string{"BACKLOG", "ACTIVE", "IN-PROGRESS", "COMPLETED", "CANCELLED"}
	m.formStatusSel = 0
	for i, st := range m.formStatusList {
		if st == t.Status {
			m.formStatusSel = i
			break
		}
	}

	m.formPriorityList = []string{"LOW", "NORMAL", "HIGH", "CRITICAL"}
	m.formPrioritySel = 1
	for i, pr := range m.formPriorityList {
		if pr == t.Priority {
			m.formPrioritySel = i
			break
		}
	}

	m.formTypeList = []string{"FEATURE", "BUG", "CHORE", "R&D", "MAINTENANCE", "INFRA"}
	m.formTypeSel = 0
	for i, tp := range m.formTypeList {
		if tp == t.Type {
			m.formTypeSel = i
			break
		}
	}

	m.formAssignedField = createFormTextInput("")
	m.formAssignedField.SetValue(t.AssignedTo)

	m.formTagsField = createFormTextInput("")
	m.formTagsField.SetValue(strings.Join(t.Tags, ", "))

	m.formChangelogList = []string{"YES", "NO"}
	m.formChangelogSel = 1
	if t.Changelog {
		m.formChangelogSel = 0
	}

	m.formDescField = createFormTextArea("")
	m.formDescField.SetValue(t.Description)

	m.formMilestoneSlug = ""
	milestones, _ := store.ListMilestones()
	for _, mil := range milestones {
		for _, tID := range mil.Tasks {
			if tID == t.ID {
				m.formMilestoneSlug = mil.Slug
				break
			}
		}
	}

	m.formMilestones = []string{"None"}
	m.formMilestonesSlugs = []string{""}
	m.formMilestonesSel = 0
	for _, mil := range milestones {
		m.formMilestones = append(m.formMilestones, mil.Title)
		m.formMilestonesSlugs = append(m.formMilestonesSlugs, mil.Slug)
		if mil.Slug == m.formMilestoneSlug {
			m.formMilestonesSel = len(m.formMilestones) - 1
		}
	}

	m.formPRsField = createFormTextInput("")
	m.formPRsField.SetValue(strings.Join(t.PullRequests, ", "))

	m.resizeFormFields()
	m.formFocusIndex = 0
	m.formTotalFields = 13
	return nil
}

func (m *dashboardModel) saveEditedTask() tea.Cmd {
	t, err := store.GetTask(m.editingID)
	if err == nil {
		oldMilestoneSlug := ""
		milestones, _ := store.ListMilestones()
		for _, mil := range milestones {
			for _, tID := range mil.Tasks {
				if tID == t.ID {
					oldMilestoneSlug = mil.Slug
					break
				}
			}
		}

		t.Title = strings.TrimSpace(m.formTitle)
		t.Slug = strings.TrimSpace(m.formSlug)
		t.Status = m.formStatus
		t.Priority = m.formPriority
		t.Type = m.formType
		t.AssignedTo = strings.TrimSpace(m.formAssignedTo)
		t.Description = strings.TrimSpace(m.formDesc)
		t.Changelog = m.formChangelog

		t.Tags = []string{}
		for _, tag := range strings.Split(m.formTags, ",") {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				t.Tags = append(t.Tags, tag)
			}
		}

		t.DependsOn = []string{}
		for _, sel := range m.formDependsOnSel {
			if sel >= 0 && sel < len(m.formDependsOnSlugs) {
				t.DependsOn = append(t.DependsOn, m.formDependsOnSlugs[sel])
			}
		}

		t.PullRequests = []string{}
		for _, id := range strings.Split(m.formPRs, ",") {
			id = strings.TrimSpace(id)
			if id != "" {
				t.PullRequests = append(t.PullRequests, id)
			}
		}

		if t.Status == "COMPLETED" {
			if t.CompletedAt == "" {
				t.CompletedAt = time.Now().Format(time.RFC3339Nano)
			}
		} else {
			t.CompletedAt = ""
		}

		_ = store.SaveTask(*t)

		paths := []string{}
		path, _ := store.GetTaskPath(t.ID)
		paths = append(paths, path)

		if oldMilestoneSlug != "" && oldMilestoneSlug != m.formMilestoneSlug {
			_ = store.UnlinkEntity(t.Slug)
			if oldMil, errM := store.GetMilestone(oldMilestoneSlug); errM == nil {
				omPath, _ := store.GetMilestonePath(oldMil.ID)
				paths = append(paths, omPath)
			}
		}

		if m.formMilestoneSlug != "" && m.formMilestoneSlug != oldMilestoneSlug {
			_ = store.LinkEntities(t.Slug, m.formMilestoneSlug)
			if newMil, errM := store.GetMilestone(m.formMilestoneSlug); errM == nil {
				nmPath, _ := store.GetMilestonePath(newMil.ID)
				paths = append(paths, nmPath)
			}
		}

		root, _ := store.GetDataRoot()
		if git.IsGitRepo(root) {
			_ = git.Commit(root, paths, fmt.Sprintf("boards: update task [%s] - %s", t.Slug, t.Title))
		}
		m.statusMsg = "OK — Task updated"
		m.statusIsError = false
	} else {
		m.statusMsg = "Error: Failed to update task"
		m.statusIsError = true
	}
	m.state = StateDashboard
	m.refresh()
	return statusCmd()
}

func (m *dashboardModel) initCreatePRForm() {
	m.formScroll = 0
	m.formTitleField = createFormTextInput("PR Description")
	m.formTitleField.Focus()

	m.formStatusList = []string{"DRAFT", "OPEN", "MERGED", "CLOSED", "REJECTED"}
	m.formStatusSel = 1 // default OPEN

	m.formHeadField = createFormTextInput("feature-branch")
	m.formBaseField = createFormTextInput("main")
	m.formSrcField = createFormTextInput("github.com/org/repo")
	m.formDestField = createFormTextInput("github.com/org/repo")

	m.formTasksField = createFormTextInput("Select linked tasks")
	m.formTasksField.SetValue("")
	m.formTaskOptions = []string{}
	m.formTaskIDs = []string{}
	m.formTaskSel = []int{}
	m.formTaskCursor = 0

	tasks, _ := store.ListTasks()
	for _, t := range tasks {
		label := t.Title
		if label == "" {
			label = t.Slug
		}
		m.formTaskOptions = append(m.formTaskOptions, label)
		m.formTaskIDs = append(m.formTaskIDs, t.ID)
	}

	m.resizeFormFields()
	m.formFocusIndex = 0
	m.formTotalFields = 8
}

func (m *dashboardModel) saveCreatedPR() tea.Cmd {
	finalSlug, err := store.CreatePullRequest(
		"",
		strings.TrimSpace(m.formBaseBranch),
		strings.TrimSpace(m.formHeadBranch),
		strings.TrimSpace(m.formSrcRepo),
		strings.TrimSpace(m.formDestRepo),
		strings.TrimSpace(m.formDesc),
	)
	if err == nil {
		pr, errPR := store.GetPullRequest(finalSlug)
		if errPR == nil {
			pr.Status = m.formStatus
			pr.SourceRepo = strings.TrimSpace(m.formSrcRepo)
			pr.DestRepo = strings.TrimSpace(m.formDestRepo)

			if pr.Status == "MERGED" {
				pr.CompletedAt = time.Now().Format(time.RFC3339Nano)
			}

			_ = store.SavePullRequest(*pr)

			paths := []string{}
			path, _ := store.GetPullRequestPath(pr.ID)
			paths = append(paths, path)
			// Also stage the markdown sidecar created by CreatePullRequest → EnsurePullRequestMarkdown
			if mdPath, errMD := store.GetPullRequestMarkdownPath(pr.ID); errMD == nil {
				paths = append(paths, mdPath)
			}

			for _, id := range strings.Split(m.formTasks, ",") {
				id = strings.TrimSpace(id)
				if id != "" {
					_ = store.LinkEntities(finalSlug, id)
					if t, errT := store.GetTask(id); errT == nil {
						tPath, _ := store.GetTaskPath(t.ID)
						paths = append(paths, tPath)
					}
				}
			}

			root, _ := store.GetDataRoot()
			if git.IsGitRepo(root) {
				_ = git.Commit(root, paths, fmt.Sprintf("boards: create pullrequest [%s] - %s -> %s", pr.Slug, pr.HeadBranch, pr.BaseBranch))
			}
		}
		m.statusMsg = "OK — Pull request created"
		m.statusIsError = false
	} else {
		m.statusMsg = "Error: Failed to create pull request: " + err.Error()
		m.statusIsError = true
	}
	m.state = StateDashboard
	m.refresh()
	return statusCmd()
}

func (m *dashboardModel) initEditPRForm(slug string) error {
	m.formScroll = 0
	pr, err := store.GetPullRequest(slug)
	if err != nil {
		return err
	}

	m.editingID = pr.ID
	m.formSlug = pr.Slug

	m.formTitleField = createFormTextInput("")
	m.formTitleField.SetValue(pr.Description)
	m.formTitleField.Focus()

	m.formSlugField = createFormTextInput("")
	m.formSlugField.SetValue(pr.Slug)

	m.formStatusList = []string{"DRAFT", "OPEN", "MERGED", "CLOSED", "REJECTED"}
	m.formStatusSel = 0
	for i, st := range m.formStatusList {
		if st == pr.Status {
			m.formStatusSel = i
			break
		}
	}

	m.formHeadField = createFormTextInput("")
	m.formHeadField.SetValue(pr.HeadBranch)

	m.formBaseField = createFormTextInput("")
	m.formBaseField.SetValue(pr.BaseBranch)

	m.formSrcField = createFormTextInput("")
	m.formSrcField.SetValue(pr.SourceRepo)

	m.formDestField = createFormTextInput("")
	m.formDestField.SetValue(pr.DestRepo)

	m.formTasksField = createFormTextInput("")
	m.formTasksField.SetValue("")
	m.formTaskOptions = []string{}
	m.formTaskIDs = []string{}
	m.formTaskSel = []int{}
	m.formTaskCursor = 0

	tasks, _ := store.ListTasks()
	if len(tasks) > 0 {
		selected := make(map[string]bool)
		for _, id := range pr.Tasks {
			selected[id] = true
		}
		for i, t := range tasks {
			label := t.Title
			if label == "" {
				label = t.Slug
			}
			m.formTaskOptions = append(m.formTaskOptions, label)
			m.formTaskIDs = append(m.formTaskIDs, t.ID)
			if selected[t.ID] {
				m.formTaskSel = append(m.formTaskSel, i)
			}
		}
	}

	m.resizeFormFields()
	m.formFocusIndex = 0
	m.formTotalFields = 9
	return nil
}

func (m *dashboardModel) saveEditedPR() tea.Cmd {
	pr, err := store.GetPullRequest(m.editingID)
	if err == nil {
		oldTasks := make(map[string]bool)
		for _, tID := range pr.Tasks {
			oldTasks[tID] = true
		}

		pr.HeadBranch = strings.TrimSpace(m.formHeadBranch)
		pr.BaseBranch = strings.TrimSpace(m.formBaseBranch)
		pr.Status = m.formStatus
		pr.SourceRepo = strings.TrimSpace(m.formSrcRepo)
		pr.DestRepo = strings.TrimSpace(m.formDestRepo)
		pr.Description = strings.TrimSpace(m.formDesc)

		newTasks := make(map[string]bool)
		pr.Tasks = []string{}
		for _, id := range strings.Split(m.formTasks, ",") {
			id = strings.TrimSpace(id)
			if id != "" {
				pr.Tasks = append(pr.Tasks, id)
				newTasks[id] = true
			}
		}

		if pr.Status == "MERGED" {
			if pr.CompletedAt == "" {
				pr.CompletedAt = time.Now().Format(time.RFC3339Nano)
			}
		} else {
			pr.CompletedAt = ""
		}

		_ = store.SavePullRequest(*pr)

		paths := []string{}
		path, _ := store.GetPullRequestPath(pr.ID)
		paths = append(paths, path)

		for oldT := range oldTasks {
			if !newTasks[oldT] {
				_ = store.UnlinkEntity(oldT)
				if t, errT := store.GetTask(oldT); errT == nil {
					tPath, _ := store.GetTaskPath(t.ID)
					paths = append(paths, tPath)
				}
			}
		}

		for newT := range newTasks {
			if !oldTasks[newT] {
				_ = store.LinkEntities(pr.Slug, newT)
				if t, errT := store.GetTask(newT); errT == nil {
					tPath, _ := store.GetTaskPath(t.ID)
					paths = append(paths, tPath)
				}
			}
		}

		root, _ := store.GetDataRoot()
		if git.IsGitRepo(root) {
			_ = git.Commit(root, paths, fmt.Sprintf("boards: update pullrequest [%s] - %s -> %s", pr.Slug, pr.HeadBranch, pr.BaseBranch))
		}
		m.statusMsg = "OK — Pull request updated"
		m.statusIsError = false
	} else {
		m.statusMsg = "Error: Failed to update pull request"
		m.statusIsError = true
	}
	m.state = StateDashboard
	m.refresh()
	return statusCmd()
}


func (m *dashboardModel) getFieldInfo(idx int) (inputType string, label string, helpText string) {
	switch m.state {
	case StateCreateMilestone:
		switch idx {
		case 0:
			return "textinput", "Milestone Title", "Enter the title of the milestone"
		case 1:
			return "select", "Status", "Select status using ← / →"
		case 2:
			return "textarea", "Description", "Enter milestone description"
		case 3:
			return "button", "Submit", "Press Enter to submit"
		}
	case StateEditMilestone:
		switch idx {
		case 0:
			return "textinput", "Milestone Title", "Enter the title of the milestone"
		case 1:
			return "textinput", "Slug", "Enter unique slug (URL-friendly identifier)"
		case 2:
			return "select", "Status", "Select status using ← / →"
		case 3:
			return "textarea", "Description", "Enter milestone description"
		case 4:
			return "tasks-list", "Linked Tasks", "Comma-separated list of task IDs/slugs"
		case 5:
			return "button", "Submit", "Press Enter to save changes"
		}
	case StateCreateTask:
		switch idx {
		case 0:
			return "textinput", "Task Title", "Enter the title of the task"
		case 1:
			return "select", "Status", "Select status using ← / →"
		case 2:
			return "priority-select", "Priority", "Select priority using ← / →"
		case 3:
			return "type-select", "Type", "Select task type using ← / →"
		case 4:
			return "assigned-input", "Assigned To", "Username of the assignee"
		case 5:
			return "tags-input", "Tags", "Comma-separated tags"
		case 6:
			return "changelog-select", "Changelog", "Include in changelog?"
		case 7:
			return "textarea", "Description", "Enter task description"
		case 8:
			return "depends-select", "Dependencies", "Select dependent tasks using space to toggle"
		case 9:
			return "milestone-select", "Link to Milestone", "Select milestone to link using ← / →"
		case 10:
			return "button", "Submit", "Press Enter to submit"
		}
	case StateEditTask:
		switch idx {
		case 0:
			return "textinput", "Task Title", "Enter the title of the task"
		case 1:
			return "textinput", "Slug", "Enter unique slug"
		case 2:
			return "select", "Status", "Select status using ← / →"
		case 3:
			return "priority-select", "Priority", "Select priority using ← / →"
		case 4:
			return "type-select", "Type", "Select task type using ← / →"
		case 5:
			return "assigned-input", "Assigned To", "Username of the assignee"
		case 6:
			return "tags-input", "Tags", "Comma-separated tags"
		case 7:
			return "changelog-select", "Changelog", "Include in changelog?"
		case 8:
			return "textarea", "Description", "Enter task description"
		case 9:
			return "depends-select", "Dependencies", "Select dependent tasks using space to toggle"
		case 10:
			return "milestone-select", "Link to Milestone", "Select milestone to link using ← / →"
		case 11:
			return "prs-list", "Linked Pull Requests", "Comma-separated list of PR IDs/slugs"
		case 12:
			return "button", "Submit", "Press Enter to save changes"
		}
	case StateCreatePR:
		switch idx {
		case 0:
			return "textinput", "PR Description", "Enter the description/title of the Pull Request"
		case 1:
			return "select", "Status", "Select status using ← / →"
		case 2:
			return "head-branch", "Head Branch", "Name of the feature/source branch"
		case 3:
			return "base-branch", "Base Branch", "Name of the target branch (e.g. main)"
		case 4:
			return "src-repo", "Source Repo", "Repository URL/path of source branch"
		case 5:
			return "dest-repo", "Destination Repo", "Repository URL/path of base branch"
		case 6:
			return "task-select", "Associated Tasks", "Select tasks using ↑ / ↓ and press space to toggle"
		case 7:
			return "button", "Submit", "Press Enter to submit"
		}
	case StateEditPR:
		switch idx {
		case 0:
			return "textinput", "PR Description", "Enter the description/title of the Pull Request"
		case 1:
			return "textinput", "Slug", "Enter unique slug"
		case 2:
			return "select", "Status", "Select status using ← / →"
		case 3:
			return "head-branch", "Head Branch", "Name of the feature/source branch"
		case 4:
			return "base-branch", "Base Branch", "Name of the target branch (e.g. main)"
		case 5:
			return "src-repo", "Source Repo", "Repository URL/path of source branch"
		case 6:
			return "dest-repo", "Destination Repo", "Repository URL/path of base branch"
		case 7:
			return "task-select", "Associated Tasks", "Select tasks using ↑ / ↓ and press space to toggle"
		case 8:
			return "button", "Submit", "Press Enter to save changes"
		}
	}
	return "", "", ""
}

func (m *dashboardModel) getActiveFieldInfo() (inputType string, label string, helpText string) {
	return m.getFieldInfo(m.formFocusIndex)
}

func (m *dashboardModel) getFieldModel(idx int) (any, bool) {
	switch m.state {
	case StateEditMilestone:
		if idx == 1 {
			return &m.formSlugField, true
		}
		if idx == 4 {
			return &m.formTasksField, true
		}
	case StateCreateTask:
		if idx == 4 {
			return &m.formAssignedField, true
		}
		if idx == 5 {
			return &m.formTagsField, true
		}
	case StateEditTask:
		if idx == 1 {
			return &m.formSlugField, true
		}
		if idx == 5 {
			return &m.formAssignedField, true
		}
		if idx == 6 {
			return &m.formTagsField, true
		}
		if idx == 12 {
			return &m.formPRsField, true
		}
	case StateCreatePR:
		if idx == 2 {
			return &m.formHeadField, true
		}
		if idx == 3 {
			return &m.formBaseField, true
		}
		if idx == 4 {
			return &m.formSrcField, true
		}
		if idx == 5 {
			return &m.formDestField, true
		}
		if idx == 6 {
			return nil, false
		}
	case StateEditPR:
		if idx == 1 {
			return &m.formSlugField, true
		}
		if idx == 3 {
			return &m.formHeadField, true
		}
		if idx == 4 {
			return &m.formBaseField, true
		}
		if idx == 5 {
			return &m.formSrcField, true
		}
		if idx == 6 {
			return &m.formDestField, true
		}
		if idx == 7 {
			return nil, false
		}
	}

	it, _, _ := m.getFieldInfo(idx)
	switch it {
	case "textinput", "assigned-input", "head-branch", "base-branch", "src-repo", "dest-repo", "stories-list", "tasks-list", "prs-list", "tags-input":
		return &m.formTitleField, true
	case "textarea":
		return &m.formDescField, false
	}
	return &m.formTitleField, true
}

func (m *dashboardModel) getActiveFieldModel() (any, bool) {
	return m.getFieldModel(m.formFocusIndex)
}

// formScrollToFocus adjusts the form scroll offset so the currently focused
// field is visible in the viewport. Uses a rough estimate — the user can
// fine-tune with PgDn/PgUp.
func (m *dashboardModel) formScrollToFocus() {
	if m.formTotalFields <= 1 {
		m.formScroll = 0
		return
	}
	// Estimate total form line count: each field ~5 lines on average
	estLinesPerField := 5
	estTotalLines := m.formTotalFields * estLinesPerField
	// Use the actual View's targetLines computation; we estimate conservatively
	estTarget := m.height - 2 // rough match for targetLines in View()
	if estTarget < 6 {
		estTarget = 6
	}
	if estTotalLines <= estTarget {
		m.formScroll = 0
		return
	}
	fieldStart := m.formFocusIndex * estLinesPerField
	maxScroll := estTotalLines - estTarget
	desired := fieldStart - estTarget/3
	if desired < 0 {
		desired = 0
	}
	if desired > maxScroll {
		desired = maxScroll
	}
	m.formScroll = desired
}

func (m *dashboardModel) updateFormFieldsFocus() {
	m.formTitleField.Blur()
	m.formSlugField.Blur()
	m.formDescField.Blur()
	m.formAssignedField.Blur()
	m.formHeadField.Blur()
	m.formBaseField.Blur()
	m.formSrcField.Blur()
	m.formDestField.Blur()
	m.formTasksField.Blur()
	m.formMilestoneField.Blur()
	m.formTypeField.Blur()
	m.formTagsField.Blur()
	m.formPRsField.Blur()

	it, _, _ := m.getActiveFieldInfo()
	if it != "select" && it != "priority-select" && it != "milestone-select" && it != "task-select" && it != "depends-select" && it != "button" {
		modelPtr, isTextInput := m.getActiveFieldModel()
		if isTextInput {
			ti := modelPtr.(*textinput.Model)
			if m.formFocusIndex != m.formLastFocusIndex {
				ti.SetCursor(len(ti.Value()))
				m.formLastFocusIndex = m.formFocusIndex
			}
			ti.Focus()
		} else {
			ta := modelPtr.(*textarea.Model)
			ta.Focus()
		}
	}
}

func (m *dashboardModel) handleSelectMove(dir int) {
	it, _, _ := m.getActiveFieldInfo()
	switch it {
	case "select":
		m.formStatusSel += dir
		if m.formStatusSel < 0 {
			m.formStatusSel = len(m.formStatusList) - 1
		} else if m.formStatusSel >= len(m.formStatusList) {
			m.formStatusSel = 0
		}
	case "priority-select":
		m.formPrioritySel += dir
		if m.formPrioritySel < 0 {
			m.formPrioritySel = len(m.formPriorityList) - 1
		} else if m.formPrioritySel >= len(m.formPriorityList) {
			m.formPrioritySel = 0
		}
	case "type-select":
		m.formTypeSel += dir
		if m.formTypeSel < 0 {
			m.formTypeSel = len(m.formTypeList) - 1
		} else if m.formTypeSel >= len(m.formTypeList) {
			m.formTypeSel = 0
		}
	case "changelog-select":
		m.formChangelogSel += dir
		if m.formChangelogSel < 0 {
			m.formChangelogSel = len(m.formChangelogList) - 1
		} else if m.formChangelogSel >= len(m.formChangelogList) {
			m.formChangelogSel = 0
		}
	case "milestone-select":
		m.formMilestonesSel += dir
		if m.formMilestonesSel < 0 {
			m.formMilestonesSel = len(m.formMilestones) - 1
		} else if m.formMilestonesSel >= len(m.formMilestones) {
			m.formMilestonesSel = 0
		}
	case "task-select":
		if len(m.formTaskOptions) == 0 {
			return
		}
		m.formTaskCursor += dir
		if m.formTaskCursor < 0 {
			m.formTaskCursor = len(m.formTaskOptions) - 1
		} else if m.formTaskCursor >= len(m.formTaskOptions) {
			m.formTaskCursor = 0
		}
	case "depends-select":
		if len(m.formDependsOnOptions) == 0 {
			return
		}
		m.formDependsOnCursor += dir
		if m.formDependsOnCursor < 0 {
			m.formDependsOnCursor = len(m.formDependsOnOptions) - 1
		} else if m.formDependsOnCursor >= len(m.formDependsOnOptions) {
			m.formDependsOnCursor = 0
		}
	}
}

func (m *dashboardModel) toggleTaskSelection() {
	if len(m.formTaskOptions) == 0 {
		return
	}
	cur := m.formTaskCursor
	selected := false
	for i, sel := range m.formTaskSel {
		if sel == cur {
			m.formTaskSel = append(m.formTaskSel[:i], m.formTaskSel[i+1:]...)
			selected = true
			break
		}
	}
	if !selected {
		m.formTaskSel = append(m.formTaskSel, cur)
	}
}

func (m *dashboardModel) toggleDependsOnSelection() {
	if len(m.formDependsOnOptions) == 0 {
		return
	}
	cur := m.formDependsOnCursor
	selected := false
	for i, sel := range m.formDependsOnSel {
		if sel == cur {
			m.formDependsOnSel = append(m.formDependsOnSel[:i], m.formDependsOnSel[i+1:]...)
			selected = true
			break
		}
	}
	if !selected {
		m.formDependsOnSel = append(m.formDependsOnSel, cur)
	}
}

func (m *dashboardModel) renderTaskSelectionList(isFocused bool) string {
	if len(m.formTaskOptions) == 0 {
		return "  (no tasks available)"
	}

	var builder strings.Builder
	for i, opt := range m.formTaskOptions {
		selected := false
		for _, sel := range m.formTaskSel {
			if sel == i {
				selected = true
				break
			}
		}

		cursor := isFocused && m.formTaskCursor == i
		icon := " "
		if selected {
			icon = "x"
		}

		line := fmt.Sprintf("  [%s] %s", icon, opt)
		if cursor {
			line = lipgloss.NewStyle().Background(catMochaGreen).Foreground(catMochaBase).Bold(true).Render(line)
		}
		builder.WriteString(line)
		if i < len(m.formTaskOptions)-1 {
			builder.WriteString("\n")
		}
	}
	return builder.String()
}

func (m *dashboardModel) renderDependsOnSelectionList(isFocused bool) string {
	if len(m.formDependsOnOptions) == 0 {
		return "  (no tasks available)"
	}

	var builder strings.Builder
	for i, opt := range m.formDependsOnOptions {
		selected := false
		for _, sel := range m.formDependsOnSel {
			if sel == i {
				selected = true
				break
			}
		}

		cursor := isFocused && m.formDependsOnCursor == i
		icon := " "
		if selected {
			icon = "x"
		}

		line := fmt.Sprintf("  [%s] %s", icon, opt)
		if cursor {
			line = lipgloss.NewStyle().Background(catMochaGreen).Foreground(catMochaBase).Bold(true).Render(line)
		}
		builder.WriteString(line)
		if i < len(m.formDependsOnOptions)-1 {
			builder.WriteString("\n")
		}
	}
	return builder.String()
}

func (m *dashboardModel) submitForm() tea.Cmd {
	m.formTitle = strings.TrimSpace(m.formTitleField.Value())
	m.formDesc = strings.TrimSpace(m.formDescField.Value())
	m.formSlug = strings.TrimSpace(m.formSlugField.Value())

	if m.state == StateCreatePR || m.state == StateEditPR {
		m.formDesc = strings.TrimSpace(m.formTitleField.Value())
	}

	if len(m.formStatusList) > 0 && m.formStatusSel >= 0 && m.formStatusSel < len(m.formStatusList) {
		m.formStatus = m.formStatusList[m.formStatusSel]
	}
	if len(m.formPriorityList) > 0 && m.formPrioritySel >= 0 && m.formPrioritySel < len(m.formPriorityList) {
		m.formPriority = m.formPriorityList[m.formPrioritySel]
	}
	if len(m.formTypeList) > 0 && m.formTypeSel >= 0 && m.formTypeSel < len(m.formTypeList) {
		m.formType = m.formTypeList[m.formTypeSel]
	}
	if len(m.formChangelogList) > 0 && m.formChangelogSel >= 0 && m.formChangelogSel < len(m.formChangelogList) {
		m.formChangelog = m.formChangelogList[m.formChangelogSel] == "YES"
	}

	m.formAssignedTo = strings.TrimSpace(m.formAssignedField.Value())
	m.formTags = strings.TrimSpace(m.formTagsField.Value())
	m.formHeadBranch = strings.TrimSpace(m.formHeadField.Value())
	m.formBaseBranch = strings.TrimSpace(m.formBaseField.Value())
	m.formSrcRepo = strings.TrimSpace(m.formSrcField.Value())
	m.formDestRepo = strings.TrimSpace(m.formDestField.Value())

	m.formTasks = strings.TrimSpace(m.formTasksField.Value())
	m.formPRs = strings.TrimSpace(m.formPRsField.Value())

	if len(m.formTaskIDs) > 0 && m.formTaskSel != nil {
		m.formTasks = ""
		for i, sel := range m.formTaskSel {
			if sel >= 0 && sel < len(m.formTaskIDs) {
				if i > 0 {
					m.formTasks += ", "
				}
				m.formTasks += m.formTaskIDs[sel]
			}
		}
	}

	// Resolve depends-on from multi-select
	m.formDependsOn = []string{}
	for _, sel := range m.formDependsOnSel {
		if sel >= 0 && sel < len(m.formDependsOnSlugs) {
			m.formDependsOn = append(m.formDependsOn, m.formDependsOnSlugs[sel])
		}
	}

	if len(m.formMilestonesSlugs) > 0 && m.formMilestonesSel >= 0 && m.formMilestonesSel < len(m.formMilestonesSlugs) {
		m.formMilestoneSlug = m.formMilestonesSlugs[m.formMilestonesSel]
	}

	// After saving, show all items so the edited item is visible regardless of new status
	m.filter = "all"

	var cmd tea.Cmd
	switch m.state {
	case StateCreateMilestone:
		cmd = m.saveCreatedMilestone()
	case StateEditMilestone:
		cmd = m.saveEditedMilestone()
	case StateCreateTask:
		cmd = m.saveCreatedTask()
	case StateEditTask:
		cmd = m.saveEditedTask()
	case StateCreatePR:
		cmd = m.saveCreatedPR()
	case StateEditPR:
		cmd = m.saveEditedPR()
	}
	return cmd
}

func renderFormField(label string, inputView string, isFocused bool, helpText string) string {
	var labelStr string
	if isFocused {
		labelStr = focusedLabelStyle.Render("▶ " + label)
	} else {
		labelStr = formLabelStyle.Render("  " + label)
	}

	res := labelStr + "\n" + inputView
	if isFocused && helpText != "" {
		res += "\n" + helperStyle.Render("    "+helpText)
	}
	return res
}

func renderFormSelectOptions(options []string, activeIndex int, isFieldFocused bool) string {
	var builder strings.Builder
	builder.WriteString("  ")

	for i, opt := range options {
		isHovered := isFieldFocused && activeIndex == i

		var icon string
		if activeIndex == i {
			icon = "⊙"
		} else {
			icon = "○"
		}

		optionText := fmt.Sprintf(" %s %s ", icon, opt)
		var styledOption string

		if isHovered {
			styledOption = lipgloss.NewStyle().
				Foreground(catMochaBase).
				Background(catMochaGreen).
				Bold(true).
				Render(optionText)
		} else if activeIndex == i {
			styledOption = lipgloss.NewStyle().
				Foreground(catMochaGreen).
				Bold(true).
				Render(optionText)
		} else {
			styledOption = lipgloss.NewStyle().
				Foreground(catMochaSubtext).
				Render(optionText)
		}

		builder.WriteString(styledOption)
		if i < len(options)-1 {
			builder.WriteString("\n  ")
		}
	}

	return builder.String()
}

func renderFormSubmitButton(isFocused bool) string {
	var button string
	if isFocused {
		button = activeButtonStyle.Render("SUBMIT FORM")
	} else {
		button = buttonStyle.Render("SUBMIT FORM")
	}
	return "  " + button
}

func (m *dashboardModel) renderFormView(width int) []string {
	var sections []string
	activeIndex := m.formFocusIndex

	renderIdxField := func(idx int) string {
		it, label, help := m.getFieldInfo(idx)
		isFocused := activeIndex == idx

		var view string
		switch it {
		case "textinput", "assigned-input", "tags-input", "prs-list", "tasks-list", "stories-list":
			ti, _ := m.getFieldModel(idx)
			view = renderFormField(label, ti.(*textinput.Model).View(), isFocused, help)
		case "textarea":
			ta, _ := m.getFieldModel(idx)
			view = renderFormField(label, ta.(*textarea.Model).View(), isFocused, help)
		case "select":
			optsView := renderFormSelectOptions(m.formStatusList, m.formStatusSel, isFocused)
			view = renderFormField(label, optsView, isFocused, help)
		case "priority-select":
			optsView := renderFormSelectOptions(m.formPriorityList, m.formPrioritySel, isFocused)
			view = renderFormField(label, optsView, isFocused, help)
		case "type-select":
			optsView := renderFormSelectOptions(m.formTypeList, m.formTypeSel, isFocused)
			view = renderFormField(label, optsView, isFocused, help)
		case "changelog-select":
			optsView := renderFormSelectOptions(m.formChangelogList, m.formChangelogSel, isFocused)
			view = renderFormField(label, optsView, isFocused, help)
		case "milestone-select":
			optsView := renderFormSelectOptions(m.formMilestones, m.formMilestonesSel, isFocused)
			view = renderFormField(label, optsView, isFocused, help)
		case "head-branch", "base-branch", "src-repo", "dest-repo":
			ti, _ := m.getFieldModel(idx)
			view = renderFormField(label, ti.(*textinput.Model).View(), isFocused, help)
		case "task-select":
			view = renderFormField(label, m.renderTaskSelectionList(isFocused), isFocused, help)
		case "depends-select":
			view = renderFormField(label, m.renderDependsOnSelectionList(isFocused), isFocused, help)
		case "button":
			view = "\n" + renderFormSubmitButton(isFocused)
		}
		return view
	}

	for i := 0; i < m.formTotalFields; i++ {
		sections = append(sections, renderIdxField(i))
	}

	return sections
}

