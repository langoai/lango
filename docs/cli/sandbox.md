# Sandbox Commands

!!! warning "Experimental"
    The OS-level sandbox is experimental. See [Configuration Reference](../configuration.md#sandbox) for sandbox settings.

Inspect sandbox configuration, platform capabilities, and run isolation smoke tests.

!!! note "OS-level Sandbox vs P2P Sandbox"
    `lango sandbox` manages **OS-level process isolation** (macOS Seatbelt; Linux bubblewrap when the `bwrap` binary is installed) for local tool execution. This is distinct from `lango p2p sandbox` which manages **container-based isolation** for remote P2P tool execution.

    **Linux requirement:** install the `bubblewrap` package (`apt install bubblewrap`, `dnf install bubblewrap`, or equivalent). The native Landlock+seccomp backend is planned but not yet implemented; selecting `backend=native` returns an unavailable isolator with a clear reason.

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

Run OS sandbox smoke tests to verify isolation is working correctly. The test
honors the configured `sandbox.backend`: when set to `none` it short-circuits
with an explanatory message; when the configured backend is unavailable it
prints the reason and exits successfully without running the cases.

The test runs four cases against the active isolator:

1. **Write restriction (deny /etc)** — invokes `/usr/bin/touch /etc/lango-sandbox-test`. Must fail (sandbox denies writes outside the policy's WritePaths).
2. **Read permission (allow system file)** — reads `/etc/hosts` (macOS) or `/etc/hostname` (Linux). Must succeed (`ReadOnlyGlobal: true` allows reading any path).
3. **Workspace write (allow tmp dir)** — creates an `os.MkdirTemp` directory, adds it to the policy's WritePaths, and touches a file inside. Must succeed.
4. **Network deny (loopback unreachable)** — opens an ephemeral `127.0.0.1:0` listener in the parent process and re-invokes the lango binary as a sandboxed child via the hidden `_probe-net <addr>` subcommand, which calls `net.DialTimeout`. Must fail (sandbox blocks the connect, even to loopback).

The network test uses no external tools (`nc`/`curl`/`bash`/`/dev/tcp`) so it
runs in minimal Docker images. Stdout/stderr are silenced via `io.Discard` in
the parent rather than shell redirection so that the sandbox's `(deny default)`
base on `/dev/null` cannot cause false negatives.

```
lango sandbox test [flags]
```

| Flag | Type | Description |
|------|------|-------------|
| `--json` | `bool` | Output results as JSON |
