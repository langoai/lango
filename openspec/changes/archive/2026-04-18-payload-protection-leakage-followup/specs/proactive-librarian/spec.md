## ADDED Requirements

### Requirement: Inquiry payload bundle preserves all fields
Inquiry persistence MUST store `{question, context, answer}` as one protected bundle so resolve operations preserve the original question and context.

#### Scenario: Save inquiry stores question and context in payload bundle
- **WHEN** `SaveInquiry` persists a pending inquiry
- **THEN** the protected payload bundle contains `question` and `context`
- **AND** the plaintext `question` and `context` columns store only redacted projections

#### Scenario: Resolve inquiry preserves question and context
- **WHEN** `ResolveInquiry` stores an answer for an existing inquiry
- **THEN** it first restores the existing bundle
- **AND** it writes back a new protected bundle containing `question`, `context`, and `answer`
- **AND** the plaintext `answer` column stores only a redacted projection

#### Scenario: Inquiry read helpers decrypt protected bundles
- **WHEN** inquiry list or read helpers return a row with ciphertext
- **THEN** they return decrypted `question`, `context`, and `answer` values from the protected bundle

#### Scenario: Inquiry protected row decrypt failure does not fall back
- **WHEN** an inquiry row has ciphertext fields present but decryption fails
- **THEN** the store returns an error or empty protected values
- **AND** it does not promote the stored plaintext projections as originals
