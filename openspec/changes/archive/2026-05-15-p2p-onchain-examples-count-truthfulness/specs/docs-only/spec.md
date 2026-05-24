## ADDED Requirements
### Requirement: P2P on-chain examples summary reflects the current shipped inventory
The `p2p-onchain-examples` main spec SHALL state the current number of shipped Docker Compose examples truthfully.

#### Scenario: Seven shipped examples are reflected
- **WHEN** seven example directories ship in `examples/`
- **THEN** the spec SHALL not describe the inventory as six examples
