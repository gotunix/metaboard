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

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"gotunix.net/metaboard/internal/interactive"
	"gotunix.net/metaboard/internal/models"
	"gotunix.net/metaboard/internal/store"
	"gotunix.net/metaboard/internal/ui"
)

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage project tasks",
}

var taskCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new task",
	Run: func(cmd *cobra.Command, args []string) {
		title, _ := cmd.Flags().GetString("title")
		slug, _ := cmd.Flags().GetString("slug")
		priority, _ := cmd.Flags().GetString("priority")
		tType, _ := cmd.Flags().GetString("type")
		assigned, _ := cmd.Flags().GetString("assigned-to")
		description, _ := cmd.Flags().GetString("description")

		if title == "" {
			fmt.Println(lipgloss.NewStyle().Foreground(ui.Red).Render("Error: --title is required"))
			return
		}

		finalSlug, err := store.CreateTask(title, slug, priority, tType, assigned, description)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Println(ui.BoldStyle.Foreground(ui.Green).Render(fmt.Sprintf("✔ Created Task [%s]", finalSlug)))
		}
	},
}

var taskViewCmd = &cobra.Command{
	Use:   "view <slug>",
	Short: "Display task details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := ui.ViewTask(args[0]); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	},
}

var taskStatusCmd = &cobra.Command{
	Use:   "status <slug> <status>",
	Short: "Update task status",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		status := strings.ToUpper(strings.ReplaceAll(args[1], "_", "-"))
		if !models.IsValidStatus(status, models.ValidTaskStatuses) {
			fmt.Printf(lipgloss.NewStyle().Foreground(ui.Red).Render("Error: invalid status %q. Allowed: %v\n"), status, models.ValidTaskStatuses)
			return
		}
		if err := store.UpdateTaskStatus(args[0], status); err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Println(ui.BoldStyle.Foreground(ui.Green).Render("✔ Task status updated"))
		}
	},
}

var taskPlanCmd = &cobra.Command{
	Use:   "plan <slug>",
	Short: "Create or edit implementation plan",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		planPath, err := store.EnsureTaskPlan(args[0])
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vim"
		}
		c := exec.Command(editor, planPath)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Run()
	},
}

var taskEditCmd = &cobra.Command{
	Use:   "edit <slug>",
	Short: "Edit a task",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Usage: metaboard task edit <slug> [flags]")
			return
		}
		idOrSlug := args[0]

		// Check if any flags were set
		if cmd.Flags().NFlag() == 0 {
			interactive.EditTaskInteractive(idOrSlug)
			return
		}

		update := store.TaskUpdate{}
		if cmd.Flags().Changed("title") {
			v, _ := cmd.Flags().GetString("title")
			update.Title = &v
		}
		if cmd.Flags().Changed("status") {
			v, _ := cmd.Flags().GetString("status")
			update.Status = &v
		}
		if cmd.Flags().Changed("priority") {
			v, _ := cmd.Flags().GetString("priority")
			update.Priority = &v
		}
		if cmd.Flags().Changed("type") {
			v, _ := cmd.Flags().GetString("type")
			update.Type = &v
		}
		if cmd.Flags().Changed("assigned-to") {
			v, _ := cmd.Flags().GetString("assigned-to")
			update.AssignedTo = &v
		}
		if cmd.Flags().Changed("description") {
			v, _ := cmd.Flags().GetString("description")
			update.Description = &v
		}
		if cmd.Flags().Changed("tags") {
			v, _ := cmd.Flags().GetString("tags")
			tags := []string{}
			for _, tag := range strings.Split(v, ",") {
				tag = strings.TrimSpace(tag)
				if tag != "" {
					tags = append(tags, tag)
				}
			}
			update.Tags = &tags
		}
		if cmd.Flags().Changed("depends") {
			v, _ := cmd.Flags().GetString("depends")
			deps := []string{}
			for _, dep := range strings.Split(v, ",") {
				dep = strings.TrimSpace(dep)
				if dep != "" {
					deps = append(deps, dep)
				}
			}
			update.DependsOn = &deps
		}

		if err := store.UpdateTask(idOrSlug, update); err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Println(ui.BoldStyle.Foreground(ui.Green).Render("✔ Task updated successfully"))
		}
	},
}

func init() {
	taskCmd.SetHelpFunc(ui.HandleHelp)
	taskCmd.SetUsageFunc(ui.HandleUsage)
	taskCreateCmd.Flags().StringP("title", "t", "", "Title of the task")
	taskCreateCmd.Flags().StringP("slug", "s", "", "Unique slug for the task")
	taskCreateCmd.Flags().StringP("priority", "p", "MEDIUM", "Priority (LOW, MEDIUM, HIGH)")
	taskCreateCmd.Flags().String("type", "TASK", "Type (FEAT, FIX, etc.)")
	taskCreateCmd.Flags().String("assigned-to", "", "Assignee for the task")
	taskCreateCmd.Flags().String("description", "", "Task description")

	taskEditCmd.Flags().StringP("title", "t", "", "New title")
	taskEditCmd.Flags().String("status", "", "New status")
	taskEditCmd.Flags().StringP("priority", "p", "", "New priority")
	taskEditCmd.Flags().String("type", "", "New type")
	taskEditCmd.Flags().String("assigned-to", "", "New assignee")
	taskEditCmd.Flags().String("description", "", "New description")
	taskEditCmd.Flags().String("tags", "", "New tags (comma separated)")
	taskEditCmd.Flags().String("depends", "", "New dependencies (comma separated IDs)")

	taskCmd.AddCommand(taskCreateCmd)
	taskCmd.AddCommand(taskViewCmd)
	taskCmd.AddCommand(taskStatusCmd)
	taskCmd.AddCommand(taskPlanCmd)
	taskCmd.AddCommand(taskEditCmd)
	rootCmd.AddCommand(taskCmd)
}
