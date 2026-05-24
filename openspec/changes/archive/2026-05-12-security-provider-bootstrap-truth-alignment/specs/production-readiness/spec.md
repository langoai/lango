## MODIFIED Requirements

### Requirement: Unsupported security provider produces actionable error
The system SHALL reject unsupported security provider names at config-time with an error message listing all valid provider options (local, rpc, aws-kms, gcp-kms, azure-kv, pkcs11).

#### Scenario: Local provider without bootstrap returns actionable wiring error
- **WHEN** security.signer.provider is set to `local`
- **AND** initSecurity is called without bootstrap-backed storage dependencies
- **THEN** initSecurity returns an error explaining that the local security provider requires bootstrap

#### Scenario: KMS provider without compiled support returns actionable build-tag error
- **WHEN** security.signer.provider is set to `aws-kms`, `gcp-kms`, `azure-kv`, or `pkcs11`
- **AND** the current build does not include the matching KMS backend
- **THEN** initSecurity returns an error that names the provider
- **AND** explains that support is not compiled
- **AND** tells the operator which `-tags` value is required to rebuild with support
