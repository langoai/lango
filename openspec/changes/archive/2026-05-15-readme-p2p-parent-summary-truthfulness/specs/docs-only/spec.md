## ADDED Requirements

### Requirement: README P2P parent summary stays aligned with the current subtree
The parent `p2p/` summary in the README internal package tree SHALL reflect the current shipped scope of the subtree.

#### Scenario: Narrow legacy summary stays removed
- **WHEN** a maintainer updates the README internal package tree
- **THEN** the parent `p2p/` summary SHALL mention collaborative workspaces, git/provenance exchange, trust policy, payments, and ZK proofs
