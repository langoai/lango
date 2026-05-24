## Overview

`zkexport` is a small standalone utility, so the simplest way to make it testable is to move parsing and routing into a helper that accepts explicit args and IO streams.

## Decision

- Replace global `flag` parsing in `main()` with a local `FlagSet` helper
- Route success output through injected stdout and usage/error text through injected stderr
- Use a seam for prover-service construction so tests do not need gnark setup

## Consequences

- The command becomes deterministic under test
- Automation can rely on stdout vs stderr separation without intercepting process-global streams
