## Overview

This is a docs/test parity change. The existing runtime contract is already encoded in `cmd/lango/main.go`: bare `lango` checks `isInteractiveFn()`, and non-interactive execution returns `cmd.Help()` instead of starting the workbench. Existing root command tests cover that behavior.

## Decisions

### Preserve Runtime Behavior

Keep the root command behavior unchanged:

- Interactive bare `lango` starts the mission workbench.
- Non-interactive bare `lango` prints help through the Cobra command output stream and returns nil.
- `lango cockpit` and `lango chat` keep their existing non-interactive actionable-error behavior.

### Use Public Docs Guard Coverage

Add a docs guard instead of adding another root command behavior test. The runtime behavior already has direct coverage; the missing contract is public documentation parity.

## Risks

- If the docs use different wording across files, the guard can become brittle. Use a small set of precise snippets that encode the behavioral contract without forcing full paragraph duplication.
- If runtime behavior changes later to return an error instead of help, this guard will intentionally fail until docs are updated with the new contract.
