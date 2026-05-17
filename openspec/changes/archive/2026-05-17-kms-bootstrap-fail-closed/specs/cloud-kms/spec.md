## MODIFIED Requirements

### Requirement: KMS KEK Slot Wrapping
The system SHALL support using any CryptoProvider (AWS KMS, GCP KMS, Azure KV,
PKCS#11) to wrap and unwrap the Master Key as a KEK slot in the envelope. The
KMS provider used for MK unwrap SHALL be a bare provider, NOT a
CompositeCryptoProvider with local fallback.

#### Scenario: KMS bootstrap fails closed when local fallback is disabled
- **WHEN** the envelope has a hardware slot, KMS bootstrap config is present,
  and active `KMSConfig.FallbackToLocal` is false
- **AND** KMS provider initialization or MK unwrap fails
- **THEN** bootstrap SHALL return an error before passphrase acquisition
- **AND** the error SHALL preserve the original KMS failure context

#### Scenario: KMS bootstrap may fall back when local fallback is enabled
- **WHEN** the envelope has a hardware slot, KMS bootstrap config is present,
  and active `KMSConfig.FallbackToLocal` is true
- **AND** KMS provider initialization or MK unwrap fails
- **THEN** bootstrap MAY warn and continue to the next credential path
