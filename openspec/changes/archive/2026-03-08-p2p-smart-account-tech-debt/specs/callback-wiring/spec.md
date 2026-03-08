# Spec: Callback Wiring Completion

## Requirements

### REQ-1: Session on-chain registration/revocation callbacks

When `SessionValidatorAddress` is configured, the session manager must wire `WithOnChainRegistration` and `WithOnChainRevocation` options that call the `SessionValidatorClient`.

**Scenarios:**
- Given SessionValidator address configured, when a session key is created, then `RegisterSessionKey` is called on-chain.
- Given SessionValidator address configured, when a session key is revoked, then `RevokeSessionKey` is called on-chain.

### REQ-2: Budget engine sync via OnChainTracker

The `OnChainTracker.SetCallback` must forward spending data to the budget engine's `Record()` method, not just log.

### REQ-3: P2P CardFn provides agent info

The protocol handler must receive a `CardFn` that returns the agent's name, DID, and peer ID.

### REQ-4: Gossip service must be started

After creation, `gossip.Start()` must be called to begin the publish/subscribe loops.

### REQ-5: Team invoke must use handler

The team coordinator's `invokeFn` must route through the P2P protocol handler to send real remote tool invocation requests, not return a stub error.

### REQ-6: SmartAccount components must be accessible

All smart account sub-components (session manager, policy engine, module registry, bundler, paymaster, on-chain tracker) must be accessible via public accessor methods from the App struct.
