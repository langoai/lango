# mission-workbench-tui Specification

## Purpose
TBD - created by archiving change surface-split-wave-six. Update Purpose after archive.
## Requirements
### Requirement: Bare `lango` launches the mission workbench

Running `lango` without a subcommand on an interactive terminal SHALL launch a standalone mission workbench rather than the cockpit shell or the focused chat surface.

#### Scenario: Bare `lango` launches the workbench
- **WHEN** the user runs `lango` on an interactive terminal
- **THEN** the application SHALL open the mission workbench surface
- **AND** that surface SHALL become the bare-`lango` contract for Wave 6

### Requirement: The mission workbench hosts Mission Control content without cockpit chrome

The mission workbench SHALL present Mission Control content directly without the full cockpit sidebar or context-panel shell. It reuses Mission Control behavior, but it is not itself the cockpit shell.

#### Scenario: Workbench shows Mission Control content directly
- **WHEN** the workbench renders successfully
- **THEN** the user SHALL see Mission Control content such as missions, live decision state, loops, activity, and the shared composer
- **AND** the workbench SHALL NOT require the full cockpit sidebar or context-panel chrome to expose that content

### Requirement: The workbench remains a lighter local surface than cockpit

The first Wave 6 workbench slice SHALL stay local and mission-native. It may reuse the same Mission Control projection assets as cockpit, but it SHALL NOT imply that cockpit-only surfaces or channel startup belong to bare `lango`.

#### Scenario: Workbench hints to the other explicit surfaces
- **WHEN** the workbench renders first-screen copy or help
- **THEN** it SHALL hint to `lango chat` as the focused chat surface
- **AND** it SHALL hint to `lango cockpit` as the advanced dashboard

#### Scenario: Cockpit-only channel startup is not implied by bare `lango`
- **WHEN** the user launches bare `lango`
- **THEN** the first Wave 6 slice SHALL NOT imply that `--with-channels` or live channel startup belongs to the workbench contract

