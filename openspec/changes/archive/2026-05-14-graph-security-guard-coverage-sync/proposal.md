## Why

The generic migrated-family output guards were added before the graph and security inspection CLI families completed their `--output table|json` migration. Without expanding guard scope, those families would remain uncovered despite now following the same UX contract.

## What Changes

- expand migrated-family CLI output guards to include graph and security inspection command families
- record that coverage explicitly in the main test-coverage spec

## Impact

- keeps graph and security under the same anti-regression umbrella as the other migrated CLI families
- reduces the chance of graph/security drift back to boolean `--json`
