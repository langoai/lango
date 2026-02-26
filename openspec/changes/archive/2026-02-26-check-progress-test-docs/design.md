# Design: Test Coverage & Documentation Sync

## Approach

### Test Strategy
- Follow existing test patterns (testify assertions, zap nop logger, mock dependencies)
- Focus on unit tests that don't require external services (no Docker, no network)
- Use table-driven tests where applicable
- Test error paths and edge cases, not just happy paths

### Documentation Strategy
- Fix incorrect references (owner-shield CLI command → configuration-only)
- Add missing configuration keys by cross-referencing `internal/config/types.go`
- Maintain existing documentation format and style

### Prioritization
1. P2P discovery/identity (highest risk — network-facing code with no tests)
2. CLI commands (user-facing code with no validation)
3. Infrastructure (workflow, cron, background — core scheduling)
4. Security/sandbox (defense-in-depth validation)
5. Remaining packages (librarian, payment CLI, p2p routes)
6. Documentation fixes (lowest risk but important for coherence)

## Key Design Decisions
- Tests should be self-contained (no external dependencies like Docker, DHT, blockchain)
- Use mocks/stubs for external interfaces (libp2p host, DHT, wallet)
- CLI tests verify command tree structure, not full execution (avoids bootstrap dependency)
