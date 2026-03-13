## Purpose

Requirements for keeping downstream artifacts (documentation, prompts, Docker config, Makefile) synchronized with core feature changes.

## Requirements

### Requirement: Prompt files reflect all tool categories
The system prompts SHALL list all current tool categories including the Team category, with accurate tool counts and descriptions for all 7 team tools.

#### Scenario: AGENTS.md tool count
- **WHEN** a user reads `prompts/AGENTS.md`
- **THEN** the document SHALL state "fifteen" tool categories and include a Team category section

#### Scenario: TOOL_USAGE.md team tools
- **WHEN** a user reads `prompts/TOOL_USAGE.md`
- **THEN** the document SHALL contain a Team Tool section documenting `team_form`, `team_delegate`, `team_status`, `team_list`, `team_disband`, `team_form_with_budget`, `team_complete_milestone` with parameters and return values

### Requirement: README reflects all implemented features
The README SHALL list all implemented features including Team Health Monitoring, Incremental Git Bundles, Task Branch Management, Config Presets, Event-Driven Bridges, EventMonitor Reorg Protection, and Escrow Hub V2.

#### Scenario: New features in README
- **WHEN** a user reads `README.md`
- **THEN** all 7 new feature areas SHALL be listed in the features section

#### Scenario: CLI commands in README
- **WHEN** a user reads the CLI commands section of `README.md`
- **THEN** `lango status`, `lango onboard --preset`, and cron `--timeout` SHALL be documented

### Requirement: Config presets documentation exists
A dedicated documentation page SHALL exist for config presets describing all 4 presets with feature matrices.

#### Scenario: Presets doc page
- **WHEN** a user navigates to `docs/features/config-presets.md`
- **THEN** the page SHALL document `minimal`, `researcher`, `collaborator`, `full` presets with feature flags

### Requirement: Status command documentation exists
A dedicated CLI reference page SHALL exist for the `lango status` command.

#### Scenario: Status doc page
- **WHEN** a user navigates to `docs/cli/status.md`
- **THEN** the page SHALL document `--output` flag, `--addr` flag, output sections, and JSON schema

### Requirement: Feature index pages updated
The feature index pages SHALL include cards for P2P Workspaces, P2P Teams, and Config Presets.

#### Scenario: Feature index cards
- **WHEN** a user reads `docs/features/index.md` or `docs/index.md`
- **THEN** cards for Workspaces, Teams, and Config Presets SHALL be present

### Requirement: Makefile test targets for new packages
The Makefile SHALL provide dedicated test targets for team, economy, and bridge packages.

#### Scenario: Makefile test-team target
- **WHEN** a user runs `make test-team`
- **THEN** tests in `./internal/p2p/team/...` SHALL execute

#### Scenario: Makefile test-economy target
- **WHEN** a user runs `make test-economy`
- **THEN** tests in `./internal/economy/...` SHALL execute

#### Scenario: Makefile test-bridges target
- **WHEN** a user runs `make test-bridges`
- **THEN** bridge-related tests SHALL execute

### Requirement: Docker config supports workspaces
The Docker Compose configuration SHALL include a workspace volume and team/economy environment variables.

#### Scenario: Docker workspace volume
- **WHEN** a user reads `docker-compose.yml`
- **THEN** a `lango-workspaces` volume SHALL be defined
