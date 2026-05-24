## Why

EIP-3009 `transferWithAuthorization` requires a fresh random nonce. The current `NewUnsigned` helper ignores the result of `crypto/rand.Read`, so an entropy-source failure can still produce an authorization with a zero or partial nonce. That turns a security-critical failure into a signed payment artifact.

## What Changes

- Make unsigned EIP-3009 authorization creation return an error when nonce generation fails.
- Ensure callers stop before signing when authorization creation fails.
- Add focused tests for nonce entropy failure and successful nonce generation.

## Impact

- Buyer-side paid invocation now fails closed if secure nonce generation fails.
- The EIP-3009 helper API gains an error return; current in-repo callers are updated in the same change.
- No configuration or schema changes.
