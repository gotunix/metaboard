// SPDX-License-Identifier: GPL-3.0-or-later
// SPDX-FileCopyrightText: 2026 The MetaBoard authors
// =============================================================================================== //
//                                                                                                 //
//   /$$      /$$             /$$               /$$$$$$$                                      /$$  //
//  | $$$    /$$$            | $$              | $$__  $$                                    | $$  //
//  | $$$$  /$$$$  /$$$$$$  /$$$$$$    /$$$$$$ | $$  \ $$  /$$$$$$   /$$$$$$   /$$$$$$   /$$$$$$$  //
//  | $$ $$/$$ $$ /$$__  $$|_  $$_/   |____  $$| $$$$$$$  /$$__  $$ |____  $$ /$$__  $$ /$$__  $$  //
//  | $$  $$$| $$| $$$$$$$$  | $$      /$$$$$$$| $$__  $$| $$  \ $$  /$$$$$$$| $$  \__/| $$  | $$  //
//  | $$\  $ | $$| $$_____/  | $$ /$$ /$$__  $$| $$  \ $$| $$  | $$ /$$__  $$| $$      | $$  | $$  //
//  | $$ \/  | $$|  $$$$$$$  |  $$$$/|  $$$$$$$| $$$$$$$/|  $$$$$$/|  $$$$$$$| $$      |  $$$$$$$  //
//  |__/     |__/ \_______/   \___/   \_______/|_______/  \______/  \_______/|__/       \_______/  //
//                                                                                                 //
// =============================================================================================== //
// This program is free software: you can redistribute it and/or modify                            //
// it under the terms of the GNU General Public License as                                  //
// published by the Free Software Foundation, either version 3 of the                              //
// License, or (at your option) any later version.                                                 //
//                                                                                                 //
// This program is distributed in the hope that it will be useful,                                 //
// but WITHOUT ANY WARRANTY; without even the implied warranty of                                  //
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the                                   //
// GNU General Public License for more details.                                             //
//                                                                                                 //
// You should have received a copy of the GNU General Public License                        //
// along with this program.  If not, see <https://www.gnu.org/licenses/>.                          //
// =============================================================================================== //

package ui

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
	"github.com/spf13/cobra"
	"gotunix.net/metaboard/internal/version"
	"gotunix.net/metaboard/internal/models"
	"gotunix.net/metaboard/internal/store"
)

const Logo = `
   /$$      /$$             /$$               /$$$$$$$                                      /$$
  | $$$    /$$$            | $$              | $$__  $$                                    | $$
  | $$$$  /$$$$  /$$$$$$  /$$$$$$    /$$$$$$ | $$  \ $$  /$$$$$$   /$$$$$$   /$$$$$$   /$$$$$$$
  | $$ $$/$$ $$ /$$__  $$|_  $$_/   |____  $$| $$$$$$$  /$$__  $$ |____  $$ /$$__  $$ /$$__  $$
  | $$  $$$| $$| $$$$$$$$  | $$      /$$$$$$$| $$__  $$| $$  \ $$  /$$$$$$$| $$  \__/| $$  | $$
  | $$\  $ | $$| $$_____/  | $$ /$$ /$$__  $$| $$  \ $$| $$  | $$ /$$__  $$| $$      | $$  | $$
  | $$ \/  | $$|  $$$$$$$  |  $$$$/|  $$$$$$$| $$$$$$$/|  $$$$$$/|  $$$$$$$| $$      |  $$$$$$$
  |__/     |__/ \_______/   \___/   \_______/|_______/  \______/  \_______/|__/       \_______/
`

func HandleHelp(cmd *cobra.Command, args []string) {
	_ = cmd.Usage()
}

func HandleUsage(cmd *cobra.Command) error {
	fmt.Println(LogoStyle.Render(Logo))
	fmt.Println(HelpTitleStyle.Render(cmd.Short))
	fmt.Println(HelpDescStyle.Render(cmd.Long))

	fmt.Println(HelpSectionStyle.Render("USAGE"))
	fmt.Printf("  %s [command]\n", cmd.CommandPath())

	if len(cmd.Commands()) > 0 {
		fmt.Println(HelpSectionStyle.Render("AVAILABLE COMMANDS"))
		for _, c := range cmd.Commands() {
			if !c.Hidden {
				fmt.Printf("  %-15s %s\n", c.Name(), c.Short)
			}
		}
	}

	if cmd.Flags().HasFlags() {
		fmt.Println(HelpSectionStyle.Render("FLAGS"))
		fmt.Println(HelpFlagStyle.Render(cmd.Flags().FlagUsages()))
	}

	fmt.Println(HelpSectionStyle.Render("LEARN MORE"))
	fmt.Printf("  Use \"%s [command] --help\" for more information about a command.\n", cmd.CommandPath())
	return nil
}

func GetVersionString(width int) string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "Error loading version information"
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
	appVersion := fmt.Sprintf("%s %s", version.AppName, version.AppVersion)

	// Title Banner
	sb.WriteString(TitleStyle.Copy().Width(width).Render(fmt.Sprintf("VERSION DETAILS: %s", strings.ToUpper(appVersion))) + "\n")
	dividerText := repeat("─", col1Width+1) + "┬" + repeat("─", width-col1Width-2)
	sb.WriteString(lipgloss.NewStyle().Foreground(catMochaBase).Background(TableBg).Render(dividerText) + "\n")

	labelStyle := lipgloss.NewStyle().Foreground(LabelFg).Bold(true).Background(TableBg)
	gridValStyle := lipgloss.NewStyle().Foreground(ValFg).Background(TableBg)
	bgStyle := lipgloss.NewStyle().Background(TableBg)
	sepStyle := lipgloss.NewStyle().Foreground(catMochaBase).Background(TableBg)

	formatField := func(label, val string, colW int) string {
		lWidth := runewidth.StringWidth(label)
		vWidth := colW - lWidth - 1
		if vWidth < 0 {
			vWidth = 0
		}
		if runewidth.StringWidth(val) > vWidth {
			val = runewidth.Truncate(val, vWidth, "")
		}
		valW := runewidth.StringWidth(val)
		return labelStyle.Render(label) + bgStyle.Render(" ") + gridValStyle.Render(val) + bgStyle.Render(repeat(" ", vWidth-valW))
	}

	row1 := formatField("App:", appVersion, col1Width) + sepStyle.Render(" │ ") + formatField("Build:", info.Main.Version, col2Width)
	row2 := formatField("Go:", info.GoVersion, col1Width) + sepStyle.Render(" │ ") + formatField("OS:", runtime.GOOS, col2Width)
	row3 := formatField("Arch:", runtime.GOARCH, col1Width) + sepStyle.Render(" │ ") + formatField("Module:", info.Main.Path, col2Width)

	sb.WriteString(row1 + "\n" + row2 + "\n" + row3 + "\n\n")

	if len(info.Deps) > 0 {
		sb.WriteString(lipgloss.NewStyle().Foreground(catMochaMauve).Bold(true).Render("Dependencies:") + "\n")
		var depLines []string
		for _, dep := range info.Deps {
			depLines = append(depLines, fmt.Sprintf("  • %s %s", dep.Path, dep.Version))
		}
		sort.Strings(depLines)
		for _, line := range depLines {
			sb.WriteString(line + "\n")
		}
	}

	return sb.String()
}

func RenderVersion() error {
	totalWidth := GetTerminalWidth()
	fmt.Print(GetVersionString(totalWidth))
	return nil
}

func ViewMilestone(idOrSlug string, version int) error {
	var m *models.Milestone
	var err error

	if version > 0 {
		m, err = store.GetMilestoneVersion(idOrSlug, version)
	} else {
		m, err = store.GetMilestone(idOrSlug)
	}

	if err != nil {
		return err
	}

	ts, _ := store.ListTasks()

	taskMap := make(map[string]models.Task)
	for _, t := range ts {
		taskMap[t.ID] = t
	}

	var sb strings.Builder

	// Simple header
	sb.WriteString(fmt.Sprintf("MILESTONE: %s (v%d)\n", strings.ToUpper(m.Slug), m.Version))
	sb.WriteString(strings.Repeat("─", 60) + "\n")

	// Key fields
	printKV := func(label, value string) {
		if value != "" {
			sb.WriteString(fmt.Sprintf("  %-14s %s\n", label+":", value))
		}
	}

	printKV("TITLE", m.Title)
	printKV("STATUS", m.Status)
	printKV("CREATED", m.CreatedAt)
	printKV("UPDATED", m.UpdatedAt)
	printKV("COMPLETED", m.CompletedAt)
	printKV("VERSION", fmt.Sprintf("%d", m.Version))

	// Description
	sb.WriteString("\nDESCRIPTION\n")
	sb.WriteString(strings.Repeat("─", 60) + "\n")
	if m.Description != "" {
		sb.WriteString(WrapText(m.Description, 78) + "\n")
	} else {
		sb.WriteString("(empty)\n")
	}

	// Tasks
	sb.WriteString("\nTASKS\n")
	sb.WriteString(strings.Repeat("─", 60) + "\n")

	var mTasks []models.Task
	for _, id := range m.Tasks {
		if t, ok := taskMap[id]; ok {
			mTasks = append(mTasks, t)
		}
	}
	store.SortTasks(mTasks)

	if len(mTasks) == 0 {
		sb.WriteString("  (no tasks)\n")
	} else {
		for _, t := range mTasks {
			statusStyle := StatusStyle(strings.ToUpper(t.Status))
			sb.WriteString(fmt.Sprintf("  • %s [%s] %s\n", t.Title, t.Slug, statusStyle.Render(t.Status)))
		}
	}

	return RunPager(sb.String())
}



func ViewTask(idOrSlug string, version int) error {
	var t *models.Task
	var err error

	if version > 0 {
		t, err = store.GetTaskVersion(idOrSlug, version)
	} else {
		t, err = store.GetTask(idOrSlug)
	}

	if err != nil {
		return err
	}

	var sb strings.Builder

	// Simple header
	sb.WriteString(fmt.Sprintf("TASK: %s (v%d)\n", strings.ToUpper(t.Slug), t.Version))
	sb.WriteString(strings.Repeat("─", 60) + "\n")

	// Key fields in simple key: value format
	printKV := func(label, value string) {
		if value != "" {
			sb.WriteString(fmt.Sprintf("  %-14s %s\n", label+":", value))
		}
	}

	printKV("TITLE", t.Title)
	printKV("STATUS", t.Status)
	printKV("PRIORITY", t.Priority)
	printKV("TYPE", t.Type)
	printKV("ASSIGNED TO", t.AssignedTo)
	if len(t.Tags) > 0 {
		printKV("TAGS", strings.Join(t.Tags, ", "))
	}
	if len(t.DependsOn) > 0 {
		printKV("DEPENDS ON", strings.Join(t.DependsOn, ", "))
	}
	printKV("CREATED", t.CreatedAt)
	printKV("UPDATED", t.UpdatedAt)
	printKV("COMPLETED", t.CompletedAt)
	printKV("VERSION", fmt.Sprintf("%d", t.Version))

	// Description
	sb.WriteString("\nDESCRIPTION\n")
	sb.WriteString(strings.Repeat("─", 60) + "\n")
	if t.Description != "" {
		sb.WriteString(WrapText(t.Description, 78) + "\n")
	} else {
		sb.WriteString("(empty)\n")
	}

	// Additional Details (from plan markdown)
	planPath, err := store.GetTaskPlanPath(t.ID)
	if err == nil {
		if _, err := os.Stat(planPath); err == nil {
			data, _ := os.ReadFile(planPath)
			if len(data) > 0 {
				sb.WriteString("\nADDITIONAL DETAILS\n")
				sb.WriteString(strings.Repeat("─", 60) + "\n")
				sb.WriteString(string(data) + "\n")
			}
		}
	}

	// Linked Pull Requests
	allPRs, _ := store.ListPullRequests()
	prMap := make(map[string]models.PullRequest)
	for _, pr := range allPRs {
		prMap[pr.ID] = pr
	}
	if len(t.PullRequests) > 0 {
		sb.WriteString("\nLINKED PULL REQUESTS\n")
		sb.WriteString(strings.Repeat("─", 60) + "\n")
		for _, prID := range t.PullRequests {
			if pr, ok := prMap[prID]; ok {
				sb.WriteString(fmt.Sprintf("  %s → %s (%s)\n", pr.HeadBranch, pr.BaseBranch, pr.Status))
			}
		}
	}

	return RunPager(sb.String())
}

// Helper to render a titled window box with precise border alignment
func ViewMilestoneHistory(idOrSlug string) error {
	history, err := store.GetMilestoneHistory(idOrSlug)
	if err != nil {
		return err
	}

	totalWidth := GetTerminalWidth()
	var sb strings.Builder

	lines := []string{}
	for i := len(history) - 1; i >= 0; i-- {
		m := history[i]
		v := m.Version
		if v == 0 {
			v = 1
		}
		lines = append(lines, fmt.Sprintf("v%d | %s | %s | %s", v, m.UpdatedAt, StatusStyle(m.Status).Render(m.Status), m.Title))
	}

	renderWindow(&sb, fmt.Sprintf("MILESTONE HISTORY: %s", strings.ToUpper(idOrSlug)), strings.Join(lines, "\n"), totalWidth)

	return RunPager(sb.String())
}

func ViewTaskHistory(idOrSlug string) error {
	history, err := store.GetTaskHistory(idOrSlug)
	if err != nil {
		return err
	}

	totalWidth := GetTerminalWidth()
	var sb strings.Builder

	lines := []string{}
	for i := len(history) - 1; i >= 0; i-- {
		t := history[i]
		v := t.Version
		if v == 0 {
			v = 1
		}
		lines = append(lines, fmt.Sprintf("v%d | %s | %s | %s", v, t.UpdatedAt, StatusStyle(t.Status).Render(t.Status), t.Title))
	}

	renderWindow(&sb, fmt.Sprintf("TASK HISTORY: %s", strings.ToUpper(idOrSlug)), strings.Join(lines, "\n"), totalWidth)
	return RunPager(sb.String())
}

func renderWindow(sb *strings.Builder, title, content string, width int) {
	if content == "" {
		content = "(empty)"
	}

	border := lipgloss.RoundedBorder()
	subStyle := lipgloss.NewStyle().Foreground(Subtle)

	// 1. Top Border with Integrated Header
	tText := fmt.Sprintf(" %s ", title)
	titleW := runewidth.StringWidth(tText)
	dL := (width - 2 - titleW) / 2
	dR := width - 2 - titleW - dL

	sb.WriteString(subStyle.Render(border.TopLeft+repeat(border.Top, dL)) +
		BoldStyle.Foreground(ValFg).Render(tText) +
		subStyle.Render(repeat(border.Top, dR)+border.TopRight) + "\n")

	// 2. Content with Lipgloss managed width and padding
	contentStyle := lipgloss.NewStyle().
		Width(width-2). // Exact width between borders
		Padding(0, 2)   // Internal padding

	renderedContent := contentStyle.Render(content)
	lines := strings.Split(renderedContent, "\n")

	for _, line := range lines {
		lineW := lipgloss.Width(line)
		padding := (width - 2) - lineW

		sb.WriteString(subStyle.Render("│") + line + repeat(" ", padding) + subStyle.Render("│") + "\n")
	}

	// 3. Bottom Border
	sb.WriteString(subStyle.Render(border.BottomLeft+repeat(border.Bottom, width-2)+border.BottomRight) + "\n")
}

func RunPager(content string) error {
	fmt.Println(strings.TrimSpace(content))
	return nil
}

func ViewPullRequest(idOrSlug string, version int) error {
	var pr *models.PullRequest
	var err error

	if version > 0 {
		pr, err = store.GetPullRequestVersion(idOrSlug, version)
	} else {
		pr, err = store.GetPullRequest(idOrSlug)
	}

	if err != nil {
		return err
	}

	var sb strings.Builder

	// Simple header
	sb.WriteString(fmt.Sprintf("PULL REQUEST: %s (v%d)\n", strings.ToUpper(pr.Slug), pr.Version))
	sb.WriteString(strings.Repeat("─", 60) + "\n")

	// Key fields
	printKV := func(label, value string) {
		if value != "" {
			sb.WriteString(fmt.Sprintf("  %-14s %s\n", label+":", value))
		}
	}

	printKV("STATUS", pr.Status)
	printKV("SRC BRANCH", pr.HeadBranch)
	printKV("DEST BRANCH", pr.BaseBranch)
	printKV("SRC REPO", pr.SourceRepo)
	printKV("DEST REPO", pr.DestRepo)
	printKV("CREATED", pr.CreatedAt)
	printKV("UPDATED", pr.UpdatedAt)
	printKV("COMPLETED", pr.CompletedAt)
	printKV("VERSION", fmt.Sprintf("%d", pr.Version))

	// Description
	sb.WriteString("\nDESCRIPTION\n")
	sb.WriteString(strings.Repeat("─", 60) + "\n")
	if pr.Description != "" {
		sb.WriteString(WrapText(pr.Description, 78) + "\n")
	} else {
		sb.WriteString("(empty)\n")
	}

	// Linked Tasks
	if len(pr.Tasks) > 0 {
		tasks, _ := store.ListTasks()
		taskMap := make(map[string]models.Task)
		for _, t := range tasks {
			taskMap[t.ID] = t
		}

		var prTasks []models.Task
		for _, tID := range pr.Tasks {
			if t, ok := taskMap[tID]; ok {
				prTasks = append(prTasks, t)
			}
		}
		store.SortTasks(prTasks)

		sb.WriteString("\nASSOCIATED TASKS\n")
		sb.WriteString(strings.Repeat("─", 60) + "\n")
		for _, t := range prTasks {
			tStatus := strings.ToUpper(t.Status)
			sb.WriteString(fmt.Sprintf("  [%s] %s (%s)\n", t.Slug, t.Title, tStatus))
		}
	}

	return RunPager(sb.String())
}

func ViewPullRequestHistory(idOrSlug string) error {
	history, err := store.GetPullRequestHistory(idOrSlug)
	if err != nil {
		return err
	}

	totalWidth := GetTerminalWidth()
	var sb strings.Builder

	lines := []string{}
	for i := len(history) - 1; i >= 0; i-- {
		pr := history[i]
		v := pr.Version
		if v == 0 {
			v = 1
		}
		lines = append(lines, fmt.Sprintf("v%d | %s | %s | %s → %s", v, pr.UpdatedAt, StatusStyle(pr.Status).Render(pr.Status), pr.HeadBranch, pr.BaseBranch))
	}

	renderWindow(&sb, fmt.Sprintf("PULL REQUEST HISTORY: %s", strings.ToUpper(idOrSlug)), strings.Join(lines, "\n"), totalWidth)
	return RunPager(sb.String())
}

func truncateTitle(t string, budget int) string {
	if runewidth.StringWidth(t) > budget && budget > 3 {
		return runewidth.Truncate(t, budget, "...")
	}
	return t
}

func WrapText(text string, limit int) string {
	if limit <= 0 {
		return text
	}
	lines := strings.Split(text, "\n")
	var wrappedLines []string
	for _, line := range lines {
		if runewidth.StringWidth(line) <= limit {
			wrappedLines = append(wrappedLines, line)
			continue
		}
		words := strings.Fields(line)
		if len(words) == 0 {
			wrappedLines = append(wrappedLines, "")
			continue
		}
		var currentLine strings.Builder
		currentLine.WriteString(words[0])
		spaceLeft := limit - runewidth.StringWidth(words[0])
		for _, word := range words[1:] {
			wordW := runewidth.StringWidth(word)
			if wordW+1 > spaceLeft {
				wrappedLines = append(wrappedLines, currentLine.String())
				currentLine.Reset()
				currentLine.WriteString(word)
				spaceLeft = limit - wordW
			} else {
				currentLine.WriteString(" " + word)
				spaceLeft -= wordW + 1
			}
		}
		if currentLine.Len() > 0 {
			wrappedLines = append(wrappedLines, currentLine.String())
		}
	}
	return strings.Join(wrappedLines, "\n")
}

func repeat(char string, count int) string {
	if count <= 0 {
		return ""
	}
	return strings.Repeat(char, count)
}

