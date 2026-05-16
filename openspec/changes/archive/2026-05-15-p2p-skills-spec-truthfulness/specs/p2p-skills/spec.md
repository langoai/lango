## ADDED Requirements
### Requirement: P2P skills spec reflects the embedded placeholder state
The `p2p-skills` main spec SHALL describe the current placeholder-only embedded skill scaffold truthfully until real embedded P2P skills are added.

#### Scenario: No nonexistent embedded P2P skills are claimed
- **WHEN** the repository ships only `skills/.placeholder/SKILL.md`
- **THEN** the `p2p-skills` main spec SHALL not claim that specific `skills/p2p-*/SKILL.md` files already exist
