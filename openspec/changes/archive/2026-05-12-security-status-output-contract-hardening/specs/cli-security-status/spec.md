## MODIFIED Requirements

### Requirement: Status output field semantics stay stable
The security status surface SHALL keep its signer-provider, DB status, and KMS fallback fields aligned with the actual runtime semantics.

#### Scenario: KMS signer remains visible in status output
- **WHEN** the security status command renders a KMS-backed signer configuration
- **THEN** the output SHALL keep the active signer provider visible as the KMS provider name
- **AND** SHALL surface the KMS provider, key ID, and fallback enabled/disabled state

#### Scenario: Unavailable config read still preserves explicit DB status strings
- **WHEN** the security status command renders JSON output for a state where DB-backed config could not be read non-interactively
- **THEN** `signer_provider` SHALL be `unavailable`
- **AND** `db_encryption` SHALL preserve the current runtime status string rather than a generic placeholder
