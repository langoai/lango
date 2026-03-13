## Context

The LLM agent processes tool results and generates responses through a shared pipeline. Tool outputs (sometimes exceeding 50K+ characters) flow into model context unfiltered, causing:
1. Model confusion from massive context payloads
2. Raw tool output echo in user-facing responses
3. Internal reasoning markers (`<thought>`, `[INTERNAL]`) leaking to users

The existing middleware pattern (`internal/toolchain/`) and prompt section system (`internal/prompt/`) provide clean extension points for this filtering layer.

## Goals / Non-Goals

**Goals:**
- Prevent oversized tool results from polluting model context (upstream truncation)
- Remove internal content from model responses before they reach users (downstream sanitization)
- Instruct the model to self-regulate output quality (preventive prompt principles)
- Enable per-rule configuration with sensible defaults (enabled by default, zero-config)

**Non-Goals:**
- Streaming chunk state machine (stateful multi-chunk tag tracking deferred to future iteration)
- Content summarization (AI-based summarization of truncated output)
- Per-tool truncation limits (single global limit is sufficient for now)
- Rate limiting or output throttling

## Decisions

### 1. Three-Layer Architecture
**Decision**: Implement prevention (prompt), upstream (truncation), and downstream (sanitization) as independent layers.
**Rationale**: Defense-in-depth. Each layer catches what the others miss. Prompt instructs the model; truncation reduces noise before model sees it; sanitizer catches anything that still leaks through.
**Alternative**: Single post-processing sanitizer only — rejected because large tool outputs would still confuse the model even if cleaned from the response.

### 2. Middleware Pattern for Truncation
**Decision**: Implement truncation as a `toolchain.Middleware` using the existing `Chain/ChainAll` pattern.
**Rationale**: Zero new abstractions needed. Consistent with `WithLearning`, `WithHooks`, `WithApproval`. Applied to all tools uniformly via `ChainAll`.
**Alternative**: Truncation inside the ADK adapter — rejected because it would couple truncation logic to ADK internals.

### 3. Code Block Protection in Sanitizer
**Decision**: Use placeholder-based protection to preserve code blocks during thought tag stripping.
**Rationale**: Simple and effective. Code blocks are extracted, placeholders inserted, tags stripped, then code blocks restored. Avoids complex negative lookahead regex.
**Alternative**: Single regex with negative lookahead — rejected due to complexity and fragility with nested content.

### 4. `*bool` Pointer Config Pattern
**Decision**: Use `*bool` pointers for feature toggles (nil = enabled by default).
**Rationale**: Follows established pattern (`AgentConfig.ErrorCorrectionEnabled`). Users can explicitly disable features without needing to set all defaults.

### 5. Truncation Before Hooks
**Decision**: Place truncation middleware (step 7a) before hooks middleware (step 7b) in the wiring pipeline.
**Rationale**: Truncated output flows to hooks, not the other way around. Hooks (security filter, event publishing) should see the same output the model will see.

## Risks / Trade-offs

- **[False positive stripping]** → Code block protection prevents stripping tags inside code blocks. Marker stripping uses line-start anchoring to minimize false matches.
- **[Truncation data loss]** → Marker `"\n... [output truncated]"` tells the model that data was cut. Default 8000 chars is generous (~2000 tokens).
- **[Chunk sanitization limitations]** → Current implementation applies full `Sanitize()` per chunk, which may miss tags spanning multiple chunks. Accepted trade-off for v1; stateful `ChunkSanitizer` is a future enhancement.
- **[Config complexity]** → Six toggle fields may seem like many, but `*bool` nil-defaults mean zero-config works out of the box. Users only touch config if they need to disable specific rules.
