## Overview

`lango p2p sandbox` touches external runtimes. To keep tests deterministic, this change introduces small constructor seams around the container executor, subprocess executor, and Docker cleanup runtime.

## Decisions

### Add narrow constructor seams

- container executor constructor for status/test
- subprocess executor constructor for smoke-test fallback
- Docker runtime constructor for cleanup

These seams are only for command-level verification and do not change production behavior.

### Route all non-error output through the Cobra writer

- status text uses `fmt.Fprintln` / `fmt.Fprintf` against `cmd.OutOrStdout()`
- smoke-test progress and success output use `fmt.Fprintln` / `fmt.Fprintf` against `cmd.OutOrStdout()`
- cleanup success output uses `fmt.Fprintln(cmd.OutOrStdout(), ...)`

## Non-Goals

- No change to sandbox runtime semantics
- No change to Docker cleanup logic beyond testability seams
