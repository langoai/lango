## MODIFIED Requirements

### Requirement: Handshake authentication

#### Scenario: Bundle cached only after authentication
- **WHEN** a handshake challenge or response is received
- **THEN** the bundle SHALL be cached only after signature verification succeeds
- **AND** alias registration SHALL occur only after approval succeeds

#### Scenario: v2 DID requires bundle with matching signing key
- **WHEN** a handshake peer claims a `did:lango:v2:` DID
- **THEN** the handshake SHALL require a non-nil bundle
- **AND** `ComputeDIDv2(bundle) == SenderDID` SHALL be verified
- **AND** `bytes.Equal(PublicKey, Bundle.SigningKey.PublicKey)` SHALL be verified
- **AND** failure of any check SHALL reject the handshake

#### Scenario: v1 DID matches public key
- **WHEN** a handshake peer claims a v1 `did:lango:` DID and provides a public key
- **THEN** `DIDFromPublicKey(PublicKey).ID == SenderDID` SHALL be verified
- **AND** mismatch SHALL reject the handshake

#### Scenario: Auto-approve uses existing alias only
- **WHEN** `AutoApproveKnownPeers` is enabled
- **THEN** session lookup SHALL use `DIDAlias.CanonicalDID` only if an alias was previously registered
- **AND** `bundle.LegacyDID` SHALL NOT be used for session lookup (unverifiable)
