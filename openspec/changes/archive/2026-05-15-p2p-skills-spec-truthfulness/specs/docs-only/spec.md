## ADDED Requirements
### Requirement: P2P embedded-skill claims stay truthful
Main specs SHALL not advertise embedded P2P skill files that are absent from the repository.

#### Scenario: Stale embedded P2P skill claims are rejected
- **WHEN** a maintainer updates the `p2p-skills` main spec while the repository still ships only the placeholder embedded skill scaffold
- **THEN** the spec SHALL not claim that `skills/p2p-*/SKILL.md` files already exist
