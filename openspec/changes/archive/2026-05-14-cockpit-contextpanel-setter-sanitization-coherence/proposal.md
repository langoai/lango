## Why

The cockpit context panel now receives sanitized labels from the current channel/runtime trackers, but `SetChannelStatuses` and `SetRuntimeStatus` themselves still accept and cache raw label values. That leaves the panel cache dependent on every caller being well-behaved instead of the setter boundary enforcing a display-safe snapshot.

## What Changes

- Sanitize channel names and runtime active-agent labels inside the context-panel setters.
- Add regression coverage for malformed setter input values.
- Record the setter-boundary replay-safety contract in OpenSpec and downstream docs.

## Impact

- Makes the context panel cache replay-safe even when fed by new callers outside the current tracker path.
- Tightens the final shared snapshot boundary around the right-side metrics surface.
