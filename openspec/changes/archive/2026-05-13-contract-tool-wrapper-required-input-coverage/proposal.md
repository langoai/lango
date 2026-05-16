## Why

The contract tool surface declares required wrapper inputs, but there is no direct regression that locks the exact missing-parameter behavior at the agent entrypoint. That leaves room for drift back toward generic downstream failures before contract parsing or execution starts.

## What Changes

- Add exact wrapper-level regressions for missing required inputs on `contract_read`, `contract_call`, and `contract_abi_load`.
- Update prompt/public contract docs to describe that missing required inputs fail at the wrapper boundary.
- Sync the contract interaction and production-readiness specs to the same fail-closed contract.

## Impact

- `contract-interaction`: required input semantics become explicitly regression-covered.
- `production-readiness`: contract tools stay aligned with the same actionable wrapper-error standard used across other hardened tool clusters.
