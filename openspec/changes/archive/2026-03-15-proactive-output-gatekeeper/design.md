## Context

The existing output gatekeeper (`WithTruncate`) applies a blunt 8000-character cut to tool results. This loses structured data (JSON arrays, log patterns, stack traces) and wastes model context tokens on irrelevant content. The system needs token-aware, content-sensitive output management.

## Goals / Non-Goals

**Goals:**
- Replace character-based truncation with token-based tiered compression
- Detect content types (JSON, Log, Code, StackTrace) and apply type-specific compression
- Store large outputs for on-demand retrieval via `tool_output_get`
- Add smart file reading (`fs_stat`, `fs_read` offset/limit) to prevent large outputs at the source
- Maintain backward compatibility (`WithTruncate` kept but unwired)

**Non-Goals:**
- Persistent output storage (in-memory TTL only)
- External tokenizer integration (using `types.EstimateTokens` character heuristic)
- Streaming/chunked output delivery

## Decisions

### 1. Three-tier classification (Small/Medium/Large)
Tiers are based on token count relative to budget: Small (≤budget), Medium (≤3×budget), Large (>3×budget). Each tier gets progressively more aggressive compression. **Why**: A single threshold creates a jarring cliff; three tiers provide graceful degradation.

### 2. Content-aware compression via `tooloutput.Compress()`
The middleware detects content type and delegates to type-specific compressors (CompressJSON, CompressLog, CompressCode, CompressStackTrace, CompressHeadTail). **Why**: Generic head/tail compression destroys structure in JSON arrays and loses critical ERROR lines in logs. Type-specific strategies preserve the most relevant information.

### 3. `_meta` injection as top-level field (not envelope wrapping)
Original string results become `{"content": "...", "_meta": {...}}`. Map results get `_meta` injected directly. **Why**: Envelope wrapping would break P2P executor compatibility. Top-level injection keeps the result shape predictable.

### 4. `OutputStorer` interface for store injection
The middleware accepts an optional `OutputStorer` interface rather than a concrete `*OutputStore`. **Why**: Testability (mock store in tests) and decoupling (middleware doesn't import lifecycle concerns).

### 5. In-memory TTL store with lifecycle.Component
`OutputStore` uses a sync.RWMutex map with background cleanup every TTL/2. Implements `lifecycle.Component` for graceful startup/shutdown. **Why**: Simple, no external dependencies. 10-minute TTL is sufficient since agents act on results immediately. The lifecycle pattern ensures the cleanup goroutine is properly managed.

### 6. `fs_stat` + offset/limit as source-side prevention
Rather than only compressing at the output, smart reading prevents large outputs from being generated. `fs_stat` lets the agent check file size before reading; `offset/limit` enables paginated reading. **Why**: Prevention is cheaper than compression.

## Risks / Trade-offs

- [Token estimation is approximate] → Acceptable for budget-based decisions; exact tokenization would require model-specific tokenizers and add latency.
- [In-memory store loses data on restart] → Acceptable; tool outputs are ephemeral and can be re-generated. No persistence needed.
- [Content detection may misclassify] → Falls back to generic CompressHeadTail which is always safe. Misclassification degrades quality, not correctness.
- [Every tool call pays detection overhead] → Fast path added: `len(text) < budget*4` skips full token estimation. Detection uses compiled package-level regexes.
