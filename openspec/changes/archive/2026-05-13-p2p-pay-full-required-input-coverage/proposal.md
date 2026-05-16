## Why

`p2p_pay` is already hardened for missing `transaction_receipt_id`, but the wrapper contract also requires `peer_did` and `amount`. Those fields are enforced in code today, yet they are not directly locked by regression tests or reflected as a full required-input contract in the main production-readiness spec.

## What Changes

- Add exact missing-parameter regressions for `p2p_pay` required `peer_did` and `amount`, alongside the existing receipt-id guard.
- Sync the production-readiness and P2P payment docs/spec wording to the full required-input set.

## Impact

- `p2p-payment`: the full wrapper contract is now regression-covered instead of only the receipt-id subset.
- `production-readiness`: the P2P payment entrypoint better matches its real required-input surface.
