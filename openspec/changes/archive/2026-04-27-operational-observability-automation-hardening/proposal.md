## Why

After the earlier execution hardening batch, the next highest-value gaps were operational visibility and automation ergonomics:

- retry/dead-letter audit persistence failures were still logged too softly
- the team reputation bridge could silently skip trust-entry and kick paths on reputation-store failures
- `lango status` dead-letter subcommands still returned plain-text errors in JSON mode
- the dead-letter CLI surface did not expose explicit `offset` / `limit` pagination even though the tool already supported it
- retry automation could not override the fallback principal with an explicit actor

## What Changes

- Promote post-adjudication retry/dead-letter evidence persistence failures from low-signal warnings to operational errors.
- Promote team-reputation bridge reputation/trust-entry failures from debug-only logging to operational error/warn logging.
- Standardize dead-letter CLI machine-mode failures as structured JSON error payloads.
- Add `--offset` / `--limit` to `lango status dead-letters`.
- Add `--actor` to `lango status dead-letter retry`.
- Truth-align CLI and architecture docs plus docs-only OpenSpec requirements.

## Impact

- Operators get louder signals when canonical retry/dead-letter evidence fails to persist.
- Automation can page dead-letter backlogs predictably and force a specific replay actor.
- JSON consumers no longer need to special-case plain-text failure output from dead-letter status commands.
