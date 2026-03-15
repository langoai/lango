## Why

Settings TUI has hidden dependency chains between features that are completely invisible to users. For example, enabling Smart Account silently fails to register tools when Payment + Security Signer are not configured. Even project developers cannot diagnose the root cause. This creates a critical UX gap that must be addressed with a 3-layer dependency discovery system.

## What Changes

- Add a dependency registry that codifies all 20+ feature dependency relationships with check functions, transitive resolution, and reverse lookup
- Display warning badges (`⚠ N`) on menu categories with unmet dependencies
- Add `@ready` smart filter to show only configurable (unblocked) categories
- Show a prerequisite panel when entering a form with unmet dependencies, with jump-to-dependency navigation
- Provide a guided setup flow (wizard) that chains prerequisite forms before the target form
- Refactor `handleMenuSelection()` to use a shared `createFormForCategory()` factory, eliminating 150+ lines of duplicated switch logic

## Capabilities

### New Capabilities
- `settings-dependency-discovery`: 3-layer dependency discovery system for Settings TUI — dependency registry, menu warning badges with @ready filter, prerequisite panel with jump navigation, and guided setup flow wizard

### Modified Capabilities

## Impact

- `internal/cli/settings/editor.go` — New fields (depIndex, depPanel, navStack, setupFlow), panel/setup flow integration in Update/View, refactored handleMenuSelection
- `internal/cli/settings/menu.go` — DependencyChecker callback, @ready filter, badge rendering in renderItem
- `internal/cli/tui/styles.go` — BadgeDependencyStyle added
- `internal/cli/settings/dependencies.go` — New: DependencyIndex, dependency graph definitions
- `internal/cli/settings/dependency_panel.go` — New: DependencyPanel component
- `internal/cli/settings/setup_flow.go` — New: SetupFlow wizard, createFormForCategory factory
- `internal/cli/settings/dependencies_test.go` — New: 16 tests covering registry, panel, setup flow, menu filter
