## MODIFIED Requirements

### Requirement: CREATE2 address computation
The Factory SHALL compute the counterfactual address using the correct CREATE2 formula: `keccak256(0xff ++ factory ++ deploymentSalt ++ keccak256(proxyCreationCode ++ abi.encode(singleton)))`. The proxyCreationCode SHALL be fetched from the factory contract's `proxyCreationCode()` view function and cached.

#### Scenario: Correct initCodeHash computation
- **WHEN** ComputeAddress is called
- **THEN** the initCodeHash SHALL be `keccak256(proxyCreationCode ++ singletonPadded)` where singletonPadded is the singleton address left-padded to 32 bytes

#### Scenario: proxyCreationCode caching
- **WHEN** ComputeAddress is called multiple times
- **THEN** the proxyCreationCode SHALL be fetched from the factory contract only once and cached for subsequent calls

#### Scenario: ComputeAddress requires context
- **WHEN** ComputeAddress is called
- **THEN** it SHALL accept a context.Context parameter and return (common.Address, error) to handle RPC failures

### Requirement: Deploy address resolution
The Factory.Deploy() SHALL first attempt to parse the actual deployed address from the contract call result. If the result does not contain a parseable address, it SHALL fall back to ComputeAddress.

#### Scenario: Deploy returns actual on-chain address
- **WHEN** the createProxyWithNonce call returns an address in the result data
- **THEN** Deploy SHALL return that address instead of computing it

#### Scenario: Deploy fallback to computed address
- **WHEN** the createProxyWithNonce result does not contain a parseable address
- **THEN** Deploy SHALL fall back to ComputeAddress with the corrected CREATE2 formula

### Requirement: UserOp receipt verification
The Manager.submitUserOp() SHALL wait for the UserOp to be included in an on-chain transaction by polling GetUserOperationReceipt(). It SHALL return the actual on-chain transaction hash instead of the UserOp hash.

#### Scenario: Successful UserOp confirmation
- **WHEN** a UserOp is submitted and included on-chain
- **THEN** submitUserOp SHALL return the on-chain transaction hash from the receipt

#### Scenario: UserOp reverted on-chain
- **WHEN** a UserOp is submitted but the receipt indicates failure (success=false)
- **THEN** submitUserOp SHALL return an error indicating on-chain revert

#### Scenario: UserOp receipt timeout
- **WHEN** no receipt is received within 2 minutes
- **THEN** submitUserOp SHALL return a timeout error

### Requirement: Session key UserOp hash correctness
The session key manager SHALL use the same UserOp hash algorithm as the EntryPoint contract (ComputeUserOpHash) instead of raw byte concatenation. The session manager SHALL accept entryPoint address and chainID via options.

#### Scenario: Session key signature matches EntryPoint
- **WHEN** a UserOp is signed with a session key
- **THEN** the signing digest SHALL match ComputeUserOpHash(op, entryPoint, chainID) which uses proper ABI encoding with 32-byte padding, keccak256 on variable-length fields, packed gas values, and double-hash with entryPoint+chainID

#### Scenario: Session manager configuration
- **WHEN** a session manager is created
- **THEN** it SHALL accept WithEntryPoint and WithChainID options for UserOp hash computation
