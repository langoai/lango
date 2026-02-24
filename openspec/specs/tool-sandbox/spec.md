# Tool Execution Process Isolation

## Overview

Subprocess-based isolation for remote P2P tool invocations. Prevents remote peers from accessing process memory containing passphrases, private keys, and session tokens.

## Interface

```go
// Executor runs tool invocations in isolation.
type Executor interface {
    Execute(ctx context.Context, toolName string, params map[string]interface{}) (map[string]interface{}, error)
}
```

## Implementations

### InProcessExecutor

Wraps an existing `ToolExecutor` function for trusted local tool calls. No isolation—direct delegation.

### SubprocessExecutor

Launches a child process using the same binary with `--sandbox-worker` flag. Communication via JSON over stdin/stdout.

**Protocol:**
- stdin → `ExecutionRequest{ToolName, Params}`
- stdout ← `ExecutionResult{Output, Error}`

**Security measures:**
- Clean environment: only `PATH` and `HOME`
- `exec.CommandContext` with configurable timeout
- Explicit `cmd.Process.Kill()` on deadline exceeded

## Configuration

```yaml
p2p:
  toolIsolation:
    enabled: false     # default (opt-in)
    timeoutPerTool: 30s
    maxMemoryMB: 256
```

## Wiring

- `handler.SetSandboxExecutor()` follows existing setter pattern
- When `sandboxExec` is set, `handleToolInvoke`/`handleToolInvokePaid` use it instead of `h.executor`
- Fallback to in-process execution when sandbox is nil

## Future (P2-8)

Phase 2 will add rlimit/cgroup/container-based resource limits on top of this subprocess foundation.
