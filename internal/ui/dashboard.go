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
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"gotunix.net/metaboard/internal/models"
	"gotunix.net/metaboard/internal/store"
)

func RenderDashboard(mode, target, filter string) error {
	milestones, err := store.ListMilestones()
	if err != nil {
		return err
	}
	store.SortMilestones(milestones)
	stories, _ := store.ListStories()
	tasks, _ := store.ListTasks()

	// Maps for easy lookup
	storyMap := make(map[string]models.Story)
	for _, s := range stories {
		storyMap[s.ID] = s
	}
	taskMap := make(map[string]models.Task)
	for _, t := range tasks {
		taskMap[t.ID] = t
	}

	if filter == "" {
		filter = "active"
	}
	filter = strings.ToLower(filter)

	isMatch := func(eType, status string) bool {
		s := strings.ToUpper(status)
		switch filter {
		case "closed":
			return s == "COMPLETED"
		case "cancelled":
			return s == "CANCELLED"
		default: // active
			return s == "ACTIVE" || s == "IN-PROGRESS" || s == "BACKLOG"
		}
	}

	// Filter milestones
	var filteredMilestones []models.Milestone
	if mode == "milestone" && target != "" {
		for _, m := range milestones {
			if m.ID == target || m.Slug == target {
				filteredMilestones = append(filteredMilestones, m)
				break
			}
		}
	} else {
		for _, m := range milestones {
			keepMilestone := false
			if isMatch("milestone", m.Status) {
				keepMilestone = true
			} else {
				// Check if any children match
				for _, sID := range m.Stories {
					if s, ok := storyMap[sID]; ok {
						if isMatch("story", s.Status) {
							keepMilestone = true
							break
						}
						for _, tID := range s.Tasks {
							if t, ok := taskMap[tID]; ok {
								if isMatch("task", t.Status) {
									keepMilestone = true
									break
								}
							}
						}
					}
					if keepMilestone {
						break
					}
				}
				if !keepMilestone {
					for _, tID := range m.Tasks {
						if t, ok := taskMap[tID]; ok && isMatch("task", t.Status) {
							keepMilestone = true
							break
						}
					}
				}
			}

			if keepMilestone {
				filteredMilestones = append(filteredMilestones, m)
			}
		}
	}

	if len(filteredMilestones) == 0 {
		msg := "No milestones found."
		if filter == "closed" {
			msg = "No closed items found."
		}
		if filter == "cancelled" {
			msg = "No cancelled items found."
		}
		fmt.Println("  " + lipgloss.NewStyle().Foreground(Gray).Render(msg))
		return nil
	}

	var sb strings.Builder

	// Styles for the new layout
	dotStyle := lipgloss.NewStyle().Foreground(Subtle)

	totalWidth := GetTerminalWidth()
	rightAlign := totalWidth - 20
	if rightAlign < 40 {
		rightAlign = 40
	}

	for i, m := range filteredMilestones {
		mStatus := strings.ToUpper(m.Status)

		// Milestone Header
		slugStyle := lipgloss.NewStyle().Foreground(Green)
		descStyle := lipgloss.NewStyle().Foreground(Yellow).Italic(true)

		desc := ""
		if len(m.Description) > 0 {
			desc = " (" + m.Description[0] + ")"
			if len(m.Description) > 1 || len(m.Description[0]) > 50 {
				if len(m.Description[0]) > 50 {
					desc = " (" + m.Description[0][:47] + "...)"
				}
			}
		}

		headerLabel := fmt.Sprintf("🏁 %s %s%s ",
			slugStyle.Render("["+m.Slug+"]"),
			BoldStyle.Foreground(lipgloss.Color("#FFFFFF")).Render(strings.ToUpper(m.Title)),
			descStyle.Render(desc),
		)

		repeatCount := rightAlign - lipgloss.Width(headerLabel)
		if repeatCount < 0 {
			repeatCount = 0
		}
		dots := dotStyle.Render(strings.Repeat("─", repeatCount))

		mStatusStyled := StatusStyle(mStatus).Render(mStatus)
		mStatusPadded := lipgloss.NewStyle().Width(12).Render(mStatusStyled)

		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(fmt.Sprintf("%s%s %s\n",
			BoldStyle.Foreground(Cyan).Render(headerLabel),
			dots,
			mStatusPadded,
		))

		// --- Stories under Milestone ---
		var activeStories []models.Story
		for _, sID := range m.Stories {
			if s, ok := storyMap[sID]; ok {
				if isMatch("story", s.Status) {
					activeStories = append(activeStories, s)
				} else {
					// Keep story if it has matching children
					hasMatchingChild := false
					for _, tID := range s.Tasks {
						if t, ok := taskMap[tID]; ok && isMatch("task", t.Status) {
							hasMatchingChild = true
							break
						}
					}
					if hasMatchingChild {
						activeStories = append(activeStories, s)
					}
				}
			}
		}

		sort.Slice(activeStories, func(i, j int) bool {
			return store.CompareNatural(activeStories[i].Slug, activeStories[j].Slug) < 0
		})

		for i, s := range activeStories {
			sStatus := strings.ToUpper(s.Status)

			branch := "├─"
			if i == len(activeStories)-1 && len(m.Tasks) == 0 {
				branch = "└─"
			}

			desc := ""
			if len(s.Description) > 0 {
				desc = " (" + s.Description[0] + ")"
				if len(s.Description) > 1 || len(s.Description[0]) > 40 {
					// Simple truncation or indicator
					if len(s.Description[0]) > 40 {
						desc = " (" + s.Description[0][:37] + "...)"
					}
				}
			}

			slugStyle := lipgloss.NewStyle().Foreground(Green)
			descStyle := lipgloss.NewStyle().Foreground(Yellow).Italic(true)

			storyLabel := fmt.Sprintf("  %s %s %s %s%s ",
				branch,
				lipgloss.NewStyle().Foreground(Magenta).Render("📖"),
				slugStyle.Render("["+s.Slug+"]"),
				BoldStyle.Foreground(lipgloss.Color("#FFFFFF")).Render(s.Title),
				descStyle.Render(desc),
			)

			sRepeat := rightAlign - lipgloss.Width(storyLabel)
			if sRepeat < 0 {
				sRepeat = 0
			}

			sStatusStyled := StatusStyle(sStatus).Render(sStatus)
			sStatusPadded := lipgloss.NewStyle().Width(12).Render(sStatusStyled)

			sb.WriteString(fmt.Sprintf("%s%s %s\n",
				storyLabel,
				dotStyle.Render(strings.Repeat(".", sRepeat)),
				sStatusPadded,
			))

			// --- Tasks under Story ---
			if mode != "stories" {
				var activeTasks []models.Task
				for _, tID := range s.Tasks {
					if t, ok := taskMap[tID]; ok {
						if isMatch("task", t.Status) {
							activeTasks = append(activeTasks, t)
						}
					}
				}

				sort.Slice(activeTasks, func(i, j int) bool {
					return store.CompareNatural(activeTasks[i].Slug, activeTasks[j].Slug) < 0
				})

				for j, t := range activeTasks {
					taskBranch := "├─"
					if j == len(activeTasks)-1 {
						taskBranch = "└─"
					}

					parentLine := "│"
					if i == len(activeStories)-1 && len(m.Tasks) == 0 {
						parentLine = " "
					}
					sb.WriteString(renderTreeTaskStr("  "+parentLine+"     "+taskBranch, t, rightAlign))
				}
			}
		}

		// --- Direct Tasks under Milestone ---
		if mode != "stories" && len(m.Tasks) > 0 {
			var activeMilestoneTasks []models.Task
			for _, tID := range m.Tasks {
				if t, ok := taskMap[tID]; ok {
					if isMatch("task", t.Status) {
						activeMilestoneTasks = append(activeMilestoneTasks, t)
					}
				}
			}

			sort.Slice(activeMilestoneTasks, func(i, j int) bool {
				return store.CompareNatural(activeMilestoneTasks[i].Slug, activeMilestoneTasks[j].Slug) < 0
			})

			for j, t := range activeMilestoneTasks {
				branch := "├─"
				if j == len(activeMilestoneTasks)-1 {
					branch = "└─"
				}
				sb.WriteString(renderTreeTaskStr("  "+branch, t, rightAlign))
			}
		}
	}

	// --- Unclaimed Items (only for closed/cancelled) ---
	if filter != "active" {
		claimedStories := make(map[string]bool)
		claimedTasks := make(map[string]bool)
		for _, m := range milestones {
			for _, sID := range m.Stories {
				claimedStories[sID] = true
			}
			for _, tID := range m.Tasks {
				claimedTasks[tID] = true
			}
		}
		for _, s := range stories {
			for _, tID := range s.Tasks {
				claimedTasks[tID] = true
			}
		}

		var unclaimed []string
		for _, s := range stories {
			if !claimedStories[s.ID] && isMatch("story", s.Status) {
				unclaimed = append(unclaimed, fmt.Sprintf("  • %s %s %s",
					lipgloss.NewStyle().Foreground(Magenta).Render("📖"),
					lipgloss.NewStyle().Foreground(Green).Render("["+s.Slug+"]"),
					BoldStyle.Render(s.Title)))
			}
		}
		for _, t := range tasks {
			if !claimedTasks[t.ID] && isMatch("task", t.Status) {
				unclaimed = append(unclaimed, fmt.Sprintf("  •  🛠  %s %s",
					lipgloss.NewStyle().Foreground(Green).Render("["+t.Slug+"]"),
					BoldStyle.Render(t.Title)))
			}
		}

		if len(unclaimed) > 0 {
			sb.WriteString("\n" + HeaderStyle.Render("UNCLAIMED "+strings.ToUpper(filter)+" ITEMS:") + "\n")
			for _, line := range unclaimed {
				sb.WriteString(line + "\n")
			}
		}
	}

	// Create Window with Border
	windowContent := strings.TrimSpace(sb.String())

	windowStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Subtle).
		Padding(1, 2).
		Width(totalWidth - 2)

	renderedWindow := windowStyle.Render(windowContent)
	actualWidth := lipgloss.Width(renderedWindow)

	// Manually inject the header into the top border line
	lines := strings.Split(renderedWindow, "\n")
	if len(lines) > 0 {
		titleText := " PROJECT DASHBOARD "
		titleWidth := len(titleText)
		border := lipgloss.RoundedBorder()

		dashesLeft := (actualWidth - 2 - titleWidth) / 2
		dashesRight := actualWidth - 2 - titleWidth - dashesLeft

		newTopLine := lipgloss.NewStyle().Foreground(Subtle).Render(border.TopLeft+strings.Repeat(border.Top, dashesLeft)) +
			BoldStyle.Foreground(lipgloss.Color("#FFFFFF")).Render(titleText) +
			lipgloss.NewStyle().Foreground(Subtle).Render(strings.Repeat(border.Top, dashesRight)+border.TopRight)

		lines[0] = newTopLine
	}
	// Output
	fmt.Println("")
	fmt.Println(strings.Join(lines, "\n"))

	return nil
}

func renderTreeTaskStr(prefix string, t models.Task, rightAlign int) string {
	tStatus := strings.ToUpper(t.Status)
	dotStyle := lipgloss.NewStyle().Foreground(Subtle)
	slugStyle := lipgloss.NewStyle().Foreground(Green)

	taskLabel := fmt.Sprintf("%s 🛠  %s %s ",
		prefix,
		slugStyle.Render("["+t.Slug+"]"),
		BoldStyle.Foreground(lipgloss.Color("#FFFFFF")).Render(t.Title),
	)

	sStyled := StatusStyle(tStatus).Render(tStatus)
	sPadded := lipgloss.NewStyle().Width(12).Render(sStyled)

	repeat := rightAlign - lipgloss.Width(taskLabel)
	if repeat < 0 {
		repeat = 0
	}

	return fmt.Sprintf("%s%s %s\n",
		taskLabel,
		dotStyle.Render(strings.Repeat(".", repeat)),
		sPadded,
	)
}

func renderTreeTask(prefix string, t models.Task, rightAlign int) {
	fmt.Print(renderTreeTaskStr(prefix, t, rightAlign))
}

func RenderBacklog() error {
	milestones, _ := store.ListMilestones()
	stories, _ := store.ListStories()
	tasks, _ := store.ListTasks()

	// Maps for lookup
	storyMap := make(map[string]models.Story)
	for _, s := range stories {
		storyMap[s.ID] = s
	}
	taskMap := make(map[string]models.Task)
	for _, t := range tasks {
		taskMap[t.ID] = t
	}

	// 1. Identify Unclaimed Stories (not in any Milestone)
	claimedStories := make(map[string]bool)
	for _, m := range milestones {
		for _, sID := range m.Stories {
			claimedStories[sID] = true
		}
	}

	var backlogStories []models.Story
	for _, s := range stories {
		if !claimedStories[s.ID] {
			backlogStories = append(backlogStories, s)
		}
	}

	// 2. Identify Unclaimed Tasks (not in any Story or Milestone)
	claimedTasks := make(map[string]bool)
	for _, m := range milestones {
		for _, tID := range m.Tasks {
			claimedTasks[tID] = true
		}
	}
	for _, s := range stories {
		for _, tID := range s.Tasks {
			claimedTasks[tID] = true
		}
	}

	var backlogTasks []models.Task
	for _, t := range tasks {
		if !claimedTasks[t.ID] {
			backlogTasks = append(backlogTasks, t)
		}
	}

	// 3. Unclaimed Plans
	linkedPlans := make(map[string]bool)
	for _, t := range tasks {
		planPath := store.GetTaskPlanPath(t.ID)
		if _, err := os.Stat(planPath); err == nil {
			linkedPlans[t.ID] = true
		}
	}

	dir := "plans"
	var backlogPlans []string
	filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}
		id := strings.TrimSuffix(d.Name(), ".md")
		if !linkedPlans[id] {
			backlogPlans = append(backlogPlans, id)
		}
		return nil
	})

	var backlogMilestones []models.Milestone
	for _, m := range milestones {
		if strings.ToUpper(m.Status) == "BACKLOG" {
			backlogMilestones = append(backlogMilestones, m)
		}
	}

	var activeMilestoneBacklogStories []models.Story
	for _, m := range milestones {
		if strings.ToUpper(m.Status) == "ACTIVE" {
			for _, sID := range m.Stories {
				if s, ok := storyMap[sID]; ok {
					if strings.ToUpper(s.Status) == "BACKLOG" {
						activeMilestoneBacklogStories = append(activeMilestoneBacklogStories, s)
					}
				}
			}
		}
	}

	if len(backlogMilestones) == 0 && len(backlogStories) == 0 && len(activeMilestoneBacklogStories) == 0 && len(backlogTasks) == 0 && len(backlogPlans) == 0 {
		fmt.Println("\n  " + lipgloss.NewStyle().Foreground(Gray).Render("Backlog is empty."))
		return nil
	}

	var sb strings.Builder
	totalWidth := GetTerminalWidth()
	rightAlign := totalWidth - 20
	if rightAlign < 40 {
		rightAlign = 40
	}

	if len(backlogMilestones) > 0 {
		store.SortMilestones(backlogMilestones)
		sb.WriteString(HeaderStyle.Render("BACKLOG MILESTONES:") + "\n")
		for _, m := range backlogMilestones {
			slugStyle := lipgloss.NewStyle().Foreground(Green)
			descStyle := lipgloss.NewStyle().Foreground(Yellow).Italic(true)

			desc := ""
			if len(m.Description) > 0 {
				desc = " (" + m.Description[0] + ")"
				if len(m.Description) > 1 || len(m.Description[0]) > 50 {
					if len(m.Description[0]) > 50 {
						desc = " (" + m.Description[0][:47] + "...)"
					}
				}
			}

			headerLabel := fmt.Sprintf("  🏁 %s %s%s ",
				slugStyle.Render("["+m.Slug+"]"),
				BoldStyle.Foreground(lipgloss.Color("#FFFFFF")).Render(strings.ToUpper(m.Title)),
				descStyle.Render(desc),
			)
			sb.WriteString(headerLabel + "\n")

			// Stories under Milestone
			sort.Slice(m.Stories, func(i, j int) bool {
				s1, ok1 := storyMap[m.Stories[i]]
				s2, ok2 := storyMap[m.Stories[j]]
				if ok1 && ok2 {
					return store.CompareNatural(s1.Slug, s2.Slug) < 0
				}
				return m.Stories[i] < m.Stories[j]
			})

			for i, sID := range m.Stories {
				if s, ok := storyMap[sID]; ok {
					branch := "├─"
					if i == len(m.Stories)-1 && len(m.Tasks) == 0 {
						branch = "└─"
					}

					sDesc := ""
					if len(s.Description) > 0 {
						sDesc = " (" + s.Description[0] + ")"
						if len(s.Description) > 1 || len(s.Description[0]) > 40 {
							if len(s.Description[0]) > 40 {
								sDesc = " (" + s.Description[0][:37] + "...)"
							}
						}
					}

					storyLabel := fmt.Sprintf("    %s %s %s %s%s ",
						branch,
						lipgloss.NewStyle().Foreground(Magenta).Render("📖"),
						slugStyle.Render("["+s.Slug+"]"),
						BoldStyle.Foreground(lipgloss.Color("#FFFFFF")).Render(s.Title),
						descStyle.Render(sDesc),
					)
					sb.WriteString(storyLabel + "\n")

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
							taskBranch := "├─"
							if j == len(s.Tasks)-1 {
								taskBranch = "└─"
							}
							parentLine := "│"
							if i == len(m.Stories)-1 && len(m.Tasks) == 0 {
								parentLine = " "
							}

							tSlugStyle := lipgloss.NewStyle().Foreground(Green)
							sb.WriteString(fmt.Sprintf("    %s     %s 🛠  %s %s\n",
								parentLine, taskBranch,
								tSlugStyle.Render("["+t.Slug+"]"),
								BoldStyle.Foreground(lipgloss.Color("#FFFFFF")).Render(t.Title)))
						}
					}
				}
			}

			// Direct Tasks under Milestone
			if len(m.Tasks) > 0 {
				sort.Slice(m.Tasks, func(i, j int) bool {
					t1, ok1 := taskMap[m.Tasks[i]]
					t2, ok2 := taskMap[m.Tasks[j]]
					if ok1 && ok2 {
						return store.CompareNatural(t1.Slug, t2.Slug) < 0
					}
					return m.Tasks[i] < m.Tasks[j]
				})
				for j, tID := range m.Tasks {
					if t, ok := taskMap[tID]; ok {
						branch := "├─"
						if j == len(m.Tasks)-1 {
							branch = "└─"
						}
						tSlugStyle := lipgloss.NewStyle().Foreground(Green)
						sb.WriteString(fmt.Sprintf("    %s 🛠  %s %s\n",
							branch, tSlugStyle.Render("["+t.Slug+"]"),
							BoldStyle.Foreground(lipgloss.Color("#FFFFFF")).Render(t.Title)))
					}
				}
			}
		}
		sb.WriteString("\n")
	}

	if len(activeMilestoneBacklogStories) > 0 {
		sort.Slice(activeMilestoneBacklogStories, func(i, j int) bool {
			return store.CompareNatural(activeMilestoneBacklogStories[i].Slug, activeMilestoneBacklogStories[j].Slug) < 0
		})

		sb.WriteString(HeaderStyle.Render("ACTIVE MILESTONE BACKLOG STORIES:") + "\n")
		for _, s := range activeMilestoneBacklogStories {
			slugStyle := lipgloss.NewStyle().Foreground(Green)
			descStyle := lipgloss.NewStyle().Foreground(Yellow).Italic(true)

			sStatus := strings.ToUpper(s.Status)
			sStatusStyled := StatusStyle(sStatus).Render(sStatus)
			sStatusPadded := lipgloss.NewStyle().Width(12).Render(sStatusStyled)

			desc := ""
			if len(s.Description) > 0 {
				desc = " (" + s.Description[0] + ")"
				if len(s.Description) > 1 || len(s.Description[0]) > 40 {
					if len(s.Description[0]) > 40 {
						desc = " (" + s.Description[0][:37] + "...)"
					}
				}
			}

			storyLabel := fmt.Sprintf("  • %s %s %s%s ",
				lipgloss.NewStyle().Foreground(Magenta).Render("📖"),
				slugStyle.Render("["+s.Slug+"]"),
				BoldStyle.Foreground(lipgloss.Color("#FFFFFF")).Render(s.Title),
				descStyle.Render(desc))

			sRepeat := rightAlign - lipgloss.Width(storyLabel)
			if sRepeat < 0 {
				sRepeat = 0
			}

			sb.WriteString(fmt.Sprintf("%s%s %s\n",
				storyLabel,
				lipgloss.NewStyle().Foreground(Subtle).Render(strings.Repeat(".", sRepeat)),
				sStatusPadded))

			// Tasks under these stories
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
					tSlugStyle := lipgloss.NewStyle().Foreground(Green)
					sb.WriteString(fmt.Sprintf("      %s 🛠  %s %s\n",
						branch, tSlugStyle.Render("["+t.Slug+"]"),
						BoldStyle.Foreground(lipgloss.Color("#FFFFFF")).Render(t.Title)))
				}
			}
		}
		sb.WriteString("\n")
	}

	if len(backlogStories) > 0 {
		sort.Slice(backlogStories, func(i, j int) bool {
			return store.CompareNatural(backlogStories[i].Slug, backlogStories[j].Slug) < 0
		})

		sb.WriteString(HeaderStyle.Render("UNCLAIMED STORIES:") + "\n")
		for _, s := range backlogStories {
			sStatus := strings.ToUpper(s.Status)
			sStatusStyled := StatusStyle(sStatus).Render(sStatus)
			sStatusPadded := lipgloss.NewStyle().Width(12).Render(sStatusStyled)

			desc := ""
			if len(s.Description) > 0 {
				desc = " (" + s.Description[0] + ")"
				if len(s.Description) > 1 || len(s.Description[0]) > 40 {
					if len(s.Description[0]) > 40 {
						desc = " (" + s.Description[0][:37] + "...)"
					}
				}
			}

			slugStyle := lipgloss.NewStyle().Foreground(Green)
			descStyle := lipgloss.NewStyle().Foreground(Yellow).Italic(true)

			storyLabel := fmt.Sprintf("  • %s %s %s%s ",
				lipgloss.NewStyle().Foreground(Magenta).Render("📖"),
				slugStyle.Render("["+s.Slug+"]"),
				BoldStyle.Foreground(lipgloss.Color("#FFFFFF")).Render(s.Title),
				descStyle.Render(desc))

			sRepeat := rightAlign - lipgloss.Width(storyLabel)
			if sRepeat < 0 {
				sRepeat = 0
			}

			sb.WriteString(fmt.Sprintf("%s%s %s\n",
				storyLabel,
				lipgloss.NewStyle().Foreground(Subtle).Render(strings.Repeat(".", sRepeat)),
				sStatusPadded))

			// Tasks under Unclaimed Story
			store.NaturalSort(s.Tasks)
			for j, tID := range s.Tasks {
				if t, ok := taskMap[tID]; ok {
					branch := "├─"
					if j == len(s.Tasks)-1 {
						branch = "└─"
					}
					sb.WriteString(renderTreeTaskStr("      "+branch, t, rightAlign))
				}
			}
		}
		sb.WriteString("\n")
	}

	if len(backlogTasks) > 0 {
		sort.Slice(backlogTasks, func(i, j int) bool {
			return store.CompareNatural(backlogTasks[i].Slug, backlogTasks[j].Slug) < 0
		})

		sb.WriteString(HeaderStyle.Render("INDEPENDENT UNCLAIMED TASKS:") + "\n")
		for _, t := range backlogTasks {
			sb.WriteString(renderTreeTaskStr("  • ", t, rightAlign))
		}
		sb.WriteString("\n")
	}

	if len(backlogPlans) > 0 {
		sb.WriteString(HeaderStyle.Render("UNCLAIMED PLANS:") + "\n")
		for _, id := range backlogPlans {
			sb.WriteString(fmt.Sprintf("  • %s %s\n",
				lipgloss.NewStyle().Foreground(Cyan).Render("📝"),
				lipgloss.NewStyle().Foreground(Gray).Render("["+id+"]")))
		}
	}

	content := strings.TrimSpace(sb.String())

	windowStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Subtle).
		Padding(1, 2).
		Width(totalWidth - 2)

	renderedWindow := windowStyle.Render(content)
	actualWidth := lipgloss.Width(renderedWindow)

	lines := strings.Split(renderedWindow, "\n")
	if len(lines) > 0 {
		titleText := " PROJECT BACKLOG "
		titleWidth := len(titleText)
		border := lipgloss.RoundedBorder()
		dashesLeft := (actualWidth - 2 - titleWidth) / 2
		dashesRight := actualWidth - 2 - titleWidth - dashesLeft
		newTopLine := lipgloss.NewStyle().Foreground(Subtle).Render(border.TopLeft+strings.Repeat(border.Top, dashesLeft)) +
			BoldStyle.Foreground(lipgloss.Color("#FFFFFF")).Render(titleText) +
			lipgloss.NewStyle().Foreground(Subtle).Render(strings.Repeat(border.Top, dashesRight)+border.TopRight)
		lines[0] = newTopLine
	}

	fmt.Println("")
	fmt.Println(strings.Join(lines, "\n"))
	return nil
}
