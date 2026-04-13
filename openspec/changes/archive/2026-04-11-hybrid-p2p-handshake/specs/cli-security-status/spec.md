## MODIFIED Requirements

### Requirement: Security status command

#### Scenario: Display PQ KEM status
- **WHEN** user runs `lango security status` and PQ handshake is enabled
- **THEN** the output SHALL include "PQ Handshake: enabled (X25519-MLKEM768)"

#### Scenario: Display PQ KEM status when disabled
- **WHEN** user runs `lango security status` and PQ handshake is not enabled
- **THEN** the output SHALL include "PQ Handshake: disabled"

#### Scenario: JSON output includes PQ KEM status
- **WHEN** user runs `lango security status --json`
- **THEN** the JSON output SHALL include `"pq_handshake_enabled": true/false` and `"pq_handshake_algorithm": "X25519-MLKEM768"` (when enabled)
