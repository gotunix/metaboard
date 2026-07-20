# Metaboard

A TUI-based project management tool that keeps tasks, stories,
milestones, and pull requests versioned directly alongside your code
in simple JSON and Markdown files.

## What is this for?

Metaboard is designed for developers who want to manage their project
lifecycle without leaving the terminal or relying on external
heavy-weight platforms. By storing project data as sharded files in
your repository, you get:
- **Full Version Control**: Track how your plan and tasks evolve
  alongside your source code with automatic git commits.
- **Offline First**: No internet connection required for project
  management.
- **Developer-Centric**: Highly scanable data structures and a
  responsive TUI dashboard.

## Key Features

- **Interactive TUI Dashboard**: A polished tree view of milestones,
  stories, tasks, and pull requests with keyboard navigation, inline
  editing, and progress bars.
- **Batch Create/Edit Forms**: Full CRUD for all entities through
  responsive terminal forms with scrollable layouts.
- **Git Integration**: Automatic commits on changes, with fetch and
  push support using SSH agent auth (encrypted keys supported).
- **Change Log Viewer**: Built-in changelog generation and display
  from task metadata.
- **Task Plan Sidecars**: Markdown files (`.md`) alongside task JSON
  for implementation plans and additional details.
- **Init & Push Wizards**: First-run init dialog and push-to-remote
  prompt with configurable branch name.

## Installation

### Prerequisites
- Go 1.21 or higher

### Build
From the project directory:
```bash
make build    # Build the metaboard binary
```

## Usage

```bash
./metaboard    # Launch the TUI dashboard (auto-inits on first run)
```

Keyboard shortcuts from the dashboard:
- `↑/↓/j/k` — Navigate the tree
- `Space` — Expand/collapse items
- `Enter` — View item details
- `t/s/m/p` — Create task/story/milestone/PR
- `e` — Edit selected item
- `d/x` — Delete selected item
- `c` — View changelog
- `v` — View version info
- `f` — Git fetch
- `g` — Git push dialog

## Directory Structure

The tool creates the following folders in the project root or
`./metadata/`:
- `milestones/` — Milestone JSON files
- `stories/` — Story JSON files
- `tasks/` — Sharded task JSON and Markdown sidecar files
  (e.g., `tasks/ab/uuid.json`)
- `pullrequests/` — Pull request JSON and Markdown files

## License

SPDX-License-Identifier: GPL-3.0-or-later
See `HEADER` for copyright details.
