## Why

Some auxiliary docs and older main specs still described stale `--json` behavior even though the implementation had already moved on or never supported that flag. These low-visibility drifts weaken trust in the CLI and make the project feel less production-ready than the code actually is.

## What Changes

- sync auxiliary docs and legacy specs to the actual output contracts
- extend executable docs guards so these stale patterns cannot quietly return

## Impact

- public documentation matches real command behavior more closely
- fewer broken copy-paste examples
- stronger regression coverage for long-tail docs/spec drift
