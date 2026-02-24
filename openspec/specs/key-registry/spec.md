## ADDED Requirements

### Requirement: Key Entity Schema
The system SHALL store key metadata using entgo.io Key entity.

#### Scenario: Key entity fields
- **WHEN** a Key entity is created
- **THEN** it SHALL have fields: id (uuid), name (unique string), remoteKeyId (string), type (enum: signing/encryption), createdAt, lastUsedAt

### Requirement: Secret Entity Schema
The system SHALL store encrypted secrets using entgo.io Secret entity.

#### Scenario: Secret entity fields
- **WHEN** a Secret entity is created
- **THEN** it SHALL have fields: id (uuid), name (unique string), encryptedValue (bytes), createdAt, updatedAt, accessCount
- **AND** it SHALL have a required edge to Key entity

### Requirement: Key Registration
The system SHALL allow registering new encryption keys.

#### Scenario: Register companion key
- **WHEN** companion app connects and provides a new key
- **THEN** the system stores key metadata in Key entity
- **AND** remoteKeyId references the companion's key identifier

#### Scenario: Register local fallback key
- **WHEN** local fallback is initialized with passphrase
- **THEN** the system stores key metadata with type "encryption"
- **AND** remoteKeyId is set to "local"

### Requirement: Default Key Selection
The system SHALL support a default key for operations without explicit keyId.

#### Scenario: Get default encryption key
- **WHEN** an encryption operation is requested without keyId
- **THEN** the system uses the most recently registered encryption key

#### Scenario: No keys available
- **WHEN** an operation is requested but no keys are registered
- **THEN** the system returns an error "no encryption keys available"

### Requirement: KMS Key Registration in Wiring
When a KMS provider is initialized, the system SHALL register the KMS key in KeyRegistry with the cloud key ARN/ID as `RemoteKeyID` and name `kms-default`.

#### Scenario: KMS provider wiring registers key
- **WHEN** `initSecurity()` initializes a KMS provider (aws-kms, gcp-kms, azure-kv, pkcs11)
- **THEN** a key named `kms-default` SHALL be registered in KeyRegistry
- **AND** its RemoteKeyID SHALL be set to `security.kms.keyId`
- **AND** its type SHALL be `encryption`
