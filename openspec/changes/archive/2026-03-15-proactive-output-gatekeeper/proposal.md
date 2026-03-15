## Why

The existing `WithTruncate` middleware uses a blunt 8000-character string cut that discards structure and context. This wastes model tokens on irrelevant tail data while losing important error messages and patterns. A token-based, content-aware output management system preserves the most relevant information within a configurable token budget.

## What Changes

- Replace `WithTruncate` character-based truncation with `WithOutputManager` token-based tiered middleware
- Add content-aware compression library (`tooloutput`) that detects JSON/Log/Code/StackTrace content and applies type-specific compression
- Add in-memory TTL output store so agents can retrieve full compressed outputs via `tool_output_get`
- Add `fs_stat` tool and `offset`/`limit` parameters to `fs_read` for smart file reading
- Add `OutputManagerConfig` to config system with `tokenBudget`, `headRatio`, `tailRatio` settings

## Capabilities

### New Capabilities
- `proactive-output-gatekeeper`: Token-based tiered output management with content-aware compression, output store, and retrieval tool

### Modified Capabilities
- `output-gatekeeper`: Extended with token-based management replacing character truncation
- `tool-filesystem`: Added `fs_stat` tool and `offset`/`limit` parameters to `fs_read`

## Impact

- `internal/toolchain/mw_output_manager.go` — new middleware
- `internal/tooloutput/` — new package (detect, compress, store)
- `internal/app/app.go` — wiring change (WithTruncate → WithOutputManager)
- `internal/app/tools_output.go` — new tool_output_get
- `internal/app/tools_filesystem.go` — fs_stat + fs_read enhancement
- `internal/tools/filesystem/filesystem.go` — Stat/ReadWithMeta methods
- `internal/config/types.go` — OutputManagerConfig
- `prompts/TOOL_USAGE.md` — new tool documentation
