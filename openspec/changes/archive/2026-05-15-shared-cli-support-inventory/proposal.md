## Why

The repository ships shared CLI support packages `internal/cli/cliboot` and `internal/cli/clihttp`, but the public inventory docs either omit them entirely or only cover one of them.

## What Changes

- add `cli/cliboot/` and `cli/clihttp/` rows to the architecture project-structure inventory
- add the missing `clihttp/` row to the README internal tree
- add an executable guard for those shared support package rows
- sync the main docs-only and test-coverage specs

## Impact

- more complete public architecture inventory
- better discoverability of shared CLI support layers
- stronger regression protection against future omissions
