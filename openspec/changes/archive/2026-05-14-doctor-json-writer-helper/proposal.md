## Why

The doctor CLI JSON renderer still constructed its own pretty-JSON encoder even though the repository already has a shared CLI JSON writer helper. That duplication increases maintenance cost and can let formatting behavior drift.

## What Changes

- route doctor CLI JSON rendering through the shared CLI JSON writer helper
- preserve the current payload shape and trailing-newline trimming behavior while removing duplicate encoder setup

## Impact

- lower maintenance cost for doctor CLI JSON rendering
- more uniform pretty-JSON writer behavior across CLI surfaces
