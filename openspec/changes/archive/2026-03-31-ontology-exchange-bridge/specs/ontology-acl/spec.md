## MODIFIED Requirements

### Requirement: System principal default
Empty string or `"system"` principal SHALL always receive `PermAdmin` access. Unknown principals (not in roles map) SHALL receive `PermRead` access. Principals with `peer:` prefix SHALL receive the permission level configured in `P2PPermission` (default `PermWrite`).

#### Scenario: Peer principal gets P2PPermission
- **WHEN** `RoleBasedPolicy.Check` is called with principal `"peer:did:lango:abc123"` and P2PPermission is PermWrite
- **THEN** it SHALL allow PermRead and PermWrite but deny PermAdmin
