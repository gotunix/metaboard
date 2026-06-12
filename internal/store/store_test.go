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

package store

import (
	"os"
	"testing"
)

func setupTest(t *testing.T) func() {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory to temp: %v", err)
	}
	return func() {
		os.Chdir(oldDir)
	}
}

func TestMilestoneCRUD(t *testing.T) {
	cleanup := setupTest(t)
	defer cleanup()

	title := "Test Milestone"
	slug, err := CreateMilestone(title, "", "First line\\nSecond line")
	if err != nil {
		t.Fatalf("CreateMilestone failed: %v", err)
	}
	if slug != "m-1" {
		t.Errorf("expected slug m-1, got %s", slug)
	}

	m, err := GetMilestone(slug)
	if err != nil {
		t.Fatalf("GetMilestone failed: %v", err)
	}
	if m.Title != title {
		t.Errorf("expected title %q, got %q", title, m.Title)
	}
	if len(m.Description) != 2 || m.Description[1] != "Second line" {
		t.Errorf("description not parsed correctly: %v", m.Description)
	}

	err = UpdateMilestoneStatus(slug, "COMPLETED")
	if err != nil {
		t.Fatalf("UpdateMilestoneStatus failed: %v", err)
	}
	m, _ = GetMilestone(slug)
	if m.Status != "COMPLETED" || m.CompletedAt == "" {
		t.Errorf("status update failed: status=%s, completed_at=%s", m.Status, m.CompletedAt)
	}
}

func TestStoryCRUD(t *testing.T) {
	cleanup := setupTest(t)
	defer cleanup()

	slug, err := CreateStory("Test Story", "", "Story desc")
	if err != nil {
		t.Fatalf("CreateStory failed: %v", err)
	}
	if slug != "s-1" {
		t.Errorf("expected slug s-1, got %s", slug)
	}

	s, err := GetStory(slug)
	if err != nil {
		t.Fatalf("GetStory failed: %v", err)
	}
	if s.Title != "Test Story" {
		t.Errorf("expected title Test Story, got %q", s.Title)
	}
}

func TestTaskCRUD(t *testing.T) {
	cleanup := setupTest(t)
	defer cleanup()

	slug, err := CreateTask("Test Task", "", "HIGH", "FEAT", "user1", "Task desc")
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}
	if slug != "t-1" {
		t.Errorf("expected slug t-1, got %s", slug)
	}

	tObj, err := GetTask(slug)
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}
	if tObj.Priority != "HIGH" || tObj.AssignedTo != "user1" {
		t.Errorf("task fields not set correctly: %+v", tObj)
	}

	// Test UpdateTask
	newTitle := "Updated Task"
	err = UpdateTask(slug, TaskUpdate{Title: &newTitle})
	if err != nil {
		t.Fatalf("UpdateTask failed: %v", err)
	}
	tObj, _ = GetTask(slug)
	if tObj.Title != newTitle {
		t.Errorf("UpdateTask failed to update title")
	}
}

func TestLinking(t *testing.T) {
	cleanup := setupTest(t)
	defer cleanup()

	mSlug, _ := CreateMilestone("M1", "", "")
	sSlug, _ := CreateStory("S1", "", "")
	tSlug, _ := CreateTask("T1", "", "MED", "TASK", "", "")

	m, _ := GetMilestone(mSlug)
	s, _ := GetStory(sSlug)
	tObj, _ := GetTask(tSlug)

	// Link S1 to M1
	if err := LinkEntities(s.ID, m.ID); err != nil {
		t.Fatalf("Link S1->M1 failed: %v", err)
	}
	// Link T1 to S1
	if err := LinkEntities(tObj.ID, s.ID); err != nil {
		t.Fatalf("Link T1->S1 failed: %v", err)
	}

	m, _ = GetMilestone(m.ID)
	if len(m.Stories) != 1 || m.Stories[0] != s.ID {
		t.Errorf("Story not linked to Milestone")
	}

	s, _ = GetStory(s.ID)
	if len(s.Tasks) != 1 || s.Tasks[0] != tObj.ID {
		t.Errorf("Task not linked to Story")
	}

	// Unlink T1
	if err := UnlinkEntity(tObj.ID); err != nil {
		t.Fatalf("Unlink T1 failed: %v", err)
	}
	s, _ = GetStory(s.ID)
	if len(s.Tasks) != 0 {
		t.Errorf("Task not unlinked from Story")
	}
}

func TestStoryUnlinking(t *testing.T) {
	cleanup := setupTest(t)
	defer cleanup()

	mSlug, _ := CreateMilestone("M1", "", "")
	sSlug, _ := CreateStory("S1", "", "")
	
	m, _ := GetMilestone(mSlug)
	s, _ := GetStory(sSlug)

	LinkEntities(s.ID, m.ID)
	
	// Unlink S1
	if err := UnlinkEntity(s.ID); err != nil {
		t.Fatalf("Unlink S1 failed: %v", err)
	}
	
	m, _ = GetMilestone(m.ID)
	if len(m.Stories) != 0 {
		t.Errorf("Story not unlinked from Milestone")
	}
}

func TestEnsureTaskPlan(t *testing.T) {
	cleanup := setupTest(t)
	defer cleanup()

	tSlug, _ := CreateTask("T1", "", "MED", "TASK", "", "")

	planPath, err := EnsureTaskPlan(tSlug)
	if err != nil {
		t.Fatalf("EnsureTaskPlan failed: %v", err)
	}
	if _, err := os.Stat(planPath); os.IsNotExist(err) {
		t.Errorf("Plan file was not created")
	}

	// Try to create plan for story (should fail)
	sSlug, _ := CreateStory("S1", "", "")
	_, err = EnsureTaskPlan(sSlug)
	if err == nil {
		t.Errorf("Expected error when creating plan for Story, got nil")
	}
}

func TestSortingAndListing(t *testing.T) {
	cleanup := setupTest(t)
	defer cleanup()

	CreateMilestone("M2", "m-2", "")
	CreateMilestone("M1", "m-1", "")
	CreateMilestone("M10", "m-10", "")

	ms, _ := ListMilestones()
	if len(ms) != 3 {
		t.Errorf("Expected 3 milestones, got %d", len(ms))
	}

	SortMilestones(ms)
	if ms[0].Slug != "m-1" || ms[1].Slug != "m-2" || ms[2].Slug != "m-10" {
		t.Errorf("Milestones not sorted correctly: %v, %v, %v", ms[0].Slug, ms[1].Slug, ms[2].Slug)
	}
	
	items := []string{"t-10", "t-2", "t-1"}
	NaturalSort(items)
	if items[0] != "t-1" || items[1] != "t-2" || items[2] != "t-10" {
		t.Errorf("NaturalSort failed: %v", items)
	}
}

func TestErrorCases(t *testing.T) {
	cleanup := setupTest(t)
	defer cleanup()

	_, _ = CreateTask("T1", "", "MED", "TASK", "", "")
	tObj, _ := GetTask("t-1")
	
	// Link T1 to T1 (should fail)
	err := LinkEntities(tObj.ID, tObj.ID)
	if err == nil {
		t.Errorf("Expected error when linking task to task, got nil")
	}
	
	// Resolve non-existent entity
	_, err = ResolveEntity("non-existent")
	if err == nil {
		t.Errorf("Expected error when resolving non-existent entity, got nil")
	}
}

func TestGenerateNextSlug(t *testing.T) {
	cleanup := setupTest(t)
	defer cleanup()

	CreateTask("T1", "t-1", "MED", "TASK", "", "")
	CreateTask("T2", "t-2", "MED", "TASK", "", "")
	CreateTask("T10", "t-10", "MED", "TASK", "", "")

	next := GenerateNextTaskSlug("t")
	if next != "t-11" {
		t.Errorf("Expected next slug t-11, got %s", next)
	}
}

func TestNaturalSortMore(t *testing.T) {
	items := []string{"t-10", "t-2", "t-1"}
	NaturalSort(items)
	if items[0] != "t-1" || items[1] != "t-2" || items[2] != "t-10" {
		t.Errorf("NaturalSort failed: %v", items)
	}
}

func TestCompareNatural(t *testing.T) {
	tests := []struct {
		s1, s2   string
		expected int
	}{
		{"a-1", "a-2", -1},
		{"a-10", "a-2", 1},
		{"a-2", "a-10", -1},
		{"a-1-b", "a-1-a", 1},
		{"task-1", "task-1", 0},
	}

	for _, tt := range tests {
		got := CompareNatural(tt.s1, tt.s2)
		if got != tt.expected {
			t.Errorf("CompareNatural(%q, %q) = %d; want %d", tt.s1, tt.s2, got, tt.expected)
		}
	}
}
