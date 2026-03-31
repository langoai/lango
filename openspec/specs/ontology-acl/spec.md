## ADDED Requirements

### Requirement: Permission model
The system SHALL define three ordered permission levels: `PermRead` (1), `PermWrite` (2), `PermAdmin` (3). A principal with permission level N SHALL be allowed to perform any operation requiring level <= N.

#### Scenario: Admin principal performs read operation
- **WHEN** a principal with `PermAdmin` calls a `PermRead` method
- **THEN** the call SHALL succeed

#### Scenario: Read principal denied write operation
- **WHEN** a principal with `PermRead` calls a `PermWrite` method
- **THEN** the call SHALL return `ErrPermissionDenied`

### Requirement: ACLPolicy interface
The system SHALL define an `ACLPolicy` interface with a `Check(principal string, required Permission) error` method. Two implementations SHALL be provided: `AllowAllPolicy` (always returns nil) and `RoleBasedPolicy` (checks principal's role against required permission).

#### Scenario: AllowAllPolicy permits everything
- **WHEN** `AllowAllPolicy.Check` is called with any principal and any permission
- **THEN** it SHALL return nil

#### Scenario: RoleBasedPolicy grants matching permission
- **WHEN** `RoleBasedPolicy` has role `"ontologist" → PermAdmin` and Check is called with principal `"ontologist"` and required `PermWrite`
- **THEN** it SHALL return nil (PermAdmin >= PermWrite)

#### Scenario: RoleBasedPolicy denies insufficient permission
- **WHEN** `RoleBasedPolicy` has role `"chronicler" → PermRead` and Check is called with principal `"chronicler"` and required `PermWrite`
- **THEN** it SHALL return `ErrPermissionDenied`

### Requirement: System principal default
Empty string or `"system"` principal SHALL always receive `PermAdmin` access. Unknown principals (not in roles map) SHALL receive `PermRead` access. Principals with `peer:` prefix SHALL receive the permission level configured in `P2PPermission` (default `PermWrite`).

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

### Requirement: Service-layer permission guards
`ServiceImpl` SHALL check permissions via `checkPermission(ctx, perm)` as the first operation in every public method except `PredicateValidator`. The method SHALL extract the principal from context via `ctxkeys.PrincipalFromContext(ctx)`.

#### Scenario: Guard on write method
- **WHEN** `AssertFact` is called with a context carrying principal `"chronicler"` (PermRead) and ACL is enabled
- **THEN** it SHALL return `ErrPermissionDenied` without executing any logic

#### Scenario: Guard on read method allows read principal
- **WHEN** `ListTypes` is called with a context carrying principal `"chronicler"` (PermRead) and ACL is enabled
- **THEN** it SHALL succeed

#### Scenario: Nil ACL policy allows all
- **WHEN** any method is called and `ServiceImpl.acl` is nil
- **THEN** `checkPermission` SHALL return nil (backward compatible)

### Requirement: Permission method classification
The system SHALL classify OntologyService methods as follows:
- **PermRead**: GetType, ListTypes, GetPredicate, ListPredicates, ValidateTriple, SchemaVersion, ConflictSet, FactsAt, OpenConflicts, Resolve, Aliases, QueryTriples, GetEntityProperties, QueryEntities, GetEntity (15 methods + PredicateValidator unguarded)
- **PermWrite**: RegisterType, RegisterPredicate, StoreTriple, AssertFact, RetractFact, SetEntityProperty, DeclareSameAs (7 methods)
- **PermAdmin**: DeprecateType, DeprecatePredicate, MergeEntities, SplitEntity, ResolveConflict, DeleteEntityProperties (6 methods)

#### Scenario: All read methods accessible with PermRead
- **WHEN** a principal with `PermRead` calls any of the 15 PermRead methods
- **THEN** all calls SHALL pass the permission check

#### Scenario: Write methods require PermWrite
- **WHEN** a principal with `PermRead` calls any of the 7 PermWrite methods
- **THEN** all calls SHALL return `ErrPermissionDenied`

### Requirement: Principal context propagation
The `ctxkeys` package SHALL provide `WithPrincipal(ctx, principal)` and `PrincipalFromContext(ctx)` functions. A `WithPrincipal()` toolchain middleware SHALL copy `AgentNameFromContext(ctx)` to the principal context key for every tool execution.

#### Scenario: Middleware sets principal from agent name
- **WHEN** a tool handler is invoked with `AgentNameFromContext(ctx) == "ontologist"` and no explicit principal set
- **THEN** after `WithPrincipal` middleware, `PrincipalFromContext(ctx)` SHALL return `"ontologist"`

#### Scenario: No agent name yields empty principal
- **WHEN** a tool handler is invoked with no agent name in context
- **THEN** `PrincipalFromContext(ctx)` SHALL return `""`

### Requirement: ACL configuration
The system SHALL support ACL configuration via `ontology.acl.enabled` (bool) and `ontology.acl.roles` (map of principal name to permission string). When `ontology.acl.enabled` is false or absent, no ACL policy SHALL be injected (nil = allow all).

#### Scenario: ACL enabled with roles
- **WHEN** config has `ontology.acl.enabled: true` and `ontology.acl.roles: {"ontologist": "admin", "chronicler": "read"}`
- **THEN** `ServiceImpl` SHALL have a `RoleBasedPolicy` with the corresponding permission mapping

#### Scenario: ACL disabled
- **WHEN** config has `ontology.acl.enabled: false`
- **THEN** `ServiceImpl.acl` SHALL be nil
