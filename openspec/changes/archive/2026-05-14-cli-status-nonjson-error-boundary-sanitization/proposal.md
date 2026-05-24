## Why

Even after hardening many individual status surfaces, the root non-JSON error boundary still returns raw downstream errors. That means any missed subcommand or bootstrap error can still leak control text directly into operator-visible CLI failures.

## What Changes

- sanitize all non-JSON status command errors at the shared boundary
- preserve `errors.Is`/`errors.Unwrap` behavior for callers and tests
- add regression coverage for malformed bootstrap errors
- sync the status spec and docs with the shared non-JSON error contract

## Impact

- provides a fail-safe safety net for operator-facing status errors
- keeps structured cause identity intact for tests and programmatic checks
