## Why

The A2A feature documentation still shows a stale config export/import workflow that omits the required profile name arguments. That creates immediate friction for operators following the guide.

## What Changes

- update the A2A feature docs to use the real `lango config export <name>` and `lango config import <file> --profile <name>` contract
- keep the example aligned with the default profile workflow already documented elsewhere

## Impact

- fewer copy-paste failures during A2A setup
- better alignment between feature docs and the actual CLI
