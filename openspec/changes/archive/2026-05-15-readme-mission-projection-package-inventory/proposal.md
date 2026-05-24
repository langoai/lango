## Why

The repository ships `internal/proposal`, `internal/loopview`, and `internal/collabview`, but the README internal package tree omits all three packages even though they are core to the mission/proposal/collaboration runtime.

## What Changes

- add `proposal/`, `loopview/`, and `collabview/` to the README internal package tree
- add an executable guard that requires those package rows
- sync the main docs-only and test-coverage specs

## Impact

- more complete internal package inventory
- better discoverability of the mission projection stack
- stronger regression protection against future omissions
