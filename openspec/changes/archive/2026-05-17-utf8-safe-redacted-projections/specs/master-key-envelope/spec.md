## MODIFIED Requirements

### Requirement: Payload projection text remains UTF-8 safe

Payload-protection redacted projections SHALL preserve valid UTF-8 when truncating text for plaintext search/display columns. Projection byte limits SHALL remain upper bounds, and truncation SHALL return the largest valid UTF-8 prefix that does not exceed the configured limit.

#### Scenario: Multibyte projection truncation remains valid
- **WHEN** a redacted projection containing multibyte UTF-8 text is truncated with a byte limit inside a multibyte rune
- **THEN** the returned projection SHALL remain valid UTF-8
- **AND** it SHALL NOT include replacement runes caused by splitting a character
- **AND** its byte length SHALL NOT exceed the configured limit

#### Scenario: Redaction still occurs before truncation
- **WHEN** a projection contains email addresses, long numbers, or long secret-like tokens
- **THEN** those sensitive values SHALL be replaced before any truncation limit is applied
