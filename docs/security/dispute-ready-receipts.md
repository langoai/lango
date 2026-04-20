---
title: Dispute-Ready Receipts
---

# Dispute-Ready Receipts

Lango's dispute-ready receipts slice is a lite operator surface for early knowledge exchange.
The underlying receipt model tracks canonical current state and append-only event history, but the exposed operator surface is still narrow: today it is primarily the `create_dispute_ready_receipt` meta tool.

## Internal Model

The internal receipt model provides:

- submission receipts linked to transaction receipts
- a current submission pointer for the canonical submission in a transaction
- canonical current state for each receipt
- an append-only event trail for receipt changes
- lite provenance and settlement references for later integration

This model exists inside `internal/receipts/*`. It is the data shape the runtime can build on, not a promise that every field is already surfaced to operators.

## What Ships in This Slice

The currently exposed operator entrypoint is the `create_dispute_ready_receipt` meta tool in `internal/app/tools_meta.go`.

What it returns today:

- `submission_receipt_id`
- `transaction_receipt_id`
- `current_submission_receipt_id`

What it does not expose yet:

- the full submission receipt payload
- the transaction receipt payload
- event trail reads
- direct operator reads of canonical approval or settlement state
- dispute adjudication or settlement execution

## Operator Use

Use the current entrypoint to create the first lite record for a submission.
It is useful when you need receipt identifiers for later follow-up, but it is not yet a full read surface.

The broader operator questions this model will eventually support are:

- what was submitted
- which transaction it belongs to
- what is currently canonical
- which events changed the receipt over time

This slice sits above exportability and approval flow.
Exportability decides whether an artifact is tradeable, approval flow decides whether release moves forward, and dispute-ready receipts preserve the durable record of that path.

## What It Is Not Yet

This is intentionally not a dispute engine.

It does not yet include:

- dispute adjudication
- human dispute UI
- full settlement execution
- settlement reconciliation
- a full evidence graph
- automatic dispute resolution logic

The receipt slice is a record layer, not a final judgment layer.

## Related Docs

- [Security Overview](index.md)
- [Exportability Policy](exportability.md)
- [Approval Flow](approval-flow.md)
- [P2P Knowledge Exchange Track](../architecture/p2p-knowledge-exchange-track.md)
- [Trust, Security & Policy Audit](../architecture/trust-security-policy-audit.md)
