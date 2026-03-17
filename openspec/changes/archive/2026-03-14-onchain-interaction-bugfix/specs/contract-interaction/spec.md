## MODIFIED Requirements

### Requirement: Transaction retry backoff
The contract caller SHALL use exponential backoff for transaction submission retries: 1s, 2s, 4s (doubling each attempt) instead of linear backoff.

#### Scenario: Exponential backoff timing
- **WHEN** a transaction submission fails and is retried
- **THEN** the delay between attempts SHALL follow exponential backoff: `2^attempt` seconds (1s, 2s, 4s)

## ADDED Requirements

### Requirement: Gas fee fallback warning
The contract caller SHALL log a WARNING when the block header's baseFee is nil and a fallback value is used.

#### Scenario: Missing baseFee triggers warning
- **WHEN** the block header does not contain a baseFee field
- **THEN** the caller SHALL log a warning message and use the default fallback value

### Requirement: X402 key material zeroing
The X402 signer SHALL zero key material using mutable byte slices. Converting a Go string to `[]byte` creates a copy; zeroing the copy does not clear the original string. The signer SHALL encode the key directly into a `[]byte` buffer using `hex.Encode` and zero that buffer after use.

#### Scenario: Key hex buffer is properly zeroed
- **WHEN** an EVM signer is created from a private key
- **THEN** the hex-encoded key material SHALL be stored in a mutable `[]byte` buffer (not derived from a string) and zeroed after the signer is created
