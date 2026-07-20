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

package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"gotunix.net/metaboard/internal/models"
)

// --- Path Helpers ---

type Store struct {
	dataDir     string
	resolvedDir string
}

func NewStore(dataDir string) *Store {
	return &Store{dataDir: dataDir}
}

var defaultStore = &Store{}

func SetDataDir(dir string) {
	defaultStore.dataDir = dir
	defaultStore.resolvedDir = ""
}

// Initialize creates the necessary directory structure in the specified path.
func (store *Store) Initialize(path string) error {
	dirs := []string{"milestones", "tasks", "pullrequests"}
	for _, d := range dirs {
		fullPath := filepath.Join(path, d)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %q: %w", fullPath, err)
		}
		// Create .gitkeep to ensure empty directories are tracked
		keepPath := filepath.Join(fullPath, ".gitkeep")
		if err := os.WriteFile(keepPath, []byte(""), 0644); err != nil {
			return fmt.Errorf("failed to create .gitkeep in %q: %w", fullPath, err)
		}
	}
	return nil
}

// GetDataRoot returns the discovered or explicitly set data root.
// It looks for milestones, stories, and tasks in '.' first, then
// './boards', then './metadata' (legacy).
func (store *Store) GetDataRoot() (string, error) {
	if store.dataDir != "" {
		return store.dataDir, nil
	}
	if store.resolvedDir != "" {
		return store.resolvedDir, nil
	}

	// Heuristic: prefer the working directory when it contains a complete
	// metaboard data layout. This avoids false positives when one of the
	// project directories exists by coincidence.
	dirs := []string{"milestones", "tasks"}
	if store.hasAllDataDirs(".", dirs) {
		store.resolvedDir = "."
		return store.resolvedDir, nil
	}

	// Next, check the boards directory.
	if store.hasAllDataDirs("boards", dirs) {
		store.resolvedDir = "boards"
		return store.resolvedDir, nil
	}

	// Finally, check the legacy metadata directory.
	if store.hasAllDataDirs("metadata", dirs) {
		store.resolvedDir = "metadata"
		return store.resolvedDir, nil
	}

	return "", fmt.Errorf("metaboard data not found. Run metaboard to launch the dashboard and initialize this repository")
}

func (store *Store) hasAllDataDirs(root string, dirs []string) bool {
	for _, d := range dirs {
		info, err := os.Stat(filepath.Join(root, d))
		if err != nil || !info.IsDir() {
			return false
		}
	}
	return true
}

func (store *Store) GetMilestonesDir() (string, error) {
	root, err := store.GetDataRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "milestones"), nil
}

func (store *Store) GetMilestonePath(id string) (string, error) {
	dir, err := store.GetMilestonesDir()
	if err != nil {
		return "", err
	}
	if len(id) < 2 {
		return filepath.Join(dir, id+".json"), nil
	}
	prefix := id[:2]
	return filepath.Join(dir, prefix, id+".json"), nil
}

func (store *Store) GetTasksDir() (string, error) {
	root, err := store.GetDataRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "tasks"), nil
}

func (store *Store) GetTaskPath(id string) (string, error) {
	dir, err := store.GetTasksDir()
	if err != nil {
		return "", err
	}
	if len(id) < 2 {
		return filepath.Join(dir, id+".json"), nil
	}
	prefix := id[:2]
	return filepath.Join(dir, prefix, id+".json"), nil
}

func (store *Store) GetTaskPlanPath(id string) (string, error) {
	dir, err := store.GetTasksDir()
	if err != nil {
		return "", err
	}
	if len(id) < 2 {
		return filepath.Join(dir, id+".md"), nil
	}
	prefix := id[:2]
	return filepath.Join(dir, prefix, id+".md"), nil
}

func (store *Store) GetPullRequestsDir() (string, error) {
	root, err := store.GetDataRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "pullrequests"), nil
}

func (store *Store) GetPullRequestPath(id string) (string, error) {
	dir, err := store.GetPullRequestsDir()
	if err != nil {
		return "", err
	}
	if len(id) < 2 {
		return filepath.Join(dir, id+".json"), nil
	}
	prefix := id[:2]
	return filepath.Join(dir, prefix, id+".json"), nil
}

func (store *Store) GetPullRequestMarkdownPath(id string) (string, error) {
	dir, err := store.GetPullRequestsDir()
	if err != nil {
		return "", err
	}
	if len(id) < 2 {
		return filepath.Join(dir, id+".md"), nil
	}
	prefix := id[:2]
	return filepath.Join(dir, prefix, id+".md"), nil
}

// --- List Functions ---

func loadHistory[T any](path string) ([]T, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var history []T
	if err := json.Unmarshal(content, &history); err == nil {
		return history, nil
	}

	// Fallback for single object (legacy format)
	var single T
	if err := json.Unmarshal(content, &single); err == nil {
		return []T{single}, nil
	}

	return nil, fmt.Errorf("failed to unmarshal history from %s", path)
}

func (store *Store) ListMilestones() ([]models.Milestone, error) {
	dir, err := store.GetMilestonesDir()
	if err != nil {
		return nil, err
	}
	var milestones []models.Milestone

	err = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".json") {
			history, err := loadHistory[models.Milestone](path)
			if err != nil || len(history) == 0 {
				return nil
			}
			// Return latest version
			m := history[len(history)-1]
			if m.Version == 0 {
				m.Version = 1
			}
			milestones = append(milestones, m)
		}
		return nil
	})

	if os.IsNotExist(err) {
		return []models.Milestone{}, nil
	}
	SortMilestones(milestones)
	return milestones, err
}

func (store *Store) ListTasks() ([]models.Task, error) {
	dir, err := store.GetTasksDir()
	if err != nil {
		return nil, err
	}
	var tasks []models.Task

	err = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".json") {
			history, err := loadHistory[models.Task](path)
			if err != nil || len(history) == 0 {
				return nil
			}
			// Return latest version
			t := history[len(history)-1]
			if t.Version == 0 {
				t.Version = 1
			}
			tasks = append(tasks, t)
		}
		return nil
	})

	if os.IsNotExist(err) {
		return []models.Task{}, nil
	}
	SortTasks(tasks)
	return tasks, err
}

func (store *Store) ListPullRequests() ([]models.PullRequest, error) {
	dir, err := store.GetPullRequestsDir()
	if err != nil {
		return nil, err
	}
	var prs []models.PullRequest

	err = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".json") {
			history, err := loadHistory[models.PullRequest](path)
			if err != nil || len(history) == 0 {
				return nil
			}
			// Return latest version
			p := history[len(history)-1]
			if p.Version == 0 {
				p.Version = 1
			}
			prs = append(prs, p)
		}
		return nil
	})

	if os.IsNotExist(err) {
		return []models.PullRequest{}, nil
	}
	SortPullRequests(prs)
	return prs, err
}

// --- CRUD Core ---

func (store *Store) CreateMilestone(title, slug, description string) (string, error) {
	id := uuid.New().String()
	if slug == "" {
		slug = "m-" + id[:8]
	}

	m := models.Milestone{
		ID:          id,
		Version:     1,
		Title:       title,
		Slug:        slug,
		Status:      "BACKLOG",
		Description: strings.ReplaceAll(description, "\\n", "\n"),
		Tasks:       []string{},
		CreatedAt:   time.Now().Format(time.RFC3339Nano),
		UpdatedAt:   time.Now().Format(time.RFC3339Nano),
	}

	path, err := store.GetMilestonePath(id)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", err
	}
	data, err := json.MarshalIndent([]models.Milestone{m}, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling milestone: %w", err)
	}
	return slug, os.WriteFile(path, data, 0644)
}

func (store *Store) CreateTask(title, slug, priority, taskType, assignedTo, description string) (string, error) {
	id := uuid.New().String()
	if slug == "" {
		slug = "t-" + id[:8]
	}

	t := models.Task{
		ID:          id,
		Version:     1,
		Slug:        slug,
		Title:       title,
		Status:      "BACKLOG",
		Priority:    strings.ToUpper(priority),
		Type:        strings.ToUpper(taskType),
		AssignedTo:  assignedTo,
		Description: strings.ReplaceAll(description, "\\n", "\n"),
		CreatedAt:   time.Now().Format(time.RFC3339Nano),
		UpdatedAt:   time.Now().Format(time.RFC3339Nano),
		Tags:        []string{},
		DependsOn:   []string{},
	}

	path, err := store.GetTaskPath(id)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", err
	}
	data, err := json.MarshalIndent([]models.Task{t}, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling task: %w", err)
	}
	return slug, os.WriteFile(path, data, 0644)
}

func (store *Store) SaveMilestone(m models.Milestone) error {
	path, err := store.GetMilestonePath(m.ID)
	if err != nil {
		return err
	}
	history, err := loadHistory[models.Milestone](path)
	if err != nil {
		return fmt.Errorf("reading milestone history: %w", err)
	}

	lastVersion := 0
	if len(history) > 0 {
		lastVersion = history[len(history)-1].Version
		if lastVersion == 0 {
			lastVersion = 1
		}
	}
	m.Version = lastVersion + 1
	m.UpdatedAt = time.Now().Format(time.RFC3339Nano)

	history = append(history, m)
	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling milestone: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

func (store *Store) SaveTask(t models.Task) error {
	path, err := store.GetTaskPath(t.ID)
	if err != nil {
		return err
	}
	history, err := loadHistory[models.Task](path)
	if err != nil {
		return fmt.Errorf("reading task history: %w", err)
	}

	lastVersion := 0
	if len(history) > 0 {
		lastVersion = history[len(history)-1].Version
		if lastVersion == 0 {
			lastVersion = 1
		}
	}
	t.Version = lastVersion + 1
	t.UpdatedAt = time.Now().Format(time.RFC3339Nano)

	history = append(history, t)
	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling task: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

func (store *Store) CreatePullRequest(slug, baseBranch, headBranch, sourceRepo, destRepo, description string) (string, error) {
	id := uuid.New().String()
	if slug == "" {
		slug = "pr-" + id[:8]
	}

	pr := models.PullRequest{
		ID:          id,
		Version:     1,
		Slug:        slug,
		Status:      "DRAFT",
		BaseBranch:  baseBranch,
		HeadBranch:  headBranch,
		SourceRepo:  sourceRepo,
		DestRepo:    destRepo,
		Description: description,
		Tasks:       []string{},
		CreatedAt:   time.Now().Format(time.RFC3339Nano),
		UpdatedAt:   time.Now().Format(time.RFC3339Nano),
	}

	path, err := store.GetPullRequestPath(id)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", err
	}
	data, err := json.MarshalIndent([]models.PullRequest{pr}, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling pull request: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", err
	}

	// Generate the initial markdown template file
	_, err = store.EnsurePullRequestMarkdown(id, baseBranch, headBranch, "DRAFT", sourceRepo, destRepo, []string{}, description, true)
	if err != nil {
		return "", err
	}

	return slug, nil
}

func (store *Store) SavePullRequest(pr models.PullRequest) error {
	path, err := store.GetPullRequestPath(pr.ID)
	if err != nil {
		return err
	}
	history, err := loadHistory[models.PullRequest](path)
	if err != nil {
		return fmt.Errorf("reading pull request history: %w", err)
	}

	lastVersion := 0
	if len(history) > 0 {
		lastVersion = history[len(history)-1].Version
		if lastVersion == 0 {
			lastVersion = 1
		}
	}
	pr.Version = lastVersion + 1
	pr.UpdatedAt = time.Now().Format(time.RFC3339Nano)

	history = append(history, pr)
	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling pull request: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

func (store *Store) UpdatePullRequestStatus(idOrSlug, newStatus string) error {
	pr, err := store.GetPullRequest(idOrSlug)
	if err != nil {
		return err
	}
	pr.Status = strings.ToUpper(newStatus)
	if pr.Status == "MERGED" || pr.Status == "CLOSED" || pr.Status == "REJECTED" {
		if pr.CompletedAt == "" {
			pr.CompletedAt = time.Now().Format(time.RFC3339Nano)
		}
	} else {
		pr.CompletedAt = ""
	}

	if err := store.SavePullRequest(*pr); err != nil {
		return err
	}

	// Update MD file header status
	_, _ = store.EnsurePullRequestMarkdown(pr.ID, pr.BaseBranch, pr.HeadBranch, pr.Status, pr.SourceRepo, pr.DestRepo, pr.Tasks, pr.Description, false)
	return nil
}

type PullRequestUpdate struct {
	Status      *string
	BaseBranch  *string
	HeadBranch  *string
	Description *string
	SourceRepo  *string
	DestRepo    *string
}

func (store *Store) UpdatePullRequest(idOrSlug string, update PullRequestUpdate) error {
	pr, err := store.GetPullRequest(idOrSlug)
	if err != nil {
		return err
	}

	if update.Status != nil {
		pr.Status = strings.ToUpper(*update.Status)
	}
	if update.BaseBranch != nil {
		pr.BaseBranch = *update.BaseBranch
	}
	if update.HeadBranch != nil {
		pr.HeadBranch = *update.HeadBranch
	}
	if update.Description != nil {
		pr.Description = strings.ReplaceAll(*update.Description, "\\n", "\n")
	}
	if update.SourceRepo != nil {
		pr.SourceRepo = *update.SourceRepo
	}
	if update.DestRepo != nil {
		pr.DestRepo = *update.DestRepo
	}

	if pr.Status == "MERGED" || pr.Status == "CLOSED" || pr.Status == "REJECTED" {
		if pr.CompletedAt == "" {
			pr.CompletedAt = time.Now().Format(time.RFC3339Nano)
		}
	} else {
		pr.CompletedAt = ""
	}

	if err := store.SavePullRequest(*pr); err != nil {
		return err
	}

	_, _ = store.EnsurePullRequestMarkdown(pr.ID, pr.BaseBranch, pr.HeadBranch, pr.Status, pr.SourceRepo, pr.DestRepo, pr.Tasks, pr.Description, update.Description != nil)
	return nil
}

func (store *Store) UpdateMilestoneStatus(idOrSlug, newStatus string) error {
	m, err := store.GetMilestone(idOrSlug)
	if err != nil {
		return err
	}
	m.Status = strings.ToUpper(newStatus)
	if m.Status == "COMPLETED" {
		if m.CompletedAt == "" {
			m.CompletedAt = time.Now().Format(time.RFC3339Nano)
		}
	} else {
		m.CompletedAt = ""
	}
	return store.SaveMilestone(*m)
}

func (store *Store) UpdateTaskStatus(idOrSlug, newStatus string) error {
	t, err := store.GetTask(idOrSlug)
	if err != nil {
		return err
	}
	t.Status = strings.ToUpper(newStatus)
	if t.Status == "COMPLETED" {
		if t.CompletedAt == "" {
			t.CompletedAt = time.Now().Format(time.RFC3339Nano)
		}
	} else {
		t.CompletedAt = ""
	}
	return store.SaveTask(*t)
}

type TaskUpdate struct {
	Title       *string
	Status      *string
	Priority    *string
	Type        *string
	AssignedTo  *string
	Description *string
	Tags        *[]string
	DependsOn   *[]string
	Changelog   *bool
}

func (store *Store) UpdateTask(idOrSlug string, update TaskUpdate) error {
	t, err := store.GetTask(idOrSlug)
	if err != nil {
		return err
	}

	if update.Title != nil {
		t.Title = *update.Title
	}
	if update.Status != nil {
		t.Status = strings.ToUpper(*update.Status)
	}
	if update.Priority != nil {
		t.Priority = strings.ToUpper(*update.Priority)
	}
	if update.Type != nil {
		t.Type = strings.ToUpper(*update.Type)
	}
	if update.AssignedTo != nil {
		t.AssignedTo = *update.AssignedTo
	}
	if update.Description != nil {
		t.Description = strings.ReplaceAll(*update.Description, "\\n", "\n")
	}
	if update.Tags != nil {
		t.Tags = *update.Tags
	}
	if update.DependsOn != nil {
		t.DependsOn = *update.DependsOn
	}
	if update.Changelog != nil {
		t.Changelog = *update.Changelog
	}

	if t.Status == "COMPLETED" {
		if t.CompletedAt == "" {
			t.CompletedAt = time.Now().Format(time.RFC3339Nano)
		}
	} else {
		t.CompletedAt = ""
	}

	return store.SaveTask(*t)
}

func (store *Store) GetMilestone(idOrSlug string) (*models.Milestone, error) {
	ms, err := store.ListMilestones()
	if err != nil {
		return nil, err
	}
	for _, m := range ms {
		if m.ID == idOrSlug || m.Slug == idOrSlug || m.Title == idOrSlug || (len(idOrSlug) >= 4 && strings.HasPrefix(m.ID, idOrSlug)) {
			return &m, nil
		}
	}
	return nil, fmt.Errorf("milestone %q not found", idOrSlug)
}

func (store *Store) GetMilestoneHistory(idOrSlug string) ([]models.Milestone, error) {
	m, err := store.GetMilestone(idOrSlug)
	if err != nil {
		return nil, err
	}
	path, _ := store.GetMilestonePath(m.ID)
	return loadHistory[models.Milestone](path)
}

func (store *Store) GetMilestoneVersion(idOrSlug string, version int) (*models.Milestone, error) {
	history, err := store.GetMilestoneHistory(idOrSlug)
	if err != nil {
		return nil, err
	}
	for _, m := range history {
		if m.Version == version {
			return &m, nil
		}
	}
	return nil, fmt.Errorf("milestone %q version %d not found", idOrSlug, version)
}

func (store *Store) GetTask(idOrSlug string) (*models.Task, error) {
	ts, err := store.ListTasks()
	if err != nil {
		return nil, err
	}
	for _, t := range ts {
		if t.ID == idOrSlug || t.Slug == idOrSlug || t.Title == idOrSlug || (len(idOrSlug) >= 4 && strings.HasPrefix(t.ID, idOrSlug)) {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("task %q not found", idOrSlug)
}

func (store *Store) GetTaskHistory(idOrSlug string) ([]models.Task, error) {
	t, err := store.GetTask(idOrSlug)
	if err != nil {
		return nil, err
	}
	path, _ := store.GetTaskPath(t.ID)
	return loadHistory[models.Task](path)
}

func (store *Store) GetTaskVersion(idOrSlug string, version int) (*models.Task, error) {
	history, err := store.GetTaskHistory(idOrSlug)
	if err != nil {
		return nil, err
	}
	for _, t := range history {
		if t.Version == version {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("task %q version %d not found", idOrSlug, version)
}

func (store *Store) GetPullRequest(idOrSlug string) (*models.PullRequest, error) {
	prs, err := store.ListPullRequests()
	if err != nil {
		return nil, err
	}
	for _, pr := range prs {
		if pr.ID == idOrSlug || pr.Slug == idOrSlug || (len(idOrSlug) >= 4 && strings.HasPrefix(pr.ID, idOrSlug)) {
			return &pr, nil
		}
	}
	return nil, fmt.Errorf("pull request %q not found", idOrSlug)
}

func (store *Store) GetPullRequestHistory(idOrSlug string) ([]models.PullRequest, error) {
	pr, err := store.GetPullRequest(idOrSlug)
	if err != nil {
		return nil, err
	}
	path, _ := store.GetPullRequestPath(pr.ID)
	return loadHistory[models.PullRequest](path)
}

func (store *Store) GetPullRequestVersion(idOrSlug string, version int) (*models.PullRequest, error) {
	history, err := store.GetPullRequestHistory(idOrSlug)
	if err != nil {
		return nil, err
	}
	for _, pr := range history {
		if pr.Version == version {
			return &pr, nil
		}
	}
	return nil, fmt.Errorf("pull request %q version %d not found", idOrSlug, version)
}

func (store *Store) EnsurePullRequestMarkdown(idOrSlug string, base, head, status string, sourceRepo, destRepo string, tasks []string, description string, forceUpdateDesc bool) (string, error) {
	resolved, err := store.ResolveEntity(idOrSlug)
	if err != nil {
		return "", err
	}
	if resolved.Type != TypePullRequest {
		return "", fmt.Errorf("only pull requests can have templates")
	}

	mdPath, err := store.GetPullRequestMarkdownPath(resolved.ID)
	if err != nil {
		return "", err
	}

	taskSlugs := ""
	if len(tasks) > 0 {
		slugs := []string{}
		for _, tID := range tasks {
			if t, err := store.GetTask(tID); err == nil {
				slugs = append(slugs, t.Slug)
			} else {
				slugs = append(slugs, tID)
			}
		}
		taskSlugs = strings.Join(slugs, ", ")
	}

	// Use head → base as the label
	prLabel := fmt.Sprintf("%s → %s", head, base)

	var existingDescSection string
	if !forceUpdateDesc {
		if _, statErr := os.Stat(mdPath); statErr == nil {
			if data, readErr := os.ReadFile(mdPath); readErr == nil {
				lines := strings.Split(string(data), "\n")
				for i, line := range lines {
					trimmed := strings.TrimSpace(line)
					if strings.HasPrefix(trimmed, "## ") {
						headerName := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(trimmed, "## ")))
						if headerName == "description" || headerName == "desc" || headerName == "why" || headerName == "what" || headerName == "how" {
							existingDescSection = strings.Join(lines[i:], "\n")
							break
						}
					}
				}
			}
		}
	}

	var template string
	if existingDescSection != "" {
		header := fmt.Sprintf("# Pull Request: %s\n- Dest Branch: %s\n- Source Branch/Tag: %s\n- Source Repo: %s\n- Dest Repo: %s\n- Status: %s\n- Tasks: %s\n\n",
			prLabel, base, head, sourceRepo, destRepo, status, taskSlugs)
		template = header + existingDescSection
	} else {
		template = fmt.Sprintf("# Pull Request: %s\n- Dest Branch: %s\n- Source Branch/Tag: %s\n- Source Repo: %s\n- Dest Repo: %s\n- Status: %s\n- Tasks: %s\n\n## Description\n%s\n",
			prLabel, base, head, sourceRepo, destRepo, status, taskSlugs,
			strings.TrimSpace(description))
		if description == "" {
			template = fmt.Sprintf("# Pull Request: %s\n- Dest Branch: %s\n- Source Branch/Tag: %s\n- Source Repo: %s\n- Dest Repo: %s\n- Status: %s\n- Tasks: %s\n\n## Description\n- \n",
				prLabel, base, head, sourceRepo, destRepo, status, taskSlugs)
		}
	}

	if err := os.MkdirAll(filepath.Dir(mdPath), 0755); err != nil {
		return "", err
	}
	if err := os.WriteFile(mdPath, []byte(template), 0644); err != nil {
		return "", err
	}
	return mdPath, nil
}

func (store *Store) ParsePullRequestMarkdown(idOrSlug string) error {
	pr, err := store.GetPullRequest(idOrSlug)
	if err != nil {
		return err
	}

	mdPath, err := store.GetPullRequestMarkdownPath(pr.ID)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(mdPath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	baseBranch := pr.BaseBranch
	headBranch := pr.HeadBranch
	status := pr.Status
	sourceRepo := pr.SourceRepo
	destRepo := pr.DestRepo
	var taskSlugs []string
	var descriptionLines []string
	var currentSection string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# Pull Request:") {
			// Skip — title is now derived from head → base
			currentSection = ""
			continue
		}
		if strings.HasPrefix(trimmed, "- Dest Branch:") {
			baseBranch = strings.TrimSpace(strings.TrimPrefix(trimmed, "- Dest Branch:"))
			currentSection = ""
			continue
		}
		if strings.HasPrefix(trimmed, "- Base Branch:") {
			baseBranch = strings.TrimSpace(strings.TrimPrefix(trimmed, "- Base Branch:"))
			currentSection = ""
			continue
		}
		if strings.HasPrefix(trimmed, "- Source Branch/Tag:") {
			headBranch = strings.TrimSpace(strings.TrimPrefix(trimmed, "- Source Branch/Tag:"))
			currentSection = ""
			continue
		}
		if strings.HasPrefix(trimmed, "- Head Branch:") {
			headBranch = strings.TrimSpace(strings.TrimPrefix(trimmed, "- Head Branch:"))
			currentSection = ""
			continue
		}
		if strings.HasPrefix(trimmed, "- Source Repo:") {
			sourceRepo = strings.TrimSpace(strings.TrimPrefix(trimmed, "- Source Repo:"))
			currentSection = ""
			continue
		}
		if strings.HasPrefix(trimmed, "- Dest Repo:") {
			destRepo = strings.TrimSpace(strings.TrimPrefix(trimmed, "- Dest Repo:"))
			currentSection = ""
			continue
		}
		if strings.HasPrefix(trimmed, "- Status:") {
			status = strings.ToUpper(strings.TrimSpace(strings.TrimPrefix(trimmed, "- Status:")))
			currentSection = ""
			continue
		}
		if strings.HasPrefix(trimmed, "- Tasks:") {
			tasksVal := strings.TrimSpace(strings.TrimPrefix(trimmed, "- Tasks:"))
			if tasksVal != "" {
				for _, t := range strings.Split(tasksVal, ",") {
					t = strings.TrimSpace(t)
					if t != "" {
						taskSlugs = append(taskSlugs, t)
					}
				}
			}
			currentSection = ""
			continue
		}

		if strings.HasPrefix(trimmed, "## ") {
			header := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(trimmed, "## ")))
			if header == "description" || header == "desc" || header == "why" || header == "what" || header == "how" {
				currentSection = "description"
			} else {
				currentSection = ""
			}
			continue
		}

		if currentSection == "description" {
			descriptionLines = append(descriptionLines, line)
		}
	}

	pr.BaseBranch = baseBranch
	pr.HeadBranch = headBranch
	pr.Status = status
	pr.SourceRepo = sourceRepo
	pr.DestRepo = destRepo
	pr.Description = strings.TrimSpace(strings.Join(descriptionLines, "\n"))

	if pr.Status == "MERGED" || pr.Status == "CLOSED" || pr.Status == "REJECTED" {
		if pr.CompletedAt == "" {
			pr.CompletedAt = time.Now().Format(time.RFC3339Nano)
		}
	} else {
		pr.CompletedAt = ""
	}

	var newTaskIDs []string
	for _, slug := range taskSlugs {
		if resolved, err := store.ResolveEntity(slug); err == nil && resolved.Type == TypeTask {
			newTaskIDs = append(newTaskIDs, resolved.ID)
		} else {
			newTaskIDs = append(newTaskIDs, slug)
		}
	}

	oldTasks := pr.Tasks
	pr.Tasks = newTaskIDs

	if err := store.SavePullRequest(*pr); err != nil {
		return err
	}

	for _, oldID := range oldTasks {
		found := false
		for _, newID := range newTaskIDs {
			if oldID == newID {
				found = true
				break
			}
		}
		if !found {
			if t, err := store.GetTask(oldID); err == nil {
				newPRs := []string{}
				for _, prID := range t.PullRequests {
					if prID != pr.ID {
						newPRs = append(newPRs, prID)
					}
				}
				t.PullRequests = newPRs
				_ = store.SaveTask(*t)
			}
		}
	}

	for _, newID := range newTaskIDs {
		found := false
		for _, oldID := range oldTasks {
			if oldID == newID {
				found = true
				break
			}
		}
		if !found {
			if t, err := store.GetTask(newID); err == nil {
				alreadyLinked := false
				for _, prID := range t.PullRequests {
					if prID == pr.ID {
						alreadyLinked = true
						break
					}
				}
				if !alreadyLinked {
					t.PullRequests = append(t.PullRequests, pr.ID)
					_ = store.SaveTask(*t)
				}
			}
		}
	}

	return nil
}

func (store *Store) EnsureTaskPlan(idOrSlug string) (string, error) {
	resolved, err := store.ResolveEntity(idOrSlug)
	if err != nil {
		return "", err
	}
	if resolved.Type != TypeTask {
		return "", fmt.Errorf("only tasks can have additional details")
	}

	planPath, err := store.GetTaskPlanPath(resolved.ID)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(planPath); os.IsNotExist(err) {
		template := "# Implementation Plan: " + idOrSlug + "\n\n" +
			"> **Task ID:** " + idOrSlug + "\n\n" +
			"## PHASE 0: PRE-FLIGHT HANDSHAKE & AGENT LOCKOUT\n\n" +
			"> **CRITICAL AGENT INSTRUCTION:** You are currently locked in \"READ-ONLY\n" +
			"> PLANNING MODE\". You are strictly unauthorized to create, delete, or modify\n" +
			"> any project code files, configuration files, or scripts. You may only use\n" +
			"> file reading tools to gather context and output text planning.\n\n" +
			"### Step 1: Rule Ingestion Requirement\n\n" +
			"To unlock the execution phases, you must first read and parse AGENTS.md (and\n" +
			"its required initialization guardrails like ARCHITECTURE.md), as well as\n" +
			"docs/TASK_MANAGEMENT_GUIDE.md for task tracking workflows.\n\n" +
			"- **Action Required**: Provide the human with a brief 2-sentence summary of how\n" +
			"  the AGENTS.md rules govern this specific task.\n" +
			"- **Action Required**: Quote the exact line from the rules/architecture docs\n" +
			"  that dictates how this specific feature should be structured or placed.\n\n" +
			"### Step 2: Human Authorization Token\n\n" +
			"- **Current Status:** LOCKED\n" +
			"- **Instructions for AI**: To request modification authorization, you must\n" +
			"  draft your proposed approach in Phase 1 below and then output the exact\n" +
			"  string: [REQUESTING_WRITE_AUTHORIZATION] to the user.\n" +
			"- **Do not modify code until the human explicitly replies with:**\n" +
			"  [WRITE_AUTHORIZATION_GRANTED]. Once granted, you may update the status\n" +
			"  above to UNLOCKED and proceed to Phase 2.\n\n" +
			"---\n\n" +
			"## PHASE 1: Architectural Blueprint\n\n" +
			"_(To be populated by the AI during the READ-ONLY planning phase)_\n\n" +
			"### Objective\n\n" +
			"- **Current Behavior:** [What does the system do right now?]\n" +
			"- **Expected Behavior:**\n" +
			"  [What should the system do after this task is complete?]\n" +
			"- **Acceptance Criteria:**\n" +
			"  - [ ] [Criteria 1 (e.g., The API returns a 200 OK with the new JSON payload)]\n" +
			"  - [ ] [Criteria 2]\n\n" +
			"### Files to Modify\n\n" +
			"- **[file_path]**:\n" +
			"  [Brief description of what is changing in this file (e.g., adding X function, updating Y struct)]\n" +
			"- **[file_path]**: [Brief description of what is changing in this file]\n\n" +
			"### Proposed Implementation & Mockups\n\n" +
			"_[Provide detailed step-by-step logic of what will change. YOU MUST INCLUDE\n" +
			"pseudo-code, structural mockups, or diff-style representations of the\n" +
			"specific functions, interfaces, or components being modified. Ensure this\n" +
			"aligns with project rules.]_\n\n" +
			"---\n\n" +
			"## PHASE 2: Execution\n\n" +
			"_(Agent executes steps here ONLY after status is changed to UNLOCKED)_\n\n" +
			"- [ ] Run `metaboard task status [slug] IN-PROGRESS` to mark the start of work.\n" +
			"- [ ] [Execution Step 1]\n" +
			"- [ ] [Execution Step 2]\n" +
			"- [ ] [Execution Step 3]\n\n" +
			"---\n\n" +
			"## PHASE 3: Testing & Verification\n\n" +
			"_(Agent ensures changes comply with TESTING.md rules)_\n\n" +
			"- [ ] All new logic has corresponding test coverage.\n" +
			"- [ ] No regression introduced.\n" +
			"- [ ] Adheres to Git Hygiene guidelines (e.g., no transient artifacts left\n" +
			"      behind).\n\n" +
			"### Step 3: Human Completion Token\n\n" +
			"- **Current Status:** PENDING VERIFICATION\n" +
			"- **Instructions for AI**: Once execution and testing are complete, you must\n" +
			"  output the exact string [REQUESTING_COMPLETION_AUTHORIZATION] and wait.\n" +
			"- **Do not mark the task as completed until the human explicitly replies\n" +
			"  with:** [COMPLETION_AUTHORIZATION_GRANTED]. Once granted, update the status\n" +
			"  above to VERIFIED.\n\n" +
			"- [ ] Run `metaboard task status [slug] COMPLETED` (or link PR) as instructed\n" +
			"      in docs/TASK_MANAGEMENT_GUIDE.md.\n"
		if err := os.WriteFile(planPath, []byte(template), 0644); err != nil {
			return "", err
		}
	}
	return planPath, nil
}

// --- Resolution ---

type EntityType string

const (
	TypeMilestone   EntityType = "MILESTONE"
	TypeTask        EntityType = "TASK"
	TypePullRequest EntityType = "PULLREQUEST"
)

type ResolvedEntity struct {
	ID   string
	Type EntityType
	Path string
}

func (store *Store) ResolveEntity(idOrSlug string) (*ResolvedEntity, error) {
	ms, err := store.ListMilestones()
	if err == nil {
		for _, m := range ms {
			if m.ID == idOrSlug || m.Slug == idOrSlug || m.Title == idOrSlug || (len(idOrSlug) >= 4 && strings.HasPrefix(m.ID, idOrSlug)) {
				path, _ := store.GetMilestonePath(m.ID)
				return &ResolvedEntity{ID: m.ID, Type: TypeMilestone, Path: path}, nil
			}
		}
	}
	ts, err := store.ListTasks()
	if err == nil {
		for _, t := range ts {
			if t.ID == idOrSlug || t.Slug == idOrSlug || t.Title == idOrSlug || (len(idOrSlug) >= 4 && strings.HasPrefix(t.ID, idOrSlug)) {
				path, _ := store.GetTaskPath(t.ID)
				return &ResolvedEntity{ID: t.ID, Type: TypeTask, Path: path}, nil
			}
		}
	}
	prs, err := store.ListPullRequests()
	if err == nil {
		for _, pr := range prs {
			if pr.ID == idOrSlug || pr.Slug == idOrSlug || (len(idOrSlug) >= 4 && strings.HasPrefix(pr.ID, idOrSlug)) {
				path, _ := store.GetPullRequestPath(pr.ID)
				return &ResolvedEntity{ID: pr.ID, Type: TypePullRequest, Path: path}, nil
			}
		}
	}
	return nil, fmt.Errorf("entity %q not found", idOrSlug)
}

func (store *Store) UnlinkEntity(childIDOrSlug string) error {
	child, err := store.ResolveEntity(childIDOrSlug)
	if err != nil {
		return err
	}

	if child.Type == TypeMilestone {
		return fmt.Errorf("milestones cannot be unlinked (they have no parents)")
	}

	if child.Type == TypeTask {
		// Remove from Milestones
		ms, _ := store.ListMilestones()
		for _, m := range ms {
			newTasks := []string{}
			found := false
			for _, id := range m.Tasks {
				if id == child.ID {
					found = true
					continue
				}
				newTasks = append(newTasks, id)
			}
			if found {
				m.Tasks = newTasks
				_ = store.SaveMilestone(m)
			}
		}
	} else if child.Type == TypePullRequest {
		// Remove from Tasks
		ts, _ := store.ListTasks()
		for _, t := range ts {
			newPRs := []string{}
			found := false
			for _, id := range t.PullRequests {
				if id == child.ID {
					found = true
					continue
				}
				newPRs = append(newPRs, id)
			}
			if found {
				t.PullRequests = newPRs
				_ = store.SaveTask(t)
			}
		}

		// Clean the PR's Tasks list
		pr, err := store.GetPullRequest(child.ID)
		if err == nil {
			pr.Tasks = []string{}
			_ = store.SavePullRequest(*pr)
			_, _ = store.EnsurePullRequestMarkdown(pr.ID, pr.BaseBranch, pr.HeadBranch, pr.Status, pr.SourceRepo, pr.DestRepo, pr.Tasks, pr.Description, false)
		}
	}
	return nil
}

func (store *Store) LinkEntities(childID, parentID string) error {
	child, err := store.ResolveEntity(childID)
	if err != nil {
		return err
	}
	parent, err := store.ResolveEntity(parentID)
	if err != nil {
		return err
	}

	if parent.Type == TypeTask && child.Type != TypePullRequest {
		return fmt.Errorf("cannot link to a Task as a parent unless linking a Pull Request")
	}
	if parent.Type == TypePullRequest {
		return fmt.Errorf("cannot link to a Pull Request as a parent")
	}
	if child.Type == TypeMilestone {
		return fmt.Errorf("milestones cannot be children")
	}
	if child.Type == TypePullRequest && parent.Type != TypeTask {
		return fmt.Errorf("pull requests can only be linked to tasks")
	}

	// Pull requests may be linked to multiple tasks, so do not unlink prior associations.
	if child.Type != TypePullRequest {
		// First, unlink child from any existing parent to support "updating" links
		_ = store.UnlinkEntity(child.ID)
	}

	if parent.Type == TypeMilestone {
		m, err := store.GetMilestone(parent.ID)
		if err != nil {
			return err
		}
		m.Tasks = append(m.Tasks, child.ID)
		return store.SaveMilestone(*m)
	} else if parent.Type == TypeTask {
		t, err := store.GetTask(parent.ID)
		if err != nil {
			return err
		}

		tAlreadyLinked := false
		for _, prID := range t.PullRequests {
			if prID == child.ID {
				tAlreadyLinked = true
				break
			}
		}
		if !tAlreadyLinked {
			t.PullRequests = append(t.PullRequests, child.ID)
			_ = store.SaveTask(*t)
		}

		pr, err := store.GetPullRequest(child.ID)
		if err != nil {
			return err
		}
		prAlreadyLinked := false
		for _, taskID := range pr.Tasks {
			if taskID == parent.ID {
				prAlreadyLinked = true
				break
			}
		}
		if !prAlreadyLinked {
			pr.Tasks = append(pr.Tasks, parent.ID)
			_ = store.SavePullRequest(*pr)
			_, _ = store.EnsurePullRequestMarkdown(pr.ID, pr.BaseBranch, pr.HeadBranch, pr.Status, pr.SourceRepo, pr.DestRepo, pr.Tasks, pr.Description, false)
		}
		return nil
	}
	return nil
}

// --- Sorting ---

func isDigit(r uint8) bool { return r >= '0' && r <= '9' }

func CompareNatural(s1, s2 string) int {
	l1, l2 := len(s1), len(s2)
	i, j := 0, 0
	for i < l1 && j < l2 {
		if isDigit(s1[i]) && isDigit(s2[j]) {
			var n1, n2 int
			for i < l1 && isDigit(s1[i]) {
				n1 = n1*10 + int(s1[i]-'0')
				i++
			}
			for j < l2 && isDigit(s2[j]) {
				n2 = n2*10 + int(s2[j]-'0')
				j++
			}
			if n1 != n2 {
				if n1 < n2 {
					return -1
				}
				return 1
			}
		} else {
			if s1[i] != s2[j] {
				if s1[i] < s2[j] {
					return -1
				}
				return 1
			}
			i++
			j++
		}
	}
	if l1 < l2 {
		return -1
	}
	if l1 > l2 {
		return 1
	}
	return 0
}

func NaturalSort(items []string) {
	sort.Slice(items, func(i, j int) bool {
		return CompareNatural(items[i], items[j]) < 0
	})
}

func SortMilestones(ms []models.Milestone) {
	sort.Slice(ms, func(i, j int) bool {
		if ms[i].CreatedAt != ms[j].CreatedAt {
			return ms[i].CreatedAt < ms[j].CreatedAt
		}
		return ms[i].Slug < ms[j].Slug
	})
}

func SortPullRequests(prs []models.PullRequest) {
	sort.Slice(prs, func(i, j int) bool {
		if prs[i].CreatedAt != prs[j].CreatedAt {
			return prs[i].CreatedAt < prs[j].CreatedAt
		}
		return prs[i].Slug < prs[j].Slug
	})
}

func SortTasks(ts []models.Task) {
	priorityRank := map[string]int{"HIGH": 1, "MEDIUM": 2, "LOW": 3, "BACKLOG": 4}
	sort.Slice(ts, func(i, j int) bool {
		p1 := priorityRank[strings.ToUpper(ts[i].Priority)]
		p2 := priorityRank[strings.ToUpper(ts[j].Priority)]
		if p1 == 0 {
			p1 = 5
		}
		if p2 == 0 {
			p2 = 5
		}

		if p1 != p2 {
			return p1 < p2
		}

		if ts[i].CreatedAt != ts[j].CreatedAt {
			return ts[i].CreatedAt < ts[j].CreatedAt
		}
		return ts[i].Slug < ts[j].Slug
	})
}

// --- Backward Compatibility Wrappers ---

// Initialize wrapper for backward compatibility
func Initialize(path string) error {
	return defaultStore.Initialize(path)
}

// GetDataRoot wrapper for backward compatibility
func GetDataRoot() (string, error) {
	return defaultStore.GetDataRoot()
}

// GetMilestonesDir wrapper for backward compatibility
func GetMilestonesDir() (string, error) {
	return defaultStore.GetMilestonesDir()
}

// GetMilestonePath wrapper for backward compatibility
func GetMilestonePath(id string) (string, error) {
	return defaultStore.GetMilestonePath(id)
}

// GetTasksDir wrapper for backward compatibility
func GetTasksDir() (string, error) {
	return defaultStore.GetTasksDir()
}

// GetTaskPath wrapper for backward compatibility
func GetTaskPath(id string) (string, error) {
	return defaultStore.GetTaskPath(id)
}

// GetTaskPlanPath wrapper for backward compatibility
func GetTaskPlanPath(id string) (string, error) {
	return defaultStore.GetTaskPlanPath(id)
}

// GetPullRequestsDir wrapper for backward compatibility
func GetPullRequestsDir() (string, error) {
	return defaultStore.GetPullRequestsDir()
}

// GetPullRequestPath wrapper for backward compatibility
func GetPullRequestPath(id string) (string, error) {
	return defaultStore.GetPullRequestPath(id)
}

// GetPullRequestMarkdownPath wrapper for backward compatibility
func GetPullRequestMarkdownPath(id string) (string, error) {
	return defaultStore.GetPullRequestMarkdownPath(id)
}

// ListMilestones wrapper for backward compatibility
func ListMilestones() ([]models.Milestone, error) {
	return defaultStore.ListMilestones()
}

// ListTasks wrapper for backward compatibility
func ListTasks() ([]models.Task, error) {
	return defaultStore.ListTasks()
}

// ListPullRequests wrapper for backward compatibility
func ListPullRequests() ([]models.PullRequest, error) {
	return defaultStore.ListPullRequests()
}

// CreateMilestone wrapper for backward compatibility
func CreateMilestone(title, slug, description string) (string, error) {
	return defaultStore.CreateMilestone(title, slug, description)
}

// CreateTask wrapper for backward compatibility
func CreateTask(title, slug, priority, taskType, assignedTo, description string) (string, error) {
	return defaultStore.CreateTask(title, slug, priority, taskType, assignedTo, description)
}

// SaveMilestone wrapper for backward compatibility
func SaveMilestone(m models.Milestone) error {
	return defaultStore.SaveMilestone(m)
}

// SaveTask wrapper for backward compatibility
func SaveTask(t models.Task) error {
	return defaultStore.SaveTask(t)
}

// CreatePullRequest wrapper for backward compatibility
func CreatePullRequest(slug, baseBranch, headBranch, sourceRepo, destRepo, description string) (string, error) {
	return defaultStore.CreatePullRequest(slug, baseBranch, headBranch, sourceRepo, destRepo, description)
}

// SavePullRequest wrapper for backward compatibility
func SavePullRequest(pr models.PullRequest) error {
	return defaultStore.SavePullRequest(pr)
}

// UpdatePullRequestStatus wrapper for backward compatibility
func UpdatePullRequestStatus(idOrSlug, newStatus string) error {
	return defaultStore.UpdatePullRequestStatus(idOrSlug, newStatus)
}

// UpdatePullRequest wrapper for backward compatibility
func UpdatePullRequest(idOrSlug string, update PullRequestUpdate) error {
	return defaultStore.UpdatePullRequest(idOrSlug, update)
}

// UpdateMilestoneStatus wrapper for backward compatibility
func UpdateMilestoneStatus(idOrSlug, newStatus string) error {
	return defaultStore.UpdateMilestoneStatus(idOrSlug, newStatus)
}

// UpdateTaskStatus wrapper for backward compatibility
func UpdateTaskStatus(idOrSlug, newStatus string) error {
	return defaultStore.UpdateTaskStatus(idOrSlug, newStatus)
}

// UpdateTask wrapper for backward compatibility
func UpdateTask(idOrSlug string, update TaskUpdate) error {
	return defaultStore.UpdateTask(idOrSlug, update)
}

// GetMilestone wrapper for backward compatibility
func GetMilestone(idOrSlug string) (*models.Milestone, error) {
	return defaultStore.GetMilestone(idOrSlug)
}

// GetMilestoneHistory wrapper for backward compatibility
func GetMilestoneHistory(idOrSlug string) ([]models.Milestone, error) {
	return defaultStore.GetMilestoneHistory(idOrSlug)
}

// GetMilestoneVersion wrapper for backward compatibility
func GetMilestoneVersion(idOrSlug string, version int) (*models.Milestone, error) {
	return defaultStore.GetMilestoneVersion(idOrSlug, version)
}

// GetTask wrapper for backward compatibility
func GetTask(idOrSlug string) (*models.Task, error) {
	return defaultStore.GetTask(idOrSlug)
}

// GetTaskHistory wrapper for backward compatibility
func GetTaskHistory(idOrSlug string) ([]models.Task, error) {
	return defaultStore.GetTaskHistory(idOrSlug)
}

// GetTaskVersion wrapper for backward compatibility
func GetTaskVersion(idOrSlug string, version int) (*models.Task, error) {
	return defaultStore.GetTaskVersion(idOrSlug, version)
}

// GetPullRequest wrapper for backward compatibility
func GetPullRequest(idOrSlug string) (*models.PullRequest, error) {
	return defaultStore.GetPullRequest(idOrSlug)
}

// GetPullRequestHistory wrapper for backward compatibility
func GetPullRequestHistory(idOrSlug string) ([]models.PullRequest, error) {
	return defaultStore.GetPullRequestHistory(idOrSlug)
}

// GetPullRequestVersion wrapper for backward compatibility
func GetPullRequestVersion(idOrSlug string, version int) (*models.PullRequest, error) {
	return defaultStore.GetPullRequestVersion(idOrSlug, version)
}

// EnsurePullRequestMarkdown wrapper for backward compatibility
func EnsurePullRequestMarkdown(idOrSlug string, base, head, status string, sourceRepo, destRepo string, tasks []string, description string) (string, error) {
	return defaultStore.EnsurePullRequestMarkdown(idOrSlug, base, head, status, sourceRepo, destRepo, tasks, description, true)
}

// ParsePullRequestMarkdown wrapper for backward compatibility
func ParsePullRequestMarkdown(idOrSlug string) error {
	return defaultStore.ParsePullRequestMarkdown(idOrSlug)
}

// EnsureTaskPlan wrapper for backward compatibility
func EnsureTaskPlan(idOrSlug string) (string, error) {
	return defaultStore.EnsureTaskPlan(idOrSlug)
}

// ResolveEntity wrapper for backward compatibility
func ResolveEntity(idOrSlug string) (*ResolvedEntity, error) {
	return defaultStore.ResolveEntity(idOrSlug)
}

// UnlinkEntity wrapper for backward compatibility
func UnlinkEntity(childIDOrSlug string) error {
	return defaultStore.UnlinkEntity(childIDOrSlug)
}

// LinkEntities wrapper for backward compatibility
func LinkEntities(childID, parentID string) error {
	return defaultStore.LinkEntities(childID, parentID)
}
