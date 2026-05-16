## Purpose

Capability spec for p2p-skills. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: No embedded P2P skill bundle ships by default
The repository SHALL NOT claim that an embedded bundle of ready-to-use P2P skills ships by default when the runtime only embeds the placeholder skill scaffold.

#### Scenario: Placeholder-only embedded skill tree
- **WHEN** the embedded `skills/` tree is scanned in the repository
- **THEN** it contains only the placeholder scaffold needed for `//go:embed`
- **AND** it does not contain `skills/p2p-*/SKILL.md` files

### Requirement: P2P skill docs stay truthful
The `p2p-skills` spec SHALL describe the current placeholder-only embedded state truthfully until real embedded P2P skills are added.

#### Scenario: No nonexistent embedded P2P skill files are claimed
- **WHEN** a maintainer reads the `p2p-skills` main spec
- **THEN** it does not claim that specific `skills/p2p-*/SKILL.md` files already exist
