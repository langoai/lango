## Context

`SessionServiceAdapter.Get()` auto-creates sessions when the requested ID doesn't exist. The check (store.Get) and create (store.Create) are separate operations — not atomic. When multiple Telegram messages arrive simultaneously for the same user, parallel goroutines all observe "session not found" and race to create the same session, causing UNIQUE constraint violations on the second+ attempts.

## Goals / Non-Goals

**Goals:**
- Concurrent auto-create calls for the same session key all succeed without error
- At most one session is created; other callers retrieve the already-created session
- No changes to the store layer or Ent schema

**Non-Goals:**
- Distributed locking or cluster-wide coordination (single-process only)
- Changing the session store's UNIQUE constraint behavior
- Preventing duplicate Telegram update processing at the channel layer

## Decisions

**Get-or-create with conflict recovery over pessimistic locking**

Approach: Let the first Create win and have losers catch the UNIQUE constraint error, then retry Get.

Rationale: This is the standard optimistic concurrency pattern. A pessimistic approach (e.g., `singleflight` or a sync.Mutex in the adapter) would add complexity, block goroutines unnecessarily, and couple the adapter to concurrency primitives. The database already enforces uniqueness — we simply need to handle the expected conflict gracefully.

Alternative considered: `golang.org/x/sync/singleflight` to deduplicate concurrent Creates. Rejected because it adds a dependency and the conflict-retry pattern is simpler and sufficient.

## Risks / Trade-offs

- [String matching on error message] → Brittle if Ent changes error format. Mitigation: the "UNIQUE constraint" text comes from SQLite/Ent and is stable across versions. Can be hardened with structured error types later if needed.
- [Extra Get on conflict] → One additional DB read per losing goroutine. Mitigation: This only occurs on the first message for a new session; negligible impact.
