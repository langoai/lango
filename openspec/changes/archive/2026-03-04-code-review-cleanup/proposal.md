## Why

Code review of the `feature/tui-cli-cmd-update` branch (16K+ lines, 279 files) identified 22 issues across correctness, reuse, quality, and efficiency. Three bugs cause silent failures (agent memory sees empty name, settlement misses events, wrong sentinel error). Multiple functions are duplicated 2–7 times. Fire-and-forget goroutines risk data loss on shutdown.

## What Changes

- Fix dual context key bug: `toolchain.AgentNameFromContext` now delegates to `ctxkeys.AgentNameFromContext` (single canonical key)
- Fix dual event bus: P2P subsystem uses the global event bus instead of creating a separate one
- Fix `ErrTeamFull` reuse: new `ErrBudgetExceeded` sentinel for budget exceeded in `AddSpend()`
- Extract `buildInputSchema()` in `adk/tools.go` to eliminate 3× schema builder duplication
- Deduplicate `ParseUSDC` (paygate delegates to wallet)
- Deduplicate `splitFrontmatter` via shared `internal/mdparse/` package
- Deduplicate `truncate()` across 5 call sites (delegate to `toolchain.Truncate`)
- Add `sync.WaitGroup` + `Close()` to settlement service for graceful shutdown
- Return cloned `Member` copies from `Team.Members()` to prevent data races
- Group `initAgent`'s 14 parameters into `agentDeps` struct
- Align trust threshold to `DefaultPostPayThreshold = 0.7` constant in both paygate and team
- Replace manual string prefix check with `strings.HasPrefix`
- Parallelize health checker probes with `sync.WaitGroup`
- Add `DeferredLedger.Cleanup()` to remove settled entries
- Replace `regexp.MustCompile(\[REJECT\])` with `strings.Contains`
- Add `parentIndex` secondary index to `InMemoryChildStore.ChildrenOf()`

## Capabilities

### New Capabilities
- `shared-mdparse`: Shared markdown frontmatter parsing package (`internal/mdparse/`)

### Modified Capabilities
- `agent-context-propagation`: toolchain context key functions now delegate to ctxkeys (single canonical key)
- `agent-self-correction`: `containsRejectPattern` uses `strings.Contains` instead of regex

## Impact

- `internal/toolchain/hooks.go` — context key functions replaced with var aliases
- `internal/app/app.go` — single event bus, `strings.HasPrefix`, `agentDeps` call site
- `internal/app/wiring.go` — `agentDeps` struct, `initAgent` signature change
- `internal/adk/tools.go` — `buildInputSchema` extracted, `AdaptTool`/`AdaptToolWithTimeout` simplified
- `internal/adk/agent.go` — `containsRejectPattern` uses strings.Contains
- `internal/p2p/team/team.go` — `ErrBudgetExceeded`, `Member.Clone()`, `Members()` returns copies
- `internal/p2p/team/payment.go` — `DefaultPostPayThreshold` constant
- `internal/p2p/paygate/gate.go` — `ParseUSDC` delegates to wallet
- `internal/p2p/paygate/trust.go` — `DefaultPostPayThreshold` constant
- `internal/p2p/paygate/ledger.go` — `Cleanup()` method
- `internal/p2p/settlement/service.go` — `WaitGroup` + `Close()`
- `internal/p2p/agentpool/pool.go` — parallel health checks
- `internal/session/child_store.go` — `parentIndex` secondary index
- `internal/skill/parser.go`, `internal/agentregistry/parser.go` — delegate to `mdparse`
- `internal/mdparse/frontmatter.go` — new shared package
- `internal/background/notification.go`, `internal/cli/cron/cron.go`, `internal/cli/workflow/workflow.go`, `internal/cli/agent/list.go` — delegate `truncate` to `toolchain.Truncate`
