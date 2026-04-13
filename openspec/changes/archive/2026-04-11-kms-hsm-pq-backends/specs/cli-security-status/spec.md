## MODIFIED Requirements

### Requirement: Security status command

#### Scenario: Display KMS protection status
- **WHEN** user runs `lango security status` and the envelope has a KMS KEK slot
- **THEN** the output SHALL include "KMS Protection: enabled (<provider>)" showing the KMS provider name

#### Scenario: Display KMS protection disabled
- **WHEN** user runs `lango security status` and no KMS KEK slot exists
- **THEN** the output SHALL include "KMS Protection: disabled"

#### Scenario: JSON output includes KMS protection
- **WHEN** user runs `lango security status --json`
- **THEN** the JSON output SHALL include `"kms_protected": true/false` and `"kms_provider": "<provider>"` (when protected)
