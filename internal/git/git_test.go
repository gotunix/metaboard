// SPDX-License-Identifier: GPL-3.0-or-later
// SPDX-FileCopyrightText: 2026 The MetaBoard authors

package git

import (
	"testing"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
)

func TestSSHUser(t *testing.T) {
	tmp := t.TempDir()

	// Initialize a dummy repo
	repo, err := gogit.PlainInit(tmp, false)
	if err != nil {
		t.Fatalf("failed to init temp git repo: %v", err)
	}

	tests := []struct {
		url  string
		want string
	}{
		{"git@github.com:org/repo.git", "git"},
		{"ssh://gitea@git.example.com:2222/org/repo.git", "gitea"},
		{"ssh://git.example.com/org/repo.git", "git"},
		{"https://github.com/org/repo.git", "git"},
		{"ssh://custom-user@host.name:port/repo", "custom-user"},
	}

	for _, tt := range tests {
		// Set the origin remote URL
		_, err := repo.CreateRemote(&config.RemoteConfig{
			Name: "origin",
			URLs: []string{tt.url},
		})
		if err != nil && err != gogit.ErrRemoteExists {
			t.Fatalf("failed to create remote: %v", err)
		}
		if err == gogit.ErrRemoteExists {
			// Update the URL if it already existed
			err = repo.DeleteRemote("origin")
			if err != nil {
				t.Fatalf("failed to delete remote: %v", err)
			}
			_, err = repo.CreateRemote(&config.RemoteConfig{
				Name: "origin",
				URLs: []string{tt.url},
			})
			if err != nil {
				t.Fatalf("failed to re-create remote: %v", err)
			}
		}

		got := sshUser(repo)
		if got != tt.want {
			t.Errorf("sshUser(%q) = %q; want %q", tt.url, got, tt.want)
		}
	}
}

func TestIsSSHRemote(t *testing.T) {
	tmp := t.TempDir()
	repo, err := gogit.PlainInit(tmp, false)
	if err != nil {
		t.Fatalf("failed to init: %v", err)
	}

	tests := []struct {
		url  string
		want bool
	}{
		{"git@github.com:org/repo.git", true},
		{"ssh://gitea@git.example.com:2222/org/repo.git", true},
		{"https://github.com/org/repo.git", false},
		{"http://github.com/org/repo.git", false},
	}

	for _, tt := range tests {
		_ = repo.DeleteRemote("origin")
		_, _ = repo.CreateRemote(&config.RemoteConfig{
			Name: "origin",
			URLs: []string{tt.url},
		})

		got := isSSHRemote(repo)
		if got != tt.want {
			t.Errorf("isSSHRemote(%q) = %t; want %t", tt.url, got, tt.want)
		}
	}
}
