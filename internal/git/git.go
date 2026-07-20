// SPDX-License-Identifier: GPL-3.0-or-later
// SPDX-FileCopyrightText: 2026 The MetaBoard authors

package git

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/skeema/knownhosts"
	gossh "golang.org/x/crypto/ssh"
	sshknownhosts "golang.org/x/crypto/ssh/knownhosts"
)

// ErrAlreadyUpToDate is returned when no new data was fetched or pushed.
var ErrAlreadyUpToDate = gogit.NoErrAlreadyUpToDate

// IsGitRepo reports whether dataDir (or any ancestor) is inside a git repository.
func IsGitRepo(dataDir string) bool {
	abs, err := filepath.Abs(dataDir)
	if err != nil {
		return false
	}
	_, err = gogit.PlainOpenWithOptions(abs, &gogit.PlainOpenOptions{DetectDotGit: true})
	return err == nil
}

// openRepo opens the git repository that contains dataDir by walking up to
// find the repo root. go-git opens from the worktree root, not a subdirectory.
func openRepo(dataDir string) (*gogit.Repository, error) {
	abs, err := filepath.Abs(dataDir)
	if err != nil {
		return nil, fmt.Errorf("resolving data dir: %w", err)
	}
	repo, err := gogit.PlainOpenWithOptions(abs, &gogit.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		return nil, fmt.Errorf("opening git repo at %q: %w", abs, err)
	}
	return repo, nil
}

// isSSHRemote returns true when origin uses an SSH URL (git@ or ssh://).
func isSSHRemote(repo *gogit.Repository) bool {
	remote, err := repo.Remote("origin")
	if err != nil {
		return false
	}
	for _, u := range remote.Config().URLs {
		if strings.HasPrefix(u, "git@") || strings.HasPrefix(u, "ssh://") {
			return true
		}
	}
	return false
}

// sshUser extracts the username from the repository's origin remote URL,
// falling back to "git".
func sshUser(repo *gogit.Repository) string {
	remote, err := repo.Remote("origin")
	if err != nil {
		return "git"
	}
	for _, u := range remote.Config().URLs {
		if strings.HasPrefix(u, "ssh://") {
			trimmed := strings.TrimPrefix(u, "ssh://")
			if idx := strings.Index(trimmed, "@"); idx != -1 {
				return trimmed[:idx]
			}
		} else if idx := strings.Index(u, "@"); idx != -1 {
			if colIdx := strings.Index(u, ":"); colIdx == -1 || idx < colIdx {
				return u[:idx]
			}
		}
	}
	return "git"
}

// hostKeyCallback returns a HostKeyCallback that implements TOFU (Trust On First Use):
//   - validates remote host keys against ~/.ssh/known_hosts
//   - if host exists but key differs, rejects connection (blocks potential MITM attack)
//   - if host is missing, appends the new key to ~/.ssh/known_hosts and allows connection
func hostKeyCallback() gossh.HostKeyCallback {
	home, err := os.UserHomeDir()
	if err != nil {
		return gossh.InsecureIgnoreHostKey()
	}
	khPath := filepath.Join(home, ".ssh", "known_hosts")

	return func(hostname string, remote net.Addr, key gossh.PublicKey) error {
		kh, _ := knownhosts.New(khPath)

		var keyErr *sshknownhosts.KeyError
		if kh != nil {
			err = kh(hostname, remote, key)
			if err == nil {
				return nil // Key matches
			}
			if !errors.As(err, &keyErr) {
				return err // Generic file or parsing error
			}
		} else {
			keyErr = &sshknownhosts.KeyError{Want: nil}
		}

		if len(keyErr.Want) > 0 {
			// Host is known but key has changed! MITM risk!
			return fmt.Errorf("git: host key mismatch for %s. Potential Man-in-the-Middle attack!", hostname)
		}

		// TOFU: Host is not known. Save key to known_hosts
		normalizedHost := knownhosts.Normalize(hostname)
		pubKeyBytes := gossh.MarshalAuthorizedKey(key)
		line := fmt.Sprintf("%s %s", normalizedHost, string(pubKeyBytes))

		_ = os.MkdirAll(filepath.Dir(khPath), 0700)
		f, err := os.OpenFile(khPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return nil // Fallback to allowing connection if unable to write
		}
		defer f.Close()

		if _, err := f.WriteString(line); err != nil {
			return nil
		}

		return nil
	}
}

// sshAuth builds SSH authentication using the system ssh-agent (preferred),
// falling back to common private-key files in ~/.ssh.
// Returns nil, nil when the remote is not SSH.
func sshAuth(repo *gogit.Repository) (transport.AuthMethod, error) {
	if !isSSHRemote(repo) {
		return nil, nil
	}

	cb := hostKeyCallback()
	user := sshUser(repo)

	// Try the ssh-agent first — matches system ssh behaviour and works
	// with encrypted keys loaded into the agent.
	if agentAuth, err := gitssh.NewSSHAgentAuth(user); err == nil {
		agentAuth.HostKeyCallback = cb
		return agentAuth, nil
	}

	// Fall back to well-known private-key files.
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolving home directory: %w", err)
	}

	for _, keyFile := range []string{"id_ed25519", "id_ecdsa", "id_rsa"} {
		keyPath := filepath.Join(home, ".ssh", keyFile)
		if _, err := os.Stat(keyPath); err != nil {
			continue
		}
		pk, err := gitssh.NewPublicKeysFromFile(user, keyPath, "")
		if err != nil {
			continue
		}
		pk.HostKeyCallback = cb
		return pk, nil
	}

	return nil, fmt.Errorf("no SSH auth available for user %q: ssh-agent not running and no key files found in ~/.ssh", user)
}

// resolveAuthor reads the name/email from the repo's local or global git config.
// Falls back to sensible defaults if nothing is configured.
func resolveAuthor(repo *gogit.Repository) *object.Signature {
	cfg, err := repo.ConfigScoped(0) // 0 = worktree (local) scope
	name := "MetaBoard"
	email := "metaboard@localhost"
	if err == nil {
		if cfg.User.Name != "" {
			name = cfg.User.Name
		}
		if cfg.User.Email != "" {
			email = cfg.User.Email
		}
	}
	return &object.Signature{
		Name:  name,
		Email: email,
		When:  time.Now(),
	}
}

// Fetch downloads objects and refs from origin without merging.
func Fetch(dataDir string) error {
	repo, err := openRepo(dataDir)
	if err != nil {
		return fmt.Errorf("git fetch failed: %w", err)
	}

	auth, _ := sshAuth(repo)

	fetchErr := repo.Fetch(&gogit.FetchOptions{
		RemoteName: "origin",
		Auth:       auth,
	})
	if fetchErr != nil {
		if errors.Is(fetchErr, gogit.NoErrAlreadyUpToDate) {
			return nil
		}
		return fmt.Errorf("git fetch failed: %w", fetchErr)
	}

	return nil
}

// Pull fetches and merges the upstream branch into the current branch.
// Note: go-git v5 does not support rebase; this is a merge-based pull.
func Pull(dataDir string) error {
	repo, err := openRepo(dataDir)
	if err != nil {
		return fmt.Errorf("git pull failed: %w", err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("git pull failed (worktree): %w", err)
	}

	auth, _ := sshAuth(repo)

	pullErr := wt.Pull(&gogit.PullOptions{
		RemoteName: "origin",
		Auth:       auth,
		Force:      false,
	})
	if pullErr != nil {
		if errors.Is(pullErr, gogit.NoErrAlreadyUpToDate) {
			return nil
		}
		return fmt.Errorf("git pull failed: %w", pullErr)
	}
	return nil
}

// stageFiles resolves file paths relative to the repo root and runs
// "git add" for them. If files is []string{"."} it stages everything.
// Returns the absolute repo root so the caller can reuse it.
func stageFiles(dataDir string, files []string) (repoRoot string, err error) {
	repo, err := openRepo(dataDir)
	if err != nil {
		return "", fmt.Errorf("opening repo: %w", err)
	}
	wt, err := repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("getting worktree: %w", err)
	}
	root := wt.Filesystem.Root()
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolving repo root: %w", err)
	}
	repoRoot = absRoot

	addArgs := []string{"add"}
	if len(files) == 1 && files[0] == "." {
		addArgs = append(addArgs, ".")
	} else {
		for _, f := range files {
			var absFile string
			if filepath.IsAbs(f) {
				absFile = f
			} else {
				absFile = filepath.Join(repoRoot, f)
			}
			rel, err := filepath.Rel(repoRoot, absFile)
			if err != nil {
				return "", fmt.Errorf("git add failed: %w", err)
			}
			addArgs = append(addArgs, rel)
		}
	}

	addCmd := exec.Command("git", addArgs...)
	addCmd.Dir = repoRoot
	addOut, err := addCmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(addOut))
		return "", fmt.Errorf("git add %v failed: %s", addArgs[1:], msg)
	}
	return repoRoot, nil
}

// Stage stages the given file paths (relative to dataDir or absolute) without
// committing. Useful for tracking newly created files immediately so a later
// Commit call can include them even after an external process (e.g. an editor)
// has run between the two calls.
// If files contains "." all changes in the worktree are staged.
func Stage(dataDir string, files []string) error {
	_, err := stageFiles(dataDir, files)
	return err
}

// Commit stages the given file paths relative to dataDir and commits them with
// message. If files contains "." all changes in the worktree are staged.
// Returns nil if there are no staged changes (nothing to commit).
// Use Push separately to send commits to the remote.
func Commit(dataDir string, files []string, message string) error {
	repoRoot, err := stageFiles(dataDir, files)
	if err != nil {
		return fmt.Errorf("git commit failed: %w", err)
	}

	commitCmd := exec.Command("git", "commit", "-m", message)
	commitCmd.Dir = repoRoot
	commitOut, err := commitCmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(commitOut))
		if strings.Contains(msg, "nothing to commit") {
			return nil
		}
		return fmt.Errorf("git commit failed: %s", msg)
	}

	return nil
}

// CurrentBranch returns the name of the currently checked-out branch.
// Returns an error if HEAD is detached.
func CurrentBranch(dataDir string) (string, error) {
	repo, err := openRepo(dataDir)
	if err != nil {
		return "", fmt.Errorf("getting current branch: %w", err)
	}
	ref, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("getting HEAD: %w", err)
	}
	if !ref.Name().IsBranch() {
		return "", fmt.Errorf("HEAD is detached (%s)", ref.Hash())
	}
	return ref.Name().Short(), nil
}

// Push sends local commits to origin on the specified branch.
// Uses the git CLI to support git worktrees (go-git's worktree support
// is incomplete and can't resolve remotes in worktree contexts).
func Push(dataDir string, branch string) error {
	if branch == "" {
		var err error
		branch, err = CurrentBranch(dataDir)
		if err != nil {
			return fmt.Errorf("determining current branch: %w", err)
		}
	}
	cmd := exec.Command("git", "push", "origin", branch)
	cmd.Dir = dataDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := string(out)
		if strings.Contains(msg, "Everything up-to-date") ||
			strings.Contains(msg, "up to date") {
			return nil
		}
		return fmt.Errorf("git push failed: %s", msg)
	}
	return nil
}
