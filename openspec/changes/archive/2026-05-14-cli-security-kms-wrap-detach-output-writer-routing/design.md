## Overview

`lango security kms wrap` and `detach` mutate the envelope and emit only a small amount of operator-facing output. The existing KMS provider seam is enough for wrap tests; detach can be exercised with a local envelope fixture.

## Decisions

### Reuse the KMS provider seam for wrap tests

The existing constructor seam allows tests to inject a reversible fake KMS provider and avoid any real cloud dependency.

### Use local envelope fixtures for detach tests

Tests create a local envelope, add one or more KMS slots, and then exercise:

- single-slot detach success
- multi-slot `--slot-id` guidance

### Route all non-error output through the Cobra writer

- wrap success uses `fmt.Fprintf` / `fmt.Fprintln` against `cmd.OutOrStdout()`
- detach success and multi-slot guidance use `fmt.Fprintf` / `fmt.Fprintln` against `cmd.OutOrStdout()`

## Non-Goals

- No change to envelope slot semantics
- No change to KMS provider selection rules
