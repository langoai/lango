## ADDED Requirements

### Requirement: Features index reference-catalog guard stays executable
Repository-level regressions that let the features landing page drop links to dedicated feature references SHALL be enforced by an executable test.

#### Scenario: Every feature reference remains linked
- **WHEN** the repository still ships dedicated feature pages under `docs/features/`
- **THEN** an executable repository test SHALL fail if `docs/features/index.md` stops linking any of those pages
- **AND** it SHALL therefore catch omissions affecting references such as `agent-format.md`, `learning.md`, `knowledge.md`, `knowledge-graph.md`, `ontology.md`, `p2p-network.md`, `provenance.md`, `run-ledger.md`, or `zkp.md`
