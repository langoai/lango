## ADDED Requirements

### Requirement: Graph CLI reference stays aligned with the current command surface
The dedicated graph CLI reference SHALL describe the implemented `status`, `query`, `stats`, `clear`, `add`, `export`, and `import` commands and their current output/format contracts.

#### Scenario: Implemented graph command contract stays documented
- **WHEN** a maintainer updates `docs/cli/graph.md`
- **THEN** it SHALL document the implemented `lango graph status`, `query`, `stats`, `clear`, `add`, `export`, and `import <file>` commands, the `table|json` output contract, the `export --format json|csv` contract, and the `clear --force` confirmation bypass
