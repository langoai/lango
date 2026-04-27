# Design

## Context

The knowledge-exchange escrow dispute path now has:

- release-vs-refund adjudication
- adjudication-aware release/refund execution gating

What is still missing is the first inline convenience slice that lets adjudication and matching execution happen in one call without introducing a new lifecycle state or background runner.

## Goals / Non-Goals

**Goals**

- extend `adjudicate_escrow_dispute` with optional `auto_execute`
- inline the matching release or refund executor after successful adjudication
- preserve adjudication as the canonical write layer even if execution fails
- return both adjudication and nested execution results
- publish a bounded public architecture page and keep docs navigation aligned
- sync the OpenSpec requirements and archive the completed slice

**Non-Goals**

- no background execution
- no retry orchestration
- no automatic execution by default policy
- no broader dispute engine behavior
