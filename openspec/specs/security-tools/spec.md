## Purpose

Capability spec for security-tools. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Secrets tool registration
The app SHALL register secrets tool with store/get/list/delete operations.

#### Scenario: Tool available
- **WHEN** agent runtime is initialized
- **THEN** secrets tool is available with operations: store, get, list, delete

### Requirement: Crypto tool registration
The app SHALL register crypto tool with encrypt/decrypt/sign/hash/keys operations.

#### Scenario: Tool available
- **WHEN** agent runtime is initialized
- **THEN** crypto tool is available with operations: encrypt, decrypt, sign, hash, keys

### Requirement: Secrets get requires approval
The secrets.get operation SHALL require user approval via ApprovalMiddleware.

#### Scenario: Approval required
- **WHEN** AI calls secrets.get
- **THEN** ApprovalMiddleware intercepts and requests approval

#### Scenario: Approval granted
- **WHEN** user approves
- **THEN** an opaque reference token SHALL be returned (not the plaintext value)

### Requirement: Secrets get returns reference token
The secrets.get operation SHALL return an opaque reference token instead of the plaintext secret value. The plaintext SHALL be stored in the RefStore and resolved at execution time by the exec tool.

#### Scenario: Get secret returns reference
- **WHEN** AI calls secrets_get with name "api-key"
- **AND** the secret exists
- **THEN** the response SHALL contain `value: "{{secret:api-key}}"`
- **AND** the response SHALL contain a `note` field explaining the token usage
- **AND** the plaintext SHALL NOT appear in the response

#### Scenario: Get secret registers with scanner
- **WHEN** AI calls secrets_get with name "api-key"
- **AND** the secret exists
- **AND** a SecretScanner is configured
- **THEN** the plaintext value SHALL be registered with the SecretScanner for output scanning

### Requirement: Crypto decrypt returns reference token
The crypto.decrypt operation SHALL return an opaque reference token instead of the plaintext decrypted data. The plaintext SHALL be stored in the RefStore and resolved at execution time.

#### Scenario: Decrypt returns reference
- **WHEN** AI calls crypto_decrypt with valid ciphertext
- **THEN** the response SHALL contain `data: "{{decrypt:<uuid>}}"` where uuid is a unique identifier
- **AND** the response SHALL contain a `note` field explaining the token usage
- **AND** the plaintext SHALL NOT appear in the response

#### Scenario: Decrypt registers with scanner
- **WHEN** AI calls crypto_decrypt with valid ciphertext
- **AND** a SecretScanner is configured
- **THEN** the plaintext value SHALL be registered with the SecretScanner for output scanning

### Requirement: LocalCryptoProvider initialization
The app SHALL initialize LocalCryptoProvider with passphrase when security.signer.provider is "local".

#### Scenario: First-time setup
- **WHEN** no salt exists in database
- **THEN** prompt user for passphrase and store salt

#### Scenario: Subsequent startup
- **WHEN** salt exists in database
- **THEN** prompt for passphrase and derive key
