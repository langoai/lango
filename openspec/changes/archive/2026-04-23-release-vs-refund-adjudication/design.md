# Design

## Context

The knowledge-exchange escrow dispute path now has:

- funded escrow through `create + fund`
- release and refund execution slices
- dispute hold evidence

What is still missing is the first canonical decision that says whether a held, dispute-ready escrow should proceed toward release or refund, without yet executing either branch.

## Goals / Non-Goals

**Goals**

- expose an adjudication service and receipts-backed meta tool
- gate adjudication on funded escrow, dispute-ready settlement progression, and recorded hold evidence
- record canonical release-vs-refund branching on the transaction receipt
- append adjudication evidence without mutating settlement progression or escrow execution state
- publish a bounded public architecture page and keep docs navigation aligned
- sync the OpenSpec requirements and archive the completed slice

**Non-Goals**

- no automatic release or refund execution
- no keep-hold or re-escalation outcomes
- no richer dispute scoring or policy arbitration
- no human adjudication UI
