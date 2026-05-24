## Overview

The remaining metrics breakdown commands all share the same rendering model: fetch one JSON endpoint, optionally emit raw JSON, otherwise render a small table or an empty-state message. That makes a shared writer-aware helper the right boundary.

## Decisions

### Make shared helpers writer-aware

- `printJSON` now accepts an `io.Writer`
- `newTabWriter` now accepts an `io.Writer`

This allows subcommands to keep their logic while routing through Cobra-managed streams.

### Cover representative command paths with `httptest`

Tests use lightweight HTTP handlers to verify:

- sessions table output
- sessions JSON output
- tools empty-state output
- agents empty-state output
- history table output

## Non-Goals

- No change to `/metrics/*` payload semantics
- No change to truncation or numeric formatting logic
