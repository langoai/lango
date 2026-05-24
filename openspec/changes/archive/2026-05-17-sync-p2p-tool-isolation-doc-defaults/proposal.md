# Proposal: Sync P2P Config Doc Defaults

## Why

Public configuration documentation must stay synchronized with `DefaultConfig()` defaults for P2P network and tool isolation settings.

`DefaultConfig()` sets several P2P defaults that were stale in public docs, including listen addresses, relay enablement, ZK handshake defaults, session token TTL, ZKP cache path, and tool isolation memory limits. This gave operators conflicting guidance about production P2P behavior and resource limits.

## What Changes

- Add an executable documentation guard that compares selected P2P default rows in `README.md` and `docs/configuration.md` against `config.DefaultConfig()`.
- Correct stale public documentation values discovered by that guard.
- Keep runtime configuration behavior unchanged.

## Non-Goals

- Changing P2P tool isolation defaults.
- Redesigning P2P networking, ZK, sandbox, or container runtime selection behavior.
- Rewriting the full configuration reference.
