## MODIFIED Requirements

### Requirement: Security provider docs stay aligned with supported providers and bootstrap rules
Public configuration docs SHALL describe the currently supported security signer providers accurately and SHALL not mention removed provider names.

#### Scenario: KMS operation docs mention build-tag and wiring requirements
- **WHEN** a user reads README, CLI, or security docs for KMS-backed signer setup
- **THEN** those docs SHALL mention that KMS providers require the matching build tag in the running binary
- **AND** SHALL mention that the runtime still depends on bootstrap-backed storage wiring for the key registry and secrets store
