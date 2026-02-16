## Context

The system prompt is currently a single hardcoded string (`_defaultSystemPrompt`) in `internal/app/wiring.go`. It contains only an identity statement with no conversation behavior rules. When the full conversation history is passed to the LLM, it tends to repeat and accumulate prior answers. The `ContextAwareModelAdapter` dynamically appends knowledge/memory/RAG sections at runtime, but the static base prompt has no extensibility.

## Goals / Non-Goals

**Goals:**
- Introduce a section-based prompt builder (`internal/prompt`) with priority ordering
- Include conversation rules by default to prevent answer repetition
- Support file-based section overrides from a configurable directory
- Maintain backward compatibility with the existing `SystemPromptPath` config

**Non-Goals:**
- Dynamic per-request prompt modification (handled by `ContextAwareModelAdapter`)
- Per-session or per-user prompt customization
- Template variable interpolation within prompt sections

## Decisions

### 1. Interface-based sections with priority ordering
**Choice**: `PromptSection` interface with `ID()`, `Priority()`, `Render()` methods; sections sorted by priority at build time.
**Rationale**: Enables clean replacement by ID (last-writer-wins) while maintaining deterministic output order. Alternative of ordered maps was considered but provides less flexibility for custom section types.

### 2. StaticSection as the primary implementation
**Choice**: Single `StaticSection` struct for all current use cases.
**Rationale**: All sections are currently static text. The interface allows future dynamic sections (e.g., time-aware, session-aware) without changing the builder.

### 3. Directory-based loader with known filename mapping
**Choice**: Map specific filenames (AGENTS.md, SAFETY.md, etc.) to section IDs; unknown `.md` files become custom sections at priority 900+.
**Rationale**: Simple convention-over-configuration. Users only need to create a file with the right name. Alternative of a manifest file was rejected as unnecessarily complex.

### 4. Three-tier precedence: PromptsDir > SystemPromptPath > defaults
**Choice**: `buildPromptBuilder()` checks `PromptsDir` first, then falls back to legacy `SystemPromptPath` (replacing only Identity), then built-in defaults.
**Rationale**: Full backward compatibility. Existing configs work unchanged; new `PromptsDir` provides full control.

### 5. Builder passed to ContextAwareModelAdapter constructor
**Choice**: Constructor accepts `*prompt.Builder` and calls `Build()` internally to set `basePrompt`.
**Rationale**: Keeps the dynamic context injection logic unchanged. The builder is only used once at construction time.

## Risks / Trade-offs

- **[Risk]** File-based overrides could introduce empty or malformed prompts → **Mitigation**: Empty files are ignored; loader logs warnings for read errors
- **[Risk]** Custom sections may conflict with dynamic sections appended by `ContextAwareModelAdapter` → **Mitigation**: Custom sections have priority 900+ (after all defaults); dynamic sections are appended after `Build()` output
- **[Trade-off]** No hot-reload of prompt files — changes require restart → Acceptable for current use case; hot-reload can be added later if needed
