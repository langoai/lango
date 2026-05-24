## Context

`internal/paymentapproval` classifies upfront payment requests using three fixed USDC thresholds. The current implementation initializes those thresholds through `mustParseUSDC`, which panics if a threshold string is invalid. Since the values are deterministic package constants, errors should be caught by tests and code review rather than represented as a runtime panic path.

## Design

Threshold values remain package-level variables, but initialization will use a helper that returns `*big.Int` without an error branch. The helper will construct values from exact integer micro-USDC units so there is no string parsing failure mode during package initialization.

The evaluator API remains unchanged:

- `EvaluateUpfrontPayment(Input) Outcome` continues to return `Outcome`.
- Invalid request amounts and invalid `UserMaxPrepay` values continue to return reject outcomes.
- Amount classification thresholds remain 5.00, 50.00, and 100.00 USDC.

## Testing

Add a package-level quality guard test that scans non-test Go files in `internal/paymentapproval` and fails if `panic(` is reintroduced. Existing behavioral tests continue to verify approval, rejection, escalation, and threshold boundaries.

## Risks

The primary risk is accidentally changing threshold scale while removing string parsing. Mitigation: keep the boundary tests around 99.99 and 100.00 USDC and add no behavior changes beyond initialization hygiene.
