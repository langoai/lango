## ADDED Requirements

### Requirement: CLI index gives graph commands a dedicated section
The public CLI index SHALL keep implemented `lango graph` commands in a dedicated graph section once the repository ships a dedicated graph CLI reference.

#### Scenario: Graph commands stay separated from Agent & Memory
- **WHEN** a maintainer updates `docs/cli/index.md`
- **THEN** it SHALL keep a dedicated `Graph Store` section for the implemented `lango graph status`, `query`, `stats`, `clear`, `add`, `export`, and `import` commands
- **AND** it SHALL hand off detailed graph coverage to `docs/cli/graph.md` instead of leaving those command rows embedded inside the `Agent & Memory` section
