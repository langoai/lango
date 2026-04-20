---
title: Dispute-Ready Receipts
---

# Dispute-Ready Receipts

Lango's dispute-ready receipts slice is a lite operator surface for early knowledge exchange.
It gives operators a canonical record for what was submitted, what transaction it belongs to, and what the current state is without pretending to be a full dispute system.

## What Ships in This Slice

This first slice includes:

- submission receipts linked to transaction receipts
- a current submission pointer for the canonical submission in a transaction
- canonical current state for each receipt
- an append-only event trail for receipt changes
- lite provenance and settlement references for later integration

## Operator Use

Use this surface to answer narrow operational questions:

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
