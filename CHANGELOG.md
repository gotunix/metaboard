# Changelog

All notable changes to this project will be documented in this file.

## [0.4.1] - 2026-07-18

### Fixed
- Fixed CLI commands (`task`, `milestone`, `pullrequest`, `link`) failing to stage and commit data files due to incorrect path handling. Deletions via CLI now also correctly stage the deleted files.

## [0.3.1] - 2026-07-05

### Added
- Form scrolling (PgDn/PgUp) for tall create/edit forms with auto-scroll
  on focus change.
- Terminal resize responsiveness in all states (forms, detail views,
  prompts, confirm dialogs).
- Scroll position percentage indicator in detail view help bar.

### Fixed
- SSH auth: known_hosts key mismatch no longer blocks git operations;
  host key warnings are silently accepted (matches system SSH ergonomics).
- SSH agent now always tried first, supporting encrypted keys loaded
  into the agent without passphrase prompts.

### Changed
- Detail view scroll step increased from 2 to 3 lines per keypress.
- Help text now references PgDn/PgUp alongside j/k for scrolling.

## [0.3.3] - 2026-07-05

### Added
- PR detail view: linked Stories/Tasks section showing parent story
  resolved from linked task IDs.
- PR edit form: cursor resets to start of value when entering a field.
- `formLastFocusIndex` tracking to detect focus changes.
- `catMochaSurface0` (#363a4f) color for lighter text field backgrounds.

### Changed
- Form text fields: dark grey (#595959) background with white text for
  clear visual distinction from the terminal background; prompt and
  textarea fields styled consistently.
- Submit button: green background in both focused and unfocused states
  (active state uses bold for emphasis).
- PR detail view labels: "Head:" → "Source Branch:", "Base:" →
  "Destination Branch:", "Source Repositories:" → "Repositories:".
- Form field width minimum raised from 40 to 50 to accommodate repo
  URLs without truncation.
- Textarea text color changed to white for readability on dark bg.

### Fixed
- Arrow keys, Home, End now work correctly in form text fields (cursor
  was being reset to 0 on every render cycle via `updateFormFieldsFocus`
  calling `SetCursor(0)` unconditionally; now only resets on actual
  focus change).
- Background ANSI artifact removed from text input fields (grey overlay
  fieldBg wrapper conflicted with internal textinput ANSI reset codes;
  background is now set directly on PromptStyle/TextStyle).

## [0.3.2] - 2026-07-05

### Added
- Status messages displayed below the help footer for all form save
  operations, fetch, push, and plan edits.
- Plan editor now commits the plan file automatically after closing
  the editor with message `"boards: plan created for task <slug>"`.
- Form text fields now scale dynamically with terminal width (40–100
  for inputs, 50–120 for textareas).

### Fixed
- Git operations (fetch, push, commit) no longer print to stdout,
  preventing TUI corruption from async goroutine output.
- Fetch and push now use `tea.Cmd` with result messages for proper
  status feedback instead of fire-and-forget goroutines.

### Changed
- Status bar moved below the help footer, styled with the same
  mauve background as the footer.

## [0.1.1] - 2026-07-03

### Added
- Dashboard `v` hotkey to show version and dependency details in the same
  styled detail view.
- CLI `-v` flag for printing build and version metadata.
- Scrollable dependency list support in the version detail view.

## [0.1.0] - 2026-06-11

### Added
- Comprehensive test suite for `internal/models` and `internal/store`.
- `make test` target in Makefile for easy test execution.
- `make install` and `make uninstall` targets with `PREFIX` support.
- Project ASCII logo and custom license headers to all source files.
- `internal/store` helpers: `GetMilestone`, `GetStory`, `GetTask`, and `EnsureTaskPlan`.
- `TaskUpdate` struct for centralized task field mapping.

### Changed
- Rebranded project from `metadata` to `metaboard`.
- Updated Go module path to `gotunix.net/metaboard`.
- Refactored business logic from `cmd/` layer to `internal/store`.
- Updated command handlers to be thin wrappers for CLI parsing.
- Standardized help and usage outputs using custom `ui` handlers.
- Simplified entity creation with atomic functions in `internal/store`.

### Fixed
- Improved slug generation logic to handle numeric sorting correctly.
- Enhanced status update logic with automatic completion timestamps.
