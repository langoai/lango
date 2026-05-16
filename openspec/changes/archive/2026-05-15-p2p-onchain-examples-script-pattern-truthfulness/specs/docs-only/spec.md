## ADDED Requirements
### Requirement: P2P on-chain examples summary stays truthful about discovery scripts
The `p2p-onchain-examples` main spec SHALL describe the current discovery-script pattern truthfully.

#### Scenario: Mixed polling and fixed-sleep behavior is reflected
- **WHEN** `p2p-trading` still uses a fixed `sleep 15` warm-up
- **THEN** the spec SHALL not describe polling loops as a universal pattern across all examples
