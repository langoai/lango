## Context

Currently, `LocalCryptoProvider` reads the passphrase from the config file's `security.passphrase`. Since AI Agents can access the config file via filesystem tools, the passphrase is at risk of exposure. This violates Zero Trust security principles.

Additionally, if the passphrase is changed or entered incorrectly, access to existing encrypted data becomes impossible.

## Goals / Non-Goals

**Goals:**
- Remove passphrase reading from config file
- Accept passphrase input via interactive terminal prompt
- Detect incorrect passphrase input early via checksum
- Support data migration when passphrase is changed
- Document security modes (Local vs RPC)

**Non-Goals:**
- RPCProvider/Companion logic changes
- Introducing new encryption algorithms
- GUI-based passphrase input

## Decisions

### 1. Interactive Prompt Implementation

**Choice**: Use `golang.org/x/term` package
```go
func promptPassphrase() (string, error) {
    fmt.Print("Enter passphrase: ")
    bytes, err := term.ReadPassword(int(syscall.Stdin))
    return string(bytes), err
}
```

**Alternatives considered**:
- `bufio.Scanner`: Input is displayed on terminal as-is → security vulnerability
- Third-party library: Unnecessary dependency addition

### 2. Checksum Storage Method

**Choice**: Store passphrase hash with salt
```
security_config table:
- key: "default" 
- salt: <random bytes>
- checksum: SHA256(passphrase + salt)  ← newly added
```

**Behavior**:
1. First setup: passphrase input → salt generation → checksum storage
2. Subsequent starts: passphrase input → checksum verification → error on mismatch

### 3. Migration Process

**Choice**: Provided as CLI command
```bash
lango security migrate-passphrase
```

**Behavior**:
```
┌─────────────────────────────────────────┐
│ 1. Enter current passphrase (verify)    │
│ 2. Enter new passphrase (confirm 2x)    │
│ 3. Retrieve all Secrets                 │
│ 4. Each Secret: decrypt with old key →  │
│              re-encrypt with new key    │
│ 5. Update Salt/Checksum                 │
│ 6. Completion message                   │
└─────────────────────────────────────────┘
```

### 4. Config Passphrase Removal

**Choice**: Deprecated with warning
- Show warning at startup if passphrase exists in existing config
- Not actually used (ignored)
- Field removal in next major version

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Cannot start in headless environments | Recommend using RPCProvider+Companion |
| Data loss on passphrase loss | Document that no recovery method exists, recommend backups |
| Data corruption on migration failure | Transaction processing + rollback support |
| Brute-force possible on checksum leak | Salt + strong hash algorithm used |

## Open Questions

1. Are passphrase minimum length/complexity requirements needed? (Answer: required, 12 characters minimum, at least one uppercase letter, one lowercase letter, one number, and one special character)
2. Should migration failure automatically roll back, or provide a manual recovery guide? (Answer: automatically roll back all changes)
