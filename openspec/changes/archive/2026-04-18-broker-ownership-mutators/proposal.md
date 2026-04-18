## Why

Broker-owned bootstrap and runtime readers are in place, but the remaining payment and settlement mutator paths still reconstruct raw Ent dependencies in the parent process. That leaves the broker boundary incomplete for the most obvious stateful payment flows and keeps production code coupled to parent-side ORM handles.

## What Changes

- Replace raw Ent-backed payment dependency construction with storage-facing transaction store and limiter capabilities.
- Update payment CLI `send`, `balance`, and `info` initialization to use storage-provided collaborators instead of extracting `*ent.Client` from `session.EntStore`.
- Update app payment and P2P settlement wiring to depend on storage-backed transaction persistence interfaces instead of direct parent-side Ent clients.
- Keep service construction in app/payment packages while moving all storage-specific payment transaction persistence behind explicit interfaces.

## Capabilities

### New Capabilities

- `payment-tx-store`: Explicit storage-facing payment transaction persistence capability used by payment and settlement services.

### Modified Capabilities

- `brokered-storage`: Broker/runtime ownership now covers stateful payment mutator setup through storage-facing capabilities rather than raw parent-side ORM access.
- `cli-payment-management`: Payment CLI stateful commands initialize through storage-facing payment capabilities instead of direct Ent extraction.
- `payment-service`: Payment service persistence and spending-limit tracking use explicit transaction store interfaces rather than direct Ent clients.

## Impact

- Affected code:
  - `internal/storage/*`
  - `internal/payment/*`
  - `internal/wallet/*`
  - `internal/p2p/settlement/*`
  - `internal/cli/payment/*`
  - `internal/app/*`
- Affected systems:
  - Payment CLI stateful commands
  - App payment initialization
  - P2P settlement persistence
- Dependencies:
  - No new external dependencies
  - New internal interfaces and facade capabilities
