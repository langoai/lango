# Sandbox Commands

!!! warning "Experimental"
    The OS-level sandbox is experimental. See [Configuration Reference](../configuration.md#sandbox) for sandbox settings.

Inspect sandbox configuration, platform capabilities, and run isolation smoke tests.

!!! note "OS-level Sandbox vs P2P Sandbox"
    `lango sandbox` manages **OS-level process isolation** (macOS Seatbelt; Linux: planned, not yet enforced) for local tool execution. This is distinct from `lango p2p sandbox` which manages **container-based isolation** for remote P2P tool execution.

## lango sandbox status

Show sandbox configuration, active isolation backend, platform capabilities, and backend availability.

The output includes:

- **Sandbox Configuration**: enabled, fail-closed mode, selected backend, network mode
- **Active Isolation**: which isolator is running and why (if unavailable)
- **Platform Capabilities**: kernel-level primitives (Seatbelt, Landlock, seccomp)
- **Backend Availability**: status of each isolation backend (seatbelt, bwrap, native)

```
lango sandbox status [flags]
```

| Flag | Type | Description |
|------|------|-------------|
| `--json` | `bool` | Output results as JSON |

## lango sandbox test

Run OS sandbox smoke tests to verify isolation is working correctly.

```
lango sandbox test [flags]
```

| Flag | Type | Description |
|------|------|-------------|
| `--json` | `bool` | Output results as JSON |
