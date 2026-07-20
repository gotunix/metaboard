// SPDX-License-Identifier: GPL-3.0-or-later
// SPDX-FileCopyrightText: 2026 The MetaBoard authors

package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"gotunix.net/metaboard/internal/git"
	"gotunix.net/metaboard/internal/store"
	"gotunix.net/metaboard/internal/ui"
)

var milestoneCmd = &cobra.Command{
	Use:   "milestone",
	Short: "Manage project milestones",
}

var milestoneCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new milestone",
	Run: func(cmd *cobra.Command, args []string) {
		title, _ := cmd.Flags().GetString("title")
		description, _ := cmd.Flags().GetString("description")
		status, _ := cmd.Flags().GetString("status")
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

		finalSlug, err := store.CreateMilestone(title, slug, description)
		if err != nil {
			fmt.Println(lipgloss.NewStyle().Foreground(ui.Red).Render(fmt.Sprintf("Error: %v", err)))
			return
		}

		if status != "" && status != "BACKLOG" {
			_ = store.UpdateMilestoneStatus(finalSlug, status)
		}

		m, errM := store.GetMilestone(finalSlug)
		if errM == nil {
			path, _ := store.GetMilestonePath(m.ID)
			if git.IsGitRepo(root) {
				_ = git.Commit(root, []string{path}, fmt.Sprintf("boards: create milestone [%s] - %s", m.Slug, m.Title))
			}
		}
		fmt.Println(ui.BoldStyle.Foreground(ui.Green).Render(fmt.Sprintf("✔ Created Milestone [%s]", finalSlug)))
	},
}

var milestoneViewCmd = &cobra.Command{
	Use:   "view [slug]",
	Short: "Display milestone details",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var idOrSlug string
		if len(args) > 0 {
			idOrSlug = args[0]
		} else {
			fmt.Println(lipgloss.NewStyle().Foreground(ui.Red).Render("Error: milestone slug is required"))
			return
		}
		version, _ := cmd.Flags().GetInt("version")
		if err := ui.ViewMilestone(idOrSlug, version); err != nil {
			fmt.Println(lipgloss.NewStyle().Foreground(ui.Red).Render(fmt.Sprintf("Error: %v", err)))
		}
	},
}

var milestoneListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all milestones",
	Run: func(cmd *cobra.Command, args []string) {
		ms, err := store.ListMilestones()
		if err != nil {
			fmt.Println(lipgloss.NewStyle().Foreground(ui.Red).Render(fmt.Sprintf("Error: %v", err)))
			return
		}
		if len(ms) == 0 {
			fmt.Println("No milestones found.")
			return
		}
		for _, m := range ms {
			statusStyle := ui.StatusStyle(m.Status)
			fmt.Printf("[%s] %s %s\n", m.Slug, m.Title, statusStyle.Render(m.Status))
		}
	},
}

var milestoneEditCmd = &cobra.Command{
	Use:   "edit <slug>",
	Short: "Edit a milestone",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		slug := args[0]
		title, _ := cmd.Flags().GetString("title")
		description, _ := cmd.Flags().GetString("description")
		status, _ := cmd.Flags().GetString("status")
		newSlug, _ := cmd.Flags().GetString("slug")

		m, err := store.GetMilestone(slug)
		if err != nil {
			fmt.Println(lipgloss.NewStyle().Foreground(ui.Red).Render(fmt.Sprintf("Error: %v", err)))
			return
		}

		if title != "" {
			m.Title = title
		}
		if newSlug != "" {
			m.Slug = newSlug
		}
		if status != "" {
			m.Status = strings.ToUpper(status)
		}
		if description != "" {
			m.Description = description
		}

		if m.Status == "COMPLETED" && m.CompletedAt == "" {
			// set completed timestamp
		} else if m.Status != "COMPLETED" {
			m.CompletedAt = ""
		}

		if err := store.SaveMilestone(*m); err != nil {
			fmt.Println(lipgloss.NewStyle().Foreground(ui.Red).Render(fmt.Sprintf("Error: %v", err)))
			return
		}

		root, _ := store.GetDataRoot()
		if git.IsGitRepo(root) {
			path, _ := store.GetMilestonePath(m.ID)
			_ = git.Commit(root, []string{path}, fmt.Sprintf("boards: update milestone [%s] - %s", m.Slug, m.Title))
		}
		fmt.Println(ui.BoldStyle.Foreground(ui.Green).Render(fmt.Sprintf("✔ Updated Milestone [%s]", m.Slug)))
	},
}

var milestoneDeleteCmd = &cobra.Command{
	Use:   "delete <slug>",
	Short: "Delete a milestone",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		slug := args[0]
		root, _ := store.GetDataRoot()
		if git.IsGitRepo(root) {
			_ = git.Pull(root)
		}
		if err := store.DeleteMilestone(slug); err != nil {
			fmt.Println(lipgloss.NewStyle().Foreground(ui.Red).Render(fmt.Sprintf("Error: %v", err)))
			return
		}
		fmt.Println(ui.BoldStyle.Foreground(ui.Green).Render(fmt.Sprintf("✔ Deleted Milestone [%s]", slug)))
		if git.IsGitRepo(root) {
			_ = git.Commit(root, []string{"."}, fmt.Sprintf("boards: delete milestone [%s]", slug))
		}
	},
}

var milestoneHistoryCmd = &cobra.Command{
	Use:   "history <slug>",
	Short: "Show milestone version history",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := ui.ViewMilestoneHistory(args[0]); err != nil {
			fmt.Println(lipgloss.NewStyle().Foreground(ui.Red).Render(fmt.Sprintf("Error: %v", err)))
		}
	},
}

var milestoneStatusCmd = &cobra.Command{
	Use:   "status <slug> <status>",
	Short: "Update milestone status",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		slug := args[0]
		newStatus := strings.ToUpper(args[1])
		if err := store.UpdateMilestoneStatus(slug, newStatus); err != nil {
			fmt.Println(lipgloss.NewStyle().Foreground(ui.Red).Render(fmt.Sprintf("Error: %v", err)))
			return
		}
		fmt.Println(ui.BoldStyle.Foreground(ui.Green).Render(fmt.Sprintf("✔ Updated Milestone [%s] status to %s", slug, newStatus)))
	},
}

func init() {
	milestoneCreateCmd.Flags().String("title", "", "Milestone title")
	milestoneCreateCmd.Flags().String("description", "", "Milestone description")
	milestoneCreateCmd.Flags().String("status", "BACKLOG", "Milestone status (BACKLOG, ACTIVE, COMPLETED, CANCELLED)")
	milestoneCreateCmd.Flags().String("slug", "", "Custom slug (optional)")
	milestoneCreateCmd.SetHelpFunc(ui.HandleHelp)
	milestoneCreateCmd.SetUsageFunc(ui.HandleUsage)

	milestoneViewCmd.Flags().Int("version", 0, "View specific version")
	milestoneViewCmd.SetHelpFunc(ui.HandleHelp)
	milestoneViewCmd.SetUsageFunc(ui.HandleUsage)

	milestoneListCmd.SetHelpFunc(ui.HandleHelp)
	milestoneListCmd.SetUsageFunc(ui.HandleUsage)

	milestoneEditCmd.Flags().String("title", "", "New title")
	milestoneEditCmd.Flags().String("description", "", "New description")
	milestoneEditCmd.Flags().String("status", "", "New status (BACKLOG, ACTIVE, COMPLETED, CANCELLED)")
	milestoneEditCmd.Flags().String("slug", "", "New slug")
	milestoneEditCmd.SetHelpFunc(ui.HandleHelp)
	milestoneEditCmd.SetUsageFunc(ui.HandleUsage)

	milestoneDeleteCmd.SetHelpFunc(ui.HandleHelp)
	milestoneDeleteCmd.SetUsageFunc(ui.HandleUsage)

	milestoneHistoryCmd.SetHelpFunc(ui.HandleHelp)
	milestoneHistoryCmd.SetUsageFunc(ui.HandleUsage)

	milestoneStatusCmd.SetHelpFunc(ui.HandleHelp)
	milestoneStatusCmd.SetUsageFunc(ui.HandleUsage)

	milestoneCmd.AddCommand(milestoneCreateCmd, milestoneViewCmd, milestoneListCmd, milestoneEditCmd, milestoneDeleteCmd, milestoneHistoryCmd, milestoneStatusCmd)
	rootCmd.AddCommand(milestoneCmd)
}
