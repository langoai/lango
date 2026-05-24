## Overview

`RedactedProjection` is a shared helper for plaintext searchable/display projections used by session messages, agent memory, learning entries, inquiries, and recall indexes. It should preserve valid UTF-8 even when applying a byte limit.

## Decisions

### Byte Limit Remains the Contract

The existing `limit` argument is treated as a maximum byte count because callers use it to bound database/search projection columns. The fix does not reinterpret the value as runes.

### Trim to Rune Boundary

When a normalized projection exceeds the limit, truncation walks backward from the byte limit until it reaches a UTF-8 rune boundary. This returns the largest valid UTF-8 prefix within the existing byte limit.

### Focused Test Coverage

Tests live in `internal/security/projection_test.go` and cover:

- Email/number/secret redaction and whitespace normalization.
- Truncating multilingual input without returning invalid UTF-8 or replacement runes.

## Risks

- A projection may be a few bytes shorter than the configured limit if the limit falls inside a multi-byte character. That is the intended tradeoff to preserve valid text.
