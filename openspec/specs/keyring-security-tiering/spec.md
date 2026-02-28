# Keyring Security Tiering

## Purpose

Hardware-backed security tier detection with biometric (macOS Touch ID) and TPM 2.0 (Linux) keyring providers, plus deny-fallback for environments without secure hardware. Prevents same-UID passphrase exposure by requiring user presence verification before keyring auto-unlock.
## Requirements
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
The system SHALL provide a `BiometricProvider` that stores secrets in the macOS login Keychain (NOT the Data Protection Keychain) using `kSecAccessControlBiometryCurrentSet` access control with `kSecAttrAccessibleWhenPasscodeSetThisDeviceOnly` protection. All Keychain queries SHALL set `kSecUseDataProtectionKeychain = kCFBooleanFalse` to explicitly target the login Keychain. This provider SHALL require Touch ID authentication for every read operation, and SHALL invalidate stored items when biometric enrollment changes.

#### Scenario: Store and retrieve with biometric
- **WHEN** a secret is stored via `BiometricProvider.Set()` and later retrieved via `BiometricProvider.Get()`
- **THEN** the Set SHALL create a login Keychain item with `BiometryCurrentSet` ACL and `kSecUseDataProtectionKeychain = false`, and Get SHALL trigger Touch ID before returning the value

#### Scenario: Biometric not available on non-Darwin platform
- **WHEN** `NewBiometricProvider()` is called on a non-Darwin or non-CGO platform
- **THEN** it SHALL return `ErrBiometricNotAvailable`

#### Scenario: Ad-hoc signed binary works without entitlement
- **WHEN** a `go build` ad-hoc signed binary calls `BiometricProvider.Set()` or `BiometricProvider.Get()`
- **THEN** the operation SHALL succeed without requiring `keychain-access-groups` entitlement

#### Scenario: Fingerprint enrollment change invalidates stored items
- **WHEN** a user changes their biometric enrollment (adds or removes fingerprints) after storing a secret
- **THEN** attempts to retrieve the secret SHALL fail because `BiometryCurrentSet` invalidates the access control

#### Scenario: Device passcode not set
- **WHEN** the device does not have a passcode configured
- **THEN** `NewBiometricProvider()` SHALL return `ErrBiometricNotAvailable` because the Keychain probe will fail

### Requirement: TPMProvider seals secrets with TPM 2.0

The TPM provider SHALL use `TPMTSymDefObject` for the SRK template symmetric parameters. The provider SHALL use `tpm2.Marshal` with single-return signature and `tpm2.Unmarshal` with generic type parameter signature `Unmarshal[T]([]byte) (*T, error)` as required by go-tpm v0.9.8.

#### Scenario: SRK template uses correct symmetric type
- **WHEN** the TPM provider creates a primary key with ECC P256 SRK template
- **THEN** the template's `Symmetric` field SHALL be of type `TPMTSymDefObject`

#### Scenario: Marshal sealed blob without error return
- **WHEN** the TPM provider marshals `TPM2BPublic` and `TPM2BPrivate` to bytes
- **THEN** the system SHALL call `tpm2.Marshal` which returns `[]byte` directly

#### Scenario: Unmarshal sealed blob with generic type parameter
- **WHEN** the TPM provider unmarshals bytes into `TPM2BPublic` or `TPM2BPrivate`
- **THEN** the system SHALL call `tpm2.Unmarshal[T](data)` returning `(*T, error)` and dereference the result

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

### Requirement: BiometricProvider SHALL zero C heap buffers before freeing
The `BiometricProvider` SHALL zero all C heap buffers containing plaintext secrets before calling `free()`. Zeroing MUST use a volatile pointer pattern to prevent compiler optimization from eliding the memory wipe.

#### Scenario: Get zeroes C buffer via secure_free
- **WHEN** `BiometricProvider.Get()` retrieves a secret from the Keychain
- **THEN** the C heap buffer SHALL be zeroed via `secure_free()` (volatile pointer loop + free) before control returns to Go

#### Scenario: Set zeroes CString buffer before freeing
- **WHEN** `BiometricProvider.Set()` stores a secret in the Keychain
- **THEN** the `C.CString` buffer containing the plaintext value SHALL be zeroed with `memset` before `free` is called

### Requirement: BiometricProvider SHALL zero intermediate Go byte slices
The `BiometricProvider.Get()` method SHALL copy Keychain data into a Go `[]byte` via `C.GoBytes`, extract the string, and then zero every byte of the `[]byte` slice before it becomes unreachable.

#### Scenario: Get zeroes Go byte slice after string extraction
- **WHEN** `BiometricProvider.Get()` copies data from C heap to Go heap
- **THEN** it SHALL use `C.GoBytes` (not `C.GoStringN`), extract the string via `string(data)`, and zero the `[]byte` with a range loop

### Requirement: secure_free C helper prevents compiler optimization
The C `secure_free` helper function SHALL cast the pointer to `volatile char *` before zeroing to prevent the compiler from optimizing away the memset as a dead store.

#### Scenario: Volatile pointer prevents optimization
- **WHEN** `secure_free(ptr, len)` is called
- **THEN** it SHALL iterate through the buffer using a `volatile char *` pointer, set each byte to zero, and then call `free(ptr)`

#### Scenario: Null pointer safety
- **WHEN** `secure_free(NULL, 0)` is called
- **THEN** it SHALL return without error (NULL guard)

### Requirement: BiometricProvider availability probe uses real Keychain write
The `keychain_biometric_available` function SHALL verify biometric support by performing a real `SecItemAdd` probe to the login Keychain with biometric ACL, rather than only checking `SecAccessControlCreateWithFlags`. The probe item SHALL be cleaned up immediately after the test.

#### Scenario: Probe succeeds on capable hardware
- **WHEN** `keychain_biometric_available()` is called on a macOS device with Touch ID and device passcode set
- **THEN** it SHALL add a probe item to the login Keychain, delete it, and return 1

#### Scenario: Probe fails without passcode
- **WHEN** `keychain_biometric_available()` is called on a macOS device without a passcode
- **THEN** the `SecItemAdd` SHALL fail and the function SHALL return 0

#### Scenario: Probe does not trigger Touch ID
- **WHEN** the probe item is added via `SecItemAdd`
- **THEN** it SHALL NOT trigger a Touch ID prompt because Keychain writes bypass ACL evaluation

### Requirement: All Keychain queries target login Keychain explicitly
Every Keychain query dictionary (set, get, has, delete) SHALL include `kSecUseDataProtectionKeychain = kCFBooleanFalse` to ensure operations target the login Keychain and never fall through to the Data Protection Keychain.

#### Scenario: Set targets login Keychain
- **WHEN** `keychain_set_biometric()` builds its query dictionaries
- **THEN** both the delete-existing and add-new dictionaries SHALL include `kSecUseDataProtectionKeychain = kCFBooleanFalse`

#### Scenario: Get targets login Keychain
- **WHEN** `keychain_get_biometric()` builds its query dictionary
- **THEN** it SHALL include `kSecUseDataProtectionKeychain = kCFBooleanFalse`

#### Scenario: Has targets login Keychain
- **WHEN** `keychain_has_biometric()` builds its query dictionary
- **THEN** it SHALL include `kSecUseDataProtectionKeychain = kCFBooleanFalse`

#### Scenario: Delete targets login Keychain
- **WHEN** `keychain_delete_biometric()` builds its query dictionary
- **THEN** it SHALL include `kSecUseDataProtectionKeychain = kCFBooleanFalse`

