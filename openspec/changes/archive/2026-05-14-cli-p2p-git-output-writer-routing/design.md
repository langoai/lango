## Overview

`lango p2p git` is guidance-oriented and only exposes one JSON-producing path today: `log --json`. That keeps the implementation narrow.

## Decisions

### Route all non-error output through the Cobra writer

- Guidance text uses `fmt.Fprintln` against `cmd.OutOrStdout()`
- `log --json` uses `json.NewEncoder(cmd.OutOrStdout())`

### Extend tests minimally

Existing tests already assert truthful guidance text. This change adds one command-level JSON capture test for `git log --json`.

## Non-Goals

- No change to git guidance wording
- No change to the set of git guidance commands
