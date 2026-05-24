## Why

`lango security change-passphrase` and `lango security recovery restore` already route their success messages through Cobra command writers, but their keyfile/keyring update notices and warnings still bypass command streams by writing directly to process stderr.

## What Changes

- Route change-passphrase keyfile/keyring notices and warnings through `cmd.ErrOrStderr()`
- Route recovery restore keyfile/keyring notices and warnings through `cmd.ErrOrStderr()`
- Add regression coverage for these warning/error-stream paths
- Update docs and OpenSpec with the explicit stderr contract

## Impact

- Makes the security CLI stream contract consistent across both success and warning paths
- Improves wrapper and test harness capture fidelity
