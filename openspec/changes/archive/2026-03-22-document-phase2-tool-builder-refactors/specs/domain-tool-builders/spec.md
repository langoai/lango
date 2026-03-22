## ADDED Requirements

### Requirement: Automation packages own their tool builder functions
The `cron`, `background`, and `workflow` packages SHALL export `BuildTools()` functions that create their automation tools without importing `internal/app/`.

#### Scenario: Automation tool builders are package-owned
- **WHEN** automation tools are registered
- **THEN** `cron.BuildTools()`, `background.BuildTools()`, and `workflow.BuildTools()` SHALL provide those tool definitions

#### Scenario: Automation packages have no app import
- **WHEN** checking imports of the automation tool builder files
- **THEN** none of them SHALL import `internal/app/`

### Requirement: Data and collaboration packages own their tool builders
The `agentmemory`, `graph`, `embedding`, `memory`, `librarian`, `p2p/team`, and `economy/escrow/sentinel` packages SHALL own their exported tool builder functions.

#### Scenario: Data package builders do not depend on app
- **WHEN** `agentmemory.BuildTools()`, `graph.BuildTools()`, `embedding.BuildRAGTools()`, `memory.BuildObservationTools()`, and `librarian.BuildTools()` are compiled
- **THEN** their files SHALL NOT import `internal/app/`

#### Scenario: Team and sentinel builders do not depend on app
- **WHEN** `team.BuildTools()`, `team.BuildEscrowTools()`, and `sentinel.BuildTools()` are compiled
- **THEN** their files SHALL NOT import `internal/app/`

### Requirement: Foundation packages own their tool builder functions
The `tools/browser`, `tools/filesystem`, `tools/exec`, `tooloutput`, `tools/crypto`, and `tools/secrets` packages SHALL export builder functions for their tool definitions. The app layer MAY keep app-specific guard callbacks, but SHALL pass them into package builders instead of defining legacy app-local tool builder functions.

#### Scenario: Foundation builders are package-owned
- **WHEN** foundation tools are registered
- **THEN** browser, filesystem, exec, output, crypto, and secrets tools SHALL come from their owning packages' builder functions

#### Scenario: Exec package accepts app-owned guards as callbacks
- **WHEN** the app wires exec tools
- **THEN** it SHALL pass the lango-command guard and protected-path guard as callback functions into `exec.BuildTools()`

### Requirement: Cycle-bound builders may remain in app
Builders that still require cross-package knowledge or import-cycle-sensitive glue MAY remain in `internal/app/` until a separate boundary redesign is completed.

#### Scenario: Meta and on-chain escrow builders remain app-owned
- **WHEN** tool builder ownership is reviewed after the recent refactors
- **THEN** `buildMetaTools()` and `buildOnChainEscrowTools()` MAY remain in `internal/app/`
- **AND** this SHALL be treated as an explicit exception, not an undocumented leftover
