## Why

The root `lango status --output json` path still assigns `tui.GetVersion()` directly into the serialized status model. If injected build metadata contains control text or embedded newlines, the version field can bypass the otherwise sanitized JSON baseline.

## What Changes

- sanitize the root status version string before storing it in `StatusInfo`
- add command-level regression coverage for malformed injected version text
- sync the status spec and docs with the top-level version sanitization contract

## Impact

- closes the remaining top-level raw-text gap in root status JSON output
- keeps build-time metadata safe for wrappers and automated consumers
