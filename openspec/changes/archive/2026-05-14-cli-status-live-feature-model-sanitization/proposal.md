## Why

`lango status --output json` still preserves raw live health feature text inside `serverInfo.features`. Top-level status fields are sanitized already, but live feature names, reasons, and suggestions can still carry ANSI/OSC control text or embedded newlines from the health endpoint.

## What Changes

- sanitize live feature status strings before storing them in `StatusInfo.ServerInfo`
- add regression coverage using a real `/health` response fixture
- sync the status spec and docs with the replay-safe live feature model contract

## Impact

- keeps root status JSON output consistent with the rest of the sanitized status model
- protects downstream automation that reads live health feature metadata
