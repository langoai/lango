## Why

The cockpit dead-letter filter bar already supports query, adjudication, and subtype filtering, but operators still cannot narrow by the latest manual replay actor or dead-letter time window without leaving the cockpit.

## What Changes

- add `manual_replay_actor`, `dead_lettered_after`, and `dead_lettered_before` to the cockpit filter bar
- forward actor/time filters through the existing dead-letter list bridge
- document the landed actor/time-filtering slice in public docs and main OpenSpec specs

## Impact

- better cockpit triage for operator-facing recovery work
- no new backend endpoints
- no change to the current apply/reset model
