# Design

## Context

The knowledge-exchange replay path now has:

- canonical replay gating
- dead-letter evidence
- manual replay dispatch

What is still missing is the first replay-authorization slice that can restrict replay by actor and replay outcome.

## Goals / Non-Goals

**Goals**

- resolve actor from runtime context
- enforce config-backed replay allowlists
- fail closed when actor cannot be resolved or is not allowed
- publish a bounded public architecture page and keep docs navigation aligned
- sync the OpenSpec requirements and archive the completed slice

**Non-Goals**

- no human approval UI
- no org-level policy editor
- no per-transaction policy snapshots
- no amount-tier replay rules
