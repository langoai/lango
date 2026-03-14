## MODIFIED Requirements

### Requirement: EntryPoint nonce for UserOps
The bundler client's GetNonce() SHALL retrieve the nonce from the EntryPoint contract via eth_call to EntryPoint.getNonce(address, uint192 key=0) using selector 0x35567e1a. The method SHALL NOT use eth_getTransactionCount (EOA nonce).

#### Scenario: Retrieve EntryPoint nonce
- **WHEN** GetNonce() is called for a smart account address
- **THEN** it calls eth_call with the EntryPoint.getNonce ABI-encoded calldata and returns the decoded uint256 nonce

### Requirement: Gas fee parameters for UserOps
The bundler client SHALL provide a GetGasFees() method that retrieves EIP-1559 gas fee parameters from the network. MaxFeePerGas SHALL be calculated as 2 * baseFee + priorityFee. The manager SHALL call GetGasFees() before gas estimation and set the fee parameters on the UserOp.

#### Scenario: Gas fees fetched successfully
- **WHEN** submitUserOp() constructs a UserOperation
- **THEN** MaxFeePerGas and MaxPriorityFeePerGas are set to non-zero values from GetGasFees() before gas estimation

#### Scenario: eth_maxPriorityFeePerGas not supported
- **WHEN** the RPC endpoint does not support eth_maxPriorityFeePerGas
- **THEN** GetGasFees() falls back to a default priority fee of 1.5 gwei

### Requirement: Deployment detection via eth_getCode
The Factory's IsDeployed() method SHALL check for contract code at the address using ethclient.CodeAt(). It SHALL return true if the code length is greater than zero. Errors from CodeAt() SHALL be propagated to the caller.

#### Scenario: Account is deployed
- **WHEN** IsDeployed() is called for an address with on-chain code
- **THEN** it returns true with nil error

#### Scenario: Account is not deployed
- **WHEN** IsDeployed() is called for an address with no on-chain code
- **THEN** it returns false with nil error

#### Scenario: RPC error during check
- **WHEN** IsDeployed() encounters an RPC error from CodeAt()
- **THEN** it returns the error to the caller (not silently returning false)

### Requirement: Post-deploy on-chain verification
After factory.Deploy() succeeds, GetOrDeploy() SHALL call IsDeployed() to verify the contract actually exists on-chain. If verification fails, GetOrDeploy() SHALL return an error.

#### Scenario: Deploy succeeds and verification passes
- **WHEN** factory.Deploy() returns success and IsDeployed() returns true
- **THEN** GetOrDeploy() returns AccountInfo with IsDeployed=true

#### Scenario: Deploy succeeds but verification fails
- **WHEN** factory.Deploy() returns success but IsDeployed() returns false
- **THEN** GetOrDeploy() returns an error indicating on-chain verification found no code

### Requirement: UserOp signing without double-hashing
The manager SHALL use wallet.SignTransaction() (raw sign) for UserOp hash signing instead of wallet.SignMessage() (which internally applies keccak256). Since computeUserOpHash() already returns a keccak256 digest, SignTransaction() produces the correct signature.

#### Scenario: UserOp is signed correctly
- **WHEN** submitUserOp() signs the UserOp hash
- **THEN** it calls SignTransaction() with the opHash bytes (no additional hashing)
