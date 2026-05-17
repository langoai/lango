## Summary

Remove the production `panic` path from `internal/paymentapproval` threshold initialization while preserving the existing upfront-payment evaluation API and decision behavior.

## Motivation

The payment approval evaluator is part of the runtime settlement path. Even though the current threshold strings are hard-coded constants, production package initialization should not contain panic-based failure paths for deterministic policy setup. This change makes the thresholds explicit and adds a regression guard so future threshold edits fail in tests instead of crashing at runtime.

## Scope

- Replace panic-based threshold parsing with explicit, checked threshold construction.
- Keep `EvaluateUpfrontPayment(Input) Outcome` behavior unchanged for valid and invalid runtime inputs.
- Add executable coverage that rejects panic reintroduction in `internal/paymentapproval`.
- Sync and archive the OpenSpec change after implementation.

## Non-Goals

- No changes to payment approval decision thresholds.
- No changes to receipt storage, payment execution, or CLI/TUI surfaces.
- No user-facing documentation changes unless implementation changes public behavior.
