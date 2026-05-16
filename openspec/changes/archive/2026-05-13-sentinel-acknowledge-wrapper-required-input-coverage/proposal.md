## Why

The Security Sentinel tool cluster already declares `alertId` as required for `sentinel_acknowledge`, but there is no direct regression that locks the exact missing-parameter behavior at the wrapper boundary. That leaves room for drift back toward generic downstream failures before alert acknowledgment begins.

## What Changes

- Add an exact wrapper-level regression for missing `alertId` on `sentinel_acknowledge`.
- Update prompt and feature docs to state that missing `alertId` fails before alert-store mutation.
- Sync the sentinel and production-readiness specs to the same fail-closed contract.

## Impact

- `escrow-sentinel`: acknowledgment input semantics become explicitly regression-covered.
- `production-readiness`: sentinel tools align with the same actionable wrapper-error standard used across other hardened tool clusters.
