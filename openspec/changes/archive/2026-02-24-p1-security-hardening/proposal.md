# P1 Security Hardening

## Problem

The P2P network layer has three medium-priority security gaps:

1. **Passphrase storage**: Passphrases are stored only as disk-based keyfiles, which are vulnerable to filesystem access. No OS-native secure storage integration exists.

2. **Tool isolation**: Remote peer tool invocations execute in the same process as the main application, exposing passphrases, private keys, and session tokens in shared process memory.

3. **Session management**: Sessions rely solely on TTL-based expiration with no explicit invalidation. There is no mechanism to revoke sessions on security events (repeated failures, reputation drops) or via CLI.

## Solution

Three independent security hardening items:

- **P1-4: OS Keyring Integration** — Use macOS Keychain / Linux secret-service / Windows DPAPI as the highest-priority passphrase source, with graceful fallback to keyfile/interactive/stdin.

- **P1-5: Tool Execution Process Isolation** — Execute remote P2P tool invocations in isolated subprocesses with clean environments, timeout enforcement, and JSON protocol communication.

- **P1-6: Session Explicit Invalidation** — Add explicit session invalidation methods, auto-invalidation on security events (repeated failures, reputation drops), and CLI management commands.

## Impact

- Prerequisite for P2 work (P1-4 → P2-7 SQLCipher, P1-5 → P2-8 Container Sandbox)
- Strengthens security posture for production P2P deployments
- Minimal shared-file conflicts between the three items
