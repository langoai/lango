## ADDED Requirements

### Requirement: SecurityTier enum represents hardware security levels
The system SHALL define a `SecurityTier` enum with values `TierNone` (0), `TierTPM` (1), and `TierBiometric` (2), ordered by security strength.

#### Scenario: SecurityTier string representation
- **WHEN** `SecurityTier.String()` is called
- **THEN** it SHALL return `"none"`, `"tpm"`, or `"biometric"` respectively

#### Scenario: Unknown tier defaults to none
- **WHEN** an unknown `SecurityTier` value calls `String()`
- **THEN** it SHALL return `"none"`

### Requirement: DetectSecureProvider probes hardware backends
The system SHALL provide a `DetectSecureProvider()` function that returns the highest-tier available `(Provider, SecurityTier)` pair by probing biometric first, then TPM, then returning `(nil, TierNone)`.

#### Scenario: macOS with Touch ID available
- **WHEN** `DetectSecureProvider()` is called on macOS with Touch ID hardware
- **THEN** it SHALL return a `BiometricProvider` and `TierBiometric`

#### Scenario: Linux with TPM 2.0 device
- **WHEN** `DetectSecureProvider()` is called on Linux with accessible `/dev/tpmrm0`
- **THEN** it SHALL return a `TPMProvider` and `TierTPM`

#### Scenario: No secure hardware available
- **WHEN** neither biometric nor TPM is available
- **THEN** it SHALL return `(nil, TierNone)`

### Requirement: BiometricProvider uses macOS Keychain with Touch ID ACL
The system SHALL provide a `BiometricProvider` that stores secrets in macOS Keychain using `kSecAccessControlBiometryAny` access control. This provider SHALL require Touch ID authentication for every read operation.

#### Scenario: Store and retrieve with biometric
- **WHEN** a secret is stored via `BiometricProvider.Set()` and later retrieved via `BiometricProvider.Get()`
- **THEN** the Set SHALL create a Keychain item with biometric ACL, and Get SHALL trigger Touch ID before returning the value

#### Scenario: Biometric not available on non-Darwin platform
- **WHEN** `NewBiometricProvider()` is called on a non-Darwin or non-CGO platform
- **THEN** it SHALL return `ErrBiometricNotAvailable`

### Requirement: TPMProvider seals secrets with TPM 2.0
The system SHALL provide a `TPMProvider` that seals secrets under the TPM's Storage Root Key and stores the sealed blob at `~/.lango/tpm/<service>_<key>.sealed`.

#### Scenario: Seal and unseal with TPM
- **WHEN** a secret is stored via `TPMProvider.Set()` and later retrieved via `TPMProvider.Get()`
- **THEN** Set SHALL seal the data with TPM2 and write the blob to disk, and Get SHALL unseal using the same TPM chip

#### Scenario: TPM not available on non-Linux platform
- **WHEN** `NewTPMProvider()` is called on a non-Linux platform
- **THEN** it SHALL return `ErrTPMNotAvailable`

#### Scenario: Delete removes sealed blob
- **WHEN** `TPMProvider.Delete()` is called for an existing sealed blob
- **THEN** it SHALL remove the sealed blob file from disk

### Requirement: Error sentinels for hardware availability
The system SHALL define `ErrBiometricNotAvailable` and `ErrTPMNotAvailable` sentinel errors for callers to distinguish hardware unavailability from other failures.

#### Scenario: Error sentinel messages
- **WHEN** error sentinels are checked
- **THEN** `ErrBiometricNotAvailable` SHALL contain "biometric authentication not available" and `ErrTPMNotAvailable` SHALL contain "TPM device not available"

### Requirement: Build-tag stubs for cross-platform compilation
The system SHALL provide stub implementations with build tags (`!darwin || !cgo` for biometric, `!linux` for TPM) that implement the `Provider` interface and return the appropriate sentinel errors.

#### Scenario: Stub methods satisfy Provider interface
- **WHEN** stub types are used on unsupported platforms
- **THEN** all `Get`, `Set`, `Delete` methods SHALL return the platform-specific sentinel error
