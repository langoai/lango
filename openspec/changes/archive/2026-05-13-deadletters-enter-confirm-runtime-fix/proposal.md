## Why

The Dead Letters retry-confirm flow now advertises `Enter` as an alternate confirm key, but the runtime still routes `Enter` through the generic filter-apply path even when retry confirmation is pending. That leaves the visible contract ahead of the real behavior.

## What Changes

- Make `Enter` submit the pending retry request when Dead Letters is in retry-confirm state.
- Add regressions for the `Enter` confirm path.
- Update the cockpit-pages spec to pin the runtime confirm behavior instead of help/docs only.

## Impact

- The visible retry-confirm contract matches the actual runtime.
- Operators can use either `r` or `Enter` to confirm a retry request, with `Esc` still canceling.
