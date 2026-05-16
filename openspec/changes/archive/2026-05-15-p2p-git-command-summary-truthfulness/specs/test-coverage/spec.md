## ADDED Requirements
### Requirement: P2P feature git-command-summary guard stays executable
Repository-level regressions that reintroduce stale `lango p2p git push` or `lango p2p git fetch` summary wording into public P2P feature docs SHALL be enforced by an executable test.

#### Scenario: Stale git command summaries are rejected
- **WHEN** `docs/features/p2p-network.md` describes `lango p2p git push` as directly creating and pushing a bundle or `lango p2p git fetch` as directly fetching and applying one
- **THEN** an executable repository test SHALL fail
