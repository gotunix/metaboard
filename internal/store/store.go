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

var (
	dataDir     string
	resolvedDir string
)

// SetDataDir explicitly sets the base directory for all data.
func SetDataDir(dir string) {
	dataDir = dir
	resolvedDir = "" // Reset cached resolution
}

// Initialize creates the necessary directory structure in the specified path.
func Initialize(path string) error {
	dirs := []string{"milestones", "stories", "tasks"}
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
// It looks for milestones, stories, and tasks in '.' first, then './metadata'.
func GetDataRoot() (string, error) {
	if dataDir != "" {
		return dataDir, nil
	}
	if resolvedDir != "" {
		return resolvedDir, nil
	}

	// Heuristic: check '.' for standard data directories
	dirs := []string{"milestones", "stories", "tasks"}
	foundInCurrent := false
	for _, d := range dirs {
		if _, err := os.Stat(d); err == nil {
			foundInCurrent = true
			break
		}
	}
	if foundInCurrent {
		resolvedDir = "."
		return resolvedDir, nil
	}

	// Check 'metadata'
	foundInMetadata := false
	for _, d := range dirs {
		if _, err := os.Stat(filepath.Join("metadata", d)); err == nil {
			foundInMetadata = true
			break
		}
	}
	if foundInMetadata {
		resolvedDir = "metadata"
		return resolvedDir, nil
	}

	return "", fmt.Errorf("metaboard data not found. Run 'metaboard init' to initialize this repository")
}

func GetMilestonesDir() (string, error) {
	root, err := GetDataRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "milestones"), nil
}

func GetStoriesDir() (string, error) {
	root, err := GetDataRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "stories"), nil
}

func GetTasksDir() (string, error) {
	root, err := GetDataRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "tasks"), nil
}

func GetTaskPath(id string) (string, error) {
	dir, err := GetTasksDir()
	if err != nil {
		return "", err
	}
	if len(id) < 2 {
		return filepath.Join(dir, id+".json"), nil
	}
	prefix := id[:2]
	return filepath.Join(dir, prefix, id+".json"), nil
}

func GetTaskPlanPath(id string) (string, error) {
	dir, err := GetTasksDir()
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

func ListMilestones() ([]models.Milestone, error) {
	dir, err := GetMilestonesDir()
	if err != nil {
		return nil, err
	}
	files, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []models.Milestone{}, nil
		}
		return nil, err
	}

	var milestones []models.Milestone
	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".json") {
			continue
		}
		content, err := os.ReadFile(filepath.Join(dir, f.Name()))
		if err != nil {
			continue
		}
		var m models.Milestone
		if err := json.Unmarshal(content, &m); err == nil {
			milestones = append(milestones, m)
		}
	}
	return milestones, nil
}

func ListStories() ([]models.Story, error) {
	dir, err := GetStoriesDir()
	if err != nil {
		return nil, err
	}
	files, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []models.Story{}, nil
		}
		return nil, err
	}

	var stories []models.Story
	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".json") {
			continue
		}
		content, err := os.ReadFile(filepath.Join(dir, f.Name()))
		if err != nil {
			continue
		}
		var s models.Story
		if err := json.Unmarshal(content, &s); err == nil {
			stories = append(stories, s)
		}
	}
	return stories, nil
}

func ListTasks() ([]models.Task, error) {
	dir, err := GetTasksDir()
	if err != nil {
		return nil, err
	}
	var tasks []models.Task

	err = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".json") {
			content, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			var t models.Task
			if err := json.Unmarshal(content, &t); err == nil {
				tasks = append(tasks, t)
			}
		}
		return nil
	})

	if os.IsNotExist(err) {
		return []models.Task{}, nil
	}
	return tasks, err
}

// --- CRUD Core ---

func CreateMilestone(title, slug, description string) (string, error) {
	if slug == "" {
		slug = GenerateNextMilestoneSlug("m")
	}
	id := uuid.New().String()

	m := models.Milestone{
		ID:          id,
		Title:       title,
		Slug:        slug,
		Status:      "BACKLOG",
		Description: strings.ReplaceAll(description, "\\n", "\n"),
		Stories:     []string{},
		Tasks:       []string{},
	}

	dir, err := GetMilestonesDir()
	if err != nil {
		return "", err
	}

	path := filepath.Join(dir, id+".json")
	data, _ := json.MarshalIndent(m, "", "  ")
	return slug, os.WriteFile(path, data, 0644)
}

func GenerateNextMilestoneSlug(prefix string) string {
	ms, _ := ListMilestones()
	maxNum := 0
	prefixWithDash := prefix + "-"

	for _, m := range ms {
		if strings.HasPrefix(m.Slug, prefixWithDash) {
			numStr := strings.TrimPrefix(m.Slug, prefixWithDash)
			var n int
			if _, err := fmt.Sscanf(numStr, "%d", &n); err == nil {
				if n > maxNum {
					maxNum = n
				}
			}
		}
	}
	return fmt.Sprintf("%s-%d", prefix, maxNum+1)
}

func CreateStory(title, slug, description string) (string, error) {
	if slug == "" {
		slug = GenerateNextStorySlug("s")
	}
	id := uuid.New().String()

	s := models.Story{
		ID:          id,
		Title:       title,
		Slug:        slug,
		Status:      "BACKLOG",
		Description: strings.ReplaceAll(description, "\\n", "\n"),
		Tasks:       []string{},
	}

	dir, err := GetStoriesDir()
	if err != nil {
		return "", err
	}

	path := filepath.Join(dir, id+".json")
	data, _ := json.MarshalIndent(s, "", "  ")
	return slug, os.WriteFile(path, data, 0644)
}

func GenerateNextStorySlug(prefix string) string {
	stories, _ := ListStories()
	maxNum := 0
	prefixWithDash := prefix + "-"

	for _, s := range stories {
		if strings.HasPrefix(s.Slug, prefixWithDash) {
			numStr := strings.TrimPrefix(s.Slug, prefixWithDash)
			var n int
			if _, err := fmt.Sscanf(numStr, "%d", &n); err == nil {
				if n > maxNum {
					maxNum = n
				}
			}
		}
	}
	return fmt.Sprintf("%s-%d", prefix, maxNum+1)
}

func CreateTask(title, slug, priority, taskType, assignedTo, description string) (string, error) {
	if slug == "" {
		slug = GenerateNextTaskSlug("t")
	}
	id := uuid.New().String()

	t := models.Task{
		ID:          id,
		Slug:        slug,
		Title:       title,
		Status:      "BACKLOG",
		Priority:    strings.ToUpper(priority),
		Type:        strings.ToUpper(taskType),
		AssignedTo:  assignedTo,
		Description: strings.ReplaceAll(description, "\\n", "\n"),
		CreatedAt:   time.Now().Format("2006-01-02T15:04:05Z"),
		Tags:        []string{},
		DependsOn:   []string{},
	}

	path, err := GetTaskPath(id)
	if err != nil {
		return "", err
	}
	os.MkdirAll(filepath.Dir(path), 0755) // Still need to create shard directory if sharding
	data, _ := json.MarshalIndent(t, "", "  ")
	return slug, os.WriteFile(path, data, 0644)
}

func GenerateNextTaskSlug(prefix string) string {
	tasks, _ := ListTasks()
	maxNum := 0
	prefixWithDash := prefix + "-"

	for _, t := range tasks {
		if strings.HasPrefix(t.Slug, prefixWithDash) {
			numStr := strings.TrimPrefix(t.Slug, prefixWithDash)
			var n int
			if _, err := fmt.Sscanf(numStr, "%d", &n); err == nil {
				if n > maxNum {
					maxNum = n
				}
			}
		}
	}
	return fmt.Sprintf("%s-%d", prefix, maxNum+1)
}

func SaveMilestone(m models.Milestone) error {
	dir, err := GetMilestonesDir()
	if err != nil {
		return err
	}
	data, _ := json.MarshalIndent(m, "", "  ")
	return os.WriteFile(filepath.Join(dir, m.ID+".json"), data, 0644)
}

func SaveStory(s models.Story) error {
	dir, err := GetStoriesDir()
	if err != nil {
		return err
	}
	data, _ := json.MarshalIndent(s, "", "  ")
	return os.WriteFile(filepath.Join(dir, s.ID+".json"), data, 0644)
}

func SaveTask(t models.Task) error {
	path, err := GetTaskPath(t.ID)
	if err != nil {
		return err
	}
	data, _ := json.MarshalIndent(t, "", "  ")
	return os.WriteFile(path, data, 0644)
}

func UpdateMilestoneStatus(idOrSlug, newStatus string) error {
	m, err := GetMilestone(idOrSlug)
	if err != nil {
		return err
	}
	m.Status = strings.ToUpper(newStatus)
	if m.Status == "COMPLETED" {
		if m.CompletedAt == "" {
			m.CompletedAt = time.Now().Format("2006-01-02T15:04:05Z")
		}
	} else {
		m.CompletedAt = ""
	}
	return SaveMilestone(*m)
}

func UpdateStoryStatus(idOrSlug, newStatus string) error {
	s, err := GetStory(idOrSlug)
	if err != nil {
		return err
	}
	s.Status = strings.ToUpper(newStatus)
	if s.Status == "COMPLETED" {
		if s.CompletedAt == "" {
			s.CompletedAt = time.Now().Format("2006-01-02T15:04:05Z")
		}
	} else {
		s.CompletedAt = ""
	}
	return SaveStory(*s)
}

func UpdateTaskStatus(idOrSlug, newStatus string) error {
	t, err := GetTask(idOrSlug)
	if err != nil {
		return err
	}
	t.Status = strings.ToUpper(newStatus)
	if t.Status == "COMPLETED" {
		if t.CompletedAt == "" {
			t.CompletedAt = time.Now().Format("2006-01-02T15:04:05Z")
		}
	} else {
		t.CompletedAt = ""
	}
	return SaveTask(*t)
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
}

func UpdateTask(idOrSlug string, update TaskUpdate) error {
	t, err := GetTask(idOrSlug)
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

	if t.Status == "COMPLETED" {
		if t.CompletedAt == "" {
			t.CompletedAt = time.Now().Format("2006-01-02T15:04:05Z")
		}
	} else {
		t.CompletedAt = ""
	}

	return SaveTask(*t)
}

func GetMilestone(idOrSlug string) (*models.Milestone, error) {
	ms, err := ListMilestones()
	if err != nil {
		return nil, err
	}
	for _, m := range ms {
		if m.ID == idOrSlug || m.Slug == idOrSlug || m.Title == idOrSlug {
			return &m, nil
		}
	}
	return nil, fmt.Errorf("milestone %q not found", idOrSlug)
}

func GetStory(idOrSlug string) (*models.Story, error) {
	ss, err := ListStories()
	if err != nil {
		return nil, err
	}
	for _, s := range ss {
		if s.ID == idOrSlug || s.Slug == idOrSlug || s.Title == idOrSlug {
			return &s, nil
		}
	}
	return nil, fmt.Errorf("story %q not found", idOrSlug)
}

func GetTask(idOrSlug string) (*models.Task, error) {
	ts, err := ListTasks()
	if err != nil {
		return nil, err
	}
	for _, t := range ts {
		if t.ID == idOrSlug || t.Slug == idOrSlug || t.Title == idOrSlug {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("task %q not found", idOrSlug)
}

func EnsureTaskPlan(idOrSlug string) (string, error) {
	resolved, err := ResolveEntity(idOrSlug)
	if err != nil {
		return "", err
	}
	if resolved.Type != TypeTask {
		return "", fmt.Errorf("only tasks can have implementation plans")
	}

	planPath, err := GetTaskPlanPath(resolved.ID)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(planPath); os.IsNotExist(err) {
		template := "# Implementation Plan: " + idOrSlug + "\n\n## Context\n- \n\n## Technical Approach\n- \n\n## Acceptance Criteria\n- [ ] \n\n## Work Tree\n- [ ] \n"
		if err := os.WriteFile(planPath, []byte(template), 0644); err != nil {
			return "", err
		}
	}
	return planPath, nil
}

// --- Resolution ---

type EntityType string

const (
	TypeMilestone EntityType = "MILESTONE"
	TypeStory     EntityType = "STORY"
	TypeTask      EntityType = "TASK"
)

type ResolvedEntity struct {
	ID   string
	Type EntityType
	Path string
}

func ResolveEntity(idOrSlug string) (*ResolvedEntity, error) {
	ms, err := ListMilestones()
	if err == nil {
		for _, m := range ms {
			if m.ID == idOrSlug || m.Slug == idOrSlug || m.Title == idOrSlug {
				dir, _ := GetMilestonesDir()
				return &ResolvedEntity{ID: m.ID, Type: TypeMilestone, Path: filepath.Join(dir, m.ID+".json")}, nil
			}
		}
	}
	ss, err := ListStories()
	if err == nil {
		for _, s := range ss {
			if s.ID == idOrSlug || s.Slug == idOrSlug || s.Title == idOrSlug {
				dir, _ := GetStoriesDir()
				return &ResolvedEntity{ID: s.ID, Type: TypeStory, Path: filepath.Join(dir, s.ID+".json")}, nil
			}
		}
	}
	ts, err := ListTasks()
	if err == nil {
		for _, t := range ts {
			if t.ID == idOrSlug || t.Slug == idOrSlug || t.Title == idOrSlug {
				path, _ := GetTaskPath(t.ID)
				return &ResolvedEntity{ID: t.ID, Type: TypeTask, Path: path}, nil
			}
		}
	}
	return nil, fmt.Errorf("entity %q not found", idOrSlug)
}

func UnlinkEntity(childIDOrSlug string) error {
	child, err := ResolveEntity(childIDOrSlug)
	if err != nil {
		return err
	}

	if child.Type == TypeMilestone {
		return fmt.Errorf("Milestones cannot be unlinked (they have no parents)")
	}

	if child.Type == TypeStory {
		ms, _ := ListMilestones()
		for _, m := range ms {
			newStories := []string{}
			found := false
			for _, id := range m.Stories {
				if id == child.ID {
					found = true
					continue
				}
				newStories = append(newStories, id)
			}
			if found {
				m.Stories = newStories
				SaveMilestone(m)
			}
		}
	} else if child.Type == TypeTask {
		// Remove from Stories
		ss, _ := ListStories()
		for _, s := range ss {
			newTasks := []string{}
			found := false
			for _, id := range s.Tasks {
				if id == child.ID {
					found = true
					continue
				}
				newTasks = append(newTasks, id)
			}
			if found {
				s.Tasks = newTasks
				SaveStory(s)
			}
		}
		// Remove from Milestones
		ms, _ := ListMilestones()
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
				SaveMilestone(m)
			}
		}
	}
	return nil
}

func LinkEntities(childID, parentID string) error {
	child, err := ResolveEntity(childID)
	if err != nil {
		return err
	}
	parent, err := ResolveEntity(parentID)
	if err != nil {
		return err
	}

	if parent.Type == TypeTask {
		return fmt.Errorf("cannot link to a Task as a parent")
	}
	if child.Type == TypeMilestone {
		return fmt.Errorf("Milestones cannot be children")
	}
	if child.Type == TypeStory && parent.Type == TypeStory {
		return fmt.Errorf("cannot link a Story to another Story")
	}

	// First, unlink child from any existing parent to support "updating" links
	UnlinkEntity(child.ID)

	if parent.Type == TypeMilestone {
		var m models.Milestone
		c, _ := os.ReadFile(parent.Path)
		json.Unmarshal(c, &m)
		if child.Type == TypeStory {
			m.Stories = append(m.Stories, child.ID)
		} else {
			m.Tasks = append(m.Tasks, child.ID)
		}
		return SaveMilestone(m)
	} else if parent.Type == TypeStory {
		var s models.Story
		c, _ := os.ReadFile(parent.Path)
		json.Unmarshal(c, &s)
		s.Tasks = append(s.Tasks, child.ID)
		return SaveStory(s)
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
		return CompareNatural(ms[i].Slug, ms[j].Slug) < 0
	})
}
