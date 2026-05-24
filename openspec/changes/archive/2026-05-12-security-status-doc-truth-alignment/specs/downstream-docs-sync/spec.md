## MODIFIED Requirements

### Requirement: Security status docs match actual field semantics
Public docs for `lango security status` SHALL describe the current field semantics exposed by the command rather than older narrower examples.

#### Scenario: Security status JSON field docs stay truth-aligned
- **WHEN** a user reads the `lango security status` field descriptions
- **THEN** `signer_provider` SHALL be documented as the active provider or `unavailable` when DB-backed config could not be read non-interactively
- **AND** `db_encryption` SHALL be documented using the current payload-protection / legacy-DB status strings rather than implying active SQLCipher page encryption
- **AND** `kms_fallback` SHALL be documented as a KMS-only enabled/disabled status
