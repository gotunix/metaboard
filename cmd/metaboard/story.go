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
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"gotunix.net/metaboard/internal/interactive"
	"gotunix.net/metaboard/internal/models"
	"gotunix.net/metaboard/internal/store"
	"gotunix.net/metaboard/internal/ui"
)

var storyCmd = &cobra.Command{
	Use:   "story",
	Short: "Manage project stories",
}

var storyCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new story",
	Run: func(cmd *cobra.Command, args []string) {
		title, _ := cmd.Flags().GetString("title")
		slug, _ := cmd.Flags().GetString("slug")
		description, _ := cmd.Flags().GetString("description")
		if title == "" {
			fmt.Println(lipgloss.NewStyle().Foreground(ui.Red).Render("Error: --title is required"))
			return
		}
		finalSlug, err := store.CreateStory(title, slug, description)
		if err != nil {
			fmt.Println(lipgloss.NewStyle().Foreground(ui.Red).Render(fmt.Sprintf("Error: %v", err)))
		} else {
			fmt.Println(ui.BoldStyle.Foreground(ui.Green).Render(fmt.Sprintf("✔ Created Story [%s]", finalSlug)))
		}
	},
}

var storyViewCmd = &cobra.Command{
	Use:   "view <slug>",
	Short: "Display story details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := ui.ViewStory(args[0]); err != nil {
			fmt.Println(lipgloss.NewStyle().Foreground(ui.Red).Render(fmt.Sprintf("Error: %v", err)))
		}
	},
}

var storyStatusCmd = &cobra.Command{
	Use:   "status <slug> <status>",
	Short: "Update story status",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		status := strings.ToUpper(args[1])
		if !models.IsValidStatus(status, models.ValidStoryStatuses) {
			fmt.Println(lipgloss.NewStyle().Foreground(ui.Red).Render(fmt.Sprintf("Error: invalid status %q. Allowed: %v", status, models.ValidStoryStatuses)))
			return
		}
		if err := store.UpdateStoryStatus(args[0], status); err != nil {
			fmt.Println(lipgloss.NewStyle().Foreground(ui.Red).Render(fmt.Sprintf("Error: %v", err)))
			return
		}
		fmt.Println(ui.BoldStyle.Foreground(ui.Green).Render(fmt.Sprintf("✔ Updated Story [%s] status to %s", args[0], status)))
	},
}

var storyEditCmd = &cobra.Command{
	Use:   "edit <slug>",
	Short: "Edit a story",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := interactive.EditStoryInteractive(args[0]); err != nil {
			fmt.Println(lipgloss.NewStyle().Foreground(ui.Red).Render(fmt.Sprintf("Error: %v", err)))
			return
		}
		fmt.Println(ui.BoldStyle.Foreground(ui.Green).Render("✔ Story updated"))
	},
}

func init() {
	storyCmd.SetHelpFunc(ui.HandleHelp)
	storyCmd.SetUsageFunc(ui.HandleUsage)
	storyCreateCmd.Flags().StringP("title", "t", "", "Title of the story")
	storyCreateCmd.Flags().StringP("slug", "s", "", "Unique slug for the story")
	storyCreateCmd.Flags().String("description", "", "Description of the story")

	storyCmd.AddCommand(storyCreateCmd)
	storyCmd.AddCommand(storyViewCmd)
	storyCmd.AddCommand(storyStatusCmd)
	storyCmd.AddCommand(storyEditCmd)
	rootCmd.AddCommand(storyCmd)
}
