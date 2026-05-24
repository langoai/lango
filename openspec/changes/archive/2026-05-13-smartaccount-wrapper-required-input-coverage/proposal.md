## Why

The smart-account tool cluster still has wrapper-contract gaps. In particular, `session_key_create` declares `targets` and `duration` as required, but its handler currently tolerates missing values and silently falls back. Other smart-account tools also lack direct regression coverage for their required wrapper inputs.

## What Changes

- Add a required string-slice extractor for wrapper validation.
- Make `session_key_create` fail closed on missing `targets` and `duration`.
- Add exact missing-parameter regressions for the smart-account tool cluster: `session_key_create`, `session_key_revoke`, `session_execute`, `policy_check`, `module_install`, `module_uninstall`, and `paymaster_approve`.
- Sync prompt/public docs and specs to the same wrapper contract.

## Impact

- `smart-account`: required tool inputs become explicitly enforced at the wrapper boundary.
- `production-readiness`: high-risk account/session/paymaster tools align with the same actionable missing-parameter standard used elsewhere.
