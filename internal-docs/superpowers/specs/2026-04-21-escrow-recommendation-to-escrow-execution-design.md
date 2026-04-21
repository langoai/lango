# Escrow Recommendation To Escrow Execution Design

## Purpose

This design defines the first product-level path that turns an existing `escrow` recommendation into real escrow execution for `knowledge exchange v1`.

The current system already has:

- structured upfront payment approval
- transaction and submission receipts
- direct payment execution gating
- an existing escrow engine and settler stack

What is still missing is the connection between:

- a transaction receipt that canonically recommends `escrow`
- an actual escrow runtime path that creates and funds the escrow

This slice closes that gap.

## Scope

This first slice includes:

- a new operator-facing meta tool
  - `execute_escrow_recommendation`
- a canonical input model based on `transaction_receipt_id`
- execution precondition checks:
  - `current_payment_approval_status = approved`
  - `canonical_settlement_hint = escrow`
- real escrow runtime execution:
  - `create`
  - `fund`
- receipt-backed canonical evidence updates:
  - transaction receipt updates
  - submission receipt trail events
  - escrow reference linkage

This first slice does not include:

- escrow `activate`
- escrow `release` or `refund`
- dispute adjudication
- execution-time backend choice
- human approval UI
- retry orchestration
- a generalized orchestration layer for every escrow use case in the economy subsystem

## Recommended Approach

Three broad approaches were considered:

### 1. Thin Tool Bridge

The tool directly loads receipts, validates state, calls escrow runtime methods, and writes receipt updates itself.

Pros:

- fastest to ship

Cons:

- pushes canonical-state rules into the tool layer
- drifts from the current receipt-service-centered architecture

### 2. Escrow Runtime Plus Receipt Service Coupling

The tool is only the entrypoint. An `escrow execution service` performs validation and runtime calls, and the `receipt service` canonicalizes the resulting evidence.

Pros:

- aligns with current approval, receipt, and payment gating structure
- keeps runtime logic and canonical evidence logic separate
- leaves a clean path for later `activate`, `release`, `refund`, and `dispute` slices

Cons:

- slightly larger than a thin bridge

### 3. Escrow-First Orchestrator

A single orchestration layer owns recommendation execution, escrow runtime, canonical evidence updates, and downstream settlement progression.

Pros:

- strong long-term control plane

Cons:

- too large for a first slice

Recommended choice: **Approach 2: Escrow Runtime Plus Receipt Service Coupling**.

## Architecture

The first slice uses this flow:

1. An operator calls `execute_escrow_recommendation(transaction_receipt_id)`.
2. The tool delegates to an `escrow execution service`.
3. The `escrow execution service` loads the transaction receipt and validates execution preconditions.
4. If valid, it calls the existing escrow engine to:
   - create escrow
   - fund escrow
5. The execution result is passed to the `receipt service`.
6. The `receipt service` updates canonical evidence:
   - transaction receipt fields
   - submission receipt trail events
   - escrow reference linkage

The tool does not directly canonicalize receipt state.

This separation is intentional:

- the execution service decides whether the escrow path may run and performs runtime calls
- the receipt service remains the canonical source for transaction and submission evidence

## Inputs

The first slice uses a single canonical input:

- `transaction_receipt_id`

The tool does not require `submission_receipt_id`.

The transaction receipt is already the canonical source for:

- the current payment approval state
- the current canonical submission
- the current settlement hint

The first slice also does not expose backend selection options. The configured escrow engine and settler combination is used as-is.

## Execution Preconditions

Escrow execution is allowed only when both conditions are true:

- `transaction_receipt.current_payment_approval_status = approved`
- `transaction_receipt.canonical_settlement_hint = escrow`

Anything else must fail closed.

This slice does not reinterpret ambiguous states. It assumes upstream approval already established that escrow is the canonical path for the transaction.

## Runtime Behavior

When preconditions are satisfied, the execution service performs:

1. escrow creation
2. escrow funding

This is real runtime execution, not a dry run and not a recorded intent.

The first slice does not continue into:

- activate
- release
- refund
- dispute

Those remain follow-on slices.

## Canonical State

The transaction receipt gains escrow-specific execution tracking:

- `escrow_execution_status`
  - `pending`
  - `created`
  - `funded`
  - `failed`
- `escrow_reference`
  - typically an `escrow_id` or equivalent runtime identifier

The transaction receipt's `canonical_settlement_status` does **not** change in this slice.

Reason:

- creating and funding escrow means the transaction has entered an escrow-backed path
- it does not yet mean settlement has been partially or fully completed

So this slice records escrow execution separately, without overstating settlement progress.

## Submission Receipt Trail

The submission receipt trail must preserve append-only evidence for this path.

At minimum, the design expects trail events for:

- escrow execution started
- escrow created
- escrow funded
- escrow execution failed

The exact event naming can follow existing receipt-event conventions, but the evidence model must make it possible to reconstruct:

- that escrow execution was attempted
- how far it progressed
- whether it failed before or after escrow creation

## Failure Semantics

Failures are first-class evidence, not silent tool errors.

If execution fails:

- `transaction_receipt.escrow_execution_status` becomes `failed`
- a failure event is appended to the submission receipt trail
- any created escrow reference remains linked when available

This slice does **not** auto-rollback a created escrow when funding fails.

If the flow reaches `created` and then fails at `fund`, the escrow remains in `created` state and the failure is recorded.

This is intentional because:

- it preserves operational truth
- it supports later retry or cleanup work
- it keeps evidence stable for operators and future dispute handling

## Evidence Model

The canonical outputs of a successful execution are:

- transaction receipt update
- submission receipt trail update
- escrow reference linkage

The canonical outputs of a failed execution are:

- transaction receipt failure state
- submission receipt failure event
- linked escrow reference when one exists

This design deliberately reuses the current receipt substrate instead of introducing a new escrow-specific receipt family in the first slice.

## Runtime Constraints

This slice is intentionally narrow.

It is `knowledge exchange` specific, not a full redesign of the economy escrow subsystem.

It does not expose:

- backend choice
- generalized escrow orchestration
- dispute logic
- settlement finalization
- operator retry flows

The main operational truth established by this slice is:

> an approved transaction whose canonical settlement hint is `escrow` can now be converted into a real escrow-backed commitment through an explicit operator entrypoint, with canonical evidence preserved in receipts.

## Testing Expectations

The implementation plan for this design should include tests for:

- allow path when approval is `approved` and settlement hint is `escrow`
- deny path when approval is not approved
- deny path when settlement hint is not `escrow`
- successful escrow create plus fund updates transaction and submission evidence
- fund failure preserves created escrow reference and records canonical failure evidence
- tool surface remains narrow and uses only `transaction_receipt_id`

## Follow-On Work

This first slice sets up the next escrow-oriented slices cleanly:

- escrow activation
- escrow release and refund
- dispute-ready linkage to escrow state
- human approval or operator UI for escrow execution
- retry and repair workflows for partially progressed escrow executions
