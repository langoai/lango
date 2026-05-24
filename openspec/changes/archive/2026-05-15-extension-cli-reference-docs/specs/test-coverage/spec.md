## ADDED Requirements

### Requirement: Extension CLI reference quality guard stays executable
Repository-level regressions that let the dedicated extension CLI reference drift away from the implemented command surface SHALL be enforced by an executable test.

#### Scenario: Implemented extension command contract remains documented
- **WHEN** the repository still ships the implemented `lango extension inspect <source>`, `install <source>`, `list`, and `remove <name>` commands with `table|json|plain` output and `--yes` scripted confirmations
- **THEN** an executable repository test SHALL fail if `docs/cli/extension.md` no longer documents that command and flag surface
