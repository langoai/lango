## ADDED Requirements

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
The system SHALL provide a `Caller.Write()` method that packs arguments, builds an EIP-1559 transaction (nonce, gas estimation, base fee), signs via `wallet.WalletProvider`, submits with retry, and polls for receipt confirmation.

#### Scenario: Successful write transaction
- **WHEN** `Write` is called with valid parameters and the RPC is available
- **THEN** a signed transaction is submitted and the result includes `TxHash` and `GasUsed`

#### Scenario: Nonce serialization prevents collisions
- **WHEN** multiple concurrent `Write` calls are made
- **THEN** nonce acquisition is serialized via mutex to prevent nonce reuse

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
