## Overview

This is a parser regression-only change. The commands already behave correctly today; the goal is to make that stability explicit.

## Decision

- Drive the tests through `newRootCmd()` instead of the individual command constructors
- Use the real persistent `--mode` flag path so future parser/layout refactors are covered

## Consequences

- Utility subcommands remain insulated from the interactive mode flag
