# Automatic Post-Adjudication Execution Design

## Purpose / Scope

This design defines the first `automatic post-adjudication execution` slice for `knowledge exchange v1`.

Its job is narrow:

- extend `adjudicate_escrow_dispute` with optional `auto_execute=true`
- inline the matching escrow release or refund executor after successful adjudication
- preserve adjudication as the canonical write layer even when nested execution fails
- return both the adjudication result and nested execution result

This slice covers:

- `adjudicate_escrow_dispute`
- inline handler orchestration
- adjudication result + nested execution result return shape

This slice does not cover:

- background queue execution
- retry orchestration
- automatic execution as a default policy
- new lifecycle states
- broader dispute engine behavior

## Trigger Model

`adjudicate_escrow_dispute` continues to canonicalize the release-vs-refund decision first.

Execution modes:

- default or omitted
  - adjudication only
- `auto_execute=true`
  - adjudication
  - then inline nested execution

Constraints:

- `auto_execute=true` and `background_execute=true` cannot be combined
- one request chooses one execution mode

`auto_execute` is an explicit convenience opt-in, not the default behavior.

## Execution Flow

1. `adjudicate_escrow_dispute(transaction_receipt_id, outcome, auto_execute?)`
2. adjudication service records canonical adjudication and progression transition
3. if `auto_execute` is false:
   - return adjudication result only
4. if `auto_execute` is true:
   - read the canonical adjudication outcome
   - `release` routes to the existing `release_escrow_settlement` path
   - `refund` routes to the existing `refund_escrow_settlement` path
5. return adjudication result plus nested execution result

The inline path does not bypass any executor gate. It reuses the same release/refund services and therefore re-runs the same adjudication-aware execution checks.

## Success / Failure Semantics

This slice keeps three layers distinct:

1. adjudication
   - canonical write layer
2. nested execution dispatch
   - inline handler orchestration layer
3. nested execution result
   - existing release/refund execution layer

Possible outcomes:

- adjudication success + nested execution success
- adjudication success + nested execution failure
- adjudication failure

Important rule:

- once adjudication succeeds, it is not rolled back if nested execution fails

This preserves the separation between canonical decision state and execution-layer failure.

## Return Shape

When `auto_execute=true`, the tool returns:

- `adjudication_result`
- `execution_result`

The nested execution result is optional:

- absent when adjudication-only mode is used
- populated when auto execution is requested

This lets callers reconstruct:

- what canonical adjudication was recorded
- whether nested release/refund execution succeeded or failed

## Implementation Shape

Recommended structure:

- adjudication service
  - canonical adjudication + progression transition only
- `adjudicate_escrow_dispute` meta tool handler
  - parse execution mode
  - call adjudication service
  - if `auto_execute=true`, invoke the existing matching release/refund service inline
- existing escrow release/refund services
  - enforce the same adjudication-aware executor gates

This keeps:

- service layer responsible for canonical writes
- handler layer responsible for convenience orchestration
- executor services responsible for money-moving execution

## Follow-On Inputs

The next follow-on work after this slice is:

1. `background post-adjudication execution`
   - async queue-based dispatch

2. `retry / recovery semantics`
   - nested execution retry and recovery after failure

3. `policy-driven defaults`
   - when auto execution should become a default behavior instead of an explicit flag
