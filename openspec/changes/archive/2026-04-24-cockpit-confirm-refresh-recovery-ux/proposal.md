## Why

The cockpit retry action is landed, but it is still too easy to invoke accidentally and leaves the UI stale after success.

## What Changes

- add inline confirm for cockpit retry
- clear confirm when context changes
- refresh backlog and selected detail after successful retry
- document the upgraded recovery UX in public docs and main OpenSpec specs

## Impact

- safer operator replay flow
- less stale cockpit state after success
- no new backend contracts
