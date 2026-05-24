## Why

The chat shell header assumes `cfg` is non-nil and dereferences `cfg.Agent` unconditionally. If a boundary path constructs chat with a missing config, the renderer can panic before any recovery UI is available.

## What Changes

- Make `renderHeader()` nil-safe and fall back to the existing `default` / `auto` labels when config is absent.
- Add regression coverage for nil-config header rendering.
- Record the renderer hardening in OpenSpec.

## Impact

- Prevents shell-header crashes in degraded or partially wired startup paths.
- Keeps the chat shell fail-closed even when config injection is missing.
