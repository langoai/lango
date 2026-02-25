# Spec: CLI Command Groups

## Overview
Improve CLI discoverability by organizing `lango --help` output into logical groups and adding cross-references between related configuration commands.

## Requirements

### R1: Command Grouping
The root command must define four Cobra groups and assign every subcommand to one:

| Group ID | Title | Commands |
|----------|-------|----------|
| `core` | Core: | serve, version, health |
| `config` | Configuration: | config, settings, onboard, doctor |
| `data` | Data & AI: | memory, graph, agent |
| `infra` | Infrastructure: | security, p2p, cron, workflow, payment |

#### Scenarios
- **lango --help**: Commands appear grouped under their titles instead of flat alphabetical list.

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
