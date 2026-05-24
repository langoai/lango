## MODIFIED Requirements

### Requirement: Unsupported security provider produces actionable error
The system SHALL reject unsupported security provider names at config-time with an error message listing all valid provider options (local, rpc, aws-kms, gcp-kms, azure-kv, pkcs11).

#### Scenario: Enclave provider is rejected as invalid config
- **WHEN** security.signer.provider is set to `enclave`
- **THEN** config validation rejects the value as invalid
- **AND** the validation error lists only `local`, `rpc`, `aws-kms`, `gcp-kms`, `azure-kv`, and `pkcs11` as valid providers
