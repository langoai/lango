## Why

The README internal CLI inventory still uses hyphen-compressed shorthand inside the security row even though the current command families are clearer and more truthful in slash-separated form.

## What Changes

- update the README internal tree security row to use slash-separated keyring, recovery, and KMS wording
- strengthen the existing README security inventory guard to reject the stale hyphen shorthand
- sync the main docs-only and test-coverage specs with that slash-form contract

## Impact

- more truthful README inventory wording
- clearer mapping from inventory docs to actual command paths
- stronger regression protection against stale shorthand returning
