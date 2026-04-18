## ADDED Requirements

### Requirement: Learning payload protection coverage
Learning entries MUST store `{error_pattern, diagnosis, fix}` as a protected bundle while keeping only redacted projections in the plaintext columns.

#### Scenario: Learning save stores protected bundle
- **WHEN** `SaveLearning` persists a learning entry
- **THEN** `error_pattern`, `diagnosis`, and `fix` are serialized into one protected bundle
- **AND** the plaintext columns store only redacted projections
- **AND** `trigger` remains plaintext

#### Scenario: Learning read decrypts protected fields
- **WHEN** `GetLearning`, `SearchLearnings`, or `SearchLearningsScored` returns a row with ciphertext
- **THEN** the returned domain object uses decrypted `error_pattern`, `diagnosis`, and `fix`
- **AND** the plaintext projections are not promoted as original values

#### Scenario: Learning FTS uses redacted projections only
- **WHEN** learning FTS rows are written or refreshed
- **THEN** the indexed content contains only redacted projection text
- **AND** original `error_pattern`, `diagnosis`, and `fix` values do not appear in plaintext FTS rows

#### Scenario: Learning protected row decrypt failure does not fall back
- **WHEN** a learning row has ciphertext fields present but decryption fails
- **THEN** the store returns an error or empty protected values
- **AND** it does not promote the stored plaintext projections as originals
