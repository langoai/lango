## MODIFIED Requirements

### Requirement: Configuration structure
The Config struct SHALL include a `Payment PaymentConfig` field after the A2A field. PaymentConfig SHALL contain: Enabled (bool), WalletProvider (string: local/rpc/composite), Network (PaymentNetworkConfig), Limits (SpendingLimitsConfig), X402 (X402Config).

PaymentNetworkConfig SHALL contain: ChainID (int64, default 84532), RPCURL (string), USDCContract (string).
SpendingLimitsConfig SHALL contain: MaxPerTx (string, default "1.00"), MaxDaily (string, default "10.00"), AutoApproveBelow (string).
X402Config SHALL contain: AutoIntercept (bool), MaxAutoPayAmount (string).

#### Scenario: Default payment config
- **WHEN** no payment config is specified
- **THEN** payment is disabled with Base Sepolia defaults (chainId 84532, maxPerTx "1.00", maxDaily "10.00")

### Requirement: Configuration validation
The Validate function SHALL check: when payment.enabled is true, payment.network.rpcUrl MUST be non-empty. payment.walletProvider MUST be one of "local", "rpc", or "composite".

#### Scenario: Payment enabled without RPC URL
- **WHEN** payment.enabled is true and rpcUrl is empty
- **THEN** validation fails with an error message

### Requirement: Environment variable substitution
The substituteEnvVars function SHALL expand `${VAR}` patterns in `payment.network.rpcUrl`.

#### Scenario: RPC URL from environment
- **WHEN** rpcUrl is set to `${BASE_RPC_URL}`
- **THEN** the environment variable value is substituted
