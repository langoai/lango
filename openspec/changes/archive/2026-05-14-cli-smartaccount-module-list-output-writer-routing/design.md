## Overview

`lango account module list` depends on a live module registry behind smart account bootstrap. A narrow seam around the final list entries is enough to make command-level tests deterministic while keeping production behavior unchanged.

## Decisions

### Introduce a small module-list seam

The seam returns the fully-shaped module list plus a cleanup callback. Tests can replace it with deterministic fixtures and avoid bootstrapping the live smart account stack.

### Route all non-error output through the Cobra writer

- JSON uses `json.NewEncoder(cmd.OutOrStdout())`
- Table output uses `tabwriter.NewWriter(cmd.OutOrStdout(), ...)`
- Empty-state uses `fmt.Fprintln(cmd.OutOrStdout(), ...)`

## Non-Goals

- No change to module registry semantics
- No change to module entry payload shape
