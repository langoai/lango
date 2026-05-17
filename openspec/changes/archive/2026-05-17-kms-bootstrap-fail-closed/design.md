# Design

## Approach

The existing bootstrap KMS unwrap branch remains the only place that changes.
When a KMS slot and KMS provider config are available:

- On KMS success, bootstrap still sets the master key and skips passphrase.
- On KMS provider creation failure or unwrap failure:
  - if `KMSConfig.FallbackToLocal` is true, bootstrap keeps the current warning
    and passphrase fallback behavior;
  - if `KMSConfig.FallbackToLocal` is false, bootstrap returns an error before
    secure-provider detection or passphrase acquisition.

Env-driven bootstrap reads `LANGO_KMS_FALLBACK_TO_LOCAL` because encrypted
profile config is loaded after credential acquisition. Invalid boolean values
are ignored and preserve the default fallback-enabled behavior.

## Testability

Introduce a package-level seam for the KMS provider factory, mirroring existing
bootstrap seams such as `acquirePassphrase` and `confirmStorePass`. Tests can
force provider initialization or decrypt failures without requiring build-tagged
cloud KMS implementations.

## Boundaries

- This change does not alter KMS provider implementations.
- This change does not alter `CompositeCryptoProvider` behavior for normal
  runtime crypto operations.
- This change does not remove passphrase fallback when the operator explicitly
  leaves fallback enabled.
