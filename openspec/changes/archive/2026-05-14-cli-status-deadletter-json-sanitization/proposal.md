## Why

`lango status dead-letters --output json` and `lango status dead-letter <id> --output json` still serialize raw bridge payloads. That leaves list/detail JSON outputs inconsistent with the replay-safe summary and retry result paths.

## What Changes

- sanitize dead-letter backlog JSON payload copies before serialization
- sanitize dead-letter detail JSON payload copies before serialization
- add regression coverage for malformed list/detail JSON fields
- sync the CLI status spec and docs with the new JSON output contract

## Impact

- closes two remaining raw-text paths in production status operator JSON output
- keeps bridge/runtime models intact while making CLI serialization safe
