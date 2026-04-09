## Purpose

Capability spec for fts5-query-safety. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: FTS5 token sanitization strips all punctuation
The `escapeFTS5Token` function SHALL remove all ASCII punctuation characters that can cause FTS5 syntax errors, including but not limited to: `.`, `?`, `!`, `@`, `#`, `$`, `%`, `&`, `=`, `|`, `~`, `<`, `>`, `;`, `,`, `[`, `]`, `\`, `/`, `'`.

#### Scenario: Single punctuation character query
- **WHEN** a search query consisting only of `?` is passed to `sanitizeFTS5Query`
- **THEN** the function SHALL return an empty string (no valid terms)

#### Scenario: Mixed punctuation and text
- **WHEN** a search query `hello? world.` is passed to `sanitizeFTS5Query`
- **THEN** the function SHALL produce `hello OR world` with punctuation stripped

### Requirement: Empty tokens are excluded after sanitization
The `sanitizeFTS5Query` function SHALL skip tokens that become empty after `escapeFTS5Token` processing, in all three code paths: normal tokens, prefix tokens (trailing `*`), and unclosed-quote fallback tokens.

#### Scenario: Prefix token with only punctuation
- **WHEN** a search query `?*` is passed to `sanitizeFTS5Query`
- **THEN** the function SHALL return an empty string (the escaped base `?` becomes empty, so `*` is not appended)

#### Scenario: Unclosed quote with only punctuation
- **WHEN** a search query `"?` is passed to `sanitizeFTS5Query`
- **THEN** the function SHALL return an empty string

### Requirement: Bulk index uses correct Ent table names
The `bulkIndexKnowledge` function SHALL query from the `knowledges` table (Ent-generated plural form), not `knowledge`.

#### Scenario: Knowledge bulk index on existing database
- **WHEN** `bulkIndexKnowledge` runs against a database with the Ent `knowledges` table
- **THEN** the query SHALL succeed without "no such table" errors

### Requirement: Bulk index handles nullable columns
The `bulkIndexLearnings` function SHALL use `COALESCE` for all nullable columns (`error_pattern`, `fix`) to prevent scan errors when values are NULL.

#### Scenario: Learning entry with NULL error_pattern and fix
- **WHEN** `bulkIndexLearnings` encounters a row where `error_pattern` and `fix` are both NULL
- **THEN** the scan SHALL succeed with empty strings substituted
