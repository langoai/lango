## MODIFIED Requirements

### Requirement: Security provider docs stay aligned with supported providers and bootstrap rules
Public configuration docs SHALL describe the currently supported security signer providers accurately and SHALL not mention removed provider names.

#### Scenario: Security provider tables list current supported values
- **WHEN** a user reads the security signer provider rows in README or configuration docs
- **THEN** those rows SHALL list `local`, `rpc`, `aws-kms`, `gcp-kms`, `azure-kv`, and `pkcs11`
- **AND** SHALL NOT mention removed values such as `enclave`
- **AND** SHALL clarify that `local` depends on bootstrap-backed storage wiring
- **AND** SHALL clarify that KMS-backed providers depend on matching build tags as well as bootstrap-backed storage wiring
