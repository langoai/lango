## Overview

`lango economy escrow show` is a config-backed inspection command with three user-visible paths: detailed config, disabled state, and ID-guidance.

## Decisions

### Route all non-error output through the Cobra writer

The show command uses `fmt.Fprintln` / `fmt.Fprintf` against `cmd.OutOrStdout()` for all three paths.

### Cover the highest-signal visible paths directly

Tests assert:

- detailed config output
- disabled-state output

Existing guidance tests already cover the `--id` path and continue to exercise the same text.

## Non-Goals

- No change to escrow show data shape
- No change to live escrow guidance semantics
