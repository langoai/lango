## Why

`lango security keyring clear` still used process stdout/stderr and `os.Stdin` directly for its confirmation prompt and result messages. That made interactive automation inconsistent with the hardened CLI command stream contract.

## What Changes

- route `keyring clear` prompt and result output through `cmd.OutOrStdout()`
- route warning output through `cmd.ErrOrStderr()`
- read confirmation input through `cmd.InOrStdin()`
- add command-level tests for abort, confirm, and `--force` flows
- sync security CLI docs and keyring specs with the command stream contract

## Impact

- improves automation compatibility and testability for keyring passphrase removal
- keeps user-visible prompt text unchanged while aligning I/O behavior with Cobra conventions
