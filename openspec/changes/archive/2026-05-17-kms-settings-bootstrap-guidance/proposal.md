# KMS Settings Bootstrap Guidance

## Summary

Improve the Security KMS settings form so operators see the same bootstrap fallback guidance that is documented in the public configuration reference.

## Problem

The KMS settings form exposes the relevant KMS fields, but the fallback description still says local fallback is only for key signing. Profile-backed KMS fallback also applies to KMS encryption/decryption after profile config is loaded, while encrypted profile bootstrap KMS unwrap happens earlier and cannot read profile-backed `security.kms.fallbackToLocal` until credentials are acquired. Operators need an in-product hint that fail-closed bootstrap requires the `LANGO_KMS_FALLBACK_TO_LOCAL=false` environment override.

## Goals

- Make the KMS fallback field description cover profile-loaded signing, encryption, and decryption fallback.
- Surface the bootstrap-time `LANGO_KMS_FALLBACK_TO_LOCAL=false` override in the settings form.
- Add tests that prevent the settings copy from drifting away from the documented KMS behavior.

## Non-Goals

- No changes to KMS runtime behavior.
- No new settings fields.
- No changes to the KMS CLI commands.
