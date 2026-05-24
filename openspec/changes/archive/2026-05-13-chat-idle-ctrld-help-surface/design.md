## Context

`handleKey()` processes `Ctrl+D` before the state-specific handlers, so the immediate quit path is always available. The only missing artifact is the idle/failed help surface, which currently shows `Enter`, `Alt+Enter`, `Ctrl+C`, and `/help` but omits `Ctrl+D`.

## Goals / Non-Goals

**Goals:**

- Make the immediate quit path visible in the idle/failed chat help bar.
- Keep the change narrow and help-surface-only.

**Non-Goals:**

- Change `Ctrl+D` behavior.
- Change streaming or approval-state help.

## Decisions

- Add `Ctrl+D quit` to the idle/failed help entries.
  - Rationale: this matches the actual control flow and complements the `Ctrl+C quit x2` entry.

## Risks / Trade-offs

- [Idle help bar becomes slightly denser] → The bar already truncates safely for narrow widths, and the added key is high-value operator guidance.
