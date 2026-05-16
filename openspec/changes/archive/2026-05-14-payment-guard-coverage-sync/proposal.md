## Why

The generic migrated-family guards were added before the payment CLI family finished moving to `--output table|json`. Without expanding the guard scope, payment would remain an uncovered gap even though it now follows the same contract.

## What Changes

- expand migrated-family CLI output guards to include the payment command family
- record that coverage explicitly in the main test-coverage spec

## Impact

- keeps payment under the same anti-regression umbrella as the other migrated CLI families
- reduces the chance of payment-specific drift back to boolean `--json`
