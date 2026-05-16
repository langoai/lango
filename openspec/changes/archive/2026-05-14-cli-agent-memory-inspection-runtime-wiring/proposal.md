## Why

`lango memory agents` and `lango memory agent <name>` currently return placeholder status text instead of querying persistent agent memory entries, even though the underlying store already supports `ListAgentNames()` and `ListAll()`. The public CLI docs also document listing behavior that the runtime does not actually provide.

## What Changes

- wire `memory agents` to the persistent agent memory store and render real agent summaries
- wire `memory agent <name>` to the persistent agent memory store and render real entries with an optional `--limit`
- keep JSON and table output available through the Cobra command writer
- sync CLI docs and specs with the actual runtime behavior

## Impact

- turns placeholder agent-memory inspection into a real operator-facing feature
- closes a long-standing docs/spec/runtime drift on per-agent memory inspection
