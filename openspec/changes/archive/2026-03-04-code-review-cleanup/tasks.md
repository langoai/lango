## 1. Critical — Correctness Bugs

- [x] 1.1 Replace `toolchain/hooks.go` private context key with `var` aliases to `ctxkeys.WithAgentName`/`ctxkeys.AgentNameFromContext`
- [x] 1.2 Delete `p2pBus := eventbus.New()` in `app.go`, pass global `bus` to `initP2P`
- [x] 1.3 Add `ErrBudgetExceeded` sentinel to `team/team.go`, use in `AddSpend()`

## 2. High — Code Reuse

- [x] 2.1 Extract `buildInputSchema()` in `adk/tools.go`, make `AdaptTool`/`AdaptToolWithTimeout` delegate to `adaptToolWithOptions`
- [x] 2.2 Replace `paygate.ParseUSDC` with `var ParseUSDC = wallet.ParseUSDC` alias
- [x] 2.3 Create `internal/mdparse/frontmatter.go` with shared `SplitFrontmatter`
- [x] 2.4 Update `skill/parser.go` to delegate `splitFrontmatter` to `mdparse.SplitFrontmatter`
- [x] 2.5 Update `agentregistry/parser.go` to delegate `splitFrontmatter` to `mdparse.SplitFrontmatter`
- [x] 2.6 Replace `truncate()` in `background/notification.go` with `toolchain.Truncate` delegation
- [x] 2.7 Replace `truncate()` in `cli/cron/cron.go` with `toolchain.Truncate` delegation
- [x] 2.8 Replace `truncate()` in `cli/workflow/workflow.go` with `toolchain.Truncate` delegation
- [x] 2.9 Replace `truncate()` in `cli/agent/list.go` with `toolchain.Truncate` delegation

## 3. Medium — Quality & Safety

- [x] 3.1 Add `sync.WaitGroup` + `Close()` to `settlement/service.go` for graceful goroutine shutdown
- [x] 3.2 Add `Member.Clone()` method, update `Team.Members()` to return copies
- [x] 3.3 Introduce `agentDeps` struct in `wiring.go`, update `initAgent` signature
- [x] 3.4 Update `initAgent` call site in `app.go` to use `agentDeps`
- [x] 3.5 Define `DefaultPostPayThreshold = 0.7` in `paygate/trust.go`, use in `DefaultTrustConfig()`
- [x] 3.6 Define `DefaultPostPayThreshold = 0.7` in `team/payment.go`, use in `SelectPaymentMode()`
- [x] 3.7 Replace manual string prefix checks with `strings.HasPrefix` in `app.go`

## 4. Low — Efficiency

- [x] 4.1 Parallelize `HealthChecker.checkAll()` with `sync.WaitGroup` in `agentpool/pool.go`
- [x] 4.2 Add `DeferredLedger.Cleanup()` method to remove settled entries in `paygate/ledger.go`
- [x] 4.3 Replace `regexp.MustCompile(\[REJECT\])` with `strings.Contains` in `adk/agent.go`
- [x] 4.4 Add `parentIndex` secondary index to `InMemoryChildStore`, update `ForkChild`/`DiscardChild`/`ChildrenOf`

## 5. Verification

- [x] 5.1 `go build ./...` passes
- [x] 5.2 `go test ./...` passes
- [x] 5.3 `go vet ./...` passes
- [x] 5.4 Update trust threshold test in `paygate/trust_test.go` to use `DefaultPostPayThreshold`
