// SPDX-License-Identifier: GPL-3.0-or-later
// SPDX-FileCopyrightText: 2026 The MetaBoard authors

package main

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"gotunix.net/metaboard/internal/git"
	"gotunix.net/metaboard/internal/store"
	"gotunix.net/metaboard/internal/ui"
)

var linkCmd = &cobra.Command{
	Use:   "link [child_slug] [parent_slug]",
	Short: "Link two entities",
	Long:  `Link a Task to a Milestone, or a Pull Request to a Task.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			fmt.Println(lipgloss.NewStyle().Foreground(ui.Red).Render("Error: child and parent slugs required"))
			return
		}
		child := args[0]
		parent := args[1]

		root, _ := store.GetDataRoot()
		if git.IsGitRepo(root) {
			_ = git.Pull(root)
		}

		if err := store.LinkEntities(child, parent); err != nil {
			fmt.Println(lipgloss.NewStyle().Foreground(ui.Red).Render(fmt.Sprintf("Error: %v", err)))
			return
		}
		fmt.Println(ui.BoldStyle.Foreground(ui.Green).Render("✔ Linked successfully"))

		if git.IsGitRepo(root) {
			cRes, errC := store.ResolveEntity(child)
			pRes, errP := store.ResolveEntity(parent)
			var paths []string
			if errC == nil {
				paths = append(paths, cRes.Path)
			}
			if errP == nil {
				paths = append(paths, pRes.Path)
			}
			if len(paths) > 0 {
				_ = git.Commit(root, paths, fmt.Sprintf("metaboard: link %s to %s", child, parent))
			}
		}
	},
}

var unlinkCmd = &cobra.Command{
	Use:   "unlink <child_slug>",
	Short: "Unlink an entity from its parent",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		child := args[0]

		root, _ := store.GetDataRoot()
		if git.IsGitRepo(root) {
			_ = git.Pull(root)
		}

		cResBefore, errCBefore := store.ResolveEntity(child)
		var parentPaths []string
		if git.IsGitRepo(root) && errCBefore == nil {
			switch cResBefore.Type {
			case store.TypeTask:
				ms, _ := store.ListMilestones()
				for _, m := range ms {
					for _, id := range m.Tasks {
						if id == cResBefore.ID {
							pPath, _ := store.GetMilestonePath(m.ID)
							parentPaths = append(parentPaths, pPath)
						}
					}
				}
			case store.TypePullRequest:
				ts, _ := store.ListTasks()
				for _, t := range ts {
					for _, id := range t.PullRequests {
						if id == cResBefore.ID {
							pPath, _ := store.GetTaskPath(t.ID)
							parentPaths = append(parentPaths, pPath)
						}
					}
				}
			}
		}

		if err := store.UnlinkEntity(child); err != nil {
			fmt.Println(lipgloss.NewStyle().Foreground(ui.Red).Render(fmt.Sprintf("Error: %v", err)))
			return
		}
		fmt.Println(ui.BoldStyle.Foreground(ui.Green).Render("✔ Unlinked successfully"))

		if git.IsGitRepo(root) {
			var paths []string
			if errCBefore == nil {
				paths = append(paths, cResBefore.Path)
			}
			for _, p := range parentPaths {
				paths = append(paths, p)
			}
			if len(paths) > 0 {
				_ = git.Commit(root, paths, fmt.Sprintf("metaboard: unlink %s", child))
			}
		}
	},
}

func init() {
	linkCmd.SetHelpFunc(ui.HandleHelp)
	linkCmd.SetUsageFunc(ui.HandleUsage)
	unlinkCmd.SetHelpFunc(ui.HandleHelp)
	unlinkCmd.SetUsageFunc(ui.HandleUsage)
	rootCmd.AddCommand(linkCmd)
	rootCmd.AddCommand(unlinkCmd)
}
