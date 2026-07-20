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

package ui

import (
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

func GetTerminalWidth() int {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w <= 0 {
		return 100 // Fallback
	}
	targetWidth := w
	if targetWidth < 80 {
		return 80
	}
	return targetWidth
}

var (
	// Catppuccin Mocha Palette
	catMochaBase    = lipgloss.Color("#1e1e2e")
	catMochaText    = lipgloss.Color("#cdd6f4")
	catMochaSubtext = lipgloss.Color("#a6adc8")
	catMochaSurface0 = lipgloss.Color("#363a4f")
	catMochaOverlay = lipgloss.Color("#6c7086")
	catMochaBlue    = lipgloss.Color("#89b4fa")
	catMochaGreen   = lipgloss.Color("#a6e3a1")
	catMochaRed     = lipgloss.Color("#f38ba8")
	catMochaMauve   = lipgloss.Color("#cba6f7")
	catMochaPeach   = lipgloss.Color("#fab387")

	TableBg = lipgloss.AdaptiveColor{Light: "#ebdcf7", Dark: "#cba6f7"} // Light mode very light lilac / Dark mode Catppuccin Mauve
	ValFg   = lipgloss.AdaptiveColor{Light: "#1e1e2e", Dark: "#1e1e2e"} // Dark text in both modes for maximum contrast
	LabelFg = lipgloss.AdaptiveColor{Light: "#45475a", Dark: "#313244"} // Dark slate in both modes for maximum contrast

	Subtle  = lipgloss.AdaptiveColor{Light: "#cba6f7", Dark: "#cba6f7"} // Mauve
	Magenta = lipgloss.AdaptiveColor{Light: "#cba6f7", Dark: "#cba6f7"} // Mauve
	Cyan    = lipgloss.AdaptiveColor{Light: "#89b4fa", Dark: "#89b4fa"} // Blue
	Green   = lipgloss.AdaptiveColor{Light: "#a6e3a1", Dark: "#a6e3a1"} // Green
	Yellow  = lipgloss.AdaptiveColor{Light: "#fab387", Dark: "#fab387"} // Peach
	Red     = lipgloss.AdaptiveColor{Light: "#f38ba8", Dark: "#f38ba8"} // Red
	Gray    = lipgloss.AdaptiveColor{Light: "#6c7086", Dark: "#6c7086"} // Overlay

	BoldStyle  = lipgloss.NewStyle().Bold(true)
	TitleStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Italic(false).
			Foreground(lipgloss.Color("#1e1e2e")).
			Background(lipgloss.Color("#cba6f7")).
			Bold(true)

	HeaderStyle = lipgloss.NewStyle().
			Foreground(Magenta).
			Bold(true).
			Underline(true)

	LabelStyle = lipgloss.NewStyle().
			Foreground(Cyan).
			Bold(true)

	StatusStyle = func(status string) lipgloss.Style {
		s := lipgloss.NewStyle().Bold(true)
		switch strings.ToUpper(status) {
		case "ACTIVE":
			return s.Foreground(Green)
		case "BACKLOG":
			return s.Foreground(Cyan)
		case "COMPLETED":
			return s.Foreground(Magenta)
		case "IN-PROGRESS":
			return s.Foreground(Yellow)
		case "CANCELLED":
			return s.Foreground(Gray)
		case "DRAFT":
			return s.Foreground(Gray)
		case "OPEN":
			return s.Foreground(Green)
		case "MERGED":
			return s.Foreground(Magenta)
		case "CLOSED":
			return s.Foreground(Red)
		case "REJECTED":
			return s.Foreground(Red)
		default:
			return s.Foreground(Gray)
		}
	}

	BorderStyle = lipgloss.NewStyle().
			Foreground(Subtle).
			Border(lipgloss.NormalBorder(), true, false, true, false)

	milestoneRowStyle = lipgloss.NewStyle().
				Foreground(catMochaMauve).
				Bold(true)

	storyRowStyle = lipgloss.NewStyle().
			Foreground(catMochaBlue)

	taskRowStyle = lipgloss.NewStyle().
			Foreground(catMochaText)

	statusDoneStyle = lipgloss.NewStyle().
			Foreground(catMochaBase).
			Background(catMochaGreen).
			Padding(0, 1).
			Bold(true)

	statusMergedStyle = lipgloss.NewStyle().
				Foreground(catMochaBase).
				Background(catMochaMauve).
				Padding(0, 1).
				Bold(true)

	statusProgressStyle = lipgloss.NewStyle().
				Foreground(catMochaBase).
				Background(catMochaPeach).
				Padding(0, 1).
				Bold(true)

	statusTodoStyle = lipgloss.NewStyle().
			Foreground(catMochaSubtext).
			Background(catMochaOverlay).
			Padding(0, 1)

	UsageStyle       = lipgloss.NewStyle().Padding(1, 2)
	CommandStyle     = lipgloss.NewStyle().Foreground(Magenta).Bold(true).Width(10)
	DescriptionStyle = lipgloss.NewStyle().Foreground(Gray)
	LogoStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#cba6f7")).Bold(true)
	HelpTitleStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#1e1e2e")).Background(lipgloss.Color("#cba6f7")).Padding(0, 1).Bold(true).MarginBottom(1)
	HelpDescStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#a6adc8")).Italic(true).MarginBottom(1)
	HelpSectionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#cba6f7")).Bold(true).MarginTop(1)
	HelpFlagStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#a6adc8"))

	focusedLabelStyle = lipgloss.NewStyle().
				Foreground(catMochaGreen).
				Bold(true)
	formLabelStyle = lipgloss.NewStyle().
			Foreground(catMochaText).
			Bold(true)
	helperStyle = lipgloss.NewStyle().
			Foreground(catMochaOverlay).
			Italic(true)
	buttonStyle = lipgloss.NewStyle().
			Foreground(catMochaBase).
			Background(catMochaGreen).
			Padding(0, 3)
	activeButtonStyle = lipgloss.NewStyle().
				Foreground(catMochaBase).
				Background(catMochaGreen).
				Padding(0, 3).
				Bold(true)
)
