## Context

The Settings TUI (`internal/cli/settings/`) uses a hierarchical menu with 60+ categories organized into 8 sections. Features have implicit dependencies (e.g., Smart Account requires Payment + Security Signer) that are invisible to users. When a feature is enabled without its prerequisites, tools silently fail to register — causing confusion even for developers.

Current architecture: `Editor` (state machine) → `MenuModel` (navigation) → `FormModel` (per-category forms). Existing patterns include `EnabledChecker` and `DirtyChecker` callbacks on `MenuModel` for smart filters.

## Goals / Non-Goals

**Goals:**
- Make feature dependency chains visible at all three interaction layers (menu, form entry, guided setup)
- Enable users to discover and resolve prerequisites without leaving the settings flow
- Codify all 20+ dependency relationships in a single, testable registry
- Eliminate duplicated form-creation logic between `handleMenuSelection()` and setup flow

**Non-Goals:**
- Auto-enabling prerequisites (users must explicitly configure each feature)
- Runtime dependency checking outside the settings TUI
- Dependency cycle support (the graph is acyclic by design)
- Changing the underlying config structure

## Decisions

### 1. Callback-based DependencyChecker on MenuModel
**Decision**: Add `DependencyChecker func(string) int` callback to `MenuModel`, following the established `EnabledChecker`/`DirtyChecker` pattern.
**Rationale**: Consistent with existing architecture. Avoids coupling `MenuModel` to the dependency system directly.
**Alternative**: Passing `DependencyIndex` directly to `MenuModel` — rejected because it breaks the callback abstraction used by other checkers.

### 2. Closure-based Check functions in dependency registry
**Decision**: Each `Dependency` holds a `Check func(cfg *config.Config) DepStatus` closure that evaluates against config.
**Rationale**: Flexible — each dependency can inspect any combination of config fields. Shared functions (`checkSmartAccountEnabled`, `checkP2PEnabled`, `checkEconomyEnabled`) avoid duplication for common patterns.
**Alternative**: Declarative field-path checking — rejected because some checks require multi-field logic (e.g., Payment enabled AND RPC URL not empty).

### 3. Depth-first transitive resolution with visited set
**Decision**: `AllTransitiveUnmet()` uses recursive depth-first traversal with a `visited map[string]bool` to prevent infinite loops.
**Rationale**: Ensures children appear before parents in the result (important for guided setup ordering). Visited set is a safety net even though the graph is acyclic.

### 4. Shared `createFormForCategory()` factory
**Decision**: Extract form creation into a single `createFormForCategory()` function used by both `handleMenuSelection()` and `SetupFlow`.
**Rationale**: Eliminates 150+ lines of duplicated switch logic. Single source of truth for category→form mapping.

### 5. Navigation stack for jump-to-dependency
**Decision**: Use `navStack []string` in `Editor` to track jumped-from categories, enabling Esc to return to the original form.
**Rationale**: Simple, bounded by graph depth (~8 max). No need for more complex navigation management.

## Risks / Trade-offs

- [Stale dependency data] → DependencyChecker re-evaluates on each render. Since checks are trivial struct field reads, performance impact is negligible.
- [New category requires two changes] → Adding a form category requires updating both `createFormForCategory()` and `NewMenuModel()`. This is a pre-existing pattern, not a regression.
- [Panel focus management] → `panelFocus bool` must be manually synced with `depPanel` lifecycle. Mitigated by always resetting both together in `attachDependencyPanel()`.
