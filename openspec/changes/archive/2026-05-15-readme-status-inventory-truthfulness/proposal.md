## Why

The README internal CLI inventory still describes the status family vaguely as "dead-letter status views" instead of listing the actual shipped dead-letter subcommands.

## What Changes

- update the README internal tree status row to the current explicit command inventory
- sync the existing remaining-inventory guard to require that exact status wording
- sync the main docs-only and test-coverage specs

## Impact

- more truthful README inventory wording
- better status-family discoverability
- stronger regression protection against vague status shorthand
