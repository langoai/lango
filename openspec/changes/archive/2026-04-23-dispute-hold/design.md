# Design

## Context

The knowledge-exchange escrow path now has:

- escrow recommendation execution through `create + fund`
- escrow release
- escrow refund

What is still missing is the first dispute-aware control slice that records when a funded escrow is held after canonical dispute handoff, without yet deciding release versus refund.

## Goals / Non-Goals

**Goals**

- expose a dispute hold service and receipts-backed meta tool
- gate hold on `funded` escrow plus `dispute-ready` settlement progression
- record hold success and failure evidence without mutating canonical transaction state
- publish a bounded public architecture page and keep docs navigation aligned
- sync the OpenSpec requirements and archive the completed slice

**Non-Goals**

- no release-vs-refund adjudication
- no explicit escrow `held` lifecycle state
- no dispute engine behavior
- no human adjudication UI
