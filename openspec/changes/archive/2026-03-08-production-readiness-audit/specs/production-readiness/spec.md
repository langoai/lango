## ADDED Requirements

### Requirement: Unsupported security provider produces actionable error
The system SHALL reject unsupported security provider names at config-time with an error message listing all valid provider options (local, rpc, aws-kms, gcp-kms, azure-kv, pkcs11).

#### Scenario: Enclave provider configured
- **WHEN** security.signer.provider is set to "enclave"
- **THEN** initSecurity returns an error containing "unsupported security provider" and listing all valid providers

#### Scenario: Unknown provider configured
- **WHEN** security.signer.provider is set to an unrecognized name
- **THEN** initSecurity returns an error containing the provider name and all valid options

### Requirement: Telegram media download completes successfully
The system SHALL download file content from Telegram's file API via HTTP GET with a 30-second timeout and return the raw bytes.

#### Scenario: Successful file download
- **WHEN** DownloadFile is called with a valid file reference
- **THEN** the system returns the file content as bytes with no error

#### Scenario: HTTP error from Telegram API
- **WHEN** the Telegram file API returns a non-200 status code
- **THEN** the system returns an error containing the HTTP status code

#### Scenario: Empty response body
- **WHEN** the Telegram file API returns a 200 status with an empty body
- **THEN** the system returns an error indicating the download produced no data

### Requirement: No dead code or context.TODO in x402 package
The x402 package SHALL contain no unused exported functions and no `context.TODO()` calls.

#### Scenario: NewX402Client removed
- **WHEN** the codebase is scanned for calls to NewX402Client
- **THEN** no references exist and the function is not present in the source

#### Scenario: No context.TODO remaining
- **WHEN** the x402 package is scanned for context.TODO()
- **THEN** zero occurrences are found

### Requirement: GVisor stub behavior is documented and tested
The GVisor runtime stub SHALL clearly document its stub nature and have tests verifying stub behavior.

#### Scenario: GVisor not available
- **WHEN** IsAvailable() is called on the GVisor stub
- **THEN** it returns false

#### Scenario: GVisor run returns unavailable error
- **WHEN** Run() is called on the GVisor stub
- **THEN** it returns ErrRuntimeUnavailable

### Requirement: Wallet package has unit test coverage
The wallet package SHALL have tests covering address derivation, transaction signing, message signing, composite fallback logic, wallet creation, and RPC dispatching.

#### Scenario: Local wallet signs transaction
- **WHEN** SignTransaction is called with a valid key in SecretsStore
- **THEN** the signature is valid and the public key can be recovered

#### Scenario: Composite wallet falls back on primary failure
- **WHEN** the primary wallet provider is disconnected
- **THEN** the composite wallet delegates to the fallback provider

#### Scenario: Wallet creation stores recoverable key
- **WHEN** CreateWallet is called
- **THEN** the stored key can be retrieved and derives the same address

### Requirement: Security KeyRegistry and SecretsStore have unit test coverage
The security package SHALL have tests covering full CRUD operations on KeyRegistry and SecretsStore with mock CryptoProvider.

#### Scenario: KeyRegistry register and retrieve
- **WHEN** a key is registered via RegisterKey
- **THEN** GetKey returns the same key with correct metadata

#### Scenario: SecretsStore encrypt and decrypt roundtrip
- **WHEN** a secret is stored via Store
- **THEN** Get returns the decrypted original value

#### Scenario: SecretsStore with no encryption key
- **WHEN** Store is called with no encryption key registered
- **THEN** it returns ErrNoEncryptionKeys

### Requirement: Payment service has unit test coverage
The payment service SHALL have tests covering Send error branches, History, RecordX402Payment, and failTx.

#### Scenario: Send with invalid address
- **WHEN** Send is called with an invalid Ethereum address
- **THEN** it returns a validation error

#### Scenario: History returns records with limit
- **WHEN** History is called with a limit
- **THEN** it returns at most that many records in descending order

### Requirement: Smart account packages have unit test coverage
The smartaccount package SHALL have tests covering factory CREATE2 computation, session key crypto, ABI encoding, paymaster errors, policy syncing, and type methods.

#### Scenario: CREATE2 address is deterministic
- **WHEN** ComputeAddress is called with identical inputs
- **THEN** it produces the same address every time

#### Scenario: Session key serialize/deserialize roundtrip
- **WHEN** a session key is serialized then deserialized
- **THEN** the restored key equals the original

#### Scenario: Policy drift detection
- **WHEN** DetectDrift is called with matching on-chain and Go-side policies
- **THEN** no drift is reported

### Requirement: Economy risk package has unit test coverage
The economy/risk package SHALL have tests covering risk factor computation and strategy selection matrix.

#### Scenario: Risk classification boundaries
- **WHEN** computeRiskScore produces boundary values
- **THEN** classifyRisk returns the correct risk level at each threshold

#### Scenario: Strategy matrix covers all combinations
- **WHEN** SelectStrategy is called with all 9 trust/verifiability combinations
- **THEN** each combination returns the expected strategy

### Requirement: P2P team conflict resolution has unit test coverage
The p2p/team package SHALL have tests covering all 4 conflict resolution strategies.

#### Scenario: TrustWeighted picks fastest successful agent
- **WHEN** ResolveConflict is called with TrustWeighted strategy
- **THEN** the fastest successful agent's result is selected

#### Scenario: FailOnConflict rejects disagreement
- **WHEN** ResolveConflict is called with FailOnConflict and conflicting results
- **THEN** an error is returned

### Requirement: P2P protocol messages and remote agent have unit test coverage
The p2p/protocol package SHALL have tests covering ResponseStatus validation, RequestType constants, and RemoteAgent accessors.

#### Scenario: ResponseStatus.Valid for all statuses
- **WHEN** Valid() is called on each defined ResponseStatus
- **THEN** it returns true for valid statuses and false for invalid ones

#### Scenario: RemoteAgent field population
- **WHEN** NewRemoteAgent is called with a config
- **THEN** all accessor methods return the configured values
