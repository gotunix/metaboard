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

package models

import "strings"

const (
	StatusBacklog    = "BACKLOG"
	StatusActive     = "ACTIVE"
	StatusInProgress = "IN-PROGRESS"
	StatusCompleted  = "COMPLETED"
	StatusCancelled  = "CANCELLED"
)

var (
	ValidMilestoneStatuses = []string{StatusBacklog, StatusActive, StatusCompleted, StatusCancelled}
	ValidStoryStatuses     = []string{StatusBacklog, StatusActive, StatusCompleted, StatusCancelled}
	ValidTaskStatuses      = []string{StatusBacklog, StatusActive, StatusInProgress, StatusCompleted, StatusCancelled}
)

func IsValidStatus(status string, allowed []string) bool {
	s := strings.ToUpper(strings.ReplaceAll(status, "_", "-"))
	for _, a := range allowed {
		if s == a {
			return true
		}
	}
	return false
}

type Milestone struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Slug        string   `json:"slug"`
	Status      string   `json:"status"`
	Description []string `json:"description"`
	Stories     []string `json:"stories"`
	Tasks       []string `json:"tasks"`
	CompletedAt string   `json:"completed_at"`
}

type Story struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Slug        string   `json:"slug"`
	Status      string   `json:"status"`
	Description []string `json:"description"`
	Tasks       []string `json:"tasks"`
	CompletedAt string   `json:"completed_at"`
}

type Task struct {
	ID          string   `json:"id"`
	Slug        string   `json:"slug"`
	Title       string   `json:"title"`
	Status      string   `json:"status"`
	Priority    string   `json:"priority"`
	Type        string   `json:"type"`
	Tags        []string `json:"tags"`
	DependsOn   []string `json:"depends_on"`
	AssignedTo  string   `json:"assigned_to"`
	Description []string `json:"description"`
	CreatedAt   string   `json:"created_at"`
	CompletedAt string   `json:"completed_at"`
}
