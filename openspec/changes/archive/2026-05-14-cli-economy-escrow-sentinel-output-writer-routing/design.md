## Overview

`lango economy escrow sentinel status` is a config-backed status command with three user-visible states:

- active sentinel
- on-chain escrow disabled
- escrow disabled

## Decisions

### Route all non-error output through the Cobra writer

The status command uses `fmt.Fprintln` / `fmt.Fprintf` against `cmd.OutOrStdout()` for all three paths.

### Reuse existing active-state guidance coverage

The existing sentinel guidance test already verifies the active path wording. This change adds explicit capture coverage for the remaining disabled-state paths.

## Non-Goals

- No change to sentinel wording
- No change to sentinel/on-chain gating semantics
