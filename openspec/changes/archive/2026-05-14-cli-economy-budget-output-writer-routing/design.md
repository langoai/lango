## Overview

`lango economy budget status` is a config-backed inspection command with three user-visible paths: enabled configuration, disabled state, and task-specific live guidance.

## Decisions

### Route all non-error output through the Cobra writer

The status command uses `fmt.Fprintln` / `fmt.Fprintf` against `cmd.OutOrStdout()` for all three paths.

### Cover all visible output paths directly

Tests assert:

- enabled configuration output
- task guidance output
- disabled-state output

## Non-Goals

- No change to budget config shape
- No change to live task-budget semantics
