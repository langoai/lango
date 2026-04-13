## Intent
Add a ProgressBus for unified progress event pub/sub across tools, agents, and background tasks.

## Scope
- New `ProgressBus` type in `internal/streamx/`
- `ProgressEvent` with Source, Type, Message, Progress, Metadata
- Pub/sub with prefix filtering
- Non-blocking emit with buffered channels

## Approach
Standalone implementation (not wrapping eventbus.Bus) with buffered subscriber channels and prefix-based filtering.
