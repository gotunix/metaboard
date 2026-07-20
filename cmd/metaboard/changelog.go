// SPDX-License-Identifier: GPL-3.0-or-later
// SPDX-FileCopyrightText: 2026 The MetaBoard authors

package main

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"gotunix.net/metaboard/internal/ui"
)

var changelogCmd = &cobra.Command{
	Use:   "changelog [output_dir]",
	Short: "Generate CHANGELOG.md",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		out := "."
		if len(args) > 0 {
			out = args[0]
		}
		if err := ui.GenerateChangelog(out); err != nil {
			fmt.Println(lipgloss.NewStyle().Foreground(ui.Red).Render(fmt.Sprintf("Error: %v", err)))
		} else {
			fmt.Println(ui.BoldStyle.Foreground(ui.Green).Render("✔ Changelog generated successfully"))
		}
	},
}

func init() {
	changelogCmd.SetHelpFunc(ui.HandleHelp)
	changelogCmd.SetUsageFunc(ui.HandleUsage)
	rootCmd.AddCommand(changelogCmd)
}
