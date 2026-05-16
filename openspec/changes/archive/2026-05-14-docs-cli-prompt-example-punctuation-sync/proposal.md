## Why

Several public CLI examples still show stale prompt punctuation such as `[y/N] y` even though the actual commands now render `[y/N]: y` through the shared prompt helper. That small mismatch is enough to erode trust in the docs for wrappers and operators.

## What Changes

- Update stale CLI prompt examples in public docs to match the current `: ` punctuation emitted by command-stream-backed prompts
- Keep the change documentation-only; no runtime behavior changes

## Capabilities

### New Capabilities

### Modified Capabilities
- `docs-only`: CLI prompt examples in public docs reflect the current command output exactly

## Impact

- Affected docs: `docs/cli/security.md`, `docs/cli/agent-memory.md`
- No code/runtime changes
