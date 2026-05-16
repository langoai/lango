## ADDED Requirements
### Requirement: P2P identity command summary reflects current CLI output
The public P2P feature-page command summary SHALL describe `lango p2p identity` as including DID output when available.

#### Scenario: Command summary includes DID
- **WHEN** a maintainer updates the P2P feature-page CLI command list
- **THEN** it SHALL not describe `lango p2p identity` as showing only peer identity and listen addresses
