# KMS Bootstrap Fail-Closed Fallback

## Why

`security.kms.fallbackToLocal` is documented and exposed as the operator
control for whether KMS failures may fall back to local/passphrase unlock. The
bootstrap KMS KEK unwrap path currently falls through to passphrase acquisition
for provider initialization and unwrap failures regardless of that setting.

For production security, disabling fallback must be fail-closed: if an operator
requires KMS unlock and KMS is unavailable, startup must stop instead of silently
accepting a weaker credential path.

## What Changes

- Make bootstrap KMS provider initialization and unwrap failures respect
  `security.kms.fallbackToLocal`.
- Add `LANGO_KMS_FALLBACK_TO_LOCAL=false` so env-driven bootstrap can request
  fail-closed behavior before encrypted profile config is loaded.
- Keep existing warning-and-passphrase fallback behavior when
  `fallbackToLocal=true`.
- Add regression tests proving fail-closed behavior does not call passphrase
  acquisition.
- Update KMS/bootstrap specs to document the fail-closed contract.

## Impact

- Modified capabilities: `cloud-kms`, `bootstrap-lifecycle`.
- Runtime behavior changes only for bootstrap attempts with an envelope KMS slot,
  KMS config/provider configured, and `fallbackToLocal=false`.
- No CLI flag or config schema change.
