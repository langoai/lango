## Overview

`eip3009.NewUnsigned` is the single construction point for buyer-side EIP-3009 authorizations. It should fail before signing if it cannot produce a full 32-byte cryptographic nonce.

## Decisions

### Return Errors from Authorization Construction

Change `NewUnsigned` from returning only `*UnsignedAuth` to returning `(*UnsignedAuth, error)`. A nonce generation failure is not recoverable at this layer because proceeding would create an authorization with weak replay protection.

### Use a Reader Boundary for Tests

Keep the production entropy source as `crypto/rand.Reader`, but route nonce generation through an unexported reader variable/helper so tests can simulate reader errors and short reads without replacing global crypto state.

### Fail Before Signing in P2P Paid Invocation

The paid invocation tool should wrap the construction error and return before `Sign` is called. This preserves the existing flow: spending checks and auto-approval still happen before building the authorization, and remote invocation is not attempted without a signed authorization.

## Risks

- The helper signature changes, but it is internal and has a small call surface.
- Tests must restore the injected reader after failure simulations to avoid cross-test contamination.
