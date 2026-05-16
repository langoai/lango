## ADDED Requirements
### Requirement: P2P identity JSON output uses explicit output selection
The identity inspection path SHALL expose machine-readable output through `--output json` instead of a boolean `--json` toggle.

#### Scenario: Identity JSON output remains machine-readable
- **WHEN** the user runs `lango p2p identity --output json`
- **THEN** the JSON SHALL include the same identity fields that were previously available through the machine-readable path
