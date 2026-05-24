## ADDED Requirements

### Requirement: Agent memory stores redacted projections only
Persistent agent memory MUST store original content as ciphertext and keep only a redacted searchable projection in the plaintext `content` column.

#### Scenario: Agent memory save stores redacted projection
- **WHEN** an agent memory entry is created or updated with payload protection enabled
- **THEN** the original content is stored as ciphertext
- **AND** the plaintext `content` column stores only a redacted projection

#### Scenario: Agent memory search uses projection only
- **WHEN** `Search` or `SearchWithContext` is executed against persistent agent memory
- **THEN** matching is performed against `key` and the plaintext projection only
- **AND** original protected content is not required for search predicates

#### Scenario: Agent memory read decrypts protected content
- **WHEN** `Get`, `ListAll`, `Search`, or `SearchWithContext` returns a row with ciphertext
- **THEN** the returned domain entry uses decrypted original content

#### Scenario: Agent memory protected row decrypt failure does not fall back
- **WHEN** an agent memory row has ciphertext fields present but decryption fails
- **THEN** the store returns an error or empty protected values
- **AND** it does not promote the stored plaintext projection as the original content
