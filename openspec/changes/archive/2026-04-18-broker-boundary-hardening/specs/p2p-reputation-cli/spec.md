## ADDED Requirements

### Requirement: P2P reputation CLI uses reputation store capability
The `lango p2p reputation` command MUST obtain its reputation store through a storage facade capability instead of constructing it from a generic Ent client in the CLI layer.

#### Scenario: Reputation command reads through facade capability
- **WHEN** the user runs `lango p2p reputation`
- **THEN** the command resolves the reputation store from the storage facade
