## ADDED Requirements

### Requirement: Dependency Registry
The system SHALL maintain a registry of feature dependency relationships as a `DependencyIndex` with O(1) lookup by category ID. Each dependency SHALL have a category ID, label, required flag, check function, and fix hint. The registry SHALL support evaluation, unmet-required counting, transitive resolution with cycle guard, and reverse lookup (dependents).

#### Scenario: Evaluate Smart Account dependencies
- **WHEN** evaluating dependencies for "smartaccount" with default (empty) config
- **THEN** the system returns 3 results: payment (NotEnabled, required), security (NotEnabled, required), economy (NotEnabled, optional)

#### Scenario: Unmet required count with partial config
- **WHEN** Payment is enabled with RPC URL but Security Signer is empty
- **THEN** `UnmetRequired("smartaccount")` returns 1

#### Scenario: All dependencies met
- **WHEN** Payment is enabled with RPC URL and Security Signer provider is set
- **THEN** `UnmetRequired("smartaccount")` returns 0

#### Scenario: Transitive resolution
- **WHEN** resolving transitive dependencies for "smartaccount_session"
- **THEN** the result includes payment, security, and smartaccount, with payment and security appearing before smartaccount (depth-first order)

#### Scenario: Cycle guard
- **WHEN** resolving transitive dependencies on any category
- **THEN** the visited set prevents infinite recursion even if a cycle exists

#### Scenario: Reverse lookup
- **WHEN** querying dependents of "payment"
- **THEN** the result includes "smartaccount" among other categories

#### Scenario: Category with no dependencies
- **WHEN** evaluating "agent" (no dependencies defined)
- **THEN** `Evaluate` returns nil and `UnmetRequired` returns 0

### Requirement: Menu Warning Badges
The system SHALL display a warning badge (`⚠ N`) next to menu categories that have N unmet required dependencies. The badge SHALL use `BadgeDependencyStyle` (warning background, bold) and appear after the ADV badge.

#### Scenario: Badge displayed for blocked category
- **WHEN** "smartaccount" has 2 unmet required dependencies
- **THEN** the menu item renders with `⚠ 2` badge

#### Scenario: No badge for unblocked category
- **WHEN** "agent" has 0 unmet required dependencies
- **THEN** no dependency badge is rendered

### Requirement: Ready Smart Filter
The system SHALL support an `@ready` smart filter in the search bar that shows only categories with zero unmet required dependencies. The filter hint text SHALL include `@ready`.

#### Scenario: Ready filter excludes blocked categories
- **WHEN** user types `@ready` in search
- **THEN** categories with unmet dependencies (e.g., "smartaccount" with 2 unmet) are excluded from results

#### Scenario: Ready filter includes unblocked categories
- **WHEN** user types `@ready` in search
- **THEN** categories with no dependencies or all dependencies met appear in results

### Requirement: Prerequisite Panel
The system SHALL display a prerequisite panel above the form when entering a category with unmet dependencies. The panel SHALL show each dependency with a status indicator (✓ met, ✗ not enabled, ⚠ misconfigured), label, optional tag, and fix hint for the selected item. The panel SHALL support cursor navigation and jump-to-dependency via Enter.

#### Scenario: Panel appears with unmet deps
- **WHEN** opening "smartaccount" form with Payment disabled
- **THEN** a prerequisite panel appears showing Payment as ✗ with fix hint

#### Scenario: Panel not created when all met
- **WHEN** opening a category with all dependencies met
- **THEN** no prerequisite panel is displayed

#### Scenario: Jump to dependency
- **WHEN** user presses Enter on an unmet dependency in the panel
- **THEN** the editor navigates to that dependency's form, pushing current form onto nav stack

#### Scenario: Return from dependency via Esc
- **WHEN** user presses Esc in a jumped-to form
- **THEN** the editor pops the nav stack and returns to the original category's form

### Requirement: Guided Setup Flow
The system SHALL offer a guided setup flow when the user presses 's' in the prerequisite panel with 1+ unmet dependencies. The flow SHALL chain prerequisite forms in dependency order (depth-first transitive resolution), with progress bar, step list, and support for next (Ctrl+N), skip (Ctrl+S), and cancel (Esc).

#### Scenario: Setup flow creation
- **WHEN** creating a setup flow with 2 unmet deps for "smartaccount"
- **THEN** the flow has 2 steps (payment, security) and state is InProgress

#### Scenario: Step progression
- **WHEN** user completes step 1 (Ctrl+N) and skips step 2 (Ctrl+S)
- **THEN** the flow state becomes Completed and editor opens the target form

#### Scenario: Deduplication of transitive deps
- **WHEN** transitive resolution produces duplicate category IDs
- **THEN** the setup flow deduplicates steps so each category appears once

#### Scenario: Cancel setup flow
- **WHEN** user presses Esc during setup flow
- **THEN** the flow is cancelled and editor returns to menu

### Requirement: Shared Form Factory
The system SHALL use a single `createFormForCategory()` function to map category IDs to form constructors, shared between `handleMenuSelection()` and `SetupFlow`. The `handleMenuSelection()` function SHALL handle only non-form selections (providers list, auth list, MCP servers list, save, cancel) separately.

#### Scenario: Form creation via factory
- **WHEN** selecting "payment" from the menu
- **THEN** `createFormForCategory("payment", cfg)` returns a non-nil form and the editor enters StepForm

#### Scenario: Non-form selection handled separately
- **WHEN** selecting "providers" from the menu
- **THEN** the editor enters StepProvidersList without calling createFormForCategory
