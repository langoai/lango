## ADDED Requirements
### Requirement: P2P pricing JSON output uses explicit output selection
The pricing inspection path SHALL expose machine-readable output through `--output json` instead of a boolean `--json` toggle.

#### Scenario: Pricing JSON output remains machine-readable
- **WHEN** the user runs `lango p2p pricing --output json`
- **THEN** the command SHALL emit the same machine-readable pricing payload fields as before
