## Why

`passphrase.acquireWithIO(...)` already accepts injected input and warning writers, but its interactive prompt branch still delegates to process-global prompt output. This leaks bootstrap passphrase prompt text outside the acquisition stream seam and can pollute stdout during CLI bootstrap.

## What Changes

- Route interactive passphrase acquisition prompt text through the writer passed to `acquireWithIO`.
- Preserve hidden terminal password reading and the existing acquisition priority chain.
- Keep keyring warnings on the same injected writer.
- Add regression tests for both existing-passphrase and first-run confirmation prompts.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `passphrase-acquisition`: Extend stream-seam behavior to interactive passphrase prompts.

## Impact

- Affected code: `internal/security/passphrase`.
- Affected tests: passphrase acquisition tests.
- No CLI flags, config fields, storage formats, or crypto behavior change.
