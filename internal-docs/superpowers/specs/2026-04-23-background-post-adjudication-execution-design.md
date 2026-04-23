# Background Post-Adjudication Execution Design

## Purpose / Scope

This design defines the first `background post-adjudication execution` slice for `knowledge exchange v1`.

Its job is narrow:

- extend `adjudicate_escrow_dispute` with optional `background_execute=true`
- enqueue a post-adjudication execution job after successful adjudication
- reuse the existing release/refund executor services from a background worker
- return both the adjudication result and a background dispatch receipt

This slice covers:

- `adjudicate_escrow_dispute`
- background dispatch receipt
- background worker execution path
- worker success and failure evidence

This slice does not cover:

- retry orchestration
- dead-letter queues
- scheduled backoff
- automatic execution as a default policy
- broader dispute engine behavior

## Trigger Model

`adjudicate_escrow_dispute` continues to canonicalize the release-vs-refund decision first.

Execution modes:

- default or omitted
  - adjudication only
- `auto_execute=true`
  - inline execution
- `background_execute=true`
  - background dispatch

Constraints:

- `auto_execute=true` and `background_execute=true` cannot be combined
- one request chooses one execution mode

`background_execute` is an explicit convenience opt-in, not the default behavior.

## Dispatch Model

When `background_execute=true`, the handler:

1. performs canonical adjudication
2. confirms adjudication success
3. enqueues a background execution task
4. records dispatch evidence
5. returns adjudication result plus dispatch receipt

The dispatch receipt should minimally include:

- `status`
  - `queued`
- `transaction_receipt_id`
- `submission_receipt_id`
- `escrow_reference`
- `outcome`
- `dispatch_reference`

This lets a caller reconstruct both the canonical adjudication and the async execution handoff in a single response.

## Worker Execution Model

The worker does not introduce a new execution stack.

It reads the canonical adjudication outcome and routes to the existing executor path:

- `release` adjudication
  - existing `release_escrow_settlement` path
- `refund` adjudication
  - existing `refund_escrow_settlement` path

Important rule:

- worker execution reuses the same adjudication-aware release/refund gates
- the background path does not bypass executor validation

This keeps the background path as the same canonical flow with different timing.

## Success / Failure Semantics

This slice separates three layers:

1. adjudication
   - canonical write layer
2. dispatch
   - background enqueue layer
3. worker execution
   - release/refund execution layer

Possible outcomes:

- adjudication success + dispatch success + worker success
- adjudication success + dispatch success + worker failure
- adjudication success + dispatch failure
- adjudication failure

Important rules:

- successful adjudication is not rolled back if dispatch fails
- successful dispatch is not rolled back if worker execution fails
- worker failures are recorded as execution evidence

This preserves the separation between canonical decision state, dispatch state, and execution-layer results.

## Evidence Model

This slice records evidence in three layers:

1. adjudication evidence
   - existing adjudication trail remains unchanged

2. dispatch evidence
   - audit log
   - submission receipt trail

3. worker execution evidence
   - existing release/refund execution evidence
   - worker failure evidence when execution fails

This allows later reconstruction of:

- what canonical adjudication was recorded
- whether background dispatch happened
- whether the worker ultimately succeeded or failed

## Return Shape

When `background_execute=true`, the tool returns:

- `adjudication_result`
- `background_dispatch_receipt`

The worker execution result is not part of the synchronous response in this first slice.
That result is observable later through execution evidence and status inspection.

## Implementation Shape

Recommended structure:

- `adjudicate_escrow_dispute` meta tool handler
  - validates execution mode
  - calls adjudication service
  - enqueues the background job

- background task payload
  - `transaction_receipt_id`
  - canonical expected outcome

- worker
  - reloads canonical adjudication
  - invokes the existing matching release/refund service

This keeps:

- the handler as dispatch orchestration
- the adjudication service as canonical write owner
- the worker as executor reuse

## Follow-On Inputs

The next follow-on work after this slice is:

1. `retry / dead-letter`
   - worker failure recovery and backoff

2. `status observation`
   - explicit status query surface for background execution

3. `policy-driven defaults`
   - when async execution should become a default mode instead of an explicit flag
