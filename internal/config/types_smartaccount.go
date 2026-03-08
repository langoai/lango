package config

import "time"

// SmartAccountConfig defines ERC-7579 smart account settings.
type SmartAccountConfig struct {
	Enabled           bool   `mapstructure:"enabled" json:"enabled"`
	FactoryAddress    string `mapstructure:"factoryAddress" json:"factoryAddress"`
	EntryPointAddress string `mapstructure:"entryPointAddress" json:"entryPointAddress"`
	Safe7579Address   string `mapstructure:"safe7579Address" json:"safe7579Address"`
	FallbackHandler   string `mapstructure:"fallbackHandler" json:"fallbackHandler"`
	BundlerURL        string `mapstructure:"bundlerURL" json:"bundlerURL"`

	Session   SmartAccountSessionConfig   `mapstructure:"session" json:"session"`
	Modules   SmartAccountModulesConfig   `mapstructure:"modules" json:"modules"`
	Paymaster SmartAccountPaymasterConfig `mapstructure:"paymaster" json:"paymaster"`
}

// SmartAccountSessionConfig defines session key settings.
type SmartAccountSessionConfig struct {
	MaxDuration     time.Duration `mapstructure:"maxDuration" json:"maxDuration"`
	DefaultGasLimit uint64        `mapstructure:"defaultGasLimit" json:"defaultGasLimit"`
	MaxActiveKeys   int           `mapstructure:"maxActiveKeys" json:"maxActiveKeys"`
}

// SmartAccountPaymasterConfig defines paymaster settings for gasless transactions.
type SmartAccountPaymasterConfig struct {
	Enabled          bool   `mapstructure:"enabled" json:"enabled"`
	Provider         string `mapstructure:"provider" json:"provider"`                 // "circle"|"pimlico"|"alchemy"
	RPCURL           string `mapstructure:"rpcURL" json:"rpcURL"`
	TokenAddress     string `mapstructure:"tokenAddress" json:"tokenAddress"`         // USDC address
	PaymasterAddress string `mapstructure:"paymasterAddress" json:"paymasterAddress"`
	PolicyID         string `mapstructure:"policyId" json:"policyId,omitempty"`
}

// SmartAccountModulesConfig defines module contract addresses.
type SmartAccountModulesConfig struct {
	SessionValidatorAddress string `mapstructure:"sessionValidatorAddress" json:"sessionValidatorAddress"`
	SpendingHookAddress     string `mapstructure:"spendingHookAddress" json:"spendingHookAddress"`
	EscrowExecutorAddress   string `mapstructure:"escrowExecutorAddress" json:"escrowExecutorAddress"`
}
