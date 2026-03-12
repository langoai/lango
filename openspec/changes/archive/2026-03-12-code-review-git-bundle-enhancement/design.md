## Context

The `feature/git-bundle-enhancement` branch adds P2P workspace/gitbundle, team health monitoring, escrow V2, team-economy bridges, and cron enhancements. Code review identified 10 issues ranging from critical context starvation bugs to minor code deduplication opportunities. All fixes are backward-compatible implementation improvements.

## Goals / Non-Goals

**Goals:**
- Fix context starvation in health monitor's git state collection
- Prevent subscription duplication on monitor restart
- Improve session token lookup performance from O(N) to O(1)
- Reduce escrow dangling detector memory usage for large datasets
- Improve API ergonomics for cron scheduler construction
- Consolidate duplicated conversion logic
- Extract repeated git command execution pattern

**Non-Goals:**
- Changing any public-facing behavior or API contracts
- Adding new features beyond the review fixes
- Modifying the P2P protocol or handshake flow
- Changing the escrow state machine transitions

## Decisions

1. **Separate contexts for ping vs git state** — The health monitor's `pingMember()` shared a single 10s context between health_ping RPC and the git state provider loop. If ping consumed most of the timeout, git state calls would starve. Decision: create a second independent 10s context for git state after successful ping. Alternative: single longer timeout — rejected because it delays unhealthy detection.

2. **sync.Once for event subscriptions** — Each `Start()` call registered new event bus subscriptions without cleanup. Decision: use `sync.Once` to register subscriptions exactly once regardless of restart cycles. Alternative: unsubscribe in `Stop()` — rejected because eventbus doesn't expose unsubscribe API.

3. **GetByToken on SessionStore** — The session validator iterated `ActiveSessions()` per request. Decision: add `GetByToken()` method that does the linear scan internally (sessions are keyed by peerDID, not token, so a reverse index would need maintenance). For the typical small N (handful of peers), this is sufficient. Alternative: maintain a token→DID reverse map — rejected as premature optimization for small N.

4. **SchedulerConfig struct** — 6 positional params is error-prone. Decision: keep `store` and `executor` positional (required), move `timezone`, `maxJobs`, `defaultTimeout`, `logger` into a config struct with sensible defaults. This follows the project's functional options pattern.

5. **ListByStatusBefore on escrow Store** — DanglingDetector loaded ALL pending escrows every scan interval. Decision: add a `ListByStatusBefore(status, time.Time)` method to filter at the DB level. The ent ORM already generates `CreatedAtLT` predicates, so the implementation is straightforward.

6. **Shared floatToMicroUSDC** — Two identical functions (`floatToBudgetAmount`, `floatToUSDC`) in separate bridge files. Decision: extract to `internal/app/convert.go` as `floatToMicroUSDC`. The original comment "kept separate to avoid coupling" was misguided — they're in the same package.

7. **runGit helper** — The same `exec.CommandContext` + stdout/stderr buffer pattern repeated 3+ times in `bundle.go`. Decision: extract a package-private `runGit(ctx, repoPath, args...)` helper. The `CreateBundle` and `CreateIncrementalBundle` methods keep their own buffer handling due to needing raw bytes + custom error detection.

## Risks / Trade-offs

- **Store interface change** → All `escrow.Store` implementations must add `ListByStatusBefore`. Both in-tree implementations (memoryStore, EntStore) are updated. Risk: external implementations would break. Mitigation: this is an internal interface with no known external consumers.
- **Cron constructor API change** → All callers of `cron.New()` must update. Risk: missed callers. Mitigation: `go build ./...` catches all at compile time; tests updated.
- **runGit error format change** → `Diff` and `snapshotRefs` now use `runGit`'s error format (`git <cmd>: <stderr>: <err>`) instead of their previous format. Risk: error message parsing in tests. Mitigation: no tests assert on specific error message format.
