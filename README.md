# Metaboard Engine

A high-performance, professional CLI engine for git-first project
management. This tool keeps your tasks, stories, milestones, and
technical implementation plans versioned directly alongside your code
in simple JSON and Markdown files.

## What is this for?

The Metaboard Engine is designed for developers who want to manage
their project lifecycle without leaving the terminal or relying on
external heavy-weight platforms. By storing project data as sharded
files in your repository, you get:
- **Full Version Control**: Track how your plan and tasks evolve
  alongside your source code.
- **Offline First**: No internet connection required for project
  management.
- **Developer-Centric**: Highly scanable data structures and a
  responsive TUI.

## Key Features

- **Responsive Hierarchical Dashboard**: A polished tree view of your
  project status that automatically scales to your terminal width.
- **Dual-Mode Editing**:
    - **Interactive TUI**: High-polish terminal forms (using `huh`)
      for rich data entry.
    - **Power-User Flags**: Update any field (status, priority,
      description, etc.) directly from the CLI for speed.
- **Unified Task-Plan Sidecars**: Technical implementation plans live
  as Markdown files (`.md`) directly alongside your task data
  (`.json`), managed by a unified workflow.
- **Natural Alphanumeric Sorting**: Intelligent ordering of slugs
  (e.g., `sv-1` comes before `sv-2`, and `sv-9` before `sv-10`).
- **Advanced Reporting**: Integrated `backlog` (unlinked entities)
  and `changelog` (chronological release notes) generators.
- **Professional Architecture**: Built with a modular Go structure for
  maximum performance and maintainability.

## Installation

### Prerequisites
- Go 1.21 or higher

### Build & Test
From the project directory:
```bash
make build    # Build the metaboard binary
make test     # Run all tests with coverage
```

## Usage

### General Commands
```bash
./metaboard dashboard           # Show the full project tree
./metaboard dashboard <slug>    # Focus on a specific milestone
./metaboard backlog             # Show unlinked tasks and stories
./metaboard changelog           # Generate CHANGELOG.md
```

### Task Management
```bash
# Create a task with flags
./metaboard task create --title "New Feature" --slug "feat-1" --priority "HIGH"

# Create/Edit implementation plan (opens in $EDITOR)
./metaboard task plan feat-1

# View task details with inline plan
./metaboard task view feat-1

# Interactive edit
./metaboard task edit feat-1

# Quick status update
./metaboard task status feat-1 CLOSED
```

## Directory Structure

The tool expects the following folders in its current working
directory (CWD):
- `milestones/`: Milestone JSON files
- `stories/`: Story JSON files
- `tasks/`: Sharded task JSON and Markdown sidecar files
  (e.g., `tasks/ab/uuid.json`)
- `plans/`: Standalone implementation plans

## License

SPDX-License-Identifier: GPL-3.0-or-later
See `HEADER` for copyright details.
