## Why

`internal/sandbox.RunWorker` currently calls `os.Exit` directly, which makes worker behavior hard to test and violates the repository convention that process termination should be centralized at binary entrypoints. The worker protocol already has explicit exit-code semantics, so the same behavior can be preserved with a returned code.

## What Changes

- Make sandbox worker execution return an exit code instead of terminating the process directly.
- Preserve worker JSON stdout protocol and existing success/error exit-code semantics.
- Update `cmd/lango` worker-mode handling to return the worker exit code.
- Add unit tests for worker decode, missing tool, tool error, and success paths.

## Non-Goals

- No changes to container image command flags.
- No changes to subprocess request/response JSON schema.
- No changes to sandbox runtime selection or container fallback behavior.
