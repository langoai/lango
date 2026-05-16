## Why

The agent hooks CLI snapshot still used its own ad-hoc pretty-JSON encoder setup even after the agent package introduced a shared pretty-JSON helper path.

## What Changes

- route the agent hooks JSON snapshot through the shared CLI JSON helper
- keep the existing payload shape unchanged while removing repeated encoder setup

## Impact

- lower maintenance cost for the agent CLI package
- one place to adjust pretty-JSON writer behavior if needed later
