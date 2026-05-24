## MODIFIED Requirements

### Requirement: Context panel optimized snapshot handling
The ContextPanel SHALL reuse existing slice capacity in SetChannelStatuses() instead of allocating a new slice on every call. Style variables for render methods SHALL be pre-allocated at module level. The toolCountSum SHALL be cached alongside the sortedTools dirty flag.

#### Scenario: Cached tool labels are replay-safe
- **WHEN** tool names contain ANSI/OSC escape sequences or embedded newlines before entering the cached `sortedTools` slice
- **THEN** the context panel SHALL strip those control sequences
- **AND** it SHALL normalize the stored cached tool text to a single line before replay
