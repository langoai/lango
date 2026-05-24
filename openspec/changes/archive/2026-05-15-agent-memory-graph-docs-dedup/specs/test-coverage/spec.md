## ADDED Requirements

### Requirement: Agent-memory docs scope guard stays executable
Repository-level regressions that reintroduce duplicated graph command docs into the agent-and-memory CLI reference SHALL be enforced by an executable test.

#### Scenario: Agent-memory docs keep graph coverage delegated
- **WHEN** the repository still ships a dedicated `docs/cli/graph.md` reference
- **THEN** an executable repository test SHALL fail if `docs/cli/agent-memory.md` reintroduces standalone `lango graph ...` command sections instead of delegating to the graph reference
