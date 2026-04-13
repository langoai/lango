## MODIFIED Requirements

### Requirement: Security status command

#### Scenario: Display PQ signing key status
- **WHEN** user runs `lango security status` and PQ signing key is available
- **THEN** the identity bundle section SHALL include "PQ Signing Key: available (ml-dsa-65)"

#### Scenario: Display PQ signing key unavailable
- **WHEN** user runs `lango security status` and PQ signing key is not available
- **THEN** the identity bundle section SHALL include "PQ Signing Key: not available"

#### Scenario: JSON output includes PQ signing key status
- **WHEN** user runs `lango security status --json`
- **THEN** the identity bundle section SHALL include `"pq_signing_key_available": true/false` and `"pq_signing_algorithm": "ml-dsa-65"` (when available)
