// SPDX-License-Identifier: GPL-3.0-or-later
// SPDX-FileCopyrightText: 2026 The MetaBoard authors

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"gotunix.net/metaboard/internal/git"
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
		priority, _ := cmd.Flags().GetString("priority")
		tType, _ := cmd.Flags().GetString("type")
		assigned, _ := cmd.Flags().GetString("assigned-to")
		description, _ := cmd.Flags().GetString("description")
		tagsStr, _ := cmd.Flags().GetString("tags")
		depsStr, _ := cmd.Flags().GetString("depends")
		changelog, _ := cmd.Flags().GetBool("changelog")
		milestoneSlug, _ := cmd.Flags().GetString("milestone")
		slug, _ := cmd.Flags().GetString("slug")

		if title == "" && cmd.Flags().NFlag() == 0 {
			fmt.Println(lipgloss.NewStyle().Foreground(ui.Red).Render("Error: --title is required"))
			return
		}
		if title == "" {
			fmt.Println(lipgloss.NewStyle().Foreground(ui.Red).Render("Error: --title is required"))
			return
		}

		root, _ := store.GetDataRoot()
		if git.IsGitRepo(root) {
			_ = git.Pull(root)
		}

		finalSlug, err := store.CreateTask(title, slug, priority, tType, assigned, description)
		if err != nil {
			fmt.Println(lipgloss.NewStyle().Foreground(ui.Red).Render(fmt.Sprintf("Error: %v", err)))
			return
		}

		t, errT := store.GetTask(finalSlug)
		if errT == nil {
			if tagsStr != "" {
				for _, tag := range strings.Split(tagsStr, ",") {
					tag = strings.TrimSpace(tag)
					if tag != "" {
						t.Tags = append(t.Tags, tag)
					}
				}
			}
			if depsStr != "" {
				for _, dep := range strings.Split(depsStr, ",") {
					dep = strings.TrimSpace(dep)
					if dep != "" {
						t.DependsOn = append(t.DependsOn, dep)
					}
				}
			}
			t.Changelog = changelog
			_ = store.SaveTask(*t)
		}

		if milestoneSlug != "" {
			_ = store.LinkEntities(finalSlug, milestoneSlug)
		}

		if git.IsGitRepo(root) {
			t, _ := store.GetTask(finalSlug)
			path, _ := store.GetTaskPath(t.ID)
			_ = git.Commit(root, []string{path}, fmt.Sprintf("boards: create task [%s] - %s", t.Slug, t.Title))
		}

		fmt.Println(ui.BoldStyle.Foreground(ui.Green).Render(fmt.Sprintf("✔ Created Task [%s]", finalSlug)))
	},
}

var taskViewCmd = &cobra.Command{
	Use:   "view [slug]",
	Short: "Display task details",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var idOrSlug string
		if len(args) > 0 {
			idOrSlug = args[0]
		} else {
			fmt.Println(lipgloss.NewStyle().Foreground(ui.Red).Render("Error: task slug is required"))
			return
		}
		version, _ := cmd.Flags().GetInt("version")
		if err := ui.ViewTask(idOrSlug, version); err != nil {
			fmt.Println(lipgloss.NewStyle().Foreground(ui.Red).Render(fmt.Sprintf("Error: %v", err)))
		}
	},
}

var taskListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tasks",
	Run: func(cmd *cobra.Command, args []string) {
		ts, err := store.ListTasks()
		if err != nil {
			fmt.Println(lipgloss.NewStyle().Foreground(ui.Red).Render(fmt.Sprintf("Error: %v", err)))
			return
		}
		if len(ts) == 0 {
			fmt.Println("No tasks found.")
			return
		}
		for _, t := range ts {
			statusStyle := ui.StatusStyle(t.Status)
			fmt.Printf("[%s] %s %s\n", t.Slug, t.Title, statusStyle.Render(t.Status))
		}
	},
}

var taskEditCmd = &cobra.Command{
	Use:   "edit <slug>",
	Short: "Edit a task",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		slug := args[0]
		title, _ := cmd.Flags().GetString("title")
		priority, _ := cmd.Flags().GetString("priority")
		tType, _ := cmd.Flags().GetString("type")
		assigned, _ := cmd.Flags().GetString("assigned-to")
		description, _ := cmd.Flags().GetString("description")
		status, _ := cmd.Flags().GetString("status")
		newSlug, _ := cmd.Flags().GetString("slug")
		tagsStr, _ := cmd.Flags().GetString("tags")
		depsStr, _ := cmd.Flags().GetString("depends")
		changelog, _ := cmd.Flags().GetBool("changelog")
		milestoneSlug, _ := cmd.Flags().GetString("milestone")

		t, err := store.GetTask(slug)
		if err != nil {
			fmt.Println(lipgloss.NewStyle().Foreground(ui.Red).Render(fmt.Sprintf("Error: %v", err)))
			return
		}

		if title != "" {
			t.Title = title
		}
		if newSlug != "" {
			t.Slug = newSlug
		}
		if priority != "" {
			t.Priority = priority
		}
		if tType != "" {
			t.Type = tType
		}
		if assigned != "" {
			t.AssignedTo = assigned
		}
		if description != "" {
			t.Description = description
		}
		if status != "" {
			t.Status = strings.ToUpper(status)
		}
		if tagsStr != "" {
			t.Tags = []string{}
			for _, tag := range strings.Split(tagsStr, ",") {
				tag = strings.TrimSpace(tag)
				if tag != "" {
					t.Tags = append(t.Tags, tag)
				}
			}
		}
		if depsStr != "" {
			t.DependsOn = []string{}
			for _, dep := range strings.Split(depsStr, ",") {
				dep = strings.TrimSpace(dep)
				if dep != "" {
					t.DependsOn = append(t.DependsOn, dep)
				}
			}
		}
		if cmd.Flags().Changed("changelog") {
			t.Changelog = changelog
		}

		if t.Status == "COMPLETED" && t.CompletedAt == "" {
			// completed at would be set
		} else if t.Status != "COMPLETED" {
			t.CompletedAt = ""
		}

		if err := store.SaveTask(*t); err != nil {
			fmt.Println(lipgloss.NewStyle().Foreground(ui.Red).Render(fmt.Sprintf("Error: %v", err)))
			return
		}

		if milestoneSlug != "" {
			_ = store.LinkEntities(t.ID, milestoneSlug)
		}

		root, _ := store.GetDataRoot()
		if git.IsGitRepo(root) {
			path, _ := store.GetTaskPath(t.ID)
			_ = git.Commit(root, []string{path}, fmt.Sprintf("boards: update task [%s] - %s", t.Slug, t.Title))
		}
		fmt.Println(ui.BoldStyle.Foreground(ui.Green).Render(fmt.Sprintf("✔ Updated Task [%s]", t.Slug)))
	},
}

var taskDeleteCmd = &cobra.Command{
	Use:   "delete <slug>",
	Short: "Delete a task",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		slug := args[0]
		root, _ := store.GetDataRoot()
		if git.IsGitRepo(root) {
			_ = git.Pull(root)
		}
		if err := store.DeleteTask(slug); err != nil {
			fmt.Println(lipgloss.NewStyle().Foreground(ui.Red).Render(fmt.Sprintf("Error: %v", err)))
			return
		}
		fmt.Println(ui.BoldStyle.Foreground(ui.Green).Render(fmt.Sprintf("✔ Deleted Task [%s]", slug)))
		if git.IsGitRepo(root) {
			_ = git.Commit(root, []string{"."}, fmt.Sprintf("boards: delete task [%s]", slug))
		}
	},
}

var taskHistoryCmd = &cobra.Command{
	Use:   "history <slug>",
	Short: "Show task version history",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := ui.ViewTaskHistory(args[0]); err != nil {
			fmt.Println(lipgloss.NewStyle().Foreground(ui.Red).Render(fmt.Sprintf("Error: %v", err)))
		}
	},
}

var taskPlanCmd = &cobra.Command{
	Use:   "plan <slug>",
	Short: "Open task plan in editor",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		slug := args[0]
		planPath, err := store.EnsureTaskPlan(slug)
		if err != nil {
			fmt.Println(lipgloss.NewStyle().Foreground(ui.Red).Render(fmt.Sprintf("Error: %v", err)))
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
		_ = c.Run()
	},
}

var taskStatusCmd = &cobra.Command{
	Use:   "status <slug> <status>",
	Short: "Update task status",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		slug := args[0]
		newStatus := strings.ToUpper(args[1])
		if err := store.UpdateTaskStatus(slug, newStatus); err != nil {
			fmt.Println(lipgloss.NewStyle().Foreground(ui.Red).Render(fmt.Sprintf("Error: %v", err)))
			return
		}
		fmt.Println(ui.BoldStyle.Foreground(ui.Green).Render(fmt.Sprintf("✔ Updated Task [%s] status to %s", slug, newStatus)))
	},
}

func init() {
	taskCreateCmd.Flags().String("title", "", "Task title")
	taskCreateCmd.Flags().String("priority", "NORMAL", "Priority (LOW, NORMAL, HIGH, CRITICAL)")
	taskCreateCmd.Flags().String("type", "TASK", "Task type (FEATURE, BUG, CHORE, R&D, MAINTENANCE, INFRA)")
	taskCreateCmd.Flags().String("assigned-to", "", "Assigned user")
	taskCreateCmd.Flags().String("description", "", "Task description")
	taskCreateCmd.Flags().String("tags", "", "Comma-separated tags")
	taskCreateCmd.Flags().String("depends", "", "Comma-separated task dependencies (slugs)")
	taskCreateCmd.Flags().Bool("changelog", false, "Include in changelog")
	taskCreateCmd.Flags().String("milestone", "", "Link to milestone (slug)")
	taskCreateCmd.Flags().String("slug", "", "Custom slug (optional)")
	taskCreateCmd.SetHelpFunc(ui.HandleHelp)
	taskCreateCmd.SetUsageFunc(ui.HandleUsage)

	taskViewCmd.Flags().Int("version", 0, "View specific version")
	taskViewCmd.SetHelpFunc(ui.HandleHelp)
	taskViewCmd.SetUsageFunc(ui.HandleUsage)

	taskEditCmd.Flags().String("title", "", "Task title")
	taskEditCmd.Flags().String("priority", "", "Priority (LOW, NORMAL, HIGH, CRITICAL)")
	taskEditCmd.Flags().String("type", "", "Task type (FEATURE, BUG, CHORE, R&D, MAINTENANCE, INFRA)")
	taskEditCmd.Flags().String("assigned-to", "", "Assigned user")
	taskEditCmd.Flags().String("description", "", "Task description")
	taskEditCmd.Flags().String("status", "", "Status (BACKLOG, ACTIVE, IN-PROGRESS, COMPLETED, CANCELLED)")
	taskEditCmd.Flags().String("slug", "", "New slug")
	taskEditCmd.Flags().String("tags", "", "Comma-separated tags")
	taskEditCmd.Flags().String("depends", "", "Comma-separated task dependencies")
	taskEditCmd.Flags().Bool("changelog", false, "Include in changelog")
	taskEditCmd.Flags().String("milestone", "", "Link to milestone (slug)")
	taskEditCmd.SetHelpFunc(ui.HandleHelp)
	taskEditCmd.SetUsageFunc(ui.HandleUsage)

	taskHistoryCmd.SetHelpFunc(ui.HandleHelp)
	taskHistoryCmd.SetUsageFunc(ui.HandleUsage)

	taskStatusCmd.SetHelpFunc(ui.HandleHelp)
	taskStatusCmd.SetUsageFunc(ui.HandleUsage)

	taskDeleteCmd.SetHelpFunc(ui.HandleHelp)
	taskDeleteCmd.SetUsageFunc(ui.HandleUsage)

	taskPlanCmd.SetHelpFunc(ui.HandleHelp)
	taskPlanCmd.SetUsageFunc(ui.HandleUsage)

	taskCmd.AddCommand(taskCreateCmd)
	taskCmd.AddCommand(taskViewCmd)
	taskCmd.AddCommand(taskListCmd)
	taskCmd.AddCommand(taskEditCmd)
	taskCmd.AddCommand(taskDeleteCmd)
	taskCmd.AddCommand(taskHistoryCmd)
	taskCmd.AddCommand(taskStatusCmd)
	taskCmd.AddCommand(taskPlanCmd)
	rootCmd.AddCommand(taskCmd)
}
