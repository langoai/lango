## Why

Public quick references should be executable starting points. Several README and CLI index entries currently omit required operands or required flags that the detailed command docs and Cobra command definitions already require. That makes first-run operator guidance fail at the terminal and weakens confidence in the CLI surface.

## What Changes

- Align README, `docs/cli/index.md`, and affected feature docs with the actual CLI usage for memory, P2P firewall/session, and config get commands.
- Strengthen docs guard tests so regressions fail when quick references drop required operands or required output flags.
- Keep the detailed command docs unchanged unless a quick-reference inconsistency is found there.

## Impact

- No runtime behavior changes.
- Public docs become more accurate and immediately runnable.
- Repository tests cover the previously missed quick-reference drift.
