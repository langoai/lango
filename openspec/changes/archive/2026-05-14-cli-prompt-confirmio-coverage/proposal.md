## Why

`ConfirmIO(...)` existed and was already used by command flows, but its coverage only exercised a single positive path. The deny and read-error branches remained weakly verified.

## What Changes

- Add direct prompt package coverage for `ConfirmIO(...)` yes, no, and read-error branches

## Impact

- Improves confidence in a shared interactive helper without changing runtime behavior
