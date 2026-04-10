## MODIFIED Requirements

### Requirement: DID-to-address resolution

The `ResolveAddress` function SHALL be replaced by an `AddressResolver` interface with `ResolveAddress(did string) (common.Address, error)`. The `DefaultAddressResolver` SHALL dispatch by DID version: v1 DIDs are resolved directly (secp256k1 decompress), v2 DIDs are resolved via BundleResolver → settlement key → address derivation.

#### Scenario: v1 DID resolves directly
- **WHEN** `ResolveAddress("did:lango:<secp256k1-hex>")` is called
- **THEN** the resolver SHALL decompress the secp256k1 key and derive the Ethereum address (existing behavior)

#### Scenario: v2 DID resolves via bundle
- **WHEN** `ResolveAddress("did:lango:v2:<hash>")` is called and the bundle is available
- **THEN** the resolver SHALL look up the IdentityBundle, extract the settlement key (secp256k1), and derive the Ethereum address

#### Scenario: v2 DID without bundle returns error
- **WHEN** `ResolveAddress("did:lango:v2:<hash>")` is called and no bundle is available
- **THEN** the resolver SHALL return an error indicating the bundle is not found
