## Overview

The TTY approval provider is small and self-contained, so the safest improvement is to replace its direct use of global terminal state and streams with package-level seams that default to the same runtime behavior.

## Decisions

### Add terminal and stream seams

`ttyIsTerminal`, `ttyInput`, and `ttyError` preserve current production behavior while allowing tests to inject deterministic state.

### Cover all three decision paths

Tests now verify:
- single approval (`y` / `yes`)
- persistent approval (`a` / `always`)
- denial (anything else)

## Non-Goals

- No change to approval routing semantics
- No change to TTY fallback being unavailable in non-terminal environments
