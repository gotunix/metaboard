// SPDX-License-Identifier: GPL-3.0-or-later
// SPDX-FileCopyrightText: 2026 The MetaBoard authors

package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func setupTest(t *testing.T) func() {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}

	// Reset store state
	SetDataDir("")
	defaultStore.resolvedDir = ""

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory to temp: %v", err)
	}

	// Initialize the store in the temp dir
	if err := Initialize("."); err != nil {
		t.Fatalf("failed to initialize store: %v", err)
	}

	return func() {
		_ = os.Chdir(oldDir)
		// Reset store state after test too
		SetDataDir("")
		defaultStore.resolvedDir = ""
	}
}

func TestGetDataRootRequiresCompleteLayout(t *testing.T) {
	cleanup := setupTest(t)
	defer cleanup()

	s := NewStore("")
	defaultStore.resolvedDir = ""

	partialDir := t.TempDir()
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	if err := os.Chdir(partialDir); err != nil {
		t.Fatalf("failed to change directory to partial temp: %v", err)
	}

	if err := os.MkdirAll("tasks", 0755); err != nil {
		t.Fatalf("failed to create partial tasks dir: %v", err)
	}

	if _, err := s.GetDataRoot(); err == nil {
		t.Fatalf("expected GetDataRoot to fail with incomplete data layout")
	}

	if err := os.MkdirAll("metadata/milestones", 0755); err != nil {
		t.Fatalf("failed to create metadata milestones dir: %v", err)
	}
	if err := os.MkdirAll("metadata/stories", 0755); err != nil {
		t.Fatalf("failed to create metadata stories dir: %v", err)
	}
	if err := os.MkdirAll("metadata/tasks", 0755); err != nil {
		t.Fatalf("failed to create metadata tasks dir: %v", err)
	}

	root, err := s.GetDataRoot()
	if err != nil {
		t.Fatalf("expected GetDataRoot to succeed with metadata layout: %v", err)
	}
	if root != "metadata" {
		t.Fatalf("expected data root to be metadata, got %q", root)
	}
}

func TestMilestoneCRUD(t *testing.T) {
	cleanup := setupTest(t)
	defer cleanup()

	title := "Test Milestone"
	slug, err := CreateMilestone(title, "m-1", "First line\\nSecond line")
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
	if m.Description != "First line\nSecond line" {
		t.Errorf("description not parsed correctly: %v", m.Description)
	}
	if m.CreatedAt == "" {
		t.Errorf("CreatedAt should be populated")
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

func TestTaskCRUD(t *testing.T) {
	cleanup := setupTest(t)
	defer cleanup()

	slug, err := CreateTask("Test Task", "t-1", "HIGH", "FEAT", "user1", "Task desc")
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

	mSlug, _ := CreateMilestone("M1", "m-1", "")
	tSlug, _ := CreateTask("T1", "t-1", "MED", "TASK", "", "")

	m, _ := GetMilestone(mSlug)
	tObj, _ := GetTask(tSlug)

	// Link T1 directly to M1
	if err := LinkEntities(tObj.ID, m.ID); err != nil {
		t.Fatalf("Link T1->M1 failed: %v", err)
	}

	m, _ = GetMilestone(m.ID)
	if len(m.Tasks) != 1 || m.Tasks[0] != tObj.ID {
		t.Errorf("Task not linked to Milestone")
	}

	// Unlink T1
	if err := UnlinkEntity(tObj.ID); err != nil {
		t.Fatalf("Unlink T1 failed: %v", err)
	}
	m, _ = GetMilestone(m.ID)
	if len(m.Tasks) != 0 {
		t.Errorf("Task not unlinked from Milestone")
	}
}

func TestEnsureTaskPlan(t *testing.T) {
	cleanup := setupTest(t)
	defer cleanup()

	tSlug, _ := CreateTask("T1", "t-1", "MED", "TASK", "", "")

	planPath, err := EnsureTaskPlan(tSlug)
	if err != nil {
		t.Fatalf("EnsureTaskPlan failed: %v", err)
	}
	if _, err := os.Stat(planPath); os.IsNotExist(err) {
		t.Errorf("Plan file was not created")
	}
}

func TestSortingAndListing(t *testing.T) {
	cleanup := setupTest(t)
	defer cleanup()

	if _, err := CreateMilestone("M1", "m-1", ""); err != nil {
		t.Fatalf("CreateMilestone failed: %v", err)
	}
	time.Sleep(10 * time.Millisecond) // Ensure different CreatedAt
	if _, err := CreateMilestone("M2", "m-2", ""); err != nil {
		t.Fatalf("CreateMilestone failed: %v", err)
	}
	time.Sleep(10 * time.Millisecond)
	if _, err := CreateMilestone("M10", "m-10", ""); err != nil {
		t.Fatalf("CreateMilestone failed: %v", err)
	}

	ms, _ := ListMilestones()
	if len(ms) != 3 {
		t.Errorf("Expected 3 milestones, got %d", len(ms))
	}

	SortMilestones(ms)
	if ms[0].Slug != "m-1" || ms[1].Slug != "m-2" || ms[2].Slug != "m-10" {
		t.Errorf("Milestones not sorted correctly: %v, %v, %v", ms[0].Slug, ms[1].Slug, ms[2].Slug)
	}
}

func TestErrorCases(t *testing.T) {
	cleanup := setupTest(t)
	defer cleanup()

	slug, _ := CreateTask("T1", "t-1", "MED", "TASK", "", "")
	tObj, _ := GetTask(slug)

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

func TestPrefixMatching(t *testing.T) {
	cleanup := setupTest(t)
	defer cleanup()

	slug, _ := CreateTask("T1", "t-uuid", "MED", "TASK", "", "")
	tObj, _ := GetTask(slug)

	prefix := tObj.ID[:8]
	resolved, err := ResolveEntity(prefix)
	if err != nil {
		t.Fatalf("ResolveEntity by prefix failed: %v", err)
	}
	if resolved.ID != tObj.ID {
		t.Errorf("Expected ID %s, got %s", tObj.ID, resolved.ID)
	}
}

func TestPullRequestMarkdownWorkflow(t *testing.T) {
	cleanup := setupTest(t)
	defer cleanup()

	// Create task to link to
	tSlug, err := CreateTask("Associated Task", "t-1", "HIGH", "TASK", "", "")
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	// Create Pull Request
	slug, err := CreatePullRequest("pr-oauth", "main", "feat-oauth", "https://github.com/source/repo.git", "https://github.com/dest/repo.git", "Auth details")
	if err != nil {
		t.Fatalf("CreatePullRequest failed: %v", err)
	}

	pr, err := GetPullRequest(slug)
	if err != nil {
		t.Fatalf("GetPullRequest failed: %v", err)
	}

	if pr.SourceRepo != "https://github.com/source/repo.git" {
		t.Errorf("expected SourceRepo 'https://github.com/source/repo.git', got %q", pr.SourceRepo)
	}
	if pr.DestRepo != "https://github.com/dest/repo.git" {
		t.Errorf("expected DestRepo 'https://github.com/dest/repo.git', got %q", pr.DestRepo)
	}

	mdPath, err := GetPullRequestMarkdownPath(pr.ID)
	if err != nil {
		t.Fatalf("GetPullRequestMarkdownPath failed: %v", err)
	}

	// Verify markdown file exists
	if _, err := os.Stat(mdPath); os.IsNotExist(err) {
		t.Errorf("Markdown template file not generated at %s", mdPath)
	}

	// Mock editing of the Markdown file
	modifiedContent := `# Pull Request: OAuth fixes updated
- Dest Branch: develop
- Source Branch/Tag: feat-oauth-v2
- Source Repo: https://github.com/source/repo-updated.git
- Dest Repo: https://github.com/dest/repo-updated.git
- Status: OPEN
- Tasks: t-1

## Description
We need develop OAuth features.
Added Google and Github buttons.
Integrated gologin wrappers.
`
	if err := os.WriteFile(mdPath, []byte(modifiedContent), 0644); err != nil {
		t.Fatalf("Failed to modify mock markdown file: %v", err)
	}

	// Parse it
	if err := ParsePullRequestMarkdown(pr.ID); err != nil {
		t.Fatalf("ParsePullRequestMarkdown failed: %v", err)
	}

	// Verify updates inside JSON
	updatedPR, err := GetPullRequest(slug)
	if err != nil {
		t.Fatalf("Failed to retrieve updated PR: %v", err)
	}

	if updatedPR.BaseBranch != "develop" {
		t.Errorf("expected BaseBranch 'develop', got %q", updatedPR.BaseBranch)
	}
	if updatedPR.HeadBranch != "feat-oauth-v2" {
		t.Errorf("expected HeadBranch 'feat-oauth-v2', got %q", updatedPR.HeadBranch)
	}
	if updatedPR.SourceRepo != "https://github.com/source/repo-updated.git" {
		t.Errorf("expected SourceRepo 'https://github.com/source/repo-updated.git', got %q", updatedPR.SourceRepo)
	}
	if updatedPR.DestRepo != "https://github.com/dest/repo-updated.git" {
		t.Errorf("expected DestRepo 'https://github.com/dest/repo-updated.git', got %q", updatedPR.DestRepo)
	}
	if updatedPR.Status != "OPEN" {
		t.Errorf("expected Status 'OPEN', got %q", updatedPR.Status)
	}
	expectedDesc := "We need develop OAuth features.\nAdded Google and Github buttons.\nIntegrated gologin wrappers."
	if strings.TrimSpace(updatedPR.Description) != expectedDesc {
		t.Errorf("expected Description %q, got %q", expectedDesc, updatedPR.Description)
	}

	// Verify bidirectional task link
	tObj, _ := GetTask(tSlug)
	if len(tObj.PullRequests) != 1 || tObj.PullRequests[0] != pr.ID {
		t.Errorf("Task not bidirectionally linked: %+v", tObj.PullRequests)
	}
	if len(updatedPR.Tasks) != 1 || updatedPR.Tasks[0] != tObj.ID {
		t.Errorf("PR tasks not linked: %+v", updatedPR.Tasks)
	}
}

func TestMultiInstanceStore(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	store1 := NewStore(dir1)
	store2 := NewStore(dir2)

	if err := store1.Initialize(dir1); err != nil {
		t.Fatalf("failed to initialize store1: %v", err)
	}
	if err := store2.Initialize(dir2); err != nil {
		t.Fatalf("failed to initialize store2: %v", err)
	}

	slug1, err := store1.CreateTask("Task 1 in Store 1", "t-1", "HIGH", "FEATURE", "user1", "Desc 1")
	if err != nil {
		t.Fatalf("failed to create task in store1: %v", err)
	}

	slug2, err := store2.CreateTask("Task 2 in Store 2", "t-2", "LOW", "BUG", "user2", "Desc 2")
	if err != nil {
		t.Fatalf("failed to create task in store2: %v", err)
	}

	// Verify isolation: store1 should only see task 1
	t1, err := store1.GetTask(slug1)
	if err != nil {
		t.Errorf("store1 failed to get task 1: %v", err)
	}
	if t1.Title != "Task 1 in Store 1" {
		t.Errorf("expected store1 task 1 title 'Task 1 in Store 1', got %q", t1.Title)
	}

	_, err = store1.GetTask(slug2)
	if err == nil {
		t.Errorf("store1 should not have access to task 2 from store2")
	}

	// Verify isolation: store2 should only see task 2
	t2, err := store2.GetTask(slug2)
	if err != nil {
		t.Errorf("store2 failed to get task 2: %v", err)
	}
	if t2.Title != "Task 2 in Store 2" {
		t.Errorf("expected store2 task 2 title 'Task 2 in Store 2', got %q", t2.Title)
	}

	_, err = store2.GetTask(slug1)
	if err == nil {
		t.Errorf("store2 should not have access to task 1 from store1")
	}
}

func TestDeletionSidecarsAndUnlinking(t *testing.T) {
	cleanup := setupTest(t)
	defer cleanup()

	// 1. Set up a Milestone, Task, and PR linked together
	mSlug, err := CreateMilestone("Milestone 1", "m-1", "Desc")
	if err != nil {
		t.Fatalf("CreateMilestone failed: %v", err)
	}
	tSlug, err := CreateTask("Task 1", "t-1", "HIGH", "FEATURE", "user", "Desc")
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}
	prSlug, err := CreatePullRequest("pr-1", "main", "feature", "src", "dest", "Desc")
	if err != nil {
		t.Fatalf("CreatePullRequest failed: %v", err)
	}

	// Link them
	if err := LinkEntities(tSlug, mSlug); err != nil {
		t.Fatalf("Linking Task to Milestone failed: %v", err)
	}
	if err := LinkEntities(prSlug, tSlug); err != nil {
		t.Fatalf("Linking PR to Task failed: %v", err)
	}

	// Create task plan and verify it exists
	planPath, err := EnsureTaskPlan(tSlug)
	if err != nil {
		t.Fatalf("EnsureTaskPlan failed: %v", err)
	}
	if _, err := os.Stat(planPath); os.IsNotExist(err) {
		t.Fatalf("Task plan markdown file was not created: %s", planPath)
	}

	// Resolve the PR and get its markdown path
	prObj, err := GetPullRequest(prSlug)
	if err != nil {
		t.Fatalf("GetPullRequest failed: %v", err)
	}
	prMdPath, err := defaultStore.GetPullRequestMarkdownPath(prObj.ID)
	if err != nil {
		t.Fatalf("GetPullRequestMarkdownPath failed: %v", err)
	}
	if _, err := os.Stat(prMdPath); os.IsNotExist(err) {
		t.Fatalf("PR markdown file was not created: %s", prMdPath)
	}

	// 2. Delete the Task and verify:
	// - Unlinked from Milestone
	// - Markdown plan file is deleted
	// - JSON file is deleted
	taskObj, err := GetTask(tSlug)
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}

	if err := DeleteTask(tSlug); err != nil {
		t.Fatalf("DeleteTask failed: %v", err)
	}

	if _, err := os.Stat(planPath); !os.IsNotExist(err) {
		t.Errorf("expected task plan markdown to be deleted, but it still exists at %s", planPath)
	}

	taskPath, _ := defaultStore.GetTaskPath(taskObj.ID)
	if _, err := os.Stat(taskPath); !os.IsNotExist(err) {
		t.Errorf("expected task JSON to be deleted, but it still exists at %s", taskPath)
	}

	// Verify unlinked from parent Milestone
	parentMilestone, err := GetMilestone(mSlug)
	if err != nil {
		t.Fatalf("GetMilestone failed: %v", err)
	}
	for _, id := range parentMilestone.Tasks {
		if id == taskObj.ID {
			t.Errorf("expected task %s to be unlinked from milestone, but it was found", taskObj.ID)
		}
	}

	// 3. Delete the PR and verify:
	// - Unlinked from Tasks
	// - PR markdown file is deleted
	// - PR JSON file is deleted
	if err := DeletePR(prSlug); err != nil {
		t.Fatalf("DeletePR failed: %v", err)
	}

	if _, err := os.Stat(prMdPath); !os.IsNotExist(err) {
		t.Errorf("expected PR markdown to be deleted, but it still exists at %s", prMdPath)
	}

	prPath, _ := defaultStore.GetPullRequestPath(prObj.ID)
	if _, err := os.Stat(prPath); !os.IsNotExist(err) {
		t.Errorf("expected PR JSON to be deleted, but it still exists at %s", prPath)
	}
}

func TestStoreLock(t *testing.T) {
	cleanup := setupTest(t)
	defer cleanup()

	// 1. Test standard lock acquisition and release
	unlock1, err := AcquireLock()
	if err != nil {
		t.Fatalf("expected AcquireLock to succeed, got: %v", err)
	}
	if unlock1 == nil {
		t.Fatal("expected unlock function to be non-nil")
	}

	// 2. Test concurrent acquisition fails / times out
	// Create another store instance pointing to same directory
	s2 := NewStore("")
	_, err = s2.AcquireLock()
	if err == nil {
		t.Fatal("expected s2.AcquireLock to fail (timeout) due to active lock")
	}
	if !strings.Contains(err.Error(), "locked by another process") {
		t.Errorf("expected lock timeout error message, got: %v", err)
	}

	// Release first lock
	unlock1()

	// Third attempt should now succeed
	unlock3, err := s2.AcquireLock()
	if err != nil {
		t.Fatalf("expected AcquireLock to succeed after unlock, got: %v", err)
	}
	unlock3()

	// 3. Test stale lock recovery (dead PID)
	root, err := defaultStore.GetDataRoot()
	if err != nil {
		t.Fatalf("GetDataRoot failed: %v", err)
	}
	lockPath := filepath.Join(root, ".metaboard.lock")

	// Create a lock file simulating a dead process ID.
	// We can use a PID that is historically highly unlikely to be running, e.g. 999999.
	// PIDs on Linux are normally limited to 32768 or 4194304, so 999999 is safe if dead,
	// but to be absolutely sure we can check if it is alive first and use a confirmed dead one.
	deadPID := 999999
	for isProcessAlive(deadPID) {
		deadPID++
	}

	err = os.WriteFile(lockPath, []byte(fmt.Sprintf("%d\n", deadPID)), 0600)
	if err != nil {
		t.Fatalf("failed to create fake stale lock file: %v", err)
	}

	// Now try to acquire lock. It should detect the dead PID, remove the stale file,
	// and successfully acquire the lock!
	unlock4, err := AcquireLock()
	if err != nil {
		t.Fatalf("expected AcquireLock to clean up stale lock and succeed, got error: %v", err)
	}
	if unlock4 == nil {
		t.Fatal("expected unlock function to be non-nil after stale lock cleanup")
	}
	unlock4()
}

func TestEnsurePullRequestMarkdownPreservation(t *testing.T) {
	cleanup := setupTest(t)
	defer cleanup()

	// 1. Create a dummy Pull Request
	prSlug, err := CreatePullRequest("main", "feature", "My PR Title", "This is an initial description.", "source-repo", "dest-repo")
	if err != nil {
		t.Fatalf("failed to create PR: %v", err)
	}

	pr, err := defaultStore.GetPullRequest(prSlug)
	if err != nil {
		t.Fatalf("failed to retrieve PR: %v", err)
	}

	mdPath, err := defaultStore.GetPullRequestMarkdownPath(pr.ID)
	if err != nil {
		t.Fatalf("failed to get PR markdown path: %v", err)
	}

	// Verify markdown file was created with initial description
	contentBytes, err := os.ReadFile(mdPath)
	if err != nil {
		t.Fatalf("failed to read PR markdown file: %v", err)
	}
	content := string(contentBytes)
	if !strings.Contains(content, "This is an initial description.") {
		t.Errorf("expected markdown to contain initial description, got: %s", content)
	}

	// 2. Simulate user editing the markdown file description directly
	userDescription := "## Description\nThis is a user-edited description.\nWith custom styling and list:\n- Point A\n- Point B\n"
	headerPart := fmt.Sprintf("# Pull Request: %s → %s\n- Dest Branch: %s\n- Source Branch/Tag: %s\n- Source Repo: %s\n- Dest Repo: %s\n- Status: %s\n- Tasks: %s\n\n",
		pr.HeadBranch, pr.BaseBranch, pr.BaseBranch, pr.HeadBranch, pr.SourceRepo, pr.DestRepo, pr.Status, "")
	newFileContent := headerPart + userDescription

	err = os.WriteFile(mdPath, []byte(newFileContent), 0644)
	if err != nil {
		t.Fatalf("failed to write custom user edits: %v", err)
	}

	// 3. Update status (which calls EnsurePullRequestMarkdown with forceUpdateDesc = false)
	err = UpdatePullRequestStatus(pr.ID, "OPEN")
	if err != nil {
		t.Fatalf("failed to update status: %v", err)
	}

	// Read markdown file again
	contentBytes, err = os.ReadFile(mdPath)
	if err != nil {
		t.Fatalf("failed to read markdown file: %v", err)
	}
	content = string(contentBytes)

	// Verify status in header was updated to OPEN
	if !strings.Contains(content, "- Status: OPEN") {
		t.Errorf("expected status header to be updated to OPEN, got: %s", content)
	}

	// Verify user-edited description was PRESERVED
	if !strings.Contains(content, "This is a user-edited description.") || !strings.Contains(content, "- Point A") {
		t.Errorf("expected user-edited description to be preserved, but it was clobbered: %s", content)
	}

	// 4. Force overwrite (forceUpdateDesc = true via package wrapper or calling store method directly)
	_, err = defaultStore.EnsurePullRequestMarkdown(pr.ID, pr.BaseBranch, pr.HeadBranch, "OPEN", pr.SourceRepo, pr.DestRepo, pr.Tasks, "Forced new description.", true)
	if err != nil {
		t.Fatalf("EnsurePullRequestMarkdown forced write failed: %v", err)
	}

	// Read markdown file again
	contentBytes, err = os.ReadFile(mdPath)
	if err != nil {
		t.Fatalf("failed to read markdown file: %v", err)
	}
	content = string(contentBytes)

	// Verify description was overwritten
	if !strings.Contains(content, "Forced new description.") {
		t.Errorf("expected description to be overwritten, got: %s", content)
	}
	if strings.Contains(content, "This is a user-edited description.") {
		t.Error("expected old user-edited description to be removed on forced overwrite, but it is still present")
	}
}



