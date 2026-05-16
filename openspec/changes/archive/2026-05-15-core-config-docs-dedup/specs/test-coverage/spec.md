## ADDED Requirements

### Requirement: Core config docs scope guard stays executable
Repository-level regressions that reintroduce duplicated config command docs into the core CLI reference SHALL be enforced by an executable test.

#### Scenario: Core docs keep config coverage delegated
- **WHEN** the repository still ships a dedicated `docs/cli/config.md` reference
- **THEN** an executable repository test SHALL fail if `docs/cli/core.md` reintroduces standalone `lango config ...` command sections instead of delegating to the config reference
