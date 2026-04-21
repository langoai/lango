---
title: Actual Payment Execution Gating
---

# Actual Payment Execution Gating

Lango's first actual payment execution gating slice turns payment approval from an advisory signal into an execution control surface for direct payment tools.

Today the gated paths are:

- `payment_send`
- `p2p_pay`

This slice is intentionally narrow. It does not execute escrow, it does not add a human payment approval UI, and it does not try to orchestrate full transaction settlement. It only decides whether a direct payment execution may proceed right now.

## What Ships in This Slice

- a shared direct-payment gate service used by both `payment_send` and `p2p_pay`
- receipt-backed `allow` or `deny` decisions
- canonical state lookup through the linked transaction receipt
- optional `submission_receipt_id` override, with fallback to the transaction's current canonical submission
- allow and deny execution events written to both the audit log and the receipt trail

## Decision Model

This gate is not a new policy engine. It consumes the canonical result of earlier surfaces:

- upfront payment approval sets the current canonical payment approval state
- dispute-ready receipts preserve the linked submission and transaction evidence
- actual payment execution gating enforces that state at the direct payment tool boundary

The gate returns only:

- `allow`
- `deny`

There is no execution-time `escalate`. Ambiguous or high-risk cases must already have been resolved upstream.

## Receipt-Backed Execution Rules

Direct payment execution is receipt-backed.

- `transaction_receipt_id` is required
- `submission_receipt_id` is optional
- when `submission_receipt_id` is omitted, the gate uses the transaction receipt's current canonical submission
- when `submission_receipt_id` is provided, it must exist, belong to the transaction, and still match the current canonical submission

For the first slice, direct payment execution allows only `prepay`.

- if the canonical payment approval state is not `approved`, execution is denied
- if the canonical settlement hint is not `prepay`, execution is denied
- if the canonical submission is stale, execution is denied

## Deny Reasons

The first slice uses a small deny reason set:

- `missing_receipt`
- `approval_not_approved`
- `stale_state`
- `execution_mode_mismatch`

These reason codes are recorded for both tool families so operators can reconstruct why a direct payment did or did not execute.

## Evidence Requirements

This slice is fail-closed.

Direct payment execution requires both:

- an audit recorder
- a receipt trail sink

If either evidence sink is unavailable, execution does not proceed. The gate does not allow an audit-only success path and does not silently skip receipt evidence.

## Current Limits

This slice still does not provide:

- escrow execution
- execution-time human escalation
- full payment orchestration
- settlement finalization logic
- payment dispute adjudication

It is the first direct-payment enforcement layer, not the complete payment runtime.

## Related Docs

- [Security Overview](index.md)
- [Upfront Payment Approval](upfront-payment-approval.md)
- [Dispute-Ready Receipts](dispute-ready-receipts.md)
- [USDC Payments](../payments/usdc.md)
