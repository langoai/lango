## Overview

The prompt package is the root of a remaining stream-routing gap: `Confirm(...)` always talks to global stdin/stdout. Adding a stream-aware variant lets command handlers opt into Cobra-managed streams while preserving the existing call sites that still rely on global TTY behavior.

## Decisions

### Add a non-breaking `ConfirmIO(...)` helper

`Confirm(...)` remains as a thin wrapper around `ConfirmIO(os.Stdin, os.Stdout, ...)` so existing behavior is unchanged outside of the migrated call sites.

### Migrate recovery setup only in this change

Recovery setup is the immediate consumer because it already aims to keep non-error output on the command writer. Broader migration of other callers can follow separately.

## Non-Goals

- No redesign of passphrase prompts
- No change to bootstrap's existing global confirmation behavior in this change
