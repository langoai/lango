## Why

The `package-consolidation` main spec still described deleted pre-consolidation package directories as if they were current paths. That makes the spec less trustworthy and sends maintainers toward directories that no longer exist.

## What Changes

- sync the `package-consolidation` main spec to the current package locations
- add an executable guard so the deleted package-directory claims cannot silently return

## Impact

- main spec better matches the current repository structure
- less confusion around where the consolidated helpers now live
- stronger regression protection for deleted package-path drift
