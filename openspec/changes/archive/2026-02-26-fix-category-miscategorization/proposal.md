## Why

`mapCategory()` and `mapKnowledgeCategory()` silently fall back to `CategoryFact` for any unrecognized LLM output type. This misclassified data is later injected into the system prompt as `[fact] key: content`, causing the agent to treat unverified information as established fact — a direct hallucination vector. The same unsafe pattern exists in 5 locations across the codebase.

## What Changes

- `mapCategory()` in `proactive_buffer.go`: returns `(Category, error)` instead of silently defaulting; adds `"pattern"` and `"correction"` cases
- `mapKnowledgeCategory()` in `parse.go`: same signature change and case additions
- Caller updates in `conversation_analyzer.go` and `session_learner.go`: error handling for new signature (future-proofing)
- `InquiryProcessor` in `inquiry_processor.go`: replaces raw `entknowledge.Category()` cast with validated `mapCategory()` call
- `save_knowledge` tool in `tools.go`: adds `"pattern"` and `"correction"` to enum, validates via `CategoryValidator` before cast
- Observation analyzer prompt: adds `pattern|correction` to allowed type list
- New table-driven tests for both `mapCategory` and `mapKnowledgeCategory`

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `knowledge-store`: Category mapping functions now return errors for unrecognized types instead of silent fallback
- `proactive-librarian`: Extraction pipeline skips entries with unrecognized types and logs warnings
- `meta-tools`: `save_knowledge` tool validates category before saving and supports `pattern`/`correction`

## Impact

- `internal/librarian/proactive_buffer.go` — signature change + caller update
- `internal/learning/parse.go` — signature change
- `internal/learning/conversation_analyzer.go` — error handling added
- `internal/learning/session_learner.go` — error handling added
- `internal/librarian/inquiry_processor.go` — raw cast replaced with validated call
- `internal/app/tools.go` — enum expanded + validation added
- `internal/librarian/observation_analyzer.go` — prompt updated
- New test files: `proactive_buffer_test.go`, `parse_test.go`
