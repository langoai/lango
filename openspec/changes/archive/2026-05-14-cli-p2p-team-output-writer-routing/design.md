## Overview

`lango p2p team` is guidance-oriented and does not perform live team control. That means output-routing changes are low risk: only the rendering boundary changes.

## Decisions

### Route all non-error output through the Cobra writer

- `list --json` and `status --json` use `json.NewEncoder(cmd.OutOrStdout())`
- Text guidance output uses `fmt.Fprintln` against `cmd.OutOrStdout()`

### Extend tests to cover JSON guidance paths

Existing tests already cover the text guidance wording. This change adds JSON command-level assertions to close the capture gap.

## Non-Goals

- No change to the guidance wording beyond writer-routing
- No change to team lifecycle semantics
