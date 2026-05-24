## Overview

`lango metrics` is an HTTP-backed summary command. Tests can drive it deterministically with an `httptest` server, so no runtime seam is required.

## Decisions

### Route all non-error output through the Cobra writer

- JSON uses `json.NewEncoder(cmd.OutOrStdout())`
- Table output uses `fmt.Fprintln` / `fmt.Fprintf` against `cmd.OutOrStdout()`

### Verify against a real HTTP handler

Tests use a lightweight HTTP server to ensure:

- the right endpoint path is called
- table output is captured by the command writer
- JSON output is captured by the command writer

## Non-Goals

- No change to `/metrics` payload semantics
- No change to how token/tool values are aggregated
