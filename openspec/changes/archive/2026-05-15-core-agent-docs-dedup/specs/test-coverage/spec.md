## ADDED Requirements

### Requirement: Core docs scope guard stays executable
Repository-level regressions that reintroduce duplicated agent diagnostics into the core CLI reference SHALL be enforced by an executable test.

#### Scenario: Core docs keep agent diagnostics delegated
- **WHEN** the repository still ships a dedicated `docs/cli/agent.md` reference
- **THEN** an executable repository test SHALL fail if `docs/cli/core.md` reintroduces standalone `lango agent trace ...` or `lango agent graph ...` sections instead of delegating to the agent reference
