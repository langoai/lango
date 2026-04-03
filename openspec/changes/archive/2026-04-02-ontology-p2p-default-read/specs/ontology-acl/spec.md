## MODIFIED Requirements

### Requirement: System principal default
Empty string or `"system"` principal SHALL always receive `PermAdmin` access. Unknown principals (not in roles map) SHALL receive `PermRead` access. Principals with `peer:` prefix SHALL receive the permission level configured in `P2PPermission` (default `PermRead`).

#### Scenario: Empty principal gets full access
- **WHEN** `RoleBasedPolicy.Check` is called with principal `""` and required `PermAdmin`
- **THEN** it SHALL return nil

#### Scenario: System principal gets full access
- **WHEN** `RoleBasedPolicy.Check` is called with principal `"system"` and required `PermAdmin`
- **THEN** it SHALL return nil

#### Scenario: Unknown principal gets read-only
- **WHEN** `RoleBasedPolicy.Check` is called with principal `"unknown_agent"` (not in roles) and required `PermWrite`
- **THEN** it SHALL return `ErrPermissionDenied`

#### Scenario: Unknown principal can read
- **WHEN** `RoleBasedPolicy.Check` is called with principal `"unknown_agent"` and required `PermRead`
- **THEN** it SHALL return nil

#### Scenario: P2P peer defaults to read-only
- **WHEN** `RoleBasedPolicy.Check` is called with principal `"peer:abc123"` and required `PermWrite`
- **AND** no explicit `P2PPermission` override is configured
- **THEN** it SHALL return `ErrPermissionDenied`

#### Scenario: P2P peer can read by default
- **WHEN** `RoleBasedPolicy.Check` is called with principal `"peer:abc123"` and required `PermRead`
- **AND** no explicit `P2PPermission` override is configured
- **THEN** it SHALL return nil

#### Scenario: P2P peer with explicit write override
- **WHEN** `SetP2PPermission(PermWrite)` has been called
- **AND** `RoleBasedPolicy.Check` is called with principal `"peer:abc123"` and required `PermWrite`
- **THEN** it SHALL return nil

### Requirement: ACL configuration
The system SHALL support ACL configuration via `ontology.acl.enabled` (bool), `ontology.acl.roles` (map of principal name to permission string), and `ontology.acl.p2pPermission` (string, default `"read"`). When `ontology.acl.enabled` is false or absent, no ACL policy SHALL be injected (nil = allow all).

#### Scenario: ACL enabled with roles
- **WHEN** config has `ontology.acl.enabled: true` and `ontology.acl.roles: {"ontologist": "admin", "chronicler": "read"}`
- **THEN** `ServiceImpl` SHALL have a `RoleBasedPolicy` with the corresponding permission mapping

#### Scenario: ACL disabled
- **WHEN** config has `ontology.acl.enabled: false`
- **THEN** `ServiceImpl.acl` SHALL be nil

#### Scenario: P2P permission config default
- **WHEN** `ontology.acl.p2pPermission` is not set or empty
- **THEN** the system SHALL use `"read"` as the default P2P permission
