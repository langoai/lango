## Why

The first-run experience currently sends mixed signals. The actual `lango onboard` wizard only covers five bootstrap steps, while several docs describe nonexistent onboard submenus for advanced systems such as Embedding & RAG, Graph Store, Multi-Agent, A2A, prompts, and OIDC auth. The command's post-save output is also server-centric and does not clearly point users at the default `lango` workbench or `lango doctor` verification flow.

## What Changes

- Align `lango onboard` post-save guidance with the real primary entry points: `lango`, `lango serve`, `lango doctor`, and `lango settings`.
- Remove or rewrite false documentation that claims advanced feature configuration lives inside extra onboard submenus.
- Keep onboarding documentation explicit that the 5-step wizard is an initial bootstrap path, while advanced configuration lives in `lango settings` or `lango config import/export`.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `cli-onboard`: Refine post-save next-step messaging so first-run guidance matches the real entry points and verification flow.
- `downstream-docs-sync`: Correct README and feature docs so advanced feature setup paths match the actual onboarding and settings surfaces.

## Impact

- Affected code: `internal/cli/onboard/onboard.go`
- Affected tests: `internal/cli/onboard/onboard_test.go`
- Affected docs: `README.md`, `docs/features/embedding-rag.md`, `docs/features/a2a-protocol.md`, `docs/features/knowledge-graph.md`
