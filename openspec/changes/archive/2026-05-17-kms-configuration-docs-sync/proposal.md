# KMS Configuration Docs Sync

## Summary

Synchronize the public configuration reference with the current Cloud KMS bootstrap behavior.

## Problem

`README.md` and `docs/security/encryption.md` describe most Cloud KMS behavior, but the public configuration references are not fully synchronized with `internal/config/types_security.go` and `internal/bootstrap/kms_env.go`. `docs/configuration.md` only lists `security.signer.provider`, and the README config table omits some optional KMS fields plus the bootstrap-time environment override needed before encrypted profile config can be loaded.

## Goals

- Add accurate Cloud KMS configuration coverage to `README.md` and `docs/configuration.md`.
- Document the bootstrap-time `LANGO_KMS_FALLBACK_TO_LOCAL=false` behavior in the configuration reference.
- Add a repository test guard so the configuration reference cannot silently drop the KMS rows or bootstrap override note.

## Non-Goals

- No runtime behavior changes.
- No changes to KMS provider implementations.
- No new CLI surface.
