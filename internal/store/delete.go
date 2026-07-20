// SPDX-License-Identifier: GPL-3.0-or-later
// SPDX-FileCopyrightText: 2026 The MetaBoard authors

package store

import (
	"fmt"
	"os"
)

// DeleteMilestone deletes a milestone, ensuring it has no attached stories or tasks.
func (store *Store) DeleteMilestone(idOrSlug string) error {
	m, err := store.GetMilestone(idOrSlug)
	if err != nil {
		return err
	}
	if len(m.Tasks) > 0 {
		return fmt.Errorf("cannot delete milestone: it still has linked tasks")
	}
	path, err := store.GetMilestonePath(m.ID)
	if err != nil {
		return err
	}
	return os.Remove(path)
}

// DeleteTask deletes a task and its plan markdown file.
func (store *Store) DeleteTask(idOrSlug string) error {
	t, err := store.GetTask(idOrSlug)
	if err != nil {
		return err
	}

	if err := store.UnlinkEntity(t.ID); err != nil {
		return fmt.Errorf("unlinking task: %w", err)
	}

	// Delete the task plan markdown sidecar if it exists
	if mdPath, errMD := store.GetTaskPlanPath(t.ID); errMD == nil {
		if errRM := os.Remove(mdPath); errRM != nil && !os.IsNotExist(errRM) {
			return fmt.Errorf("deleting task plan markdown: %w", errRM)
		}
	}

	path, err := store.GetTaskPath(t.ID)
	if err != nil {
		return err
	}
	return os.Remove(path)
}

// DeletePR deletes a pull request and its markdown file.
func (store *Store) DeletePR(idOrSlug string) error {
	pr, err := store.GetPullRequest(idOrSlug)
	if err != nil {
		return err
	}

	if err := store.UnlinkEntity(pr.ID); err != nil {
		return fmt.Errorf("unlinking pull request: %w", err)
	}

	// Delete the PR markdown sidecar if it exists
	if mdPath, errMD := store.GetPullRequestMarkdownPath(pr.ID); errMD == nil {
		if errRM := os.Remove(mdPath); errRM != nil && !os.IsNotExist(errRM) {
			return fmt.Errorf("deleting pull request markdown: %w", errRM)
		}
	}

	path, err := store.GetPullRequestPath(pr.ID)
	if err != nil {
		return err
	}
	return os.Remove(path)
}

func DeleteMilestone(idOrSlug string) error {
	return defaultStore.DeleteMilestone(idOrSlug)
}

func DeleteTask(idOrSlug string) error {
	return defaultStore.DeleteTask(idOrSlug)
}

func DeletePR(idOrSlug string) error {
	return defaultStore.DeletePR(idOrSlug)
}
