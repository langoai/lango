# Operator Replay / Manual Retry Design

## Purpose / Scope

This design defines the first `operator replay / manual retry` slice for dead-lettered background post-adjudication execution in `knowledge exchange v1`.

Its job is narrow:

- target only dead-lettered post-adjudication execution
- require canonical adjudication to still be present
- replay by reusing the existing background post-adjudication execution path
- append manual retry evidence without clearing prior dead-letter evidence

This slice covers:

- a new `retry_post_adjudication_execution` operator tool
- replay gating
- replay service
- new background dispatch receipt
- manual retry evidence

This slice does not cover:

- inline replay
- arbitrary background-task replay
- generic dead-letter browsing UI
- richer replay policy
- broader dispute engine behavior

## Execution Gate

The canonical input is:

- `transaction_receipt_id`

`retry_post_adjudication_execution(transaction_receipt_id)` may proceed only when:

- dead-letter evidence already exists
- canonical adjudication is still present
- the current submission still exists
- the transaction still resolves to a valid post-adjudication execution branch

This means:

- plain background failure evidence is not enough
- replay requires a still-live canonical branch decision on the transaction receipt

## Replay Model

This slice does not create a new execution stack.

Replay uses the existing `background post-adjudication execution` path:

1. operator invokes `retry_post_adjudication_execution`
2. replay service loads transaction, current submission, and adjudication
3. replay service appends manual retry evidence
4. replay service enqueues the same post-adjudication execution path again

Replay therefore means:

- keep the old dead-letter evidence
- append a new operator replay action
- create a new async execution attempt

## Success / Failure Semantics

On replay success:

- canonical adjudication stays unchanged
- prior dead-letter evidence stays unchanged
- `manual retry requested` evidence is appended
- a new background dispatch receipt is returned

On replay failure:

- canonical adjudication remains unchanged
- no new dispatch is created
- replay failure evidence may be appended if needed

This keeps canonical branch state, dispatch behavior, and operator replay action separate.

## Evidence Model

This slice uses append-only evidence.

Existing evidence:

- dead-letter evidence remains intact

New evidence:

- `manual retry requested`
- new background dispatch evidence

This allows later reconstruction of:

- when the execution dead-lettered
- when an operator explicitly replayed it
- which new dispatch was created
- how the later execution attempt progressed

## Return Shape

On success, `retry_post_adjudication_execution` returns:

- `canonical adjudication snapshot`
- `new background dispatch receipt`

This lets the caller see:

- what branch is still canonical
- what new async execution handoff was created

## Implementation Shape

Recommended structure:

- `retry_post_adjudication_execution` meta tool
- shared replay service
- existing background dispatch reuse

Flow:

1. tool accepts `transaction_receipt_id`
2. replay service loads transaction and current submission
3. replay gate validates dead-letter evidence and canonical adjudication
4. replay service appends `manual retry requested`
5. replay service reuses existing background dispatch
6. return adjudication snapshot plus new dispatch receipt

This keeps:

- the tool as a thin entrypoint
- the replay service as the replay-policy owner
- dispatch behavior on the already-landed background execution path

## Follow-On Inputs

The next follow-on work after this slice is:

1. `dead-letter browsing / status observation`
   - surface all dead-lettered executions to operators

2. `policy-driven replay controls`
   - define who may replay and under which conditions

3. `generic replay substrate`
   - later decide whether replay should generalize beyond post-adjudication execution
