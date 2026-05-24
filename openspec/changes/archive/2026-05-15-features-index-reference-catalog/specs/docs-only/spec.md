## ADDED Requirements

### Requirement: Features index links every feature reference page
The public features landing page SHALL provide a catalog of every dedicated page under `docs/features/` so feature deep dives remain discoverable from the features section itself.

#### Scenario: Feature references stay linked from the features index
- **WHEN** a maintainer updates `docs/features/index.md`
- **THEN** it SHALL include links to every dedicated page under `docs/features/` other than `index.md`
- **AND** that catalog SHALL cover pages such as `agent-format.md`, `learning.md`, `knowledge.md`, `knowledge-graph.md`, `ontology.md`, `p2p-network.md`, `provenance.md`, `run-ledger.md`, and `zkp.md`
