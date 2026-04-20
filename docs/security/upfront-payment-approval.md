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
- append-only payment approval events for later reconstruction
- the `approve_upfront_payment` meta tool as the current operator entrypoint

What the operator entrypoint returns today:

- the approval decision
- the reason
- the suggested payment mode
- the amount class
- the risk class

## Decision Model

- `approve` means the upfront payment path is acceptable under the current slice.
- `reject` means the upfront payment path should not proceed under current policy or budget context.
- `escalate` means the decision cannot be resolved automatically and needs follow-on handling.

The suggested payment mode is a recommendation, not execution. The current slice may recommend `prepay` or `escrow`, but it does not open, move, or settle funds.

## Operator Notes

This slice is intentionally narrow.

It does not yet include:

- actual payment execution gating
- escrow execution
- human approval UI
- full transaction orchestration
- payment dispute adjudication
- partial settlement execution

The receipt update is the important durable output of this slice. It gives later surfaces one canonical payment approval state for the transaction, but it is not a settlement engine.

## Related Docs

- [Security Overview](index.md)
- [Approval Flow](approval-flow.md)
- [Dispute-Ready Receipts](dispute-ready-receipts.md)
- [P2P Knowledge Exchange Track](../architecture/p2p-knowledge-exchange-track.md)
- [Trust, Security & Policy Audit](../architecture/trust-security-policy-audit.md)
