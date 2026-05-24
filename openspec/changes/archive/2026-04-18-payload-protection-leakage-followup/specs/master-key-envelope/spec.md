## ADDED Requirements

### Requirement: Payload key version remains fixed in leakage follow-up
The corrective leakage-follow-up change MUST continue using payload key version `1` and MUST NOT introduce rotation or multi-version payload handling.

#### Scenario: Protected rows continue using key version 1
- **WHEN** payload-protected rows are written by the corrective leakage-followup change
- **THEN** their `*_key_version` fields are set to `1`
- **AND** no new payload key version is introduced
