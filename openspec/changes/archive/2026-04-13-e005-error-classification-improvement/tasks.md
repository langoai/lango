## Tasks

### 1. New cause constants and classifyError patterns

- [x] 1.1 Add `CauseProviderAuth` and `CauseProviderConnection` constants in `internal/adk/errors.go`
- [x] 1.2 Add auth classification block in `classifyError()` between 500/503 and "tool" checks, using `strings.ToLower(msg)`
- [x] 1.3 Add connection classification block in `classifyError()` after auth block
- [x] 1.4 Update nil-error path to set `CauseDetail: "classifyError called with nil error (defensive)"`
- [x] 1.5 Update E005 fallback `OperatorSummary` to include truncated error message via `msg[:min(len(msg), 200)]`

### 2. UserMessage() curated messages

- [x] 2.1 Replace `ErrModelError` case in `UserMessage()` with switch on `CauseClass` for auth/connection/default

### 3. Coordinating executor log improvement

- [x] 3.1 Add `cause_detail` and `error` fields to AgentError recovery log in `internal/agentrt/coordinating_executor.go`

### 4. Recovery policy update

- [x] 4.1 Add `CauseProviderAuth → CauseUnknown` and `CauseProviderConnection → CauseTransient` in `classifyForRetry()`
- [x] 4.2 Expand `ErrModelError` case in `Decide()` with auth escalation and connection retry branches

### 5. Tests

- [x] 5.1 Add classifyError test cases in `internal/adk/errors_test.go`: auth (401, invalid api key, uppercase), connection (connection refused, no such host, dial tcp)
- [x] 5.2 Add UserMessage test cases: auth curated message, connection curated message, E005 default no raw detail
- [x] 5.3 Add recovery policy test cases in `internal/agentrt/recovery_test.go`: auth escalates, connection retries

### 6. Verification

- [x] 6.1 `go build ./...` passes
- [x] 6.2 `go test ./...` passes
