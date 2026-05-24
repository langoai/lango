## Overview

`lango economy risk status` is a config-backed inspection command with two simple states: enabled and disabled. No seam is required; tests can execute the command directly with synthetic config.

## Decisions

### Route all non-error output through the Cobra writer

The status command uses `fmt.Fprintln` / `fmt.Fprintf` against `cmd.OutOrStdout()` for both enabled and disabled paths.

### Cover both output states directly

Tests assert:

- enabled configuration output
- disabled-state output

## Non-Goals

- No change to risk config shape
- No change to threshold semantics
