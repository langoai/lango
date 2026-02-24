## Purpose

Cloud KMS and HSM backend integration for the CryptoProvider interface. Provides build-tag-isolated implementations for AWS KMS, GCP KMS, Azure Key Vault, and PKCS#11, with retry logic, health checking, and CLI management.

## Requirements

### Requirement: KMS Provider Factory
The system SHALL provide a `NewKMSProvider(providerName, kmsConfig)` factory that dispatches to the correct KMS backend based on provider name. Supported names: `aws-kms`, `gcp-kms`, `azure-kv`, `pkcs11`.

#### Scenario: Valid provider name
- **WHEN** `NewKMSProvider("aws-kms", validConfig)` is called with a compiled build tag
- **THEN** the factory returns an initialized `CryptoProvider` and nil error

#### Scenario: Unknown provider name
- **WHEN** `NewKMSProvider("unknown", config)` is called
- **THEN** the factory returns an error containing the unknown name and lists supported providers

#### Scenario: Provider not compiled
- **WHEN** `NewKMSProvider("aws-kms", config)` is called without the `kms_aws` build tag
- **THEN** the stub returns an error indicating the provider was not compiled and which build tag is needed

### Requirement: Build Tag Isolation
Each KMS provider SHALL be gated behind build tags. The default build (no tags) SHALL compile successfully using stub files that return descriptive errors. Build tags: `kms_aws`, `kms_gcp`, `kms_azure`, `kms_pkcs11`, `kms_all`.

#### Scenario: Default build without tags
- **WHEN** `go build ./...` is run without any KMS build tags
- **THEN** the project compiles successfully using stub implementations

#### Scenario: Build with kms_all tag
- **WHEN** `go build -tags kms_all ./...` is run
- **THEN** all four KMS providers are compiled into the binary

### Requirement: Transient Error Retry
KMS operations SHALL be retried with exponential backoff (100ms base, doubled each attempt) for transient errors. Only errors classified as `ErrKMSUnavailable` or `ErrKMSThrottled` SHALL be retried.

#### Scenario: Transient error succeeds on retry
- **WHEN** a KMS operation returns `ErrKMSThrottled` on the first attempt
- **AND** succeeds on the second attempt
- **THEN** the operation returns success

#### Scenario: Non-transient error not retried
- **WHEN** a KMS operation returns `ErrKMSAccessDenied`
- **THEN** the error is returned immediately without retry

#### Scenario: Retries exhausted
- **WHEN** a KMS operation returns transient errors for all configured retry attempts
- **THEN** the last error is returned

### Requirement: KMS Health Checker
The system SHALL provide a `KMSHealthChecker` implementing `ConnectionChecker` that probes KMS availability via encrypt/decrypt roundtrip. Results SHALL be cached for 30 seconds.

#### Scenario: KMS reachable
- **WHEN** the health checker probes and the roundtrip succeeds
- **THEN** `IsConnected()` returns true

#### Scenario: KMS unreachable with cache
- **WHEN** the last probe failed less than 30 seconds ago
- **THEN** `IsConnected()` returns the cached false result without re-probing

### Requirement: KMS Error Classification
Each KMS provider SHALL classify cloud-specific errors into sentinel error types: `ErrKMSUnavailable`, `ErrKMSAccessDenied`, `ErrKMSKeyDisabled`, `ErrKMSThrottled`, `ErrKMSInvalidKey`. Errors SHALL be wrapped in `KMSError` with Provider, Op, KeyID context.

#### Scenario: AWS access denied
- **WHEN** AWS KMS returns `AccessDeniedException`
- **THEN** the error wraps `ErrKMSAccessDenied` and includes provider="aws", operation, and key ID

#### Scenario: GCP throttled
- **WHEN** GCP KMS returns gRPC `ResourceExhausted` status
- **THEN** the error wraps `ErrKMSThrottled`

### Requirement: AWS KMS Provider
The AWS KMS provider SHALL implement `CryptoProvider` using `aws-sdk-go-v2/service/kms`. Sign uses `ECDSA_SHA_256` with `MessageType: RAW`. Encrypt/Decrypt use `SYMMETRIC_DEFAULT`. Authentication uses SDK default credential chain.

#### Scenario: Encrypt and decrypt roundtrip
- **WHEN** data is encrypted with `Encrypt()` then decrypted with `Decrypt()`
- **THEN** the original plaintext is recovered

#### Scenario: Key alias resolution
- **WHEN** `keyID` is "local" or "default"
- **THEN** the configured default key ID is used

### Requirement: GCP KMS Provider
The GCP KMS provider SHALL implement `CryptoProvider` using `cloud.google.com/go/kms/apiv1`. Sign uses `AsymmetricSign` with SHA-256 digest. Encrypt/Decrypt use symmetric operations. Authentication uses Application Default Credentials.

#### Scenario: Sign with SHA-256 digest
- **WHEN** `Sign()` is called with a payload
- **THEN** the payload is SHA-256 hashed before sending to GCP AsymmetricSign

### Requirement: Azure Key Vault Provider
The Azure KV provider SHALL implement `CryptoProvider` using `azkeys`. Sign uses ES256. Encrypt/Decrypt use RSA-OAEP. Authentication uses `DefaultAzureCredential`.

#### Scenario: Missing vault URL rejected
- **WHEN** `newAzureKVProvider()` is called with empty `VaultURL`
- **THEN** an error is returned indicating vault URL is required

### Requirement: PKCS#11 Provider
The PKCS#11 provider SHALL implement `CryptoProvider` using `miekg/pkcs11`. Sign uses `CKM_ECDSA`. Encrypt/Decrypt use `CKM_AES_GCM` with 12-byte IV prepended to ciphertext. PIN is read from `LANGO_PKCS11_PIN` env var with config fallback.

#### Scenario: PIN from environment variable
- **WHEN** `LANGO_PKCS11_PIN` environment variable is set
- **THEN** it takes priority over the config pin value

#### Scenario: Session cleanup on Close
- **WHEN** `Close()` is called on the PKCS#11 provider
- **THEN** the session is logged out, closed, and the module is finalized

### Requirement: KMS Fallback to Local
When `security.kms.fallbackToLocal` is true, the system SHALL wrap the KMS provider in `CompositeCryptoProvider` with the local crypto provider as fallback and `KMSHealthChecker` as the connection checker.

#### Scenario: KMS unavailable with fallback enabled
- **WHEN** the KMS provider is unreachable and `fallbackToLocal` is true
- **THEN** operations transparently fall back to the local crypto provider

### Requirement: KMS CLI Commands
The system SHALL provide `lango security kms status`, `lango security kms test`, and `lango security kms keys` CLI commands.

#### Scenario: KMS status display
- **WHEN** `lango security kms status` is run with a KMS provider configured
- **THEN** the output shows provider type, key ID, region, fallback status, and connection status

#### Scenario: KMS roundtrip test
- **WHEN** `lango security kms test` is run
- **THEN** the system performs an encrypt/decrypt roundtrip and reports success or failure

#### Scenario: KMS keys listing
- **WHEN** `lango security kms keys` is run
- **THEN** all keys from KeyRegistry are displayed with ID, name, type, and remote key ID

### Requirement: KMS Config Structure
The system SHALL define `KMSConfig` with fields: Region, KeyID, Endpoint, FallbackToLocal, TimeoutPerOperation, MaxRetries, Azure (AzureKVConfig), PKCS11 (PKCS11Config). Defaults: FallbackToLocal=true, TimeoutPerOperation=5s, MaxRetries=3.

#### Scenario: Default config values
- **WHEN** no KMS config is provided
- **THEN** FallbackToLocal is true, TimeoutPerOperation is 5 seconds, MaxRetries is 3
