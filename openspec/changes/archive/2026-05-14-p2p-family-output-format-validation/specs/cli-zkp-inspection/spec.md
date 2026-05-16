## ADDED Requirements
### Requirement: ZKP inspection JSON output uses explicit output selection
The ZKP inspection CLI SHALL expose machine-readable output through `--output json` instead of a boolean `--json` toggle.

#### Scenario: ZKP status and circuit JSON output remain machine-readable
- **WHEN** the user runs `lango p2p zkp status --output json` or `lango p2p zkp circuits --output json`
- **THEN** the commands SHALL emit the same machine-readable ZKP payloads as before
