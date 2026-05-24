## ADDED Requirements

### Requirement: Session message payload protection coverage
Session message persistence MUST store original content and sensitive tool-call payloads as ciphertext while keeping only redacted projections in plaintext columns.

#### Scenario: Session message content stored as ciphertext
- **WHEN** a message is created, appended, or rewritten through session storage
- **THEN** the original message text is stored in `content_ciphertext`
- **AND** the plaintext `content` column stores only a redacted projection

#### Scenario: Tool-call payload projection excludes sensitive fields
- **WHEN** a message with tool calls is persisted
- **THEN** the plaintext `tool_calls` JSON stores only `id`, `name`, and `thought_signature`
- **AND** `input`, `output`, and `thought` are not stored in plaintext
- **AND** the original tool-call payloads are stored in ciphertext fields

#### Scenario: Session reload decrypts protected payloads
- **WHEN** a stored message row has ciphertext fields present
- **THEN** session reload decrypts and returns the original content and tool-call payloads
- **AND** plaintext projection values are not promoted back to the domain model as originals

#### Scenario: Legacy plaintext rows still load
- **WHEN** a stored message row has no ciphertext fields
- **THEN** session reload may use the existing plaintext columns as a legacy fallback

#### Scenario: Protected row decrypt failure does not fall back
- **WHEN** a stored message row has ciphertext fields but decryption fails
- **THEN** the store returns an error or empty protected value
- **AND** it does not promote the plaintext projection as the original content

### Requirement: Session compaction uses protected summary messages
Compaction-generated summary messages MUST follow the same payload-protection rules as regular messages.

#### Scenario: Compaction summary stored as protected message
- **WHEN** session compaction rewrites earlier messages into a summary message
- **THEN** the original summary text is stored as ciphertext
- **AND** the plaintext column stores only a redacted projection
