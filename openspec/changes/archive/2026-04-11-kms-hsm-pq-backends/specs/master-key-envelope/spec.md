## MODIFIED Requirements

### Requirement: KEK Slot types

#### Scenario: KMS KEK slot wraps MK via CryptoProvider
- **WHEN** `AddKMSSlot(ctx, label, mk, provider, kmsProviderName, kmsKeyID)` is called
- **THEN** the envelope SHALL call `provider.Encrypt(ctx, kmsKeyID, mk)` and store the returned ciphertext in `slot.WrappedMK`
- **AND** the slot SHALL have `Type: "hardware"`, `KDFAlg: "none"`, `WrapAlg: "kms-envelope"`
- **AND** `KMSProvider` and `KMSKeyID` SHALL be populated from the arguments

#### Scenario: KMS KEK slot fields use omitempty
- **WHEN** an envelope is serialized to JSON
- **THEN** `KMSProvider` and `KMSKeyID` fields SHALL use `omitempty` tags
- **AND** envelopes without KMS slots SHALL serialize identically to the pre-KMS format

### Requirement: MK unwrap from KMS with 2-tier matching

The `UnwrapFromKMS(ctx, provider, providerName, keyID)` method SHALL apply a two-tier matching strategy to find and decrypt the correct KMS slot.

#### Scenario: Tier 1 exact match (provider + keyID)
- **WHEN** `UnwrapFromKMS` is called and a slot exists with `KMSProvider == providerName && KMSKeyID == keyID`
- **THEN** the method SHALL call `provider.Decrypt(ctx, slot.KMSKeyID, slot.WrappedMK)`
- **AND** return the recovered MK on success

#### Scenario: Tier 2 provider-only fallback
- **WHEN** no Tier 1 exact match succeeds and slots exist with `KMSProvider == providerName`
- **THEN** the method SHALL try each matching slot using `slot.KMSKeyID` for the decrypt call (not the env keyID)
- **AND** return the recovered MK on first success

#### Scenario: No matching slot
- **WHEN** no `KEKSlotHardware` slot matches the configured provider
- **THEN** `UnwrapFromKMS` SHALL return `ErrKMSSlotUnavailable`

#### Scenario: Recovered MK size validation
- **WHEN** `provider.Decrypt` returns successfully
- **THEN** the recovered MK SHALL be validated to be exactly 32 bytes
- **AND** if the size is wrong, the method SHALL zero the bytes and return `ErrUnwrapFailed`

#### Scenario: KMS slot backward compatibility
- **WHEN** an envelope JSON without `kms_provider` or `kms_key_id` fields is loaded
- **THEN** those fields SHALL default to empty strings
- **AND** the envelope SHALL load and function correctly for passphrase/mnemonic slots
