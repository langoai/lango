## ADDED Requirements

### Requirement: CLI index links every dedicated CLI reference page
The public CLI index SHALL provide an explicit catalog of every dedicated page under `docs/cli/` so operators can discover the deeper command-family references from the top-level index.

#### Scenario: Dedicated CLI references stay linked from the index
- **WHEN** a maintainer updates `docs/cli/index.md`
- **THEN** it SHALL include links to every dedicated page under `docs/cli/` other than `index.md`
- **AND** that catalog SHALL cover command-family pages such as `core.md`, `status.md`, `agent.md`, `agent-memory.md`, `automation.md`, `extension.md`, `graph.md`, `payment.md`, `provenance.md`, `sandbox.md`, and `smartaccount.md`
