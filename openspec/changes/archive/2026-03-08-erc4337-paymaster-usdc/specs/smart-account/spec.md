## MODIFIED Requirements

### Requirement: Session key paymaster allowlist
The `SessionPolicy` struct SHALL include an `allowedPaymasters` field (address array). When non-empty, `validateUserOp` SHALL enforce that the paymaster address in `paymasterAndData` is in the allowlist.

#### Scenario: Paymaster in allowlist
- **WHEN** `paymasterAndData` contains a paymaster address that is in `allowedPaymasters`
- **THEN** validation SHALL pass (return packed validAfter/validUntil, not 1)

#### Scenario: Paymaster not in allowlist
- **WHEN** `paymasterAndData` contains a paymaster address NOT in `allowedPaymasters`
- **THEN** validation SHALL return 1 (SIG_VALIDATION_FAILED)

#### Scenario: Empty allowlist allows all paymasters
- **WHEN** `allowedPaymasters` is empty (length 0)
- **THEN** any paymaster SHALL be allowed (backward compatible)

#### Scenario: No paymaster with allowlist set
- **WHEN** `paymasterAndData` is empty but `allowedPaymasters` is non-empty
- **THEN** validation SHALL pass (paymaster is optional)

#### Scenario: Short paymasterAndData
- **WHEN** `paymasterAndData` has fewer than 20 bytes
- **THEN** the paymaster allowlist check SHALL be skipped

#### Scenario: Session registration with allowlist
- **WHEN** a session key is registered with `allowedPaymasters`
- **THEN** the `_setSession` function SHALL persist the `allowedPaymasters` array
