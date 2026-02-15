## Context

Security CLI commands (`lango security secrets list/set/delete`, `status`, `migrate-passphrase`) use a two-phase initialization pattern: `bootstrap.Run()` acquires the passphrase and initializes DB + crypto, then each subcommand's `RunE` opens a second DB connection and calls `initLocalCrypto()` which calls `passphrase.Acquire()` again. This results in users being prompted for the passphrase twice per command.

The onboarding menu currently lists "Providers" near the bottom (after Embedding & RAG), but providers must be configured before agents since agent configuration depends on provider selection.

## Goals / Non-Goals

**Goals:**
- Eliminate the double passphrase prompt by passing the full `*bootstrap.Result` through the security command tree instead of only `*config.Config`
- Reuse the already-initialized `CryptoProvider` and `*ent.Client` from bootstrap, removing redundant DB connections
- Place "Providers" first in the onboarding menu to reflect the natural configuration order

**Non-Goals:**
- Changing the bootstrap lifecycle or passphrase acquisition logic itself
- Modifying the memory command's loader (separate concern)
- Adding new security commands or features

## Decisions

### Pass `*bootstrap.Result` instead of `*config.Config` to security commands

The security command tree accepts a `bootLoader func() (*bootstrap.Result, error)` closure instead of `cfgLoader func() (*config.Config, error)`. Each subcommand's `RunE` calls `bootLoader()` once, gets config, DB client, and crypto provider in a single call, then defers `boot.DBClient.Close()`.

**Rationale**: This is the minimal change that eliminates the root cause. The bootstrap result already contains everything the security commands need. Alternative approaches (caching the passphrase globally, sharing a crypto singleton) would introduce mutable global state.

### Replace `initLocalCrypto` with `secretsStoreFromBoot`

The `initLocalCrypto(store)` function, which independently acquired the passphrase, is replaced by `secretsStoreFromBoot(boot)` which creates a `SecretsStore` directly from `boot.Crypto` and `boot.DBClient`. This centralizes the creation pattern and ensures the bootstrap-initialized crypto provider is reused.

### Refactor `migrateSecrets` to accept `CryptoProvider` interface

Instead of receiving a raw passphrase string and re-creating the old provider, `migrateSecrets` now accepts `security.CryptoProvider` (the already-initialized provider from bootstrap) for decryption. Only the new provider is created from the user's new passphrase input.

## Risks / Trade-offs

- [DB client lifetime shift] Each command's `RunE` now owns the `boot.DBClient.Close()` lifecycle via defer, whereas previously the loader closure handled it. → Mitigated by consistent `defer boot.DBClient.Close()` in every RunE.
- [Migrate command uses `NewEntStoreWithClient`] The migrate command wraps `boot.DBClient` in an `EntStore` for access to `MigrateSecrets()`. This creates a thin wrapper without owning the connection. → Acceptable since `NewEntStoreWithClient` does not manage connection lifecycle.
