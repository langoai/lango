# Spec: P2P Workspace Downstream Artifacts

## Purpose

Capability spec for p2p-workspace-downstream. See requirements below for scope and behavior contracts.

## Requirements

### REQ-1: TUI Settings Form
The TUI settings editor must expose P2P workspace configuration with 7 fields (enabled, dataDir, maxWorkspaces, maxBundleSizeBytes, chroniclerEnabled, autoSandbox, contributionTracking). Fields 5-7 must be conditionally visible when workspace is enabled.

#### Scenarios
- User navigates to P2P Workspace in settings menu
- User toggles workspace enabled and sees conditional fields appear
- User enters invalid maxWorkspaces (non-positive) and sees validation error

### REQ-2: Doctor Diagnostic Check
The doctor command must include a WorkspaceCheck that validates workspace configuration, git binary availability, and data directory existence. The check must be fixable (auto-create data directory).

#### Scenarios
- Workspace disabled → check is skipped
- Workspace enabled, git missing → warning
- Workspace enabled, data dir missing → fixable warning
- Workspace enabled, all good → pass with summary

### REQ-3: Tool Catalog Entry
The `lango agent tools` command must list a "workspace" category with config key `p2p.workspace.enabled`.

### REQ-4: CLI Documentation
`docs/cli/p2p.md` must document all 10 workspace/git subcommands with usage, flags, and examples.

### REQ-5: Feature Documentation
`docs/features/p2p-network.md` must describe collaborative workspaces (lifecycle, members, messages, chronicler, contributions) and git bundle exchange (bare repos, bundle protocol, DAG leaves).

### REQ-6: README Update
README must list P2P Workspaces in the features section and include 10 workspace/git CLI commands.

### REQ-7: Prompt Documentation
`prompts/TOOL_USAGE.md` must document all 12 workspace/git agent tools with usage patterns. `prompts/AGENTS.md` must include the workspace category (14 total).

### REQ-8: Unit Tests
Core packages must have table-driven tests: Manager (16 tests), ContributionTracker (5 tests), Chronicler (5 tests), BareRepoStore (7 tests), Service (4 tests).

### REQ-9: Docker & Makefile
Docker Compose must include commented workspace env/volume. Makefile must include `test-workspace` target.
