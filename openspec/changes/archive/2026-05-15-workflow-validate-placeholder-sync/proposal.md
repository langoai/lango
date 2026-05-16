## Why

The workflow family is now present in the architecture inventory and README internal tree, but both still stop at a bare `validate` token instead of the real shipped command shape `validate <file>`.

## What Changes

- update the README internal tree and architecture inventory to describe `validate <file>`
- sync the existing remaining-inventory guard and main specs with that placeholder-aware workflow contract

## Impact

- more truthful workflow inventory docs
- clearer operator guidance for the validate subcommand
- stronger regression protection against placeholder loss
