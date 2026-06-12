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
// it under the terms of the GNU Affero General Public License as                                  //
// published by the Free Software Foundation, either version 3 of the                              //
// License, or (at your option) any later version.                                                 //
//                                                                                                 //
// This program is distributed in the hope that it will be useful,                                 //
// but WITHOUT ANY WARRANTY; without even the implied warranty of                                  //
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the                                   //
// GNU Affero General Public License for more details.                                             //
//                                                                                                 //
// You should have received a copy of the GNU Affero General Public License                        //
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
	"github.com/spf13/cobra"
	"gotunix.net/metaboard/internal/models"
	"gotunix.net/metaboard/internal/store"
	"gotunix.net/metaboard/internal/version"
)

const Logo = `
   /$$$$$$$                        /$$
  | $$__  $$                      | $$
  | $$  \ $$  /$$$$$$   /$$$$$$$ /$$$$$$    /$$$$$$   /$$$$$$
  | $$$$$$$/ /$$__  $$ /$$_____/|_  $$_/   /$$__  $$ /$$__  $$
  | $$__  $$| $$  \ $$|  $$$$$$   | $$    | $$$$$$$$| $$  \__/
  | $$  \ $$| $$  | $$ \____  $$  | $$ /$$| $$_____/| $$
  | $$  | $$|  $$$$$$/ /$$$$$$$/  |  $$$$/|  $$$$$$$| $$
  |__/  |__/ \______/ |_______/    \___/   \_______/|__/
`

func HandleHelp(cmd *cobra.Command, args []string) {
	cmd.Usage()
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

func RenderVersion() error {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return fmt.Errorf("failed to read build info")
	}

	totalWidth := GetTerminalWidth()
	var sb strings.Builder

	appVersion := fmt.Sprintf("%s %s", version.AppName, version.AppVersion)

	border := lipgloss.RoundedBorder()
	subStyle := lipgloss.NewStyle().Foreground(Subtle)
	labelStyle := LabelStyle.Padding(0, 1).Width(14)
	valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Padding(0, 1)

	// Header
	titleText := " SYSTEM INFORMATION "
	dLeft := (totalWidth - 2 - len(titleText)) / 2
	dRight := totalWidth - 2 - len(titleText) - dLeft
	sb.WriteString("\n" + subStyle.Render(border.TopLeft+strings.Repeat(border.Top, dLeft)) +
		BoldStyle.Foreground(lipgloss.Color("#FFFFFF")).Render(titleText) +
		subStyle.Render(strings.Repeat(border.Top, dRight)+border.TopRight) + "\n")

	// Row Calculations
	contentWidth := totalWidth - 2
	availForSplit := contentWidth - 28 - 3
	vW1 := availForSplit / 2
	vW2 := availForSplit - vW1

	renderSplitRow := func(l1, v1, l2, v2 string, isLast bool) {
		sb.WriteString(subStyle.Render("│") + labelStyle.Render(l1) + subStyle.Render("│") + valStyle.Width(vW1).Render(v1) +
			subStyle.Render("│") + labelStyle.Render(l2) + subStyle.Render("│") + valStyle.Width(vW2).Render(v2) +
			subStyle.Render("│") + "\n")
		if !isLast {
			sb.WriteString(subStyle.Render("├──────────────┼"+strings.Repeat("─", vW1)+"┼──────────────┼"+strings.Repeat("─", vW2)+"┤") + "\n")
		}
	}

	renderSplitRow("APP:", appVersion, "GO:", info.GoVersion, false)
	renderSplitRow("OS:", runtime.GOOS, "ARCH:", runtime.GOARCH, true)

	sb.WriteString(subStyle.Render(border.BottomLeft+strings.Repeat(border.Bottom, totalWidth-2)+border.BottomRight) + "\n")
	sb.WriteString("\n")

	// Dependencies Window
	depContentWidth := totalWidth - 10
	dotStyle := lipgloss.NewStyle().Foreground(Subtle)

	var depLines []string
	for _, dep := range info.Deps {
		path := lipgloss.NewStyle().Foreground(Cyan).Render(dep.Path)
		version := lipgloss.NewStyle().Foreground(Green).Render(dep.Version)

		label := "• " + path + " "
		repeat := depContentWidth - lipgloss.Width(label) - lipgloss.Width(dep.Version) - 1
		if repeat < 0 {
			repeat = 0
		}

		depLines = append(depLines, label+dotStyle.Render(strings.Repeat(".", repeat))+" "+version)
	}
	sort.Strings(depLines)

	renderWindow(&sb, "DEPENDENCIES", strings.Join(depLines, "\n"), totalWidth)

	fmt.Print(sb.String())
	return nil
}

func ViewMilestone(idOrSlug string) error {
	ms, _ := store.ListMilestones()
	ss, _ := store.ListStories()
	ts, _ := store.ListTasks()

	var m models.Milestone
	found := false
	for _, mil := range ms {
		if mil.ID == idOrSlug || mil.Slug == idOrSlug || mil.Title == idOrSlug {
			m = mil
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("milestone %q not found", idOrSlug)
	}

	storyMap := make(map[string]models.Story)
	for _, s := range ss {
		storyMap[s.ID] = s
	}
	taskMap := make(map[string]models.Task)
	for _, t := range ts {
		taskMap[t.ID] = t
	}

	totalWidth := GetTerminalWidth()
	border := lipgloss.RoundedBorder()
	subStyle := lipgloss.NewStyle().Foreground(Subtle)

	var sb strings.Builder

	// 1. Header
	titleText := fmt.Sprintf(" MILESTONE: %s ", strings.ToUpper(m.Slug))
	dLeft := (totalWidth - 2 - len(titleText)) / 2
	dRight := totalWidth - 2 - len(titleText) - dLeft
	sb.WriteString("\n" + subStyle.Render(border.TopLeft+strings.Repeat(border.Top, dLeft)) +
		BoldStyle.Foreground(lipgloss.Color("#FFFFFF")).Render(titleText) +
		subStyle.Render(strings.Repeat(border.Top, dRight)+border.TopRight) + "\n")

	// 2. Info Grid
	labelStyle := LabelStyle.Padding(0, 1).Width(14)
	valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Padding(0, 1)

	// Total line width should be exactly totalWidth
	// Line: │ (1) + label1 (14) + │ (1) + vW1 + │ (1) + label2 (14) + │ (1) + vW2 + │ (1) = 33 + vW1 + vW2
	// We want 33 + vW1 + vW2 = totalWidth => vW1 + vW2 = totalWidth - 33
	availForSplit := totalWidth - 33

	if availForSplit < 0 {
		availForSplit = 0
	}
	vW1 := availForSplit / 2
	vW2 := availForSplit - vW1

	renderSplitRow := func(l1, v1, l2, v2 string, isLast bool) {
		sb.WriteString(subStyle.Render("│") + labelStyle.Render(l1) + subStyle.Render("│") + valStyle.Width(vW1).Render(v1) +
			subStyle.Render("│") + labelStyle.Render(l2) + subStyle.Render("│") + valStyle.Width(vW2).Render(v2) +
			subStyle.Render("│") + "\n")
		if !isLast {
			sb.WriteString(subStyle.Render("├──────────────┼"+strings.Repeat("─", vW1)+"┼──────────────┼"+strings.Repeat("─", vW2)+"┤") + "\n")
		}
	}

	renderSplitRow("TITLE:", m.Title, "STATUS:", StatusStyle(m.Status).Render(m.Status), false)
	renderSplitRow("ID:", m.ID, "COMPLETED:", m.CompletedAt, true)
	sb.WriteString(subStyle.Render(border.BottomLeft+strings.Repeat(border.Bottom, totalWidth-2)+border.BottomRight) + "\n")

	// 3. Description
	renderWindow(&sb, "DESCRIPTION", strings.Join(m.Description, "\n"), totalWidth)

	// 4. Stories
	sort.Slice(m.Stories, func(i, j int) bool {
		s1, ok1 := storyMap[m.Stories[i]]
		s2, ok2 := storyMap[m.Stories[j]]
		if ok1 && ok2 {
			return store.CompareNatural(s1.Slug, s2.Slug) < 0
		}
		return m.Stories[i] < m.Stories[j]
	})
	var storyLines []string
	slugStyle := lipgloss.NewStyle().Foreground(Green)
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(Yellow).Italic(true)

	dotStyle := lipgloss.NewStyle().Foreground(Subtle)

	for _, id := range m.Stories {
		if s, ok := storyMap[id]; ok {
			desc := ""
			if len(s.Description) > 0 {
				desc = " (" + s.Description[0] + ")"
			}
			storyLines = append(storyLines, fmt.Sprintf("• %s %s%s", slugStyle.Render("["+s.Slug+"]"), titleStyle.Render(s.Title), descStyle.Render(desc)))

			// Tasks under Story
			sort.Slice(s.Tasks, func(i, j int) bool {
				t1, ok1 := taskMap[s.Tasks[i]]
				t2, ok2 := taskMap[s.Tasks[j]]
				if ok1 && ok2 {
					return store.CompareNatural(t1.Slug, t2.Slug) < 0
				}
				return s.Tasks[i] < s.Tasks[j]
			})

			for j, tID := range s.Tasks {
				if t, ok := taskMap[tID]; ok {
					branch := "├─"
					if j == len(s.Tasks)-1 {
						branch = "└─"
					}

					tStatus := strings.ToUpper(t.Status)
					sStyled := StatusStyle(tStatus).Render(tStatus)

					label := fmt.Sprintf("    %s %s %s",
						branch,
						slugStyle.Render("["+t.Slug+"]"),
						titleStyle.Render(t.Title))

					// Calculate dots for right alignment
					cWidth := totalWidth - 6
					repeat := cWidth - lipgloss.Width(label) - lipgloss.Width(tStatus) - 2
					if repeat < 0 {
						repeat = 0
					}

					storyLines = append(storyLines, fmt.Sprintf("%s %s %s",
						label,
						dotStyle.Render(strings.Repeat(".", repeat)),
						sStyled))
				}
			}
		}
	}
	renderWindow(&sb, "LINKED STORIES", strings.Join(storyLines, "\n"), totalWidth)

	// 5. Tasks
	sort.Slice(m.Tasks, func(i, j int) bool {
		t1, ok1 := taskMap[m.Tasks[i]]
		t2, ok2 := taskMap[m.Tasks[j]]
		if ok1 && ok2 {
			return store.CompareNatural(t1.Slug, t2.Slug) < 0
		}
		return m.Tasks[i] < m.Tasks[j]
	})
	var taskLines []string
	for _, id := range m.Tasks {
		if t, ok := taskMap[id]; ok {
			tStatus := strings.ToUpper(t.Status)
			sStyled := StatusStyle(tStatus).Render(tStatus)

			label := fmt.Sprintf("• %s %s", slugStyle.Render("["+t.Slug+"]"), titleStyle.Render(t.Title))

			cWidth := totalWidth - 6
			repeat := cWidth - lipgloss.Width(label) - lipgloss.Width(tStatus) - 2
			if repeat < 0 {
				repeat = 0
			}

			taskLines = append(taskLines, fmt.Sprintf("%s %s %s",
				label,
				dotStyle.Render(strings.Repeat(".", repeat)),
				sStyled))
		}
	}
	renderWindow(&sb, "DIRECT TASKS", strings.Join(taskLines, "\n"), totalWidth)

	fmt.Print(sb.String())
	return nil
}

func ViewStory(idOrSlug string) error {
	ss, _ := store.ListStories()
	ts, _ := store.ListTasks()

	var s models.Story
	found := false
	for _, st := range ss {
		if st.ID == idOrSlug || st.Slug == idOrSlug || st.Title == idOrSlug {
			s = st
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("story %q not found", idOrSlug)
	}

	taskMap := make(map[string]models.Task)
	for _, t := range ts {
		taskMap[t.ID] = t
	}

	totalWidth := GetTerminalWidth()
	border := lipgloss.RoundedBorder()
	subStyle := lipgloss.NewStyle().Foreground(Subtle)

	var sb strings.Builder

	// 1. Header
	titleText := fmt.Sprintf(" STORY: %s ", strings.ToUpper(s.Slug))
	dLeft := (totalWidth - 2 - len(titleText)) / 2
	dRight := totalWidth - 2 - len(titleText) - dLeft
	sb.WriteString("\n" + subStyle.Render(border.TopLeft+strings.Repeat(border.Top, dLeft)) +
		BoldStyle.Foreground(lipgloss.Color("#FFFFFF")).Render(titleText) +
		subStyle.Render(strings.Repeat(border.Top, dRight)+border.TopRight) + "\n")

	// 2. Info Grid
	labelStyle := LabelStyle.Padding(0, 1).Width(14)
	valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Padding(0, 1)

	// Total line width should be exactly totalWidth
	// Line: │ (1) + label1 (14) + │ (1) + vW1 + │ (1) + label2 (14) + │ (1) + vW2 + │ (1) = 33 + vW1 + vW2
	// We want 33 + vW1 + vW2 = totalWidth => vW1 + vW2 = totalWidth - 33
	availForSplit := totalWidth - 33

	if availForSplit < 0 {
		availForSplit = 0
	}
	vW1 := availForSplit / 2
	vW2 := availForSplit - vW1

	renderSplitRow := func(l1, v1, l2, v2 string, isLast bool) {
		sb.WriteString(subStyle.Render("│") + labelStyle.Render(l1) + subStyle.Render("│") + valStyle.Width(vW1).Render(v1) +
			subStyle.Render("│") + labelStyle.Render(l2) + subStyle.Render("│") + valStyle.Width(vW2).Render(v2) +
			subStyle.Render("│") + "\n")
		if !isLast {
			sb.WriteString(subStyle.Render("├──────────────┼"+strings.Repeat("─", vW1)+"┼──────────────┼"+strings.Repeat("─", vW2)+"┤") + "\n")
		}
	}

	renderSplitRow("TITLE:", s.Title, "STATUS:", StatusStyle(s.Status).Render(s.Status), false)
	renderSplitRow("ID:", s.ID, "COMPLETED:", s.CompletedAt, true)
	sb.WriteString(subStyle.Render(border.BottomLeft+strings.Repeat(border.Bottom, totalWidth-2)+border.BottomRight) + "\n")

	// 3. Description
	renderWindow(&sb, "DESCRIPTION", strings.Join(s.Description, "\n"), totalWidth)

	// 4. Tasks
	sort.Slice(s.Tasks, func(i, j int) bool {
		t1, ok1 := taskMap[s.Tasks[i]]
		t2, ok2 := taskMap[s.Tasks[j]]
		if ok1 && ok2 {
			return store.CompareNatural(t1.Slug, t2.Slug) < 0
		}
		return s.Tasks[i] < s.Tasks[j]
	})
	var taskLines []string
	slugStyle := lipgloss.NewStyle().Foreground(Green)
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true)
	dotStyle := lipgloss.NewStyle().Foreground(Subtle)
	cWidth := totalWidth - 6

	for _, id := range s.Tasks {
		if t, ok := taskMap[id]; ok {
			tStatus := strings.ToUpper(t.Status)
			sStyled := StatusStyle(tStatus).Render(tStatus)

			label := fmt.Sprintf("• %s %s", slugStyle.Render("["+t.Slug+"]"), titleStyle.Render(t.Title))

			repeat := cWidth - lipgloss.Width(label) - lipgloss.Width(tStatus) - 2
			if repeat < 0 {
				repeat = 0
			}

			taskLines = append(taskLines, fmt.Sprintf("%s %s %s",
				label,
				dotStyle.Render(strings.Repeat(".", repeat)),
				sStyled))
		}
	}
	renderWindow(&sb, "LINKED TASKS", strings.Join(taskLines, "\n"), totalWidth)

	fmt.Print(sb.String())
	return nil
}

func ViewTask(idOrSlug string) error {
	ts, _ := store.ListTasks()

	var t models.Task
	found := false
	for _, task := range ts {
		if task.ID == idOrSlug || task.Slug == idOrSlug || task.Title == idOrSlug {
			t = task
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("task %q not found", idOrSlug)
	}

	priorityColor := Gray
	switch strings.ToUpper(t.Priority) {
	case "HIGH":
		priorityColor = Red
	case "MEDIUM":
		priorityColor = Yellow
	case "LOW":
		priorityColor = Cyan
	}

	totalWidth := GetTerminalWidth()
	border := lipgloss.RoundedBorder()
	subStyle := lipgloss.NewStyle().Foreground(Subtle)
	labelStyle := LabelStyle.Padding(0, 1).Width(14)
	valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Padding(0, 1)

	var sb strings.Builder

	// 1. Header
	titleText := fmt.Sprintf(" TASK: %s ", strings.ToUpper(t.Slug))
	dLeft := (totalWidth - 2 - len(titleText)) / 2
	dRight := totalWidth - 2 - len(titleText) - dLeft
	sb.WriteString("\n" + subStyle.Render(border.TopLeft+strings.Repeat(border.Top, dLeft)) +
		BoldStyle.Foreground(lipgloss.Color("#FFFFFF")).Render(titleText) +
		subStyle.Render(strings.Repeat(border.Top, dRight)+border.TopRight) + "\n")

	// 2. Row Calculations
	// Total line width should be exactly totalWidth
	// Full Line: │ (1) + label (14) + │ (1) + fullValW + │ (1) = 17 + fullValW
	// Split Line: │ (1) + label1 (14) + │ (1) + vW1 + │ (1) + label2 (14) + │ (1) + vW2 + │ (1) = 34 + vW1 + vW2
	fullValW := totalWidth - 17
	if fullValW < 0 {
		fullValW = 0
	}

	availForSplit := totalWidth - 33

	if availForSplit < 0 {
		availForSplit = 0
	}
	vW1 := availForSplit / 2
	vW2 := availForSplit - vW1

	renderFullRow := func(l, v string, isLast bool) {
		sb.WriteString(subStyle.Render("│") + labelStyle.Render(l) + subStyle.Render("│") + valStyle.Width(fullValW).Render(v) + subStyle.Render("│") + "\n")
		if !isLast {
			sb.WriteString(subStyle.Render("├──────────────┼"+strings.Repeat("─", fullValW)+"┤") + "\n")
		}
	}

	renderSplitRow := func(l1, v1, l2, v2 string, isLast bool) {
		sb.WriteString(subStyle.Render("│") + labelStyle.Render(l1) + subStyle.Render("│") + valStyle.Width(vW1).Render(v1) +
			subStyle.Render("│") + labelStyle.Render(l2) + subStyle.Render("│") + valStyle.Width(vW2).Render(v2) +
			subStyle.Render("│") + "\n")
		if !isLast {
			sb.WriteString(subStyle.Render("├──────────────┼"+strings.Repeat("─", vW1)+"┼──────────────┼"+strings.Repeat("─", vW2)+"┤") + "\n")
		}
	}

	renderFullRow("TITLE:", t.Title, false)
	renderSplitRow("STATUS:", StatusStyle(t.Status).Render(t.Status), "PRIORITY:", lipgloss.NewStyle().Foreground(priorityColor).Bold(true).Render(t.Priority), false)
	renderSplitRow("TYPE:", t.Type, "ASSIGNED:", t.AssignedTo, false)
	renderFullRow("TAGS:", strings.Join(t.Tags, ", "), false)
	renderFullRow("DEPENDS ON:", strings.Join(t.DependsOn, ", "), false)
	renderSplitRow("CREATED:", t.CreatedAt, "COMPLETED:", t.CompletedAt, true)

	sb.WriteString(subStyle.Render(border.BottomLeft+strings.Repeat(border.Bottom, totalWidth-2)+border.BottomRight) + "\n")
	fmt.Print(sb.String())

	// 3. Description
	var sbDesc strings.Builder
	renderWindow(&sbDesc, "DESCRIPTION", strings.Join(t.Description, "\n"), totalWidth)
	fmt.Println("") // Space before Description
	fmt.Print(sbDesc.String())

	// 4. Implementation Plan (Inline from Sidecar .md)
	planPath := store.GetTaskPlanPath(t.ID)
	planContent := ""
	if _, err := os.Stat(planPath); err == nil {
		data, _ := os.ReadFile(planPath)
		planContent = string(data)
	}

	if planContent != "" {
		var sbPlan strings.Builder
		renderWindow(&sbPlan, "IMPLEMENTATION PLAN", planContent, totalWidth)
		fmt.Println("") // Space before Plan
		fmt.Print(sbPlan.String())
	}

	return nil
}

// Helper to render a titled window box with precise border alignment
func renderWindow(sb *strings.Builder, title, content string, width int) {
	if content == "" {
		content = "(empty)"
	}

	border := lipgloss.RoundedBorder()
	subStyle := lipgloss.NewStyle().Foreground(Subtle)

	// 1. Top Border with Integrated Header
	tText := fmt.Sprintf(" %s ", title)
	dL := (width - 2 - len(tText)) / 2
	dR := width - 2 - len(tText) - dL

	sb.WriteString(subStyle.Render(border.TopLeft+strings.Repeat(border.Top, dL)) +
		BoldStyle.Foreground(lipgloss.Color("#FFFFFF")).Render(tText) +
		subStyle.Render(strings.Repeat(border.Top, dR)+border.TopRight) + "\n")

	// 2. Content with Lipgloss managed width and padding
	contentStyle := lipgloss.NewStyle().
		Width(width-2). // Width between bars
		Padding(0, 2)   // 2 cells of padding on each side

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		renderedLine := contentStyle.Render(line)
		sb.WriteString(subStyle.Render("│") + renderedLine + subStyle.Render("│") + "\n")
	}

	// 3. Bottom Border
	sb.WriteString(subStyle.Render(border.BottomLeft+strings.Repeat(border.Bottom, width-2)+border.BottomRight) + "\n")
}
