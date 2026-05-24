## Why

The economy-layer escrow tools are already covered by exact missing-parameter regressions, but the non-escrow economy entrypoints are not. That leaves `economy_budget_*`, `economy_risk_assess`, `economy_negotiate*`, and `economy_price_quote` exposed to drift back toward generic downstream failures before budget lookup, risk assessment, negotiation lookup, or pricing lookup begin.

## What Changes

- Add exact missing-parameter regressions for the non-escrow economy tool cluster.
- Update prompt/public docs to describe that these required economy inputs fail at the wrapper boundary.
- Sync production-readiness to the same fail-closed contract.

## Impact

- `economy`: the full operator-facing economy surface gains more uniform wrapper coverage.
- `production-readiness`: budget/risk/pricing/negotiation tools follow the same actionable missing-parameter standard as escrow tools.
