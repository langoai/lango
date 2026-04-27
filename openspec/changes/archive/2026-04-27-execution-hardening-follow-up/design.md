## Design Summary

This follow-up stays tactical and avoids changing the recently landed product/runtime surface area.

### Service-local transaction locks

The dispute hold, escrow adjudication, and escrow refund services each keep a small `sync.Map` of `transactionReceiptID -> *sync.Mutex`.

This does not introduce a new shared orchestration layer. It only prevents the same transaction from entering the execution path in parallel inside one process.

### Background retry-key dedup

`background.Manager.Submit` now derives the canonical retry key before allocation and reuses an existing task ID when a matching key is already:

- `pending`
- `running`
- `failed` with `next_retry_at` set

This keeps retry dedup aligned with the already-landed retry metadata model.

### Settlement escalation guard

`escalationProgressionStatus` now handles every known `SettlementProgressionStatus` explicitly and panics on unknown internal values.

That keeps future enum growth from silently degrading to `review-needed`.

### Test coverage

Focused tests cover:

- concurrent service execution serialization
- hold/refund success-record partial-commit failures
- retry-key dedup for running and scheduled tasks
- panic-on-unknown escalation mapping
