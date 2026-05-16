## Overview

`lango p2p workspace` is guidance-oriented and only exposes a few JSON-producing paths. That keeps the implementation narrow: only the rendering boundary changes.

## Decisions

### Route all non-error output through the Cobra writer

- JSON paths use `json.NewEncoder(cmd.OutOrStdout())`
- Text guidance output uses `fmt.Fprintln` / `fmt.Fprintf` against `cmd.OutOrStdout()`

### Extend tests to cover JSON guidance paths

Existing tests already cover the text guidance wording. This change adds command-level assertions for `create --json`, `list --json`, and `status --json`.

## Non-Goals

- No change to workspace guidance wording
- No change to workspace lifecycle semantics
