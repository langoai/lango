## ADDED Requirements
### Requirement: P2P reputation JSON output uses explicit output selection
The reputation inspection path SHALL expose machine-readable output through `--output json` instead of a boolean `--json` toggle.

#### Scenario: Reputation JSON output remains machine-readable
- **WHEN** the user runs `lango p2p reputation --peer-did "did:lango:abc123" --output json`
- **THEN** the command SHALL emit the same machine-readable peer reputation fields as before
