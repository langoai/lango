## Why

The `domain-tool-builders` main spec still described the removed economy builder in terms of a deleted `internal/app/tools_economy.go` file, even though the current source of truth is the domain-owned `internal/economy/tools.go` builder and app module wiring.

## What Changes

- sync the wording to the current builder ownership and wiring path
- extend the existing economy builder-path guard to cover this main spec too

## Impact

- main specs better match the current code layout
- less confusion about economy tool registration ownership
- stronger regression protection for another deleted builder-path reference
