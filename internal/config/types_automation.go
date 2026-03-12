package config

import "time"

// CronConfig defines cron scheduling settings.
type CronConfig struct {
	// Enable the cron scheduling system.
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// Default timezone for cron schedules (e.g. "Asia/Seoul").
	Timezone string `mapstructure:"timezone" json:"timezone"`

	// Maximum number of concurrently executing jobs.
	MaxConcurrentJobs int `mapstructure:"maxConcurrentJobs" json:"maxConcurrentJobs"`

	// Default session mode for jobs: "isolated" or "main".
	DefaultSessionMode string `mapstructure:"defaultSessionMode" json:"defaultSessionMode"`

	// How long to retain job execution history (e.g. "30d", "720h").
	HistoryRetention string `mapstructure:"historyRetention" json:"historyRetention"`

	// DefaultJobTimeout is the maximum duration for a single job execution (default: 30m).
	DefaultJobTimeout time.Duration `mapstructure:"defaultJobTimeout" json:"defaultJobTimeout"`

	// Default delivery channels when deliver_to is not specified (e.g. ["telegram"]).
	DefaultDeliverTo []string `mapstructure:"defaultDeliverTo" json:"defaultDeliverTo"`
}

// BackgroundConfig defines background task execution settings.
type BackgroundConfig struct {
	// Enable the background task system.
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// Time in milliseconds before an agent turn is auto-yielded to background.
	YieldMs int `mapstructure:"yieldMs" json:"yieldMs"`

	// Maximum number of concurrently running background tasks.
	MaxConcurrentTasks int `mapstructure:"maxConcurrentTasks" json:"maxConcurrentTasks"`

	// TaskTimeout is the maximum duration for a single background task (default: 30m).
	TaskTimeout time.Duration `mapstructure:"taskTimeout" json:"taskTimeout"`

	// Default delivery channels when channel is not specified (e.g. ["telegram"]).
	DefaultDeliverTo []string `mapstructure:"defaultDeliverTo" json:"defaultDeliverTo"`
}

// WorkflowConfig defines workflow engine settings.
type WorkflowConfig struct {
	// Enable the workflow engine.
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// Maximum number of concurrently executing workflow steps.
	MaxConcurrentSteps int `mapstructure:"maxConcurrentSteps" json:"maxConcurrentSteps"`

	// Default timeout for a single workflow step (e.g. "10m").
	DefaultTimeout time.Duration `mapstructure:"defaultTimeout" json:"defaultTimeout"`

	// Directory to store workflow state for resume capability.
	StateDir string `mapstructure:"stateDir" json:"stateDir"`

	// Default delivery channels when deliver_to is not specified (e.g. ["telegram"]).
	DefaultDeliverTo []string `mapstructure:"defaultDeliverTo" json:"defaultDeliverTo"`
}

// PaymentConfig defines blockchain payment settings.
type PaymentConfig struct {
	// Enable blockchain payment features.
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// WalletProvider selects the wallet backend: "local", "rpc", or "composite".
	WalletProvider string `mapstructure:"walletProvider" json:"walletProvider"`

	// Network defines blockchain network parameters.
	Network PaymentNetworkConfig `mapstructure:"network" json:"network"`

	// Limits defines spending restrictions.
	Limits SpendingLimitsConfig `mapstructure:"limits" json:"limits"`

	// X402 defines X402 protocol interception settings.
	X402 X402Config `mapstructure:"x402" json:"x402"`
}

// PaymentNetworkConfig defines blockchain network parameters.
type PaymentNetworkConfig struct {
	// ChainID is the EVM chain ID (default: 84532 = Base Sepolia).
	ChainID int64 `mapstructure:"chainId" json:"chainId"`

	// RPCURL is the JSON-RPC endpoint for the blockchain network.
	RPCURL string `mapstructure:"rpcUrl" json:"rpcUrl"`

	// USDCContract is the USDC token contract address on the target chain.
	USDCContract string `mapstructure:"usdcContract" json:"usdcContract"`
}

// SpendingLimitsConfig defines spending restrictions for payment transactions.
type SpendingLimitsConfig struct {
	// MaxPerTx is the maximum amount per transaction in USDC (e.g. "1.00").
	MaxPerTx string `mapstructure:"maxPerTx" json:"maxPerTx"`

	// MaxDaily is the maximum daily spending in USDC (e.g. "10.00").
	MaxDaily string `mapstructure:"maxDaily" json:"maxDaily"`

	// AutoApproveBelow is the amount below which transactions are auto-approved.
	AutoApproveBelow string `mapstructure:"autoApproveBelow" json:"autoApproveBelow"`
}

// X402Config defines X402 protocol interception settings.
type X402Config struct {
	// AutoIntercept enables automatic interception of HTTP 402 responses.
	AutoIntercept bool `mapstructure:"autoIntercept" json:"autoIntercept"`

	// MaxAutoPayAmount is the maximum amount to auto-pay for X402 challenges.
	MaxAutoPayAmount string `mapstructure:"maxAutoPayAmount" json:"maxAutoPayAmount"`
}

// A2AConfig defines Agent-to-Agent protocol settings.
type A2AConfig struct {
	// Enable A2A protocol support.
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// BaseURL is the external URL where this agent is reachable.
	BaseURL string `mapstructure:"baseUrl" json:"baseUrl"`

	// AgentName is the name advertised in the Agent Card.
	AgentName string `mapstructure:"agentName" json:"agentName"`

	// AgentDescription is the description in the Agent Card.
	AgentDescription string `mapstructure:"agentDescription" json:"agentDescription"`

	// RemoteAgents is a list of external A2A agents to integrate as sub-agents.
	RemoteAgents []RemoteAgentConfig `mapstructure:"remoteAgents" json:"remoteAgents"`
}

// RemoteAgentConfig defines an external A2A agent to connect to.
type RemoteAgentConfig struct {
	// Name is the local name for this remote agent.
	Name string `mapstructure:"name" json:"name"`

	// AgentCardURL is the URL to fetch the agent card from.
	// Typically: https://host/.well-known/agent.json
	AgentCardURL string `mapstructure:"agentCardUrl" json:"agentCardUrl"`
}
