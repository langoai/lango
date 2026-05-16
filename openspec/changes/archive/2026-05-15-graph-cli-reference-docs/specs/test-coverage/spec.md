## ADDED Requirements

### Requirement: Graph CLI reference quality guard stays executable
Repository-level regressions that let the dedicated graph CLI reference drift away from the implemented command surface SHALL be enforced by an executable test.

#### Scenario: Implemented graph command contract remains documented
- **WHEN** the repository still ships the implemented `lango graph status`, `query`, `stats`, `clear`, `add`, `export`, and `import <file>` commands
- **THEN** an executable repository test SHALL fail if `docs/cli/graph.md` no longer documents that command surface, the `table|json` output contract, the `export --format json|csv` contract, or the `clear --force` behavior
