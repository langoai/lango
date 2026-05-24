# Design

## Context

The knowledge-exchange escrow dispute path now has:

- canonical release-vs-refund adjudication
- adjudication-aware release/refund execution gating
- optional inline post-adjudication execution

What is still missing is the first async convenience slice that lets the selected branch be dispatched onto the existing background task substrate.

## Goals / Non-Goals

**Goals**

- extend `adjudicate_escrow_dispute` with `background_execute=true`
- enforce mutual exclusivity with `auto_execute`
- return a dispatch receipt after successful enqueue
- reuse the existing background task substrate
- publish a bounded public architecture page and keep docs navigation aligned
- sync the OpenSpec requirements and archive the completed slice

**Non-Goals**

- no retry orchestration
- no dead-letter handling
- no specialized status observation API
- no broader dispute engine behavior
