## Overview

`lango account info` depends on a live smart account manager and several bootstrapped components. A narrow seam around the final info result is enough to make command-level tests deterministic while keeping production behavior unchanged.

## Decisions

### Introduce a small info-result seam

The seam returns the fully-shaped account info payload plus a cleanup callback. Tests can replace it with deterministic fixtures and avoid bootstrapping the live smart account stack.

### Route all non-error output through the Cobra writer

- JSON uses `json.NewEncoder(cmd.OutOrStdout())`
- Table output uses `tabwriter.NewWriter(cmd.OutOrStdout(), ...)`

## Non-Goals

- No change to smart account dependency initialization semantics
- No change to account info payload shape
