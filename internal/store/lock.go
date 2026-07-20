// SPDX-License-Identifier: GPL-3.0-or-later
// SPDX-FileCopyrightText: 2026 The MetaBoard authors

package store

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

// isProcessAlive checks if a process with the given PID is currently running.
func isProcessAlive(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// Signal 0 checks for process existence without sending a signal.
	err = process.Signal(syscall.Signal(0))
	if err == nil {
		return true // Process is running
	}
	// If the error is ErrProcessDone or ESRCH (no such process), the process is dead.
	if errors.Is(err, os.ErrProcessDone) || errors.Is(err, syscall.ESRCH) {
		return false
	}
	// If we get permission denied, the process exists but is owned by another user.
	return true
}

// AcquireLock attempts to create a lock file. It blocks/retries for a short period.
// If the lock file exists but contains a dead PID, it clears the stale lock.
// It returns an unlock function and any error encountered.
func (store *Store) AcquireLock() (func(), error) {
	root, err := store.GetDataRoot()
	if err != nil {
		return nil, err
	}
	lockPath := filepath.Join(root, ".metaboard.lock")

	timeout := time.After(2 * time.Second)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return nil, errors.New("timeout: database is locked by another process")
		case <-ticker.C:
			// O_CREATE | O_EXCL ensures the file is created ONLY if it does not already exist
			file, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
			if err == nil {
				// Write current PID to the lock file
				_, _ = fmt.Fprintf(file, "%d\n", os.Getpid())
				_ = file.Close()
				return func() { _ = os.Remove(lockPath) }, nil
			}

			// If the file exists, check if it is stale
			if os.IsExist(err) {
				content, readErr := os.ReadFile(lockPath)
				if readErr == nil {
					var pid int
					_, scanErr := fmt.Sscanf(string(content), "%d", &pid)
					if scanErr != nil || pid <= 0 || !isProcessAlive(pid) {
						// Lock file is corrupted, empty, or process is dead -> remove stale lock
						_ = os.Remove(lockPath)
					}
				} else if os.IsNotExist(readErr) {
					// File was removed by the owner in the meantime, that's fine
				}
			}
		}
	}
}

// AcquireLock wrapper for backward compatibility
func AcquireLock() (func(), error) {
	return defaultStore.AcquireLock()
}
