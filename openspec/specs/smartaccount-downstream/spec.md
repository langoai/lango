# Spec: Smart Account Downstream Artifact Sync

## Purpose

Capability spec for smartaccount-downstream. See requirements below for scope and behavior contracts.

## Requirements

### REQ-1: TUI Smart Account Settings
The TUI settings editor MUST include configuration forms for all 19 SmartAccount config keys, organized into 4 categories: Smart Account (main), SA Session Keys, SA Paymaster, SA Modules.

#### Scenario: Settings editor exposes smart account categories
- **WHEN** a user opens `lango settings` and navigates to the Infrastructure section
- **THEN** the Smart Account categories SHALL be visible

#### Scenario: Main smart account form exposes editable fields
- **WHEN** a user selects "Smart Account" in the settings editor
- **THEN** all main config fields (enabled, factory, entrypoint, safe7579, fallback, bundler) SHALL be editable

#### Scenario: Edited smart account fields persist on save
- **WHEN** a user modifies a smart account field and saves
- **THEN** the config SHALL be persisted correctly

### REQ-2: Documentation Coverage
Feature docs, CLI docs, config docs, tool usage docs, and README MUST document all smart account capabilities matching the actual codebase.

#### Scenario: Feature docs describe the smart account surface
- **WHEN** a user reads `docs/features/smart-accounts.md`
- **THEN** they SHALL find architecture overview, session keys, paymaster, policy, modules, tools, and config coverage

#### Scenario: CLI docs list the smart account commands
- **WHEN** a user reads `docs/cli/smartaccount.md`
- **THEN** they SHALL find all 11 CLI commands with flags and examples

#### Scenario: Top-level smart account help includes paymaster examples
- **WHEN** a user reads `docs/cli/smartaccount.md` or runs `lango account --help`
- **THEN** the top-level overview SHALL include representative examples for info, deploy, session, module, policy, and paymaster surfaces

#### Scenario: Configuration docs list smart account keys
- **WHEN** a user reads `docs/configuration.md`
- **THEN** they SHALL find all 19 SmartAccount config keys

### REQ-3: Multi-Agent Tool Routing
All 12 smart account tools MUST be routed to the vault sub-agent in multi-agent orchestration mode.

#### Scenario: Orchestrator routes smart account work to vault
- **WHEN** multi-agent mode is enabled and a user requests smart account operations
- **THEN** the orchestrator MUST route the request to the vault agent

#### Scenario: Smart account tools do not fall into Unmatched
- **WHEN** `PartitionTools` processes smart account tools
- **THEN** none of those tools MUST fall into `Unmatched`

### REQ-4: Cross-Reference Integrity
Feature index, economy doc, and contracts doc MUST cross-reference smart accounts.

#### Scenario: Smart account docs are cross-linked from related surfaces
- **WHEN** a user reads the feature index, economy doc, or contracts doc
- **THEN** those docs MUST cross-reference smart accounts

### REQ-5: Build and Deploy
Makefile MUST include `check-abi` target. Docker compose MUST include smart account env var example.

#### Scenario: Build and deploy artifacts mention smart account setup
- **WHEN** a developer inspects the Makefile or docker compose configuration
- **THEN** the Makefile MUST include `check-abi`
- **AND** docker compose MUST include a smart account environment variable example
