## Overview

Use the existing session database and `agentmemory.EntStore` to back `lango memory agents` and `lango memory agent`.

## Decisions

- reuse `session.NewEntStore(cfg.Session.DatabasePath)` and `agentmemory.NewEntStore(store.Client())`
- `memory agents` returns per-agent summaries with name, entry count, and most recent update time
- `memory agent <name>` returns the stored entries for that agent, ordered by store order and truncated by `--limit`
- both commands use `cmd.OutOrStdout()` for text and JSON output

## Risks

- encrypted or unavailable session databases will still fail through the existing session store open path
- mitigated by preserving explicit error returns instead of hiding store failures
