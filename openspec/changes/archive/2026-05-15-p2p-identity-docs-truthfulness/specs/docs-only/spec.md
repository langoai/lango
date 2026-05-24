## ADDED Requirements
### Requirement: P2P identity feature docs reflect current CLI behavior
Public P2P feature docs SHALL describe the current `lango p2p identity` behavior truthfully.

#### Scenario: DID output is documented
- **WHEN** the CLI prints the active DID directly when available
- **THEN** `docs/features/p2p-network.md` SHALL not describe the command as omitting the DID
