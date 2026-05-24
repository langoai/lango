## Why

The shipped command surface uses `lango agent trace show <trace-id>`, but README, the public CLI index, and their guard/spec wording still referred to a nonexistent `lango agent trace <id>` shortcut. That left public quick references out of sync with the actual CLI contract.

## What Changes

- update `README.md` and `docs/cli/index.md` to use `lango agent trace show <trace-id>`
- update the executable quick-reference guard to enforce the current command shape across both files
- sync the corresponding docs-only and test-coverage requirements to the real command contract

## Impact

- public quick references now match the actual `agent trace show` command surface
- reduced operator confusion when copying agent trace commands from docs
- stronger regression protection against stale `agent trace <id>` wording
