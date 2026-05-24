## Overview

The recovery commands mix envelope logic with direct terminal output, and setup also uses a process-global confirmation-word prompt. Introducing narrow execution seams plus a stream-aware confirmation helper is enough to make the command surfaces deterministic while preserving the security flow.

## Decisions

### Introduce setup and restore execution seams

The seams own the existing recovery flows in production and can be stubbed in tests to validate command writer capture without invoking real prompts or envelope operations.

### Make confirmation-word prompts command-stream aware

`confirmWord(...)` now accepts an input reader and output writer so the setup flow can honor `cmd.InOrStdin()` and `cmd.OutOrStdout()`.

## Non-Goals

- No change to mnemonic generation or recovery semantics
- No change to keyfile/keyring warning behavior on restore
