## Purpose

Generic smart contract interaction layer for EVM chains. Provides ABI caching, read (view/pure) and write (state-changing tx) capabilities, agent tools, and CLI commands.

## Requirements

### Requirement: Contract caller provides read and write access
The contract package SHALL expose a `ContractCaller` interface with `Read` and `Write` methods that the concrete `Caller` struct implements. Consumers SHALL accept the interface type instead of the concrete struct.

#### Scenario: ContractCaller interface defined
- **WHEN** a package needs to call smart contracts
- **THEN** it SHALL depend on the `ContractCaller` interface, not the concrete `*Caller` struct

#### Scenario: Caller satisfies ContractCaller
- **WHEN** `*Caller` is used where `ContractCaller` is expected
- **THEN** it SHALL compile without error (compile-time interface check via `var _ ContractCaller = (*Caller)(nil)`)

### Requirement: ABI cache provides thread-safe parsed ABI storage
The system SHALL provide an `ABICache` that stores parsed `abi.ABI` objects keyed by `chainID:address`. The cache SHALL be safe for concurrent access via `sync.RWMutex`. The cache SHALL support `Get`, `Set`, and `GetOrParse` (lazy parse + cache) operations.

#### Scenario: Cache miss triggers parse and store
- **WHEN** `GetOrParse` is called with a valid ABI JSON for an uncached address
- **THEN** the ABI is parsed, stored in cache, and returned without error

#### Scenario: Cache hit returns existing entry
- **WHEN** `GetOrParse` is called for an address already in cache
- **THEN** the cached ABI is returned without re-parsing

#### Scenario: Invalid ABI JSON returns error
- **WHEN** `GetOrParse` is called with malformed JSON
- **THEN** an error is returned and nothing is cached

### Requirement: Contract caller reads view/pure functions
The system SHALL provide a `Caller.Read()` method that packs arguments via `abi.Pack()`, calls `ethclient.CallContract()`, and unpacks the result via `method.Outputs.Unpack()`. No transaction or gas is required.

#### Scenario: Successful read call
- **WHEN** `Read` is called with a valid ABI, method name, and arguments
- **THEN** the packed calldata is sent via `CallContract` and the decoded result is returned

#### Scenario: Method not found in ABI
- **WHEN** `Read` is called with a method name not present in the ABI
- **THEN** an error containing the method name is returned

### Requirement: Contract caller writes state-changing transactions
The system SHALL provide a `Caller.Write()` method that packs arguments, builds an EIP-1559 transaction (nonce, gas estimation, base fee), signs via `wallet.WalletProvider`, submits with retry, and polls for receipt confirmation. Write() SHALL check the receipt status after waiting for the transaction receipt. If the receipt status is not successful (ReceiptStatusSuccessful), Write() SHALL return an ErrTxReverted error with the transaction hash and status. If the receipt times out, Write() SHALL return an ErrReceiptTimeout error instead of silently returning a partial result.

#### Scenario: Successful write transaction
- **WHEN** `Write` is called with valid parameters and the RPC is available
- **THEN** a signed transaction is submitted and the result includes `TxHash` and `GasUsed`

#### Scenario: Nonce serialization prevents collisions
- **WHEN** multiple concurrent `Write` calls are made
- **THEN** nonce acquisition is serialized via mutex to prevent nonce reuse

#### Scenario: Transaction reverts on-chain
- **WHEN** a Write() call submits a transaction that gets mined but reverts
- **THEN** Write() returns an error wrapping ErrTxReverted with the tx hash and receipt status

#### Scenario: Receipt timeout
- **WHEN** a Write() call submits a transaction but the receipt is not available within the timeout period
- **THEN** Write() returns an error wrapping ErrReceiptTimeout with the tx hash

### Requirement: Agent tools expose contract interaction
The system SHALL register three agent tools: `contract_read` (SafetyLevel Safe), `contract_call` (SafetyLevel Dangerous), and `contract_abi_load` (SafetyLevel Safe). Tools SHALL be registered under the `"contract"` catalog category.

#### Scenario: contract_read tool returns decoded data
- **WHEN** the `contract_read` tool is invoked with address, ABI, and method
- **THEN** it calls `Caller.Read()` and returns the decoded data

#### Scenario: contract_call tool returns tx hash
- **WHEN** the `contract_call` tool is invoked with address, ABI, method, and optional value
- **THEN** it calls `Caller.Write()` and returns the transaction hash

### Requirement: CLI commands validate contract parameters
The system SHALL provide `lango contract read`, `lango contract call`, and `lango contract abi load` CLI commands under GroupID `"infra"`. Commands SHALL validate ABI parsing and method existence. The `blockLangoExec` guard SHALL include `"lango contract"`.

#### Scenario: CLI read validates ABI and method
- **WHEN** `lango contract read --address 0x... --abi ./erc20.json --method balanceOf` is run
- **THEN** the ABI is parsed, the method is validated, and a guidance message is shown

#### Scenario: CLI abi load parses and reports
- **WHEN** `lango contract abi load --address 0x... --file ./erc20.json` is run
- **THEN** the ABI is parsed and method/event counts are displayed

### Requirement: Contract feature documentation page
The documentation site SHALL include a `docs/features/contracts.md` page documenting smart contract interaction capabilities including ABI cache, read (view/pure), and write (state-changing) operations, with experimental warning, architecture overview, agent tools listing, and configuration reference.

#### Scenario: Contract feature docs page exists
- **WHEN** the documentation site is built
- **THEN** `docs/features/contracts.md` SHALL exist with sections for ABI cache, read operations, write operations, agent tools, and configuration

### Requirement: Contract CLI documentation page
The documentation site SHALL include a `docs/cli/contract.md` page documenting `lango contract read`, `lango contract call`, and `lango contract abi load` commands with flags tables and example output following the `docs/cli/payment.md` pattern.

#### Scenario: Contract CLI docs page exists
- **WHEN** the documentation site is built
- **THEN** `docs/cli/contract.md` SHALL exist with sections for read, call, and abi load subcommands

#### Scenario: Each subcommand documented with flags
- **WHEN** a user reads the contract CLI reference
- **THEN** each subcommand SHALL include a flags table with `--address`, `--abi`, `--method`, `--args`, `--chain-id`, and `--output` flags documented

### Requirement: Transaction retry backoff
The contract caller SHALL use exponential backoff for transaction submission retries: 1s, 2s, 4s (doubling each attempt) instead of linear backoff.

#### Scenario: Exponential backoff timing
- **WHEN** a transaction submission fails and is retried
- **THEN** the delay between attempts SHALL follow exponential backoff: `2^attempt` seconds (1s, 2s, 4s)

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
