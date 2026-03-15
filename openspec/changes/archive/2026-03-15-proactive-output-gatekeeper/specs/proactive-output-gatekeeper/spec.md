## ADDED Requirements

### Requirement: Token-based output tier classification
The system SHALL classify tool output into three tiers based on estimated token count relative to the configured token budget: Small (tokens ≤ budget), Medium (budget < tokens ≤ 3×budget), Large (tokens > 3×budget).

#### Scenario: Small output passes through
- **WHEN** a tool returns output with estimated tokens ≤ the token budget
- **THEN** the output SHALL pass through uncompressed with `_meta.tier` set to `"small"` and `_meta.compressed` set to `false`

#### Scenario: Medium output is compressed
- **WHEN** a tool returns output with estimated tokens between budget and 3×budget
- **THEN** the output SHALL be compressed using content-aware compression and `_meta.tier` SHALL be `"medium"`

#### Scenario: Large output is aggressively compressed and stored
- **WHEN** a tool returns output with estimated tokens > 3×budget
- **THEN** the output SHALL be compressed with half the budget, stored in the output store, and `_meta.storedRef` SHALL contain the storage UUID

### Requirement: Content-aware compression
The system SHALL detect the content type of tool output and apply type-specific compression: CompressJSON for JSON, CompressLog for logs, CompressCode for code, CompressStackTrace for stack traces, and CompressHeadTail as a generic fallback.

#### Scenario: JSON array compression
- **WHEN** a JSON array exceeds the token budget
- **THEN** the system SHALL show the first 2 items and a total count

#### Scenario: Log compression
- **WHEN** log output exceeds the token budget
- **THEN** the system SHALL extract ERROR/WARN lines first, then apply head/tail compression

#### Scenario: Stack trace compression
- **WHEN** a stack trace exceeds the token budget
- **THEN** the system SHALL keep the first goroutine/thread block and summarize the rest

### Requirement: Content type detection
The system SHALL detect content types using heuristics: JSON (starts with `{`/`[`), StackTrace (goroutine/panic/traceback patterns), Log (≥2 timestamps + ≥2 log levels), Code (≥2 syntax keywords), Text (default).

#### Scenario: JSON detection
- **WHEN** trimmed output starts with `{` or `[`
- **THEN** content type SHALL be `"json"`

#### Scenario: Log detection requires thresholds
- **WHEN** output contains ≥2 timestamp patterns and ≥2 log level keywords
- **THEN** content type SHALL be `"log"`

### Requirement: Output metadata injection
The system SHALL inject `_meta` as a top-level field on all processed results. For string results, the output SHALL be wrapped as `{"content": "...", "_meta": {...}}`. For map results, `_meta` SHALL be injected directly.

#### Scenario: String result wrapping
- **WHEN** the original tool result is a string
- **THEN** the result SHALL be transformed to `{"content": "<original>", "_meta": {...}}`

#### Scenario: Map result injection
- **WHEN** the original tool result is a map
- **THEN** `_meta` SHALL be added as a key directly in the map

### Requirement: Output store with TTL
The system SHALL provide an in-memory store for large tool outputs with configurable TTL. The store SHALL implement `lifecycle.Component` with a background cleanup goroutine.

#### Scenario: Store and retrieve
- **WHEN** a large output is stored
- **THEN** the full content SHALL be retrievable by the returned UUID reference

#### Scenario: TTL expiry
- **WHEN** a stored output exceeds the TTL duration
- **THEN** the entry SHALL be evicted and retrieval SHALL return not-found

### Requirement: Output retrieval tool
The system SHALL provide a `tool_output_get` tool with three modes: `full` (returns entire stored content), `range` (returns lines at offset/limit), and `grep` (returns lines matching a regex pattern).

#### Scenario: Full retrieval
- **WHEN** `tool_output_get` is called with mode `full` and a valid ref
- **THEN** the full stored content SHALL be returned

#### Scenario: Range retrieval
- **WHEN** `tool_output_get` is called with mode `range`, offset, and limit
- **THEN** the specified line range SHALL be returned with total line count

#### Scenario: Grep retrieval
- **WHEN** `tool_output_get` is called with mode `grep` and a pattern
- **THEN** matching lines SHALL be returned

### Requirement: Configuration
The system SHALL support `tools.outputManager` configuration with `enabled` (*bool, default true), `tokenBudget` (int, default 2000), `headRatio` (float64, default 0.7), and `tailRatio` (float64, default 0.3).

#### Scenario: Disabled output manager
- **WHEN** `tools.outputManager.enabled` is set to false
- **THEN** tool output SHALL pass through without any processing
