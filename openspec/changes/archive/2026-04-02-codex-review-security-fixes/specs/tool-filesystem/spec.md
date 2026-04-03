## MODIFIED Requirements

### Requirement: P2P delete restriction
The `Delete` method MUST use the canonical P2P context marker from the `ctxkeys` package (`ctxkeys.IsP2PRequest`) to detect P2P origin. When a P2P request is detected, only single-file or empty-directory deletion (`os.Remove`) SHALL be permitted. The filesystem package MUST NOT define its own P2P context key.

#### Scenario: P2P delete uses canonical context key
- **WHEN** `Delete` is called with a context carrying `ctxkeys.WithP2PRequest`
- **THEN** it SHALL use `os.Remove` (single file/empty dir only)
- **AND** it SHALL NOT use `os.RemoveAll`

#### Scenario: Non-P2P delete allows recursive removal
- **WHEN** `Delete` is called without the P2P context marker
- **THEN** it SHALL use `os.RemoveAll` for recursive deletion
