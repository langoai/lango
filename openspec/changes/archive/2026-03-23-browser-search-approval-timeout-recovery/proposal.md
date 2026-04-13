## Why

Approval replay protection currently keys on raw tool params. For `browser_search`, that makes semantically identical searches diverge into separate approval states when the model changes `limit` or whitespace only. At the same time, a single approval timeout becomes too sticky within the turn, so later recovery attempts can still be replay-blocked even after the user approves a follow-up prompt.

## What Changes

- Canonicalize `browser_search` approval replay keys to query-only semantics.
- Allow approval timeouts to reopen a bounded number of prompts within the same turn before replay-blocking.
- Preserve deny and unavailable outcomes as immediate replay-blocks.
- Add regression tests for canonical browser-search replay and timeout recovery.
- Sync approval docs and README to the new runtime behavior.

## Capabilities

### Modified Capabilities

- `channel-approval`: Turn-local replay now uses canonical browser-search matching and bounded timeout retries.

## Impact

- `internal/approval/turn_state.go`
- `internal/toolchain/mw_approval.go`
- `internal/toolchain/middleware_test.go`
- `docs/security/tool-approval.md`
- `docs/security/approval-cli.md`
- `README.md`
