# Spec: CLI Command Groups

## Overview
Improve CLI discoverability by organizing `lango --help` output into logical groups and adding cross-references between related configuration commands.

## Requirements

### R1: Command Grouping
The root command must define five Cobra groups organized by user intent and assign every subcommand to one:

| Group ID | Title | Commands |
|----------|-------|----------|
| `start` | Getting Started: | serve, onboard, doctor, settings, status, version |
| `ai` | AI & Knowledge: | agent, memory, learning, graph, librarian, a2a, metrics |
| `auto` | Automation: | cron, workflow, bg |
| `net` | Network & Economy: | p2p, payment, economy, contract, account, mcp |
| `sys` | Security & System: | security, approval, health, config |

#### Scenarios
- **lango --help**: Commands appear grouped under their titles instead of flat alphabetical list.

#### Scenario: Getting Started group
- **WHEN** user runs `lango --help`
- **THEN** Getting Started section contains: serve, onboard, doctor, settings, status, version

#### Scenario: AI & Knowledge group
- **WHEN** user runs `lango --help`
- **THEN** AI & Knowledge section contains: agent, memory, learning, graph, librarian, a2a, metrics

#### Scenario: Automation group
- **WHEN** user runs `lango --help`
- **THEN** Automation section contains: cron, workflow, bg

#### Scenario: Network & Economy group
- **WHEN** user runs `lango --help`
- **THEN** Network & Economy section contains: p2p, payment, economy, contract, account, mcp

#### Scenario: Security & System group
- **WHEN** user runs `lango --help`
- **THEN** Security & System section contains: security, approval, health, config

### R2: Cross-References (See Also)
Each configuration-related command must include a "See Also" section in its `Long` description:
- `config` → settings, onboard, doctor
- `settings` → config, onboard, doctor
- `onboard` → settings, config, doctor
- `doctor` → settings, config, onboard

#### Scenarios
- **lango config --help**: Shows "See Also" section with settings, onboard, doctor references.
- **lango doctor --help**: Shows "See Also" section with settings, config, onboard references.

## Constraints
- No behavioral changes — only `--help` output affected
- All existing commands continue to work identically
