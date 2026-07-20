// SPDX-License-Identifier: GPL-3.0-or-later
// SPDX-FileCopyrightText: 2026 The MetaBoard authors

package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"gotunix.net/metaboard/internal/store"
	"gotunix.net/metaboard/internal/ui"
)

var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Initialize a new metaboard project",
	Long:  "Create the required directory structure (milestones, tasks, pullrequests) in the specified path (defaults to ./metadata).",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := "metadata"
		if len(args) > 0 {
			path = args[0]
		}

		if err := store.Initialize(path); err != nil {
			fmt.Println(lipgloss.NewStyle().Foreground(ui.Red).Render(fmt.Sprintf("Error: %v", err)))
			os.Exit(1)
		}

		fmt.Println(ui.BoldStyle.Foreground(ui.Green).Render(fmt.Sprintf("✔ Initialized metaboard in %q", path)))
	},
}

func init() {
	initCmd.SetHelpFunc(ui.HandleHelp)
	initCmd.SetUsageFunc(ui.HandleUsage)
	rootCmd.AddCommand(initCmd)
}
