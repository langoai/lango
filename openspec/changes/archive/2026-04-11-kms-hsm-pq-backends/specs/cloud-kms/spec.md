## MODIFIED Requirements

### Requirement: KMS KEK Slot Wrapping

The system SHALL support using any CryptoProvider (AWS KMS, GCP KMS, Azure KV, PKCS#11) to wrap and unwrap the Master Key as a KEK slot in the envelope. The KMS provider used for MK unwrap SHALL be a bare provider, NOT a CompositeCryptoProvider with local fallback.

#### Scenario: KMS wraps MK during slot addition
- **WHEN** `AddKMSSlot` is called with a CryptoProvider and key ID
- **THEN** the system calls `provider.Encrypt(ctx, kmsKeyID, mk)` to wrap the MK
- **AND** stores the ciphertext in the envelope slot

#### Scenario: KMS unwraps MK during bootstrap
- **WHEN** the envelope has a hardware slot and KMS env vars are configured
- **THEN** the system creates a bare KMS provider and calls `UnwrapFromKMS`
- **AND** the provider SHALL NOT be wrapped in `CompositeCryptoProvider`

#### Scenario: CompositeCryptoProvider not used for MK unwrap
- **WHEN** KMS is unavailable during MK unwrap
- **THEN** the system SHALL NOT fall back to `LocalCryptoProvider.Decrypt` on the same slot
- **AND** SHALL instead fall through to the next credential path (mnemonic or passphrase)

### Requirement: KMS CLI wrap and detach commands

The system SHALL provide `lango security kms wrap` and `lango security kms detach` CLI commands.

#### Scenario: Wrap command adds KMS slot
- **WHEN** `lango security kms wrap --provider aws-kms --key-id <arn>` is run
- **THEN** the system bootstraps to obtain MK, creates a KMS provider, adds a KMS slot to the envelope, and persists the updated envelope

#### Scenario: Detach command removes single KMS slot
- **WHEN** `lango security kms detach` is run and exactly one hardware slot exists
- **THEN** the slot is removed and the envelope is persisted

#### Scenario: Detach command with multiple slots requires slot-id
- **WHEN** `lango security kms detach` is run and multiple hardware slots exist
- **THEN** the command SHALL require `--slot-id <uuid>` and list matching slots if not provided

#### Scenario: Detach guard preserves minimum slots
- **WHEN** `lango security kms detach` would leave zero passphrase/mnemonic slots
- **THEN** the command SHALL refuse and display an error
