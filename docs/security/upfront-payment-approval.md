---
title: Upfront Payment Approval
---

# Upfront Payment Approval

Lango's first upfront payment approval slice decides whether a `knowledge exchange v1` transaction may start with an upfront payment. It sits beside exportability and artifact release approval: exportability decides whether the artifact is tradeable, artifact release approval decides whether release moves forward, and upfront payment approval records the prepayment decision for the transaction.

## What Ships in This Slice

The landed surface is narrow and operator-facing:

- structured upfront payment decisions: `approve`, `reject`, and `escalate`
- decision records with reason, suggested payment mode, amount class, and risk class
- transaction receipt updates that store the current canonical payment approval state
- escrow execution input binding when the approved suggested mode is `escrow`
- append-only payment approval events for later reconstruction
- the `approve_upfront_payment` meta tool as the current operator entrypoint

What the operator entrypoint returns today:

- `transaction_receipt_id`
- `submission_receipt_id`
- the evaluated amount, trust score, user max prepay, and remaining budget inputs
- the approval decision and reason
- the suggested payment mode
- the amount class and risk class
- the updated canonical payment approval state on the linked transaction receipt
- the canonical decision and settlement hint stored on that receipt
- the current `escrow_execution_status` when an escrow recommendation is bound

## Decision Model

- `approve` means the upfront payment path is acceptable under the current slice.
- `reject` means the upfront payment path should not proceed under current policy or budget context.
- `escalate` means the decision cannot be resolved automatically and needs follow-on handling.

The suggested payment mode remains a recommendation at approval time. The current evaluator can recommend `prepay`, `escrow`, `reject`, or `escalate`. When it recommends `escrow`, the approval step binds the escrow execution input onto the transaction receipt, but it still does not create, fund, release, refund, or dispute the escrow by itself.

## Operator Notes

This slice is intentionally narrow.

It does not yet include:

- human approval UI
- full transaction orchestration
- payment dispute adjudication
- partial settlement execution
- escrow activation, release, refund, or dispute handling

The receipt update remains the important durable output of this slice. It gives later surfaces one canonical payment approval state for the transaction. Direct payment execution gating now consumes that state for `prepay`, and escrow execution now consumes the bound escrow input for the first `create + fund` escrow path. This approval surface is still not a settlement engine.

## Related Docs

- [Security Overview](index.md)
- [Approval Flow](approval-flow.md)
- [Actual Payment Execution Gating](actual-payment-execution-gating.md)
- [Escrow Execution](escrow-execution.md)
- [Dispute-Ready Receipts](dispute-ready-receipts.md)
- [P2P Knowledge Exchange Track](../architecture/p2p-knowledge-exchange-track.md)
- [Trust, Security & Policy Audit](../architecture/trust-security-policy-audit.md)
