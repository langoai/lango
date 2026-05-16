## Why

The main P2P tool cluster already declares required wrapper inputs, but only `p2p_pay` is directly locked by missing-parameter regressions. That leaves the rest of the operator-facing P2P entrypoints vulnerable to future drift back toward generic downstream failures before session lookup, remote invocation, or firewall mutation begins.

## What Changes

- Add exact missing-parameter regressions for `p2p_connect`, `p2p_disconnect`, `p2p_query`, `p2p_firewall_add`, `p2p_firewall_remove`, `p2p_price_query`, `p2p_reputation`, and `p2p_invoke_paid`.
- Update prompt/public docs to describe that these required P2P inputs fail at the wrapper boundary.
- Sync the production-readiness spec to the same fail-closed contract.

## Impact

- `p2p-networking`: operator-facing tool entrypoints become explicitly regression-covered.
- `production-readiness`: P2P wrapper semantics align with the same actionable missing-parameter standard used across other hardened tool clusters.
