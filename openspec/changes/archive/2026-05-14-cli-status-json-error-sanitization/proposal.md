## Why

`lango status --output json` still emits top-level error payloads using raw `err.Error()` text. That leaves one remaining status CLI path where ANSI/OSC control sequences or embedded newlines can leak into machine-readable output.

## What Changes

- sanitize top-level JSON error text before serializing `statusJSONError`
- add regression coverage for malformed retry-command JSON errors
- sync the CLI status spec and docs to include JSON error payload normalization

## Impact

- closes the remaining raw-text gap on `lango status` JSON error output
- keeps error payloads stable and replay-safe for downstream automation
