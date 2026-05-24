## ADDED Requirements
### Requirement: P2P on-chain examples summary avoids stale exact test counts
The `p2p-onchain-examples` main spec SHALL describe shipped example coverage without stale exact `Tests (N)` totals.

#### Scenario: Representative checks are used instead of stale totals
- **WHEN** shipped example scripts evolve independently
- **THEN** the spec SHALL summarize representative checks instead of hard-coded stale test counts
