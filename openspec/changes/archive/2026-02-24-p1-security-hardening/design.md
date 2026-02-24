# P1 Security Hardening — Design

## Architecture

### P1-4: OS Keyring Integration

```
passphrase.Acquire()
  1. keyring (if KeyringProvider set and available)
  2. keyfile (~/.lango/keyfile)
  3. interactive terminal prompt
  4. stdin pipe
```

- `keyring.Provider` interface: `Get/Set/Delete(service, key)`
- `keyring.OSProvider` wraps `github.com/zalando/go-keyring`
- `keyring.IsAvailable()` probes with write/read/delete cycle
- Graceful fallback: CI/headless silently skip to keyfile
- CLI: `lango security keyring store/clear/status`

### P1-5: Tool Execution Process Isolation

```
Remote peer request → handler.handleToolInvoke()
  if sandboxExec != nil:
    SubprocessExecutor.Execute()
      → os.Executable() --sandbox-worker
      → JSON stdin: ExecutionRequest{ToolName, Params}
      → JSON stdout: ExecutionResult{Output, Error}
      → Clean env (PATH, HOME only)
      → context.WithTimeout + cmd.Process.Kill()
  else:
    h.executor() (in-process, existing behavior)
```

- `sandbox.Executor` interface: `Execute(ctx, toolName, params) (map[string]interface{}, error)`
- `InProcessExecutor` for trusted local tools
- `SubprocessExecutor` for P2P remote invocations
- `sandbox.RunWorker()` entry point in child process
- Phase 1: timeout only; Phase 2 (P2-8): rlimit/container

### P1-6: Session Explicit Invalidation

```
SessionStore enhanced with:
  - Invalidate(peerDID, reason)
  - InvalidateAll(reason)
  - InvalidateByCondition(reason, predicate)
  - InvalidationHistory()
  - onInvalidate callback

SecurityEventHandler:
  - Tracks consecutive tool failures per peer
  - Auto-invalidates at threshold (default 5)
  - Listens for reputation drops via callback
```

- `InvalidationReason` enum: logout, reputation_drop, repeated_failures, manual_revoke, security_event
- Callback pattern (like EmbedCallback/GraphCallback) avoids import cycles
- `reputation.Store.SetOnChangeCallback()` fires on score updates
- CLI: `lango p2p session list/revoke/revoke-all`

## File Layout

| Component | New Files | Modified Files |
|-----------|-----------|----------------|
| P1-4 Keyring | `internal/keyring/keyring.go`, `os_keyring.go`, `keyring_test.go`; `cli/security/keyring.go` | `passphrase/acquire.go`, `bootstrap/bootstrap.go`, `config/types.go`, `config/loader.go`, `cli/security/migrate.go`, `go.mod` |
| P1-5 Sandbox | `internal/sandbox/executor.go`, `in_process.go`, `subprocess.go`, `worker.go`, `executor_test.go` | `config/types.go`, `config/loader.go`, `p2p/protocol/handler.go`, `app/app.go`, `cmd/lango/main.go` |
| P1-6 Session | `handshake/security_events.go`, `session_test.go`, `security_events_test.go`; `cli/p2p/session.go` | `handshake/session.go`, `p2p/protocol/handler.go`, `reputation/store.go`, `app/wiring.go`, `cli/p2p/p2p.go` |
