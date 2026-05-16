## Overview

The top-level root binary already uses Cobra, but a few utility subcommands still emit success output directly to process stdout instead of the command writer.

## Decision

- Keep error handling behavior unchanged
- Route only successful `version` and `health` output through `cmd.OutOrStdout()`
- Add command-level tests that capture the command output buffer directly

## Consequences

- Wrapper and harness capture becomes consistent with the rest of the CLI
- Docker-style health checks still observe the same `ok` payload on stdout
