## Why

The migrated-family docs guard was introduced before the agent inspection subset finished moving to `--output table|json`. Without expanding coverage, the public docs for `agent status`, `list`, `tools`, and `hooks` could drift back to stale `--json` examples without the guard noticing.

## What Changes

- expand the migrated-family docs guard to include the agent inspection subset
- record that coverage explicitly in the main test-coverage spec

## Impact

- keeps agent inspection docs under the same anti-regression umbrella as the other migrated CLI families
- reduces the chance of agent-inspection docs drifting back to stale boolean `--json`
