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

package interactive

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"gotunix.net/metaboard/internal/models"
	"gotunix.net/metaboard/internal/store"
	"gotunix.net/metaboard/internal/ui"
)

func EditMilestoneInteractive(idOrSlug string) error {
	ms, _ := store.ListMilestones()
	var m models.Milestone
	found := false
	for _, mil := range ms {
		if mil.ID == idOrSlug || mil.Slug == idOrSlug {
			m = mil
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("milestone %q not found", idOrSlug)
	}

	descStr := strings.Join(m.Description, "\n")
	storiesStr := strings.Join(m.Stories, ", ")
	tasksStr := strings.Join(m.Tasks, ", ")

	theme := huh.ThemeCharm()
	theme.Group.Base = theme.Group.Base.Border(lipgloss.NormalBorder()).Padding(1, 2).BorderForeground(ui.Magenta)
	theme.Group.Title = theme.Group.Title.Foreground(ui.Magenta).Bold(true).MarginBottom(1)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Title").Value(&m.Title).Validate(func(s string) error {
				if len(s) == 0 {
					return fmt.Errorf("title cannot be empty")
				}
				return nil
			}),
			huh.NewInput().Title("Slug").Value(&m.Slug).Validate(func(s string) error {
				if len(s) == 0 {
					return fmt.Errorf("slug cannot be empty")
				}
				return nil
			}),
			huh.NewSelect[string]().Title("Status").Options(
				huh.NewOption("BACKLOG", "BACKLOG"),
				huh.NewOption("ACTIVE", "ACTIVE"),
				huh.NewOption("COMPLETED", "COMPLETED"),
				huh.NewOption("CANCELLED", "CANCELLED"),
			).Value(&m.Status),
			huh.NewText().Title("Description").Value(&descStr).Lines(8),
			huh.NewInput().Title("Linked Stories (comma separated IDs)").Value(&storiesStr),
			huh.NewInput().Title("Linked Tasks (comma separated IDs)").Value(&tasksStr),
		).Title("EDIT MILESTONE: " + m.Slug),
	).WithTheme(theme).WithShowHelp(true)

	if err := form.Run(); err != nil {
		return err
	}

	m.Description = strings.Split(descStr, "\n")
	for len(m.Description) > 0 && m.Description[len(m.Description)-1] == "" {
		m.Description = m.Description[:len(m.Description)-1]
	}

	m.Stories = []string{}
	for _, id := range strings.Split(storiesStr, ",") {
		id = strings.TrimSpace(id)
		if id != "" {
			m.Stories = append(m.Stories, id)
		}
	}
	m.Tasks = []string{}
	for _, id := range strings.Split(tasksStr, ",") {
		id = strings.TrimSpace(id)
		if id != "" {
			m.Tasks = append(m.Tasks, id)
		}
	}

	if m.Status == "COMPLETED" {
		if m.CompletedAt == "" {
			m.CompletedAt = time.Now().Format("2006-01-02T15:04:05Z")
		}
	} else {
		m.CompletedAt = ""
	}
	return store.SaveMilestone(m)
}

func EditStoryInteractive(idOrSlug string) error {
	ss, _ := store.ListStories()
	var s models.Story
	found := false
	for _, st := range ss {
		if st.ID == idOrSlug || st.Slug == idOrSlug {
			s = st
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("story %q not found", idOrSlug)
	}

	descStr := strings.Join(s.Description, "\n")
	tasksStr := strings.Join(s.Tasks, ", ")

	theme := huh.ThemeCharm()
	theme.Group.Base = theme.Group.Base.Border(lipgloss.NormalBorder()).Padding(1, 2).BorderForeground(ui.Magenta)
	theme.Group.Title = theme.Group.Title.Foreground(ui.Magenta).Bold(true).MarginBottom(1)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Title").Value(&s.Title).Validate(func(str string) error {
				if len(str) == 0 {
					return fmt.Errorf("title cannot be empty")
				}
				return nil
			}),
			huh.NewInput().Title("Slug").Value(&s.Slug).Validate(func(str string) error {
				if len(str) == 0 {
					return fmt.Errorf("slug cannot be empty")
				}
				return nil
			}),
			huh.NewSelect[string]().Title("Status").Options(
				huh.NewOption("BACKLOG", "BACKLOG"),
				huh.NewOption("ACTIVE", "ACTIVE"),
				huh.NewOption("COMPLETED", "COMPLETED"),
				huh.NewOption("CANCELLED", "CANCELLED"),
			).Value(&s.Status),
			huh.NewText().Title("Description").Value(&descStr).Lines(8),
			huh.NewInput().Title("Linked Tasks (comma separated IDs)").Value(&tasksStr),
		).Title("EDIT STORY: " + s.Slug),
	).WithTheme(theme).WithShowHelp(true)

	if err := form.Run(); err != nil {
		return err
	}

	s.Description = strings.Split(descStr, "\n")
	for len(s.Description) > 0 && s.Description[len(s.Description)-1] == "" {
		s.Description = s.Description[:len(s.Description)-1]
	}

	s.Tasks = []string{}
	for _, id := range strings.Split(tasksStr, ",") {
		id = strings.TrimSpace(id)
		if id != "" {
			s.Tasks = append(s.Tasks, id)
		}
	}

	if s.Status == "COMPLETED" {
		if s.CompletedAt == "" {
			s.CompletedAt = time.Now().Format("2006-01-02T15:04:05Z")
		}
	} else {
		s.CompletedAt = ""
	}
	return store.SaveStory(s)
}

func EditTaskInteractive(idOrSlug string) error {
	ts, _ := store.ListTasks()
	var t models.Task
	found := false
	for _, tk := range ts {
		if tk.ID == idOrSlug || tk.Slug == idOrSlug {
			t = tk
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("task %q not found", idOrSlug)
	}

	descStr := strings.Join(t.Description, "\n")
	//	tagsStr := strings.Join(t.Tags, ", ")
	//	depsStr := strings.Join(t.DependsOn, ", ")

	theme := huh.ThemeCharm()
	theme.Group.Base = theme.Group.Base.Border(lipgloss.NormalBorder()).Padding(1, 2).BorderForeground(ui.Magenta)
	theme.Group.Title = theme.Group.Title.Foreground(ui.Magenta).Bold(true).MarginBottom(1)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Title").Value(&t.Title).Validate(func(str string) error {
				if len(str) == 0 {
					return fmt.Errorf("title cannot be empty")
				}
				return nil
			}),
			huh.NewInput().Title("Slug").Value(&t.Slug).Validate(func(str string) error {
				if len(str) == 0 {
					return fmt.Errorf("slug cannot be empty")
				}
				return nil
			}),
			huh.NewSelect[string]().Title("Status").Options(
				huh.NewOption("BACKLOG", "BACKLOG"),
				huh.NewOption("ACTIVE", "ACTIVE"),
				huh.NewOption("IN-PROGRESS", "IN-PROGRESS"),
				huh.NewOption("COMPLETED", "COMPLETED"),
				huh.NewOption("CANCELLED", "CANCELLED"),
			).Value(&t.Status),
			huh.NewSelect[string]().Title("Priority").Options(
				huh.NewOption("LOW", "LOW"),
				huh.NewOption("MEDIUM", "MEDIUM"),
				huh.NewOption("HIGH", "HIGH"),
			).Value(&t.Priority),
			huh.NewSelect[string]().Title("Type").Options(
				huh.NewOption("TASK", "TASK"),
				huh.NewOption("FEATURE", "FEATURE"),
				huh.NewOption("BUG", "BUG"),
				huh.NewOption("REFACTOR", "REFACTOR"),
				huh.NewOption("CHORE", "CHORE"),
				huh.NewOption("DOCS", "DOCS"),
				huh.NewOption("TEST", "TEST"),
			).Value(&t.Type),
			huh.NewInput().Title("Assigned To").Value(&t.AssignedTo),
			huh.NewText().Title("Description").Value(&descStr).Lines(8),
		).Title("EDIT TASK: " + t.Slug),
	).WithTheme(theme).WithShowHelp(true)

	if err := form.Run(); err != nil {
		return err
	}

	t.Description = strings.Split(descStr, "\n")
	for len(t.Description) > 0 && t.Description[len(t.Description)-1] == "" {
		t.Description = t.Description[:len(t.Description)-1]
	}

	//	t.Tags = []string{}
	//	for _, tag := range strings.Split(tagsStr, ",") {
	//		tag = strings.TrimSpace(tag)
	//		if tag != "" { t.Tags = append(t.Tags, tag) }
	//	}
	//	t.DependsOn = []string{}
	//	for _, dep := range strings.Split(depsStr, ",") {
	//		dep = strings.TrimSpace(dep)
	//		if dep != "" { t.DependsOn = append(t.DependsOn, dep) }
	//	}

	//	if t.Status == "COMPLETED" {
	//		if t.CompletedAt == "" { t.CompletedAt = time.Now().Format("2006-01-02T15:04:05Z") }
	//	} else {
	//		t.CompletedAt = ""
	//	}
	return store.SaveTask(t)
}
