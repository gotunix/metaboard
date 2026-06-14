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

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"gotunix.net/metaboard/internal/ui"
)

var dashboardCmd = &cobra.Command{
	Use:   "dashboard [milestone_slug]",
	Short: "Show project overview",
	Long:  `Show a hierarchical tree view of milestones, stories, and tasks.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		mode := ""
		target := ""
		if len(args) > 0 {
			mode = "milestone"
			target = args[0]
		}
		if err := ui.RenderDashboard(mode, target, "active"); err != nil {
			fmt.Println(lipgloss.NewStyle().Foreground(ui.Red).Render(fmt.Sprintf("Error: %v", err)))
		}
	},
}

var dashboardAllCmd = &cobra.Command{
	Use:     "all",
	Aliases: []string{"everything"},
	Short:   "Show all items regardless of status",
	Run: func(cmd *cobra.Command, args []string) {
		if err := ui.RenderDashboard("all", "", "all"); err != nil {
			fmt.Println(lipgloss.NewStyle().Foreground(ui.Red).Render(fmt.Sprintf("Error: %v", err)))
		}
	},
}

var dashboardClosedCmd = &cobra.Command{
	Use:     "closed",
	Aliases: []string{"completed", "done"},
	Short:   "Show closed/completed items",
	Run: func(cmd *cobra.Command, args []string) {
		if err := ui.RenderDashboard("all", "", "closed"); err != nil {
			fmt.Println(lipgloss.NewStyle().Foreground(ui.Red).Render(fmt.Sprintf("Error: %v", err)))
		}
	},
}

var dashboardCancelledCmd = &cobra.Command{
	Use:     "cancelled",
	Aliases: []string{"aborted", "dropped"},
	Short:   "Show cancelled items",
	Run: func(cmd *cobra.Command, args []string) {
		if err := ui.RenderDashboard("all", "", "cancelled"); err != nil {
			fmt.Println(lipgloss.NewStyle().Foreground(ui.Red).Render(fmt.Sprintf("Error: %v", err)))
		}
	},
}

func init() {
	dashboardCmd.SetHelpFunc(ui.HandleHelp)
	dashboardCmd.SetUsageFunc(ui.HandleUsage)
	dashboardCmd.AddCommand(dashboardAllCmd)
	dashboardCmd.AddCommand(dashboardClosedCmd)
	dashboardCmd.AddCommand(dashboardCancelledCmd)
	rootCmd.AddCommand(dashboardCmd)
}
