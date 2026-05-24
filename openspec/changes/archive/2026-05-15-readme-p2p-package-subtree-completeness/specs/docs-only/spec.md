## ADDED Requirements

### Requirement: README internal tree includes the current P2P package subtree
The README internal package tree SHALL include the currently shipped `internal/p2p` subpackages rather than a partial subset.

#### Scenario: P2P package subtree stays truthful
- **WHEN** a maintainer updates the README internal package tree
- **THEN** it SHALL include `agentpool`, `discovery`, `firewall`, `gitbundle`, `handshake`, `identity`, `ontologybridge`, `paygate`, `protocol`, `provenanceproto`, `reputation`, `settlement`, `team`, `trustpolicy`, `workspace`, and `zkp`
