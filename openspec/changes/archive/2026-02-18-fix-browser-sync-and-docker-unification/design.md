## Context

The browser tool uses `sync.Once` for lazy initialization with a retry mechanism that resets the `Once` via value assignment (`t.browserOnce = sync.Once{}`). In concurrent environments, this causes a fatal `sync: unlock of unlocked mutex` crash because `sync.Once` contains internal mutex state that cannot be safely copied or reassigned while other goroutines may be waiting on it.

The Docker setup currently supports three deployment modes (slim, browser-included, sidecar) via `WITH_BROWSER` build arg and compose profiles, which adds operational complexity. The remote browser WebSocket connection code (`RemoteBrowserURL` config + `ROD_BROWSER_WS` env var) exists only to support the sidecar pattern.

## Goals / Non-Goals

**Goals:**
- Fix the `sync.Once` data race crash in browser initialization
- Provide a retryable initialization pattern that is safe under concurrency
- Simplify Docker to a single image that always includes Chromium
- Remove all remote browser connection code and sidecar infrastructure
- Update specs and documentation to reflect the simplified architecture

**Non-Goals:**
- Changing browser tool API or functionality (navigate, screenshot, click, etc.)
- Modifying the browser panic recovery or session management logic
- Optimizing Docker image size (Chromium inclusion is accepted)

## Decisions

### Decision 1: `sync.Mutex` + `bool` instead of `sync.Once`

**Choice**: Replace `sync.Once` + value-reset with `sync.Mutex` + `initDone bool` guard.

**Rationale**: `sync.Once` is designed for exactly-once execution and provides no safe reset mechanism. Assigning a new `sync.Once{}` while another goroutine holds the internal mutex causes undefined behavior. A manual mutex + bool pattern provides the same serialization guarantee while allowing safe reset (set `initDone = false` under lock). On failure, `initDone` naturally remains `false`, enabling automatic retry without any reset logic.

**Alternative considered**: `atomic.Bool` + `sync.Mutex` — adds unnecessary complexity for no benefit since the mutex already serializes access.

### Decision 2: Remove remote browser support entirely

**Choice**: Delete `RemoteBrowserURL` config field, `ROD_BROWSER_WS` env var check, and remote connection code. Keep only local browser launch.

**Rationale**: Remote browser support exists solely for the Docker sidecar pattern. With Chromium always bundled in the image, there is no need for external browser connections. Removing this code eliminates a maintenance surface and simplifies the init path.

### Decision 3: Single Docker image with Chromium

**Choice**: Always install Chromium in the runtime image. Remove `WITH_BROWSER` build arg and all compose profiles.

**Rationale**: The slim image saved ~350MB but required either the sidecar pattern (operational complexity) or disabling browser tools entirely. A single image simplifies deployment, eliminates profile confusion, and ensures browser tools always work when enabled.

## Risks / Trade-offs

- **Larger Docker image (~550MB vs ~200MB)** → Acceptable trade-off for operational simplicity. Modern container registries handle layer caching efficiently.
- **Breaking change for users with `remoteBrowserUrl` config** → Low impact; this was a Docker-specific feature. Migration: remove the config field and use the single image.
- **Lock ordering in Close()** → `Close()` holds `t.mu` (session lock) then acquires `t.initMu` (init lock). `ensureBrowser()` only holds `t.initMu`. No deadlock risk since `ensureBrowser` never acquires `t.mu`.
