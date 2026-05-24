## MODIFIED Requirements

### Requirement: P2P feature documentation
The system SHALL provide `docs/features/p2p-network.md` covering the live external-collaboration behavior truthfully, including current DID identity modes, payment-surface ownership, and guidance-oriented team/workspace/git operator surfaces where direct live control does not yet exist.

#### Scenario: Identity documentation includes both DID modes
- **WHEN** a user reads `docs/features/p2p-network.md`
- **THEN** the identity section SHALL describe both legacy wallet-derived `did:lango:<hex>` identities and bundle-backed `did:lango:v2:<hash>` identities

#### Scenario: Team and workspace command summary is truthful
- **WHEN** a user reads the quick command list in `docs/features/p2p-network.md`
- **THEN** team, workspace, and git commands SHALL be described as guidance-oriented or inspection-oriented when they do not provide full direct live control

#### Scenario: Workspace chronicler wording reflects partial wiring
- **WHEN** a user reads the workspace chronicler section
- **THEN** the documentation SHALL explain that graph-triple persistence depends on triple-adder wiring being available and is not yet guaranteed as a default live path

### Requirement: P2P CLI reference documentation
The system SHALL provide `docs/cli/p2p.md` with usage, flags, arguments, and examples for all P2P commands, and those examples SHALL reflect the current runtime honestly.

#### Scenario: Identity docs describe active DID exposure
- **WHEN** a user reads the `lango p2p identity` section in `docs/cli/p2p.md`
- **THEN** the documentation SHALL explain that the command exposes the active DID when one is available and SHALL distinguish legacy and v2 DID modes

#### Scenario: Pricing docs describe provider-side quote surface
- **WHEN** a user reads the `lango p2p pricing` section
- **THEN** the documentation SHALL describe it as the provider-side public quote configuration surface rather than a generic pricing policy engine

#### Scenario: Team, workspace, and git examples are guidance-oriented
- **WHEN** a user reads the `team`, `workspace`, or `git` CLI sections
- **THEN** the examples SHALL match the current server-backed or tool-backed reality instead of implying direct live control
