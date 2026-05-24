# Design

## Context

The knowledge-exchange track now has:

- escrow recommendation execution through `create + fund`
- escrow release
- the first escrow refund service and `refund_escrow_settlement`

What is still missing is the public architecture page and the OpenSpec closeout that make the first escrow refund slice discoverable and keep the docs contract aligned with the landed code.

## Goals / Non-Goals

**Goals**

- publish a bounded public architecture page for escrow refund
- wire the page into the architecture landing page, P2P knowledge-exchange track, and docs navigation
- record the `refund_escrow_settlement` meta tool in OpenSpec
- archive the completed change after syncing main specs

**Non-Goals**

- no refund terminal state
- no dispute-linked refund branching
- no release reversal
- no human UI work
