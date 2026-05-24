# Spec: P2P Workspace Downstream Artifacts

## Purpose

Capability spec for p2p-workspace-downstream. See requirements below for scope and behavior contracts.

## Requirements

### REQ-1: TUI Settings Form
The TUI settings editor SHALL expose P2P workspace configuration with 7 fields (enabled, dataDir, maxWorkspaces, maxBundleSizeBytes, chroniclerEnabled, autoSandbox, contributionTracking). Fields 5-7 SHALL be conditionally visible when workspace is enabled.

#### Scenario: User can reach P2P Workspace settings
- **WHEN** a user navigates to P2P Workspace in the settings menu
- **THEN** the workspace configuration form SHALL be visible

#### Scenario: Conditional fields appear when workspace is enabled
- **WHEN** a user toggles workspace enabled on
- **THEN** the conditional downstream fields SHALL appear

#### Scenario: Invalid maxWorkspaces shows validation error
- **WHEN** a user enters a non-positive `maxWorkspaces` value
- **THEN** the form SHALL show a validation error

### REQ-2: Doctor Diagnostic Check
The doctor command SHALL include a `WorkspaceCheck` that validates workspace configuration, git binary availability, and data directory existence. The check SHALL be fixable by auto-creating the data directory.

#### Scenario: Disabled workspace skips the diagnostic
- **WHEN** workspace support is disabled
- **THEN** the workspace diagnostic SHALL be skipped

#### Scenario: Missing git emits a warning
- **WHEN** workspace support is enabled and git is unavailable
- **THEN** the doctor check SHALL report a warning

#### Scenario: Missing data dir is a fixable warning
- **WHEN** workspace support is enabled and the data directory is missing
- **THEN** the doctor check SHALL report a fixable warning

#### Scenario: Healthy workspace config passes
- **WHEN** workspace support is enabled and all prerequisites are present
- **THEN** the doctor check SHALL pass with a summary

### REQ-3: Tool Catalog Entry
The `lango agent tools` command SHALL list a `workspace` category with config key `p2p.workspace.enabled`.

#### Scenario: Workspace category appears in tool catalog output
- **WHEN** `lango agent tools` is run
- **THEN** the output SHALL include the `workspace` category with config key `p2p.workspace.enabled`

### REQ-4: CLI Documentation
`docs/cli/p2p.md` SHALL document all 10 workspace/git subcommands with usage, flags, and examples.

#### Scenario: CLI doc covers all workspace and git commands
- **WHEN** a user reads `docs/cli/p2p.md`
- **THEN** all 10 workspace/git subcommands SHALL be documented with usage, flags, and examples

### REQ-5: Feature Documentation
`docs/features/p2p-network.md` SHALL describe collaborative workspaces (lifecycle, members, messages, chronicler, contributions) and git bundle exchange (bare repos, bundle protocol, DAG leaves).

#### Scenario: Feature docs describe workspace lifecycle and git exchange
- **WHEN** a user reads `docs/features/p2p-network.md`
- **THEN** it SHALL describe collaborative workspace lifecycle and git bundle exchange behavior

### REQ-6: README Update
README SHALL list P2P Workspaces in the features section and include 10 workspace/git CLI commands.

#### Scenario: README mentions P2P workspaces and commands
- **WHEN** a user reads the README features and CLI sections
- **THEN** P2P Workspaces and the 10 workspace/git commands SHALL be present

### REQ-7: Prompt Documentation
`prompts/TOOL_USAGE.md` SHALL document all 12 workspace/git agent tools with usage patterns. `prompts/AGENTS.md` SHALL include the workspace category (14 total).

#### Scenario: Prompt docs include workspace tools and category
- **WHEN** prompt documentation is reviewed
- **THEN** `TOOL_USAGE.md` SHALL document all 12 workspace/git tools
- **AND** `AGENTS.md` SHALL include the workspace category

### REQ-8: Unit Tests
Core packages SHALL have table-driven tests: Manager (16 tests), ContributionTracker (5 tests), Chronicler (5 tests), BareRepoStore (7 tests), Service (4 tests).

#### Scenario: Core workspace packages have table-driven tests
- **WHEN** the workspace test suite is inspected
- **THEN** Manager, ContributionTracker, Chronicler, BareRepoStore, and Service SHALL have the documented table-driven coverage

### REQ-9: Docker & Makefile
Docker Compose SHALL include commented workspace env/volume entries. Makefile SHALL include a `test-workspace` target.

#### Scenario: Build and deployment artifacts mention workspace support
- **WHEN** developers inspect Docker Compose and the Makefile
- **THEN** docker compose SHALL include commented workspace env/volume examples
- **AND** the Makefile SHALL include `test-workspace`
