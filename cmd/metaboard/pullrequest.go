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

var prCmd = &cobra.Command{
	Use:     "pullrequest",
	Aliases: []string{"pr"},
	Short:   "Manage project pull requests",
}

var prCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new pull request record and open its Markdown template",
	Run: func(cmd *cobra.Command, args []string) {
		destBranch, _ := cmd.Flags().GetString("dest-branch")
		if destBranch == "main" {
			if baseFlag, _ := cmd.Flags().GetString("base"); baseFlag != "main" {
				destBranch = baseFlag
			}
		}

		sourceBranch, _ := cmd.Flags().GetString("source-branch")
		if sourceBranch == "" {
			sourceBranch, _ = cmd.Flags().GetString("head")
		}

		sourceRepo, _ := cmd.Flags().GetString("source-repo")
		destRepo, _ := cmd.Flags().GetString("dest-repo")
		description, _ := cmd.Flags().GetString("description")

		if cmd.Flags().NFlag() == 0 {
			fmt.Println(lipgloss.NewStyle().Foreground(ui.Red).Render("Error: flags required (no interactive mode)"))
			return
		}

		root, _ := store.GetDataRoot()
		if git.IsGitRepo(root) {
			_ = git.Pull(root)
		}

		finalSlug, err := store.CreatePullRequest("", destBranch, sourceBranch, sourceRepo, destRepo, description)
		if err != nil {
			fmt.Println(lipgloss.NewStyle().Foreground(ui.Red).Render(fmt.Sprintf("Error: %v", err)))
			return
		}
		fmt.Println(ui.BoldStyle.Foreground(ui.Green).Render(fmt.Sprintf("✔ Created Pull Request [%s]", finalSlug)))

		// Open Markdown template in editor automatically
		pr, err := store.GetPullRequest(finalSlug)
		if err == nil {
			mdPath, err := store.GetPullRequestMarkdownPath(pr.ID)
			if err == nil {
				editor := os.Getenv("EDITOR")
				if editor == "" {
					editor = "vim"
				}
				c := exec.Command(editor, mdPath)
				c.Stdin = os.Stdin
				c.Stdout = os.Stdout
				c.Stderr = os.Stderr
				_ = c.Run()

				if err := store.ParsePullRequestMarkdown(pr.ID); err != nil {
					fmt.Println(lipgloss.NewStyle().Foreground(ui.Red).Render(fmt.Sprintf("Warning: failed to parse PR markdown: %v", err)))
				}
			}
		}

		if git.IsGitRepo(root) {
			p, _ := store.GetPullRequest(finalSlug)
			path, _ := store.GetPullRequestPath(p.ID)
			mdPath, _ := store.GetPullRequestMarkdownPath(p.ID)
			_ = git.Commit(root, []string{path, mdPath}, fmt.Sprintf("boards: create pullrequest [%s] - %s → %s", p.Slug, p.HeadBranch, p.BaseBranch))
		}
	},
}

var prViewCmd = &cobra.Command{
	Use:   "view [slug]",
	Short: "Display pull request details",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var idOrSlug string
		if len(args) > 0 {
			idOrSlug = args[0]
		} else {
			fmt.Println(lipgloss.NewStyle().Foreground(ui.Red).Render("Error: PR slug is required"))
			return
		}
		version, _ := cmd.Flags().GetInt("version")
		if err := ui.ViewPullRequest(idOrSlug, version); err != nil {
			fmt.Println(lipgloss.NewStyle().Foreground(ui.Red).Render(fmt.Sprintf("Error: %v", err)))
		}
	},
}

var prListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all pull requests",
	Run: func(cmd *cobra.Command, args []string) {
		prs, err := store.ListPullRequests()
		if err != nil {
			fmt.Println(lipgloss.NewStyle().Foreground(ui.Red).Render(fmt.Sprintf("Error: %v", err)))
			return
		}
		if len(prs) == 0 {
			fmt.Println("No pull requests found.")
			return
		}
		for _, pr := range prs {
			statusStyle := ui.StatusStyle(pr.Status)
			fmt.Printf("[%s] %s → %s %s\n", pr.Slug, pr.HeadBranch, pr.BaseBranch, statusStyle.Render(pr.Status))
		}
	},
}

var prEditCmd = &cobra.Command{
	Use:   "edit <slug>",
	Short: "Edit a pull request",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		slug := args[0]
		headBranch, _ := cmd.Flags().GetString("head-branch")
		baseBranch, _ := cmd.Flags().GetString("base-branch")
		sourceRepo, _ := cmd.Flags().GetString("source-repo")
		destRepo, _ := cmd.Flags().GetString("dest-repo")
		status, _ := cmd.Flags().GetString("status")
		description, _ := cmd.Flags().GetString("description")

		pr, err := store.GetPullRequest(slug)
		if err != nil {
			fmt.Println(lipgloss.NewStyle().Foreground(ui.Red).Render(fmt.Sprintf("Error: %v", err)))
			return
		}

		if headBranch != "" {
			pr.HeadBranch = headBranch
		}
		if baseBranch != "" {
			pr.BaseBranch = baseBranch
		}
		if sourceRepo != "" {
			pr.SourceRepo = sourceRepo
		}
		if destRepo != "" {
			pr.DestRepo = destRepo
		}
		if status != "" {
			pr.Status = strings.ToUpper(status)
		}
		if description != "" {
			pr.Description = description
		}

		if pr.Status == "MERGED" && pr.CompletedAt == "" {
			// completed at would be set
		} else if pr.Status != "MERGED" {
			pr.CompletedAt = ""
		}

		if err := store.SavePullRequest(*pr); err != nil {
			fmt.Println(lipgloss.NewStyle().Foreground(ui.Red).Render(fmt.Sprintf("Error: %v", err)))
			return
		}

		root, _ := store.GetDataRoot()
		if git.IsGitRepo(root) {
			path, _ := store.GetPullRequestPath(pr.ID)
			mdPath, _ := store.GetPullRequestMarkdownPath(pr.ID)
			_ = git.Commit(root, []string{path, mdPath}, fmt.Sprintf("boards: update pullrequest [%s] - %s → %s", pr.Slug, pr.HeadBranch, pr.BaseBranch))
		}
		fmt.Println(ui.BoldStyle.Foreground(ui.Green).Render(fmt.Sprintf("✔ Updated Pull Request [%s]", pr.Slug)))
	},
}

var prDeleteCmd = &cobra.Command{
	Use:   "delete <slug>",
	Short: "Delete a pull request",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		slug := args[0]
		root, _ := store.GetDataRoot()
		if git.IsGitRepo(root) {
			_ = git.Pull(root)
		}
		if err := store.DeletePR(slug); err != nil {
			fmt.Println(lipgloss.NewStyle().Foreground(ui.Red).Render(fmt.Sprintf("Error: %v", err)))
			return
		}
		fmt.Println(ui.BoldStyle.Foreground(ui.Green).Render(fmt.Sprintf("✔ Deleted Pull Request [%s]", slug)))
		if git.IsGitRepo(root) {
			_ = git.Commit(root, []string{"."}, fmt.Sprintf("boards: delete pullrequest [%s]", slug))
		}
	},
}

var prHistoryCmd = &cobra.Command{
	Use:   "history <slug>",
	Short: "Show pull request version history",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := ui.ViewPullRequestHistory(args[0]); err != nil {
			fmt.Println(lipgloss.NewStyle().Foreground(ui.Red).Render(fmt.Sprintf("Error: %v", err)))
		}
	},
}

func init() {
	prCreateCmd.Flags().String("dest-branch", "main", "Destination/base branch")
	prCreateCmd.Flags().String("source-branch", "", "Source/head branch")
	prCreateCmd.Flags().String("source-repo", "", "Source repository URL")
	prCreateCmd.Flags().String("dest-repo", "", "Destination repository URL")
	prCreateCmd.Flags().String("description", "", "PR description")
	prCreateCmd.SetHelpFunc(ui.HandleHelp)
	prCreateCmd.SetUsageFunc(ui.HandleUsage)

	prViewCmd.Flags().Int("version", 0, "View specific version")
	prViewCmd.SetHelpFunc(ui.HandleHelp)
	prViewCmd.SetUsageFunc(ui.HandleUsage)

	prListCmd.SetHelpFunc(ui.HandleHelp)
	prListCmd.SetUsageFunc(ui.HandleUsage)

	prEditCmd.Flags().String("head-branch", "", "Source branch")
	prEditCmd.Flags().String("base-branch", "", "Destination branch")
	prEditCmd.Flags().String("source-repo", "", "Source repository URL")
	prEditCmd.Flags().String("dest-repo", "", "Destination repository URL")
	prEditCmd.Flags().String("status", "", "Status (DRAFT, OPEN, MERGED, CLOSED, REJECTED)")
	prEditCmd.Flags().String("description", "", "PR description")
	prEditCmd.SetHelpFunc(ui.HandleHelp)
	prEditCmd.SetUsageFunc(ui.HandleUsage)

	prDeleteCmd.SetHelpFunc(ui.HandleHelp)
	prDeleteCmd.SetUsageFunc(ui.HandleUsage)

	prHistoryCmd.SetHelpFunc(ui.HandleHelp)
	prHistoryCmd.SetUsageFunc(ui.HandleUsage)

	prCmd.AddCommand(prCreateCmd, prViewCmd, prListCmd, prEditCmd, prDeleteCmd, prHistoryCmd)
	rootCmd.AddCommand(prCmd)
}
