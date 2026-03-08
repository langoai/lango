# Spec: Stub to Real Implementations

## Requirements

### REQ-1: CLI commands must call real services

All smart account CLI commands (deploy, info, session create/list/revoke, policy show/set, module list/install, paymaster status/approve) must initialize dependencies from bootstrap and call actual service methods.

**Scenarios:**
- Given `lango account deploy`, when executed, then `manager.GetOrDeploy()` is called and the real account address is displayed.
- Given `lango account session create`, when valid flags are provided, then a real session key is created and the key ID is returned.

### REQ-2: PolicySyncer bridges Go and on-chain policies

A `PolicySyncer` must support:
- `PushToChain`: Write Go-side policy limits to the SpendingHook contract
- `PullFromChain`: Read on-chain config and update the Go-side policy
- `DetectDrift`: Compare and report differences between Go and on-chain policies

### REQ-3: Paymaster recovery with retry and fallback

A `RecoverableProvider` must wrap any `PaymasterProvider` with:
- Exponential-backoff retry for transient errors (`ErrPaymasterTimeout`)
- Immediate failure for permanent errors (`ErrPaymasterRejected`, `ErrInsufficientToken`)
- Configurable fallback: abort or switch to direct gas
- `IsTransient()`/`IsPermanent()` error classification functions
