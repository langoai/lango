# Design

## Approach

This is a UI copy and guard-test change in the settings layer.

The existing `Security KMS` form already contains all profile-backed KMS fields. We will keep the field set unchanged and update the copy around `kms_fallback_to_local` so it clearly explains:

- Profile-backed fallback applies after profile config is loaded.
- The profile-loaded fallback covers KMS signing, encryption, and decryption operations.
- Fail-closed encrypted profile bootstrap requires `LANGO_KMS_FALLBACK_TO_LOCAL=false` with `LANGO_KMS_PROVIDER`.

## Scope Control

The work is intentionally limited to the settings form, its tests, and the `settings-security-advanced` spec. Public docs were already synchronized in the preceding KMS configuration docs change.

## Verification

- A focused settings form test will fail before the copy update.
- Focused settings tests, full build/test, and OpenSpec validation will run before commit.
