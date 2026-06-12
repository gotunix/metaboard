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

	"gotunix.net/metaboard/internal/models"
	"gotunix.net/metaboard/internal/store"
)

func GenerateChangelog(outputDir string) error {
	milestones, _ := store.ListMilestones()
	stories, _ := store.ListStories()
	tasks, _ := store.ListTasks()

	storyMap := make(map[string]models.Story)
	for _, s := range stories {
		storyMap[s.ID] = s
	}
	taskMap := make(map[string]models.Task)
	for _, t := range tasks {
		taskMap[t.ID] = t
	}

	handled := make(map[string]bool)
	changelogFile := filepath.Join(outputDir, "CHANGELOG.md")
	f, err := os.Create(changelogFile)
	if err != nil {
		return err
	}
	defer f.Close()

	f.WriteString("# Changelog\n\nAll notable changes to this project will be documented in this file.\n\n")

	var completedMilestones []models.Milestone
	for _, m := range milestones {
		if strings.ToUpper(m.Status) == "COMPLETED" {
			completedMilestones = append(completedMilestones, m)
		}
	}

	sort.Slice(completedMilestones, func(i, j int) bool {
		return completedMilestones[i].CompletedAt > completedMilestones[j].CompletedAt
	})

	renderCategory := func(sb *strings.Builder, category string, tks []models.Task) {
		if len(tks) > 0 {
			sb.WriteString("### " + category + "\n")
			for _, t := range tks {
				sb.WriteString(fmt.Sprintf("- [TASK: %s] - %s\n", t.Slug, t.Title))
			}
			sb.WriteString("\n")
		}
	}

	for _, m := range completedMilestones {
		mDate := "YYYY-MM-DD"
		if len(m.CompletedAt) >= 10 {
			mDate = m.CompletedAt[:10]
		}
		f.WriteString(fmt.Sprintf("## [%s] - %s\n\n", m.Title, mDate))

		var mTasks []models.Task
		for _, sID := range m.Stories {
			if s, ok := storyMap[sID]; ok && strings.ToUpper(s.Status) == "COMPLETED" {
				handled[s.ID] = true
				for _, tID := range s.Tasks {
					if t, ok := taskMap[tID]; ok && strings.ToUpper(t.Status) == "COMPLETED" {
						mTasks = append(mTasks, t)
						handled[t.ID] = true
					}
				}
			}
		}
		for _, tID := range m.Tasks {
			if t, ok := taskMap[tID]; ok && strings.ToUpper(t.Status) == "COMPLETED" {
				mTasks = append(mTasks, t)
				handled[t.ID] = true
			}
		}

		sort.Slice(mTasks, func(i, j int) bool {
			return mTasks[i].CompletedAt < mTasks[j].CompletedAt
		})

		groups := make(map[string][]models.Task)
		for _, t := range mTasks {
			cat := "Changed"
			switch strings.ToUpper(t.Type) {
			case "FEAT", "FEATURE", "ADD":
				cat = "Added"
			case "FIX", "BUG", "HOTFIX":
				cat = "Fixed"
			case "SECURITY":
				cat = "Security"
			case "REFACTOR":
				cat = "Updated"
			case "DEPRECATED":
				cat = "Deprecated"
			case "REMOVE", "REMOVED":
				cat = "Removed"
			case "CHORE", "DOCS", "TEST", "MAINT":
				cat = "Maintenance"
			}
			groups[cat] = append(groups[cat], t)
		}

		var sb strings.Builder
		renderCategory(&sb, "Security", groups["Security"])
		renderCategory(&sb, "Deprecated", groups["Deprecated"])
		renderCategory(&sb, "Added", groups["Added"])
		renderCategory(&sb, "Fixed", groups["Fixed"])
		renderCategory(&sb, "Updated", groups["Updated"])
		renderCategory(&sb, "Changed", groups["Changed"])
		renderCategory(&sb, "Removed", groups["Removed"])
		renderCategory(&sb, "Maintenance", groups["Maintenance"])
		f.WriteString(sb.String())
	}

	var unreleasedTasks []models.Task
	for _, s := range stories {
		if strings.ToUpper(s.Status) == "COMPLETED" && !handled[s.ID] {
			for _, tID := range s.Tasks {
				if t, ok := taskMap[tID]; ok && strings.ToUpper(t.Status) == "COMPLETED" && !handled[t.ID] {
					unreleasedTasks = append(unreleasedTasks, t)
					handled[t.ID] = true
				}
			}
		}
	}
	for _, t := range tasks {
		if strings.ToUpper(t.Status) == "COMPLETED" && !handled[t.ID] {
			unreleasedTasks = append(unreleasedTasks, t)
			handled[t.ID] = true
		}
	}

	sort.Slice(unreleasedTasks, func(i, j int) bool {
		return unreleasedTasks[i].CompletedAt < unreleasedTasks[j].CompletedAt
	})

	if len(unreleasedTasks) > 0 {
		f.WriteString("## [Unreleased]\n\n")
		groups := make(map[string][]models.Task)
		for _, t := range unreleasedTasks {
			cat := "Changed"
			switch strings.ToUpper(t.Type) {
			case "FEAT", "FEATURE", "ADD":
				cat = "Added"
			case "FIX", "BUG", "HOTFIX":
				cat = "Fixed"
			case "SECURITY":
				cat = "Security"
			case "REFACTOR":
				cat = "Updated"
			case "DEPRECATED":
				cat = "Deprecated"
			case "REMOVE", "REMOVED":
				cat = "Removed"
			case "CHORE", "DOCS", "TEST", "MAINT":
				cat = "Maintenance"
			}
			groups[cat] = append(groups[cat], t)
		}
		var sb strings.Builder
		renderCategory(&sb, "Security", groups["Security"])
		renderCategory(&sb, "Deprecated", groups["Deprecated"])
		renderCategory(&sb, "Added", groups["Added"])
		renderCategory(&sb, "Fixed", groups["Fixed"])
		renderCategory(&sb, "Updated", groups["Updated"])
		renderCategory(&sb, "Changed", groups["Changed"])
		renderCategory(&sb, "Removed", groups["Removed"])
		renderCategory(&sb, "Maintenance", groups["Maintenance"])
		f.WriteString(sb.String())
	}
	return nil
}
