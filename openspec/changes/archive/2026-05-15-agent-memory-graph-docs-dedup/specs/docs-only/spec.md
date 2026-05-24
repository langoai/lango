## ADDED Requirements

### Requirement: Agent & memory CLI reference delegates graph commands to the dedicated graph page
The agent-and-memory CLI reference SHALL not keep a duplicated embedded graph command manual once a dedicated graph CLI reference exists.

#### Scenario: Graph command duplication stays removed from agent-memory docs
- **WHEN** a maintainer updates `docs/cli/agent-memory.md`
- **THEN** it SHALL hand off graph command coverage to `docs/cli/graph.md` instead of embedding standalone `lango graph ...` sections
