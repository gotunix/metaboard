# metaboard — AI Agent Usage Guide

A git-first project metaboard for tracking milestones, tasks, and pull requests directly alongside your code. Designed for AI agents to manage work programmatically via CLI.

## Quick Start

```bash
# Initialize in your project
metaboard init

# Or initialize in a specific directory
metaboard init .boards

# Create a milestone
metaboard milestone create --title "User Authentication" --description "Implement auth system" --status ACTIVE

# Create tasks under the milestone
metaboard task create --title "Design auth schema" --priority HIGH --type FEATURE --milestone user-authentication
metaboard task create --title "Implement JWT tokens" --priority HIGH --type FEATURE --milestone user-authentication
metaboard task create --title "Add login endpoint" --priority MEDIUM --type FEATURE --milestone user-authentication

# Link a PR to a task
metaboard pullrequest create --source-branch feat/login --dest-branch main --description "Add login API"
metaboard link pr-abc123 task-xyz789

# View progress
metaboard dashboard
```

## Core Concepts

| Entity | Description | Status Values |
|--------|-------------|---------------|
| **Milestone** | Top-level project phase/goal | BACKLOG, ACTIVE, COMPLETED, CANCELLED |
| **Task** | Work item under a milestone | BACKLOG, ACTIVE, IN-PROGRESS, COMPLETED, CANCELLED |
| **Pull Request** | Code change linked to tasks | DRAFT, OPEN, MERGED, CLOSED, REJECTED |

**Hierarchy**: `Milestone → Task` (direct). Tasks can also exist unlinked in the "Backlog" section.

## CLI Reference

### Milestones

```bash
# Create
metaboard milestone create --title "API v2" --description "Rewrite REST API" --status ACTIVE --slug api-v2

# List all
metaboard milestone list

# View details (shows linked tasks)
metaboard milestone view api-v2

# View specific version
metaboard milestone view api-v2 --version 3

# Edit
metaboard milestone edit api-v2 --title "API v2.1" --status COMPLETED

# Update status only
metaboard milestone status api-v2 COMPLETED

# History
metaboard milestone history api-v2

# Delete
metaboard milestone delete api-v2
```

### Tasks

```bash
# Create (link to milestone with --milestone)
metaboard task create \
  --title "Add rate limiting" \
  --priority HIGH \
  --type FEATURE \
  --assigned-to "backend-team" \
  --description "Implement token bucket rate limiter" \
  --tags "security,api" \
  --depends "task-001,task-002" \
  --changelog \
  --milestone api-v2 \
  --slug rate-limit

# List all tasks
metaboard task list

# View details
metaboard task view rate-limit

# Edit
metaboard task edit rate-limit --priority CRITICAL --status IN-PROGRESS

# Update status
metaboard task status rate-limit COMPLETED

# History
metaboard task history rate-limit

# Open implementation plan in editor
metaboard task plan rate-limit

# Delete
metaboard task delete rate-limit
```

### Pull Requests

```bash
# Create (opens markdown template in $EDITOR)
metaboard pullrequest create \
  --source-branch feat/rate-limit \
  --dest-branch main \
  --source-repo github.com/org/repo \
  --dest-repo github.com/org/repo \
  --description "Add rate limiting middleware"

# Aliases: `metaboard pr ...`
metaboard pr list
metaboard pr view pr-xyz
metaboard pr edit pr-xyz --status MERGED
metaboard pr history pr-xyz
metaboard pr delete pr-xyz
```

### Linking

```bash
# Link task to milestone
metaboard link task-slug milestone-slug

# Link PR to task
metaboard link pr-slug task-slug

# Unlink
metaboard unlink task-slug
```

### Other Commands

```bash
# Initialize new board
metaboard init [path]          # defaults to ./metadata

# Generate CHANGELOG.md from completed milestones/tasks
metaboard changelog [output_dir]  # defaults to .

# TUI Dashboard (interactive)
metaboard dashboard [milestone_slug]
metaboard dashboard all        # show all statuses
metaboard dashboard closed     # completed/merged
metaboard dashboard cancelled  # cancelled items

# Version info
metaboard version
```

## TUI Dashboard (Interactive)

Launch with `metaboard` or `metaboard dashboard`.

### Navigation
| Key | Action |
|-----|--------|
| `j` / `k` / `↑` / `↓` | Move cursor |
| `Enter` | View selected item |
| `e` | Edit selected item |
| `m` / `t` / `p` | Create Milestone / Task / PR |
| `u` | Unlink selected item |
| `d` / `x` | Delete selected item |
| `c` | View/regenerate changelog |
| `v` | View version info |
| `g` / `f` | Git fetch / push |
| `q` / `Esc` | Quit / back |

### Forms (Create/Edit)
- `Tab` / `Shift+Tab` — Next/previous field
- `←` / `→` — Change select/priority/type fields
- `Space` — Toggle multi-select (tasks, dependencies)
- `Enter` — Submit form
- `Esc` — Cancel

## AI Agent Workflow Patterns

### 1. Starting a New Feature
```bash
# Create milestone for the feature
metaboard milestone create --title "Payment Integration" --status ACTIVE

# Break down into tasks
metaboard task create --title "Research providers" --priority HIGH --milestone payment-integration
metaboard task create --title "Implement Stripe checkout" --priority HIGH --milestone payment-integration
metaboard task create --title "Add webhook handler" --priority MEDIUM --milestone payment-integration
metaboard task create --title "Write tests" --priority MEDIUM --type CHORE --milestone payment-integration
```

### 2. Tracking Progress
```bash
# Quick status check
metaboard milestone list
metaboard task list

# View specific milestone with tasks
metaboard milestone view payment-integration

# Update as you complete work
metaboard task status task-slug COMPLETED
metaboard task status task-slug IN-PROGRESS
```

### 3. Code Review Flow
```bash
# When PR is opened
metaboard pr create --source-branch feat/stripe --dest-branch main --description "Stripe integration"
metaboard link pr-abc task-xyz

# After review
metaboard pr edit pr-abc --status MERGED
metaboard task status task-xyz COMPLETED
```

### 4. Generating Release Notes
```bash
# Generates CHANGELOG.md from completed milestones with changelog=true tasks
metaboard changelog
```

## Data Storage

- Files stored under `./metadata/` (or custom `--data-dir`)
- Structure:
  ```
  metadata/
  ├── milestones/xx/uuid.json    # versioned history
  ├── tasks/xx/uuid.json
  ├── pullrequests/xx/uuid.json
  └── pullrequests/xx/uuid.md    # PR markdown template
  ```
- All changes auto-committed to git when run in a git repo
- Human-readable JSON with full version history

## Tips for AI Agents

1. **Use slugs for stable references** — Slugs don't change unless explicitly edited
2. **Prefer `--milestone` flag when creating tasks** — Avoids separate link step
3. **Check `metaboard task list` before creating** — Avoid duplicates
4. **Use `--changelog` flag on tasks** — Auto-includes in `metaboard changelog`
5. **Run `metaboard dashboard` for visual overview** — Better for complex state
6. **All commands respect `--data-dir`** — Works with multiple boards

## Example: Complete Feature Lifecycle

```bash
# 1. Initialize
metaboard init

# 2. Plan milestone
metaboard milestone create --title "Search Feature" --slug search-feature --status ACTIVE

# 3. Create tasks
metaboard task create --title "Design search schema" --priority HIGH --type FEATURE --milestone search-feature --slug design-schema
metaboard task create --title "Implement Elasticsearch index" --priority HIGH --type FEATURE --milestone search-feature --slug es-index
metaboard task create --title "Add search API endpoint" --priority HIGH --type FEATURE --milestone search-feature --slug search-api
metaboard task create --title "Add autocomplete" --priority MEDIUM --type FEATURE --milestone search-feature --slug autocomplete
metaboard task create --title "Write integration tests" --priority MEDIUM --type CHORE --milestone search-feature --slug search-tests

# 4. Work on tasks (update status as you go)
metaboard task status design-schema COMPLETED
metaboard task status es-index IN-PROGRESS
# ... work ...
metaboard task status es-index COMPLETED
metaboard task status search-api IN-PROGRESS
# ... work ...
metaboard task status search-api COMPLETED
metaboard task status autocomplete COMPLETED
metaboard task status search-tests COMPLETED

# 5. Create PR for the feature
metaboard pr create --source-branch feat/search --dest-branch main --description "Full-text search with autocomplete"
metaboard link pr-search task-es-index
metaboard link pr-search task-search-api
metaboard link pr-search task-autocomplete
metaboard link pr-search task-search-tests

# 6. After merge
metaboard pr edit pr-search --status MERGED
metaboard milestone status search-feature COMPLETED

# 7. Generate release notes
metaboard changelog
```

## Global Flags

| Flag | Description |
|------|-------------|
| `-d, --data-dir` | Base directory for metaboard data |
| `-v, --version` | Show version information |
| `-h, --help` | Show help |

All commands inherit these flags.