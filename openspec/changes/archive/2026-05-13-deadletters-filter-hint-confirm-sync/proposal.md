## Why

The Dead Letters page now routes `Enter` to retry confirmation when a retry request is pending, but the filter hint line still always says `Enter to apply`. That leaves one prominent in-page hint behind the actual confirm-state behavior.

## What Changes

- Make the Dead Letters filter hint line context-sensitive so confirm state advertises retry confirmation instead of generic filter apply.
- Add regression coverage for the confirm-state hint text.
- Update cockpit docs and spec to describe the confirm-state hint behavior.

## Impact

- The in-page hint line matches the actual `Enter` action during retry confirmation.
- Operators no longer see contradictory guidance in the same view.
