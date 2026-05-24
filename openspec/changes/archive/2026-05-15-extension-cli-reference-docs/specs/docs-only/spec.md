## ADDED Requirements

### Requirement: Extension CLI reference stays aligned with the current command surface
The dedicated extension CLI reference SHALL describe the implemented `inspect`, `install`, `list`, and `remove` commands and their current output/confirmation contracts.

#### Scenario: Implemented extension command contract stays documented
- **WHEN** a maintainer updates `docs/cli/extension.md`
- **THEN** it SHALL document the implemented `lango extension inspect <source>`, `install <source>`, `list`, and `remove <name>` commands, the `table|json|plain` output contract, and the `--yes` scripted-run path
