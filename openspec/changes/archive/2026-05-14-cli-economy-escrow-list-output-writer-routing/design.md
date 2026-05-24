## Overview

`lango economy escrow list` is a config-backed inspection command with three user-visible paths: summary output, economy-disabled state, and escrow-disabled state.

## Decisions

### Route all non-error output through the Cobra writer

The list command uses `fmt.Fprintln` / `fmt.Fprintf` against `cmd.OutOrStdout()` for all three paths.

### Cover all visible output paths directly

Tests assert:

- summary output with on-chain details
- economy-disabled output
- escrow-disabled output

## Non-Goals

- No change to escrow summary semantics
- No change to on-chain config shape
