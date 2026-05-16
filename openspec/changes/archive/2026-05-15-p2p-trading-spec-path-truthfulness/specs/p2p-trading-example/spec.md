## ADDED Requirements
### Requirement: P2P trading example spec reflects the current MockUSDC path
The `p2p-trading-example` main spec SHALL point to the current repository path for the shipped MockUSDC contract.

#### Scenario: Stale MockUSDC path is rejected
- **WHEN** the repository keeps the mock contract at `contracts/test/mocks/MockUSDC.sol`
- **THEN** the `p2p-trading-example` main spec SHALL not claim `contracts/MockUSDC.sol` as the current contract path
