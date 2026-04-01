## 1. Recovery logic cleanup (internal/agentrt/recovery.go)

- [x] 1.1 Replace `defaultRetryLimits` mutable map with switch statement in `retryLimitForClass`
- [x] 1.2 Remove `goto classCheck` — restructure `Decide` as early-return guard with computed effective retry limit

## 2. Timer leak fix (internal/agentrt/coordinating_executor.go)

- [x] 2.1 Extract `sleepWithContext(ctx, d)` helper using `time.NewTimer` + `defer timer.Stop()`
- [x] 2.2 Replace both `select { case <-ctx.Done()...; case <-time.After(...)... }` blocks with `sleepWithContext`
- [x] 2.3 Remove unnecessary inline comments on lines ~78-80

## 3. Exhaustive switch (internal/observability/collector.go)

- [x] 3.1 Add `case "allow": // no-op, exhaustive` to `RecordPolicyDecision` switch

## 4. Map copy safety (internal/app/wiring.go)

- [x] 4.1 Copy `cachedMetadata` map before closure capture in `rootSessionObserver`

## 5. Dead code removal (internal/app/modules_provenance.go)

- [x] 5.1 Remove manual map key sorting in `computeConfigFingerprint` — rely on `json.Marshal` deterministic key ordering

## 6. Targeted metrics query (internal/app/wiring_session_usage.go)

- [x] 6.1 Replace `collector.Snapshot()` with `collector.SessionMetrics(sessionKey)` in `wireSessionUsage`

## 7. Verification

- [x] 7.1 Run `go build ./...` to verify compilation
- [x] 7.2 Run `go test ./internal/agentrt/... ./internal/app/... ./internal/observability/... -v` to verify all tests pass
