## Overview

The `serve` command already runs through Cobra, but its success-path banner and summary still bypass the command writer. A few lightweight seams make that path testable without changing boot/application behavior.

## Decision

- Route the startup banner and summary through `cmd.OutOrStdout()`
- Add seams for boot loading, logging init/sync, app construction, and shutdown waiting so the command can complete under test
- Leave the broader TUI/workbench/cockpit startup stderr banners untouched in this slice

## Consequences

- Wrapper capture becomes consistent for `lango serve`
- The command can be regression-tested without starting the full app runtime
