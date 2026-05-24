## ADDED Requirements
### Requirement: P2P skills spec truthfulness guard stays executable
Repository-level regressions that reintroduce stale embedded-P2P-skill claims into the `p2p-skills` main spec SHALL be enforced by an executable repository test.

#### Scenario: Stale embedded P2P skill claims are rejected
- **WHEN** the repository still ships only the placeholder embedded skill scaffold
- **THEN** an executable repository test SHALL fail if the `p2p-skills` main spec claims specific `skills/p2p-*/SKILL.md` files already exist
