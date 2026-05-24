## ADDED Requirements

### Requirement: CLI index dedicated-reference catalog guard stays executable
Repository-level regressions that let the top-level CLI index drop links to dedicated CLI reference pages SHALL be enforced by an executable test.

#### Scenario: Every dedicated CLI reference remains linked
- **WHEN** the repository still ships dedicated CLI reference pages under `docs/cli/`
- **THEN** an executable repository test SHALL fail if `docs/cli/index.md` stops linking any of those pages
- **AND** it SHALL therefore catch omissions affecting dedicated references such as `core.md`, `status.md`, `agent.md`, `automation.md`, `extension.md`, `graph.md`, `payment.md`, `provenance.md`, `sandbox.md`, or `smartaccount.md`
