## Context

The broker protocol is newline-delimited JSON over shared stdio. That means concurrent writes must be serialized, reads must support large single-line payloads, and shutdown sequencing must preserve the method-level cleanup path before pipes are closed.

## Goals / Non-Goals

**Goals:**
- Make broker client request writes atomic with respect to other goroutines.
- Allow encrypt/decrypt payload RPCs to exceed 64 KiB without breaking the read loop.
- Ensure `Close()` executes the graceful shutdown RPC path before the transport is torn down.

**Non-Goals:**
- Changing the wire format away from newline-delimited JSON.
- Introducing request cancellation messages or broker auto-restart behavior.
- Changing broker server semantics beyond what is needed for transport correctness.

## Decisions

- Add a dedicated write mutex separate from the pending-map mutex.
- Use `json.Decoder` on the stdout stream instead of `bufio.Scanner`.
- `Close()` will attempt `methodShutdown` before flipping the closed flag; then it will close stdin and wait for process exit.
- Transport tests will focus on deterministic client behavior rather than full child-process integration.

## Risks / Trade-offs

- `json.Decoder` treats malformed output as terminal, which is acceptable because a corrupted broker stream should fail closed.
- Serializing writes may slightly reduce throughput, but the protocol is request/response over one shared pipe anyway, so correctness is more important.
