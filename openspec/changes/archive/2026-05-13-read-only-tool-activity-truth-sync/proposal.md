## Why

Several built-in inspection tools are marked read-only, but their activity metadata is either missing or too broad. `cron_list` and `cron_history` are currently classified as management activity even though they only inspect scheduler state. The P2P workspace and git inspection tools also leave activity unset despite being stable read/query surfaces.

That drift weakens capability summaries, activity-based search, and routing quality. Production metadata should describe what these tools actually do.

## What Changes

- Reclassify `cron_list` and `cron_history` as read-only query tools.
- Add explicit read/query activity metadata to the P2P workspace and git inspection tools.
- Add regressions that lock the updated capability metadata.
- Sync prompt/docs/spec wording so inspection paths are described consistently.

## Impact

- `tool-capability-layer`: activity metadata better matches actual inspection behavior.
- `automation-agent-tools`: cron inspection paths stop looking like management mutations.
- `p2p-workspace`: workspace and git review tools become easier to reason about in discovery and routing flows.
