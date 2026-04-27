---
title: Escrow Execution
---

# Escrow Execution

Lango's first escrow execution slice turns an approved escrow recommendation into a real `create + fund` runtime path for `knowledge exchange v1`.

It sits after upfront payment approval. Upfront payment approval decides whether the transaction should use escrow and binds the execution input onto the transaction receipt. Escrow execution then consumes that bound input and creates and funds the escrow while preserving canonical receipt evidence.

## What Ships in This Slice

The landed surface is intentionally narrow and operator-facing:

- escrow-recommended upfront payment approvals can bind escrow execution input onto the transaction receipt
- the `execute_escrow_recommendation` meta tool executes the approved escrow path
- the runtime creates the escrow and immediately funds it
- transaction receipts track canonical escrow execution status and escrow reference
- receipt trails append escrow execution progress and failure events

What the current execution entrypoint returns today:

- `transaction_receipt_id`
- `submission_receipt_id`
- `escrow_reference`
- `escrow_execution_status`

## Operator Entry Points

The first slice uses two operator-facing steps:

1. `approve_upfront_payment`
2. `execute_escrow_recommendation`

When upfront payment approval recommends `escrow`, the approval step binds:

- buyer DID
- seller DID
- total amount
- reason
- optional task ID
- milestone list

onto the linked transaction receipt as escrow execution input. The approval response also exposes `escrow_execution_status` so operators can see that the transaction is prepared for escrow execution.

The execution step then requires only:

- `transaction_receipt_id`

The runtime resolves the current canonical submission from the transaction receipt instead of asking operators to supply a separate submission ID at execution time.

## Execution Model

This slice is receipt-backed and fail-closed.

- execution requires a transaction receipt with canonical payment approval state `approved`
- execution requires canonical settlement hint `escrow`
- execution requires bound escrow execution input on the transaction receipt
- execution records progress in the linked receipt trail before and after runtime calls

The current execution sequence is:

1. append `escrow_execution_started`
2. create the escrow through the escrow engine
3. append `escrow_execution_created`
4. fund the created escrow
5. append `escrow_execution_funded`

If create or fund fails, the runtime appends `escrow_execution_failed` and preserves the failure reason in receipt evidence.

## Receipt Evidence

The transaction receipt now carries escrow-specific canonical state:

- `escrow_execution_status`
- `escrow_reference`
- `escrow_execution_input`

The current status values are:

- `pending`
- `created`
- `funded`
- `failed`

This keeps the transaction receipt as the canonical operator surface for "is this escrow recommendation only advisory, prepared, or already executed?"

## Current Limits

This first slice does not yet provide:

- escrow activation
- milestone completion or release
- refund handling
- dispute handling or adjudication
- human approval UI
- full settlement orchestration
- operator read APIs for complete escrow receipt history beyond the current internal/runtime surfaces

The landed scope is strictly `create + fund` execution for transactions that were already approved with an escrow recommendation.

## Related Docs

- [Security Overview](index.md)
- [Upfront Payment Approval](upfront-payment-approval.md)
- [Actual Payment Execution Gating](actual-payment-execution-gating.md)
- [Dispute-Ready Receipts](dispute-ready-receipts.md)
- [P2P Knowledge Exchange Track](../architecture/p2p-knowledge-exchange-track.md)
