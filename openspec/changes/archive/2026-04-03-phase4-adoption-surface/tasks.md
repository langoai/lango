## 1. README Restructure

- [x] 1.1 Change lead tagline from "sovereign AI agent runtime with built-in commerce" to "trustworthy multi-agent runtime in Go"
- [x] 1.2 Move early-stage warning note to after badge row with feature status table link
- [x] 1.3 Remove old "⚠️ Note" warning section
- [x] 1.4 Reorder "Why Lango?" bullets: trust/orchestration/observability first, economy/P2P second
- [x] 1.5 Condense inline CLI reference (~180 lines) to 8 key commands with link to docs/cli/index.md

## 2. Installation Documentation

- [x] 2.1 Add platform-specific C compiler setup section to docs/getting-started/installation.md (macOS, Ubuntu, Fedora, Alpine)
- [x] 2.2 Add `go install` vs `make build` difference note with build tags explanation
- [x] 2.3 Strengthen quickstart.md cross-link to installation prerequisites

## 3. TUI Experimental Badges

- [x] 3.1 Add `BadgeExperimentalStyle` to internal/cli/tui/styles.go following existing badge pattern
- [x] 3.2 Add `ExperimentalCategories` map to internal/cli/settings/menu.go with all experimental category IDs
- [x] 3.3 Add `expBadge` rendering in `renderItem()` after ADV badge and before dependency badge
- [x] 3.4 Add `@experimental` case in `applyFilter()` switch block
- [x] 3.5 Add `@experimental` to filterHint display string
- [x] 3.6 Add drift-prevention test in menu_test.go verifying ExperimentalCategories set

## 4. Roadmap Documentation

- [x] 4.1 Add Execution Progress table to docs/development/roadmap.md showing Phase 1-4 status
- [x] 4.2 Mark completed items in Impact-Effort Backlog with status column
- [x] 4.3 Update Future Strategy Tracks status to reflect Phase 1-3 completion

## 5. Verification

- [x] 5.1 Verify `go build ./...` succeeds
- [x] 5.2 Verify `go test ./internal/cli/settings/... ./internal/cli/tui/...` passes
