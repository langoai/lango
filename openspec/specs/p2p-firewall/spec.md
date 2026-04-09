## Purpose

Capability spec for p2p-firewall. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Default Deny-All ACL Policy

The `Firewall` SHALL enforce a deny-all default policy on all incoming P2P queries. A query from a peer SHALL be denied unless at least one ACL rule with `action="allow"` matches both the peer DID and tool name. An explicit `action="deny"` rule that matches SHALL immediately reject the query, overriding any prior allow. Rules SHALL be evaluated in insertion order.

#### Scenario: Query allowed by explicit rule
- **WHEN** an ACL rule `{PeerDID: "did:lango:abc", Action: "allow", Tools: ["search"]}` exists and `FilterQuery("did:lango:abc", "search")` is called
- **THEN** `FilterQuery` SHALL return nil (allowed)

#### Scenario: Query denied when no matching allow rule
- **WHEN** no ACL rule exists for the requesting peer DID and tool combination
- **THEN** `FilterQuery` SHALL return an error containing "no matching allow rule"

#### Scenario: Explicit deny rule overrides allow
- **WHEN** both an allow rule and a deny rule match the same peer DID and tool
- **THEN** the deny rule SHALL cause `FilterQuery` to return an error containing "query denied by firewall rule"

#### Scenario: Wildcard peer DID matches all peers
- **WHEN** an ACL rule has `PeerDID: "*"` and `Action: "allow"` with `Tools: ["*"]`
- **THEN** `FilterQuery` SHALL return nil for any peer DID and any tool name

---

### Requirement: Per-Peer Rate Limiting

The `Firewall` SHALL enforce per-peer rate limits using a token-bucket rate limiter keyed by peer DID. When an ACL rule specifies `RateLimit > 0`, a limiter SHALL be created allowing at most `RateLimit` requests per minute. A wildcard rate limiter on `PeerDID="*"` SHALL apply globally to all peers. Rate limit checks MUST occur before ACL evaluation.

#### Scenario: Rate limit exceeded returns error
- **WHEN** a peer DID's rate limiter has no remaining tokens
- **THEN** `FilterQuery` SHALL return an error containing "rate limit exceeded"

#### Scenario: Global wildcard rate limit applied
- **WHEN** a rule with `PeerDID="*"` and `RateLimit=60` exists and 61 requests arrive in one minute
- **THEN** the 61st request SHALL be denied with "global rate limit exceeded"

#### Scenario: Peer without rate limit rule is not throttled
- **WHEN** no rate limit rule exists for a peer DID
- **THEN** the peer SHALL not be rate-limited regardless of request frequency

---

### Requirement: Tool Name Pattern Matching

ACL rule `Tools` fields SHALL support exact matches, prefix wildcard matching (e.g. `"search*"` matches `"search_web"` and `"search_local"`), and a bare `"*"` to match all tool names. An empty `Tools` slice SHALL match all tool names.

#### Scenario: Exact tool name match
- **WHEN** a rule has `Tools: ["search_web"]` and `FilterQuery` is called with tool `"search_web"`
- **THEN** the rule SHALL match

#### Scenario: Wildcard suffix tool match
- **WHEN** a rule has `Tools: ["search*"]` and `FilterQuery` is called with tool `"search_local"`
- **THEN** the rule SHALL match

#### Scenario: Non-matching tool name
- **WHEN** a rule has `Tools: ["search"]` and `FilterQuery` is called with tool `"payment_send"`
- **THEN** the rule SHALL NOT match

---

### Requirement: Response Sanitization

`Firewall.SanitizeResponse` SHALL remove all fields from a response map whose names match sensitive key patterns (case-insensitive): `db_path`, `file_path`, `internal_id`, `_internal`, and any field containing `password`, `secret`, `private_key`, or `token`. String values containing absolute file paths of 3 or more path segments SHALL have the path replaced with `[path-redacted]`. Nested maps SHALL be sanitized recursively.

#### Scenario: Sensitive key removed from response
- **WHEN** `SanitizeResponse` is called on `{"result": "ok", "private_key": "0xdeadbeef"}`
- **THEN** the returned map SHALL contain `"result"` but SHALL NOT contain `"private_key"`

#### Scenario: File path in string value redacted
- **WHEN** a response string value contains `/home/user/.lango/data/bolt.db`
- **THEN** `SanitizeResponse` SHALL replace it with `[path-redacted]`

#### Scenario: Nested sensitive fields removed
- **WHEN** `SanitizeResponse` is called on `{"data": {"token": "abc123", "value": 42}}`
- **THEN** the nested `"token"` field SHALL be removed and `"value"` SHALL be preserved

---

### Requirement: ZK Attestation for Responses

`Firewall.AttestResponse` SHALL call the configured `ZKAttestFunc` with the SHA-256 hash of the response and the SHA-256 hash of the agent's DID, returning the serialized ZK attestation proof. If no `ZKAttestFunc` is configured, the method SHALL return `(nil, nil)`.

#### Scenario: Attestation proof generated when function configured
- **WHEN** `SetZKAttestFunc` has been called with a non-nil function and `AttestResponse` is called
- **THEN** `AttestResponse` SHALL invoke the function and return the resulting proof bytes

#### Scenario: No attestation when function not configured
- **WHEN** `SetZKAttestFunc` has not been called and `AttestResponse` is called
- **THEN** `AttestResponse` SHALL return `(nil, nil)` without error

---

### Requirement: Validate overly permissive ACL rules
The firewall SHALL provide a `ValidateRule()` function that rejects allow rules with wildcard peer (`"*"`) combined with wildcard tools (empty list or containing `"*"`). Deny rules SHALL always pass validation.

#### Scenario: Wildcard peer with empty tools (allow)
- **WHEN** `ValidateRule` is called with `{PeerDID: "*", Action: "allow", Tools: []}`
- **THEN** it SHALL return an error "overly permissive rule: allow all peers with all tools is prohibited"

#### Scenario: Wildcard peer with wildcard tool (allow)
- **WHEN** `ValidateRule` is called with `{PeerDID: "*", Action: "allow", Tools: ["*"]}`
- **THEN** it SHALL return an error

#### Scenario: Wildcard peer with specific tools (allow)
- **WHEN** `ValidateRule` is called with `{PeerDID: "*", Action: "allow", Tools: ["echo"]}`
- **THEN** it SHALL return nil (allowed)

#### Scenario: Specific peer with wildcard tools (allow)
- **WHEN** `ValidateRule` is called with `{PeerDID: "did:key:abc", Action: "allow", Tools: ["*"]}`
- **THEN** it SHALL return nil (allowed)

#### Scenario: Wildcard deny rule
- **WHEN** `ValidateRule` is called with `{PeerDID: "*", Action: "deny", Tools: ["*"]}`
- **THEN** it SHALL return nil (deny rules always safe)

### Requirement: Dynamic Rule Management

`Firewall.AddRule` SHALL validate the rule using `ValidateRule()` before adding it. If validation fails, it SHALL return the error without adding the rule. On success, it SHALL append the ACL rule, create a rate limiter if `RateLimit > 0`, and return nil. `Firewall.RemoveRule` SHALL remove all rules matching the given peer DID and delete the associated rate limiter. `Firewall.Rules` SHALL return a copy of the current rule slice to prevent external mutation.

#### Scenario: AddRule rejects overly permissive rule
- **WHEN** `AddRule` is called with a wildcard allow-all rule
- **THEN** it SHALL return an error and NOT add the rule to the firewall

#### Scenario: AddRule accepts valid rule
- **WHEN** `AddRule` is called with a specific peer allow rule
- **THEN** it SHALL add the rule and return nil

#### Scenario: Rule added at runtime takes immediate effect
- **WHEN** `AddRule` is called with an allow rule for a peer DID
- **THEN** subsequent `FilterQuery` calls for that peer DID SHALL be evaluated against the new rule

#### Scenario: Rules returns independent copy
- **WHEN** the caller modifies the slice returned by `Firewall.Rules()`
- **THEN** the internal rule list SHALL NOT be affected

### Requirement: Initial rules backward compatibility
When constructing a Firewall with `New()`, overly permissive initial rules SHALL be loaded with a warning log (not rejected). This preserves backward compatibility with existing configurations while alerting operators.

#### Scenario: Overly permissive initial rule
- **WHEN** `New()` is called with a wildcard allow-all rule in the initial rules slice
- **THEN** the rule SHALL be loaded (backward compat) and a warning SHALL be logged
