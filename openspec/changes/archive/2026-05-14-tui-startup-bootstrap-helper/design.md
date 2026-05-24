## Overview

The three interactive TUI entrypoints all perform the same sequence: init logging, redirect stdlib logging, set the active profile banner context, and emit a startup notice. A shared helper can own that sequence without changing the visible behavior.

## Decision

- Extract one helper that receives the log filename, logging seams, startup writer, and startup line
- Let callers keep ownership of app creation and shutdown behavior
- Rely on the existing startup-notice regressions for behavior preservation

## Consequences

- Less duplication in the main entrypoint
- Lower risk of one TUI startup path drifting from the others
