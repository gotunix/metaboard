// SPDX-License-Identifier: GPL-3.0-or-later
// SPDX-FileCopyrightText: 2026 The MetaBoard authors

package ui

import (
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"gotunix.net/metaboard/internal/store"
)



type ViewState int

const (
	StateDashboard ViewState = iota
	StateCreateMilestone
	StateEditMilestone
	StateCreateTask
	StateEditTask
	StateCreatePR
	StateEditPR
	StateViewMilestone
	StateViewTask
	StateViewPR
	StateViewVersion
	StateViewChangelog
	StatePromptInit
	StatePromptInitLocation
	StatePromptPush
	StateConfirmDelete
	StateConfirmUnlink
)

type dashboardLine struct {
	text         string
	isSelectable bool
	itemType     string
	slug         string
}

type FlatItemType int

const (
	TypeMilestone FlatItemType = iota
	TypeTask
	TypePR
)

type FlatRow struct {
	Type         FlatItemType
	Title        string
	ItemType     string
	Slug         string
	Status       string
	Progress     float64
	ProgressText string
	Depth        int
	Expanded     bool
}

type dashboardModel struct {
	mode          string
	target        string
	filter        string
	lines         []dashboardLine
	cursor        int
	viewStart     int
	previewScroll int
	width         int
	height        int
	expanded      map[string]bool
	widths        []int
	flatRows      []FlatRow

	// Sub-states
	state ViewState

	// Form data storage
	editingID string
	formSlug  string // Slug of the entity being edited

	formTitle  string
	formStatus string
	formDesc   string
	formTasks  string

	// Story fields
	formMilestoneSlug string

	// Task fields
	formPriority   string
	formType       string
	formAssignedTo string
	formTags       string
	formDependsOn  []string
	formChangelog bool
	formPRs       string

	// Pull Request fields
	formHeadBranch string
	formBaseBranch string
	formSrcRepo    string
	formDestRepo   string

	viewSlug         string
	viewContent      string // cached rendered content for detail views to avoid O(n) scan in View()
	changelogMessage string
	initError        string
	initLocation     string // chosen init path: "." or "boards"

	pushBranchField textinput.Model // branch name input for push dialog
	pushError       string           // last push error message

	statusMsg        string // transient status/error shown in the bottom bar
	statusIsError    bool   // true = red, false = green
	pendingPlanPath  string // plan file path to commit after editor closes (absolute)
	pendingPlanSlug  string // task slug for the commit message
	pendingPlanRoot  string // absolute data root captured at editor-launch time

	deleteViewSlug string
	deleteViewType string
	deleteErrorMsg string

	unlinkViewSlug string
	unlinkViewType string
	unlinkErrorMsg string

	// Custom Form Models
	formTitleField     textinput.Model
	formSlugField      textinput.Model
	formDescField      textarea.Model
	formAssignedField  textinput.Model
	formHeadField      textinput.Model
	formBaseField      textinput.Model
	formSrcField       textinput.Model
	formDestField      textinput.Model
	formTasksField     textinput.Model
	formMilestoneField textinput.Model
	formTypeField      textinput.Model
	formTagsField      textinput.Model
	formPRsField       textinput.Model

	formDependsOnOptions []string
	formDependsOnSlugs   []string
	formDependsOnSel     []int
	formDependsOnCursor  int

	formTaskOptions []string
	formTaskIDs     []string
	formTaskSel     []int
	formTaskCursor  int

	// Custom Form States
	formFocusIndex    int
	formLastFocusIndex int
	formTotalFields   int
	formScroll       int // scrolling offset for tall forms
	formStatusList   []string
	formStatusSel    int
	formPriorityList []string
	formPrioritySel  int

	formMilestones     []string
	formMilestonesSlugs []string
	formMilestonesSel  int
	formTypeList      []string
	formTypeSel       int
	formChangelogList []string
	formChangelogSel  int
}

func formFieldWidth(termWidth int) int {
	if termWidth <= 0 {
		return 50
	}
	w := termWidth - 50
	if w < 50 {
		w = 50
	}
	if w > 100 {
		w = 100
	}
	return w
}

func (m *dashboardModel) resizeFormFields() {
	fw := formFieldWidth(m.width)
	m.formTitleField.Width = fw
	m.formDescField.SetWidth(fw)
	m.formSlugField.Width = fw
	m.formAssignedField.Width = fw
	m.formTagsField.Width = fw
	m.formHeadField.Width = fw
	m.formBaseField.Width = fw
	m.formSrcField.Width = fw
	m.formDestField.Width = fw
	m.formTasksField.Width = fw
	m.formMilestoneField.Width = fw
	m.formTypeField.Width = fw
	m.formPRsField.Width = fw

	m.resetFormFieldCursors()
}

func (m *dashboardModel) resetFormFieldCursors() {
	inputs := []*textinput.Model{
		&m.formTitleField, &m.formSlugField, &m.formAssignedField,
		&m.formTagsField, &m.formHeadField, &m.formBaseField,
		&m.formSrcField, &m.formDestField,
		&m.formTasksField, &m.formMilestoneField,
		&m.formTypeField, &m.formPRsField,
	}
	for _, f := range inputs {
		f.SetCursor(len(f.Value()))
	}
}

func createFormTextInput(placeholder string) textinput.Model {
	t := textinput.New()
	t.Placeholder = placeholder
	t.CharLimit = 64
	t.Width = 30
	t.Prompt = "  "
	t.PromptStyle = lipgloss.NewStyle().Foreground(catMochaBlue)
	t.TextStyle = lipgloss.NewStyle().Foreground(catMochaText)
	t.Cursor.Style = lipgloss.NewStyle().Foreground(catMochaMauve)
	return t
}

func createFormTextArea(placeholder string) textarea.Model {
	t := textarea.New()
	t.Placeholder = placeholder
	t.CharLimit = 128
	t.SetWidth(40)
	t.SetHeight(3)
	t.ShowLineNumbers = false
	t.Prompt = "  "
	t.Cursor.Style = lipgloss.NewStyle().Foreground(catMochaMauve)
	return t
}

func newDashboardModel(mode, target, filter string) *dashboardModel {
	m := &dashboardModel{
		mode:               mode,
		target:             target,
		filter:             filter,
		state:              StateDashboard,
		expanded:           make(map[string]bool),
		formLastFocusIndex: -1,
	}
	m.formTitleField = createFormTextInput("")
	m.formSlugField = createFormTextInput("")
	m.formDescField = createFormTextArea("")
	m.formAssignedField = createFormTextInput("")
	m.formHeadField = createFormTextInput("")
	m.formBaseField = createFormTextInput("")
	m.formSrcField = createFormTextInput("")
	m.formDestField = createFormTextInput("")
	m.formTasksField = createFormTextInput("")
	m.formMilestoneField = createFormTextInput("")
	m.formTypeField = createFormTextInput("")
	m.formTagsField = createFormTextInput("")
	m.formPRsField = createFormTextInput("")
	m.pushBranchField = createFormTextInput("feature-branch-name")
	_, err := store.GetDataRoot()
	if err != nil {
		m.state = StatePromptInit
	} else {
		m.refresh()
	}
	return m
}

