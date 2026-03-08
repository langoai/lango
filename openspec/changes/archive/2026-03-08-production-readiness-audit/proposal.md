## Why

The codebase accumulated production-readiness issues under an MVP mindset: unimplemented stubs that crash at runtime, `context.TODO()` in production handlers, dead code, and zero test coverage on security-critical packages (wallet, security, payment, smart account). This audit eliminates all findings from a three-pronged review (stubs, test gaps, broken flows) before the codebase moves to production.

## What Changes

- **Fix enclave provider crash**: Replace hard `fmt.Errorf` with config-time validation listing valid providers
- **Implement Telegram media download**: Complete `DownloadFile` stub with HTTP GET, 30s timeout, error handling
- **Remove dead `NewX402Client`**: Eliminate unused factory function and its `context.TODO()` usage
- **Document GVisor stub**: Improve doc comments, add stub behavior tests
- **Add wallet tests**: 5 test files covering local, composite, create, RPC wallet, and utilities
- **Add security tests**: KeyRegistry CRUD, SecretsStore CRUD with mock CryptoProvider
- **Add payment tests**: Service.Send error branches, History, RecordX402Payment, failTx
- **Add smartaccount tests**: Factory CREATE2, session crypto roundtrip, errors unwrap, ABI encoder, paymaster, policy syncer, types
- **Add economy risk tests**: Risk factors (trust, amount sigmoid, verifiability), strategy matrix (9 combinations)
- **Add P2P tests**: Team conflict resolution (4 strategies), protocol messages, remote agent accessors

## Capabilities

### New Capabilities
- `production-readiness`: Covers stub elimination, dead code removal, and comprehensive test coverage for security-critical packages

### Modified Capabilities
- `blockchain-wallet`: Test coverage added for local/composite/RPC wallet and create flows
- `security-fixes`: Test coverage added for KeyRegistry and SecretsStore
- `payment-service`: Test coverage added for Send, Balance, History, RecordX402Payment
- `smart-account`: Test coverage added for factory, session crypto, ABI encoder, paymaster, policy syncer, types
- `economy-risk`: Test coverage added for risk factors and strategy selection
- `p2p-team-coordination`: Test coverage added for conflict resolution
- `p2p-protocol`: Test coverage added for messages and remote agent
- `channel-telegram`: DownloadFile stub implemented
- `x402-protocol`: Dead code removed, context.TODO eliminated

## Impact

- **14 packages affected**: wallet, security, payment, smartaccount (5 sub-packages), economy/risk, p2p/team, p2p/protocol, channels/telegram, x402, app, sandbox
- **No API changes**: All fixes are internal; no public interfaces modified
- **No dependency changes**: No new imports required
- **Risk**: Low — primarily test additions and stub fixes with no behavioral changes to existing working code
