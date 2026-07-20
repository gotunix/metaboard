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
// it under the terms of the GNU General Public License as                                  //
// published by the Free Software Foundation, either version 3 of the                              //
// License, or (at your option) any later version.                                                 //
//                                                                                                 //
// This program is distributed in the hope that it will be useful,                                 //
// but WITHOUT ANY WARRANTY; without even the implied warranty of                                  //
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the                                   //
// GNU General Public License for more details.                                             //
//                                                                                                 //
// You should have received a copy of the GNU General Public License                        //
// along with this program.  If not, see <https://www.gnu.org/licenses/>.                          //
// =============================================================================================== //

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gotunix.net/metaboard/internal/store"
	"gotunix.net/metaboard/internal/ui"
)

var dataDir string
var versionFlag bool

var rootCmd = &cobra.Command{
	Use:   "metaboard",
	Short: "A git-first project metaboard engine",
	Long: `A high-performance CLI engine for git-first project management.
Version your tasks and milestones directly alongside your code.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if dataDir != "" {
			store.SetDataDir(dataDir)
		}
		if versionFlag {
			if err := ui.RenderVersion(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			os.Exit(0)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		if err := ui.RenderDashboard("", "", "all"); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.SetHelpFunc(ui.HandleHelp)
	rootCmd.SetUsageFunc(ui.HandleUsage)

	rootCmd.PersistentFlags().StringVarP(&dataDir, "data-dir", "d", "", "base directory for metaboard data")
	rootCmd.PersistentFlags().BoolVarP(&versionFlag, "version", "v", false, "show version information")
}
