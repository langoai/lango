## Why

`lango metrics` and `lango alerts` advertise `--output table|json`, but they do not actually validate that contract before contacting the gateway. They also duplicate the same HTTP/JSON helper logic in separate packages.

## What Changes

- add a shared CLI HTTP/output helper for gateway-backed commands
- validate `--output` early for `lango metrics` and `lango alerts`
- reject unknown output formats before any gateway call
- add regressions and sync docs/specs

## Impact

- more predictable operator and automation UX
- less duplicated CLI plumbing
- fewer unnecessary gateway calls on invalid invocations
