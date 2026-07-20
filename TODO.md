# Metaboard Reminders & TODOs

These are the remaining medium and low priority tasks from the codebase review, ready to be picked up in a future session.

## 🟡 Medium Priority

- [x] **Simplify `pushBranch` Dialog Input** — replaced with `textinput.Model`
- [ ] **Enhance `Makefile` Build & Test Targets** ([Makefile](Makefile)):
  - Add the `-race` detector flag to the `test` target.
  - Inject the app version and git commit hash dynamically into the build target using linker flags.
- [x] **Optimize `SortTasks` Performance** — priority map lifted out of sort comparator

---

## 🟢 Low Priority

- [x] **Resolve License Header Inconsistencies** — standardized on GPL-3.0-or-later
- [ ] **Fix CHANGELOG Version Ordering** ([CHANGELOG.md](CHANGELOG.md)):
  - Correct the version headers in the changelog file to ensure releases are sorted in strict chronological/semantic ordering.
