# Changelog

All notable changes to this project will be documented in this file.

## [0.1.0] - 2026-06-11

### Added
- Comprehensive test suite for `internal/models` and `internal/store`.
- `make test` target in Makefile for easy test execution.
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
