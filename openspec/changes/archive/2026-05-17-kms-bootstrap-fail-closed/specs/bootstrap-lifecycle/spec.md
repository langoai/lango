## MODIFIED Requirements

### Requirement: KMS bootstrap env config
The system SHALL provide a `KMSConfigFromEnv()` function that reads KMS KEK
configuration from environment variables and SHALL make bootstrap KMS fallback
behavior honor the active `KMSConfig`.

#### Scenario: KMS fallback env config
- **WHEN** `LANGO_KMS_PROVIDER` is set and `LANGO_KMS_FALLBACK_TO_LOCAL=false`
- **THEN** `KMSConfigFromEnv()` SHALL return a KMS config with
  `FallbackToLocal` set to false

#### Scenario: KMS fail-closed bootstrap
- **WHEN** bootstrap attempts KMS KEK unwrap with `KMSConfig.FallbackToLocal`
  set to false
- **AND** the configured KMS provider cannot be created or cannot unwrap the
  master key
- **THEN** bootstrap SHALL fail immediately
- **AND** SHALL NOT fall back to passphrase acquisition

#### Scenario: KMS fallback-enabled bootstrap
- **WHEN** bootstrap attempts KMS KEK unwrap with `KMSConfig.FallbackToLocal`
  set to true
- **AND** the configured KMS provider cannot be created or cannot unwrap the
  master key
- **THEN** bootstrap SHALL continue to the existing passphrase credential path
