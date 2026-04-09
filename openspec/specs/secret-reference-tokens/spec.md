## Purpose

Capability spec for secret-reference-tokens. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Reference token storage
The system SHALL provide a RefStore that maps opaque reference tokens to secret plaintext values in memory.

#### Scenario: Store secret reference
- **WHEN** a secret value is stored with name "api-key"
- **THEN** the RefStore SHALL return a token in the format `{{secret:api-key}}`
- **AND** the plaintext value SHALL be stored internally

#### Scenario: Store decrypted data reference
- **WHEN** decrypted data is stored with id "uuid-123"
- **THEN** the RefStore SHALL return a token in the format `{{decrypt:uuid-123}}`
- **AND** the plaintext value SHALL be stored internally

### Requirement: Reference token resolution
The system SHALL resolve reference tokens to their plaintext values.

#### Scenario: Resolve known token
- **WHEN** `Resolve` is called with `{{secret:api-key}}`
- **AND** the token was previously stored
- **THEN** the system SHALL return the plaintext value and true

#### Scenario: Resolve unknown token
- **WHEN** `Resolve` is called with `{{secret:unknown}}`
- **AND** the token was never stored
- **THEN** the system SHALL return nil and false

#### Scenario: Resolve all tokens in string
- **WHEN** `ResolveAll` is called with a string containing multiple tokens
- **THEN** all known tokens SHALL be replaced with their plaintext values
- **AND** unknown tokens SHALL remain unchanged

### Requirement: Defensive copy semantics
The system SHALL copy values on ingress and egress to prevent external mutation of internal state.

#### Scenario: External mutation after store
- **WHEN** a caller stores a value and then mutates the original byte slice
- **THEN** the internally stored value SHALL remain unchanged

#### Scenario: External mutation after resolve
- **WHEN** a caller resolves a value and then mutates the returned byte slice
- **THEN** subsequent resolves SHALL return the original value

### Requirement: Thread safety
The RefStore SHALL be safe for concurrent access from multiple goroutines.

#### Scenario: Concurrent store and resolve
- **WHEN** multiple goroutines call Store and Resolve concurrently
- **THEN** no data races SHALL occur
- **AND** all stored values SHALL be resolvable after concurrent operations complete

### Requirement: Value enumeration for scanning
The RefStore SHALL provide methods to enumerate stored values for output scanning.

#### Scenario: List all values
- **WHEN** `Values` is called
- **THEN** the system SHALL return copies of all stored plaintext values

#### Scenario: Name mapping
- **WHEN** `Names` is called
- **THEN** the system SHALL return a mapping of plaintext value (as string) to its reference name
