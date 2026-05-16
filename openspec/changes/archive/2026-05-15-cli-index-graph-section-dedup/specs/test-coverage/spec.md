## ADDED Requirements

### Requirement: CLI index graph-section guard stays executable
Repository-level regressions that put graph quick-reference rows back inside the Agent & Memory section instead of keeping a dedicated graph section SHALL be enforced by an executable test.

#### Scenario: CLI index keeps graph coverage in its own section
- **WHEN** the repository still ships a dedicated `docs/cli/graph.md` reference and the implemented `lango graph status`, `query`, `stats`, `clear`, `add`, `export`, and `import <file>` commands
- **THEN** an executable repository test SHALL fail if `docs/cli/index.md` drops the dedicated `Graph Store` section, loses the handoff to `docs/cli/graph.md`, or reintroduces graph command rows inside the `Agent & Memory` section
