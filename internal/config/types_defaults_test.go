package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig_Server(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()

	assert.Equal(t, "localhost", cfg.Server.Host)
	assert.Equal(t, 18789, cfg.Server.Port)
	assert.True(t, cfg.Server.HTTPEnabled)
	assert.True(t, cfg.Server.WebSocketEnabled)
}

func TestDefaultConfig_Agent(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()

	assert.Equal(t, "anthropic", cfg.Agent.Provider)
	assert.Empty(t, cfg.Agent.Model)
	assert.Equal(t, 4096, cfg.Agent.MaxTokens)
	assert.InDelta(t, 0.7, cfg.Agent.Temperature, 1e-9)
	assert.Equal(t, 5*time.Minute, cfg.Agent.RequestTimeout)
	assert.Equal(t, 2*time.Minute, cfg.Agent.ToolTimeout)
}

func TestDefaultConfig_Logging(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()

	assert.Equal(t, "info", cfg.Logging.Level)
	assert.Equal(t, "console", cfg.Logging.Format)
}

func TestDefaultConfig_Session(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()

	assert.Equal(t, "~/.lango/lango.db", cfg.Session.DatabasePath)
	assert.Equal(t, 24*time.Hour, cfg.Session.TTL)
	assert.Equal(t, 50, cfg.Session.MaxHistoryTurns)
}

func TestDefaultConfig_Tools(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()

	assert.Equal(t, 30*time.Second, cfg.Tools.Exec.DefaultTimeout)
	assert.True(t, cfg.Tools.Exec.AllowBackground)
	assert.Equal(t, int64(10*1024*1024), cfg.Tools.Filesystem.MaxReadSize)
	assert.False(t, cfg.Tools.Browser.Enabled)
	assert.True(t, cfg.Tools.Browser.Headless)
	assert.Equal(t, 5*time.Minute, cfg.Tools.Browser.SessionTimeout)
}

func TestDefaultConfig_Security(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()

	assert.True(t, cfg.Security.Interceptor.Enabled)
	assert.Equal(t, ApprovalPolicyDangerous, cfg.Security.Interceptor.ApprovalPolicy)
	assert.False(t, cfg.Security.DBEncryption.Enabled)
	assert.Equal(t, 4096, cfg.Security.DBEncryption.CipherPageSize)
	assert.True(t, cfg.Security.KMS.FallbackToLocal)
	assert.Equal(t, 5*time.Second, cfg.Security.KMS.TimeoutPerOperation)
	assert.Equal(t, 3, cfg.Security.KMS.MaxRetries)
}

func TestDefaultConfig_Graph(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()

	assert.False(t, cfg.Graph.Enabled)
	assert.Equal(t, "bolt", cfg.Graph.Backend)
	assert.Equal(t, 2, cfg.Graph.MaxTraversalDepth)
	assert.Equal(t, 10, cfg.Graph.MaxExpansionResults)
}

func TestDefaultConfig_MCP(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()

	assert.False(t, cfg.MCP.Enabled)
	assert.Equal(t, 30*time.Second, cfg.MCP.DefaultTimeout)
	assert.Equal(t, 25000, cfg.MCP.MaxOutputTokens)
	assert.Equal(t, 30*time.Second, cfg.MCP.HealthCheckInterval)
	assert.True(t, cfg.MCP.AutoReconnect)
	assert.Equal(t, 5, cfg.MCP.MaxReconnectAttempts)
}

func TestDefaultConfig_P2P(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()

	assert.False(t, cfg.P2P.Enabled)
	assert.Len(t, cfg.P2P.ListenAddrs, 2)
	assert.Equal(t, "~/.lango/p2p", cfg.P2P.KeyDir)
	assert.True(t, cfg.P2P.EnableRelay)
	assert.True(t, cfg.P2P.EnableMDNS)
	assert.Equal(t, 50, cfg.P2P.MaxPeers)
	assert.Equal(t, 30*time.Second, cfg.P2P.HandshakeTimeout)
	assert.Equal(t, 24*time.Hour, cfg.P2P.SessionTokenTTL)
	assert.True(t, cfg.P2P.ZKHandshake)
	assert.True(t, cfg.P2P.ZKAttestation)
	assert.Equal(t, "plonk", cfg.P2P.ZKP.ProvingScheme)
}

func TestDefaultConfig_Payment(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()

	assert.False(t, cfg.Payment.Enabled)
	assert.Equal(t, "local", cfg.Payment.WalletProvider)
	assert.Equal(t, int64(84532), cfg.Payment.Network.ChainID)
	assert.Equal(t, "1.00", cfg.Payment.Limits.MaxPerTx)
	assert.Equal(t, "10.00", cfg.Payment.Limits.MaxDaily)
	assert.Equal(t, "0.10", cfg.Payment.Limits.AutoApproveBelow)
}

func TestDefaultConfig_Automation(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()

	// Cron
	assert.False(t, cfg.Cron.Enabled)
	assert.Equal(t, "UTC", cfg.Cron.Timezone)
	assert.Equal(t, 5, cfg.Cron.MaxConcurrentJobs)

	// Background
	assert.False(t, cfg.Background.Enabled)
	assert.Equal(t, 30000, cfg.Background.YieldMs)
	assert.Equal(t, 3, cfg.Background.MaxConcurrentTasks)

	// Workflow
	assert.False(t, cfg.Workflow.Enabled)
	assert.Equal(t, 4, cfg.Workflow.MaxConcurrentSteps)
	assert.Equal(t, 10*time.Minute, cfg.Workflow.DefaultTimeout)
}

func TestDefaultConfig_Skill(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()

	assert.True(t, cfg.Skill.Enabled)
	assert.Equal(t, "~/.lango/skills", cfg.Skill.SkillsDir)
	assert.True(t, cfg.Skill.AllowImport)
	assert.Equal(t, 50, cfg.Skill.MaxBulkImport)
	assert.Equal(t, 5, cfg.Skill.ImportConcurrency)
	assert.Equal(t, 2*time.Minute, cfg.Skill.ImportTimeout)
}

func TestDefaultConfig_Alerting(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()

	assert.False(t, cfg.Alerting.Enabled)
	assert.Equal(t, 10, cfg.Alerting.PolicyBlockRate)
	assert.Equal(t, 5, cfg.Alerting.RecoveryRetries)
}

func TestDefaultConfig_Replay(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()

	assert.Empty(t, cfg.Replay.AllowedActors)
	assert.Empty(t, cfg.Replay.ReleaseAllowedActors)
	assert.Empty(t, cfg.Replay.RefundAllowedActors)
}

func TestValidate_ValidLogLevels(t *testing.T) {
	t.Parallel()

	validLevels := []string{"debug", "info", "warn", "error"}
	for _, level := range validLevels {
		t.Run(level, func(t *testing.T) {
			t.Parallel()
			cfg := DefaultConfig()
			cfg.Logging.Level = level
			assert.NoError(t, Validate(cfg))
		})
	}
}

func TestValidate_ValidLogFormats(t *testing.T) {
	t.Parallel()

	validFormats := []string{"json", "console"}
	for _, format := range validFormats {
		t.Run(format, func(t *testing.T) {
			t.Parallel()
			cfg := DefaultConfig()
			cfg.Logging.Format = format
			assert.NoError(t, Validate(cfg))
		})
	}
}

func TestValidate_InvalidLogFormat(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	cfg.Logging.Format = "xml"
	err := Validate(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid log format")
}

func TestValidate_PortBoundaries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give    int
		wantErr bool
	}{
		{give: 0, wantErr: true},
		{give: -1, wantErr: true},
		{give: 1, wantErr: false},
		{give: 65535, wantErr: false},
		{give: 65536, wantErr: true},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			t.Parallel()
			cfg := DefaultConfig()
			cfg.Server.Port = tt.give
			err := Validate(cfg)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidate_SecuritySignerProviders(t *testing.T) {
	t.Parallel()

	validProviders := []string{"local", "rpc", "enclave", "aws-kms", "gcp-kms", "azure-kv", "pkcs11"}
	for _, p := range validProviders {
		t.Run(p, func(t *testing.T) {
			t.Parallel()
			cfg := DefaultConfig()
			cfg.Security.Signer.Provider = p
			// Fill required fields for specific providers
			switch p {
			case "rpc":
				cfg.Security.Signer.RPCUrl = "http://localhost:8080"
			case "aws-kms", "gcp-kms":
				cfg.Security.KMS.KeyID = "key-123"
			case "azure-kv":
				cfg.Security.KMS.Azure.VaultURL = "https://vault.azure.net"
				cfg.Security.KMS.KeyID = "key-123"
			case "pkcs11":
				cfg.Security.KMS.PKCS11.ModulePath = "/usr/lib/pkcs11.so"
			}
			assert.NoError(t, Validate(cfg))
		})
	}
}

func TestValidate_InvalidSignerProvider(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	cfg.Security.Signer.Provider = "bogus"
	err := Validate(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid security.signer.provider")
}

func TestValidate_GraphBackend(t *testing.T) {
	t.Parallel()

	t.Run("bolt is valid", func(t *testing.T) {
		t.Parallel()
		cfg := DefaultConfig()
		cfg.Graph.Enabled = true
		cfg.Graph.Backend = "bolt"
		assert.NoError(t, Validate(cfg))
	})

	t.Run("unknown backend is invalid", func(t *testing.T) {
		t.Parallel()
		cfg := DefaultConfig()
		cfg.Graph.Enabled = true
		cfg.Graph.Backend = "neo4j"
		err := Validate(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "graph.backend")
	})
}

func TestValidate_MCPServerTransports(t *testing.T) {
	t.Parallel()

	t.Run("stdio requires command", func(t *testing.T) {
		t.Parallel()
		cfg := DefaultConfig()
		cfg.MCP.Enabled = true
		cfg.MCP.Servers = map[string]MCPServerConfig{
			"test": {Transport: "stdio"},
		}
		err := Validate(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "command is required")
	})

	t.Run("http requires url", func(t *testing.T) {
		t.Parallel()
		cfg := DefaultConfig()
		cfg.MCP.Enabled = true
		cfg.MCP.Servers = map[string]MCPServerConfig{
			"test": {Transport: "http"},
		}
		err := Validate(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "url is required")
	})

	t.Run("unsupported transport", func(t *testing.T) {
		t.Parallel()
		cfg := DefaultConfig()
		cfg.MCP.Enabled = true
		cfg.MCP.Servers = map[string]MCPServerConfig{
			"test": {Transport: "grpc"},
		}
		err := Validate(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not supported")
	})
}

func TestApprovalPolicy_Valid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give ApprovalPolicy
		want bool
	}{
		{give: ApprovalPolicyDangerous, want: true},
		{give: ApprovalPolicyAll, want: true},
		{give: ApprovalPolicyConfigured, want: true},
		{give: ApprovalPolicyNone, want: true},
		{give: ApprovalPolicy("unknown"), want: false},
		{give: ApprovalPolicy(""), want: false},
	}

	for _, tt := range tests {
		t.Run(string(tt.give), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.give.Valid())
		})
	}
}

func TestApprovalPolicy_Values(t *testing.T) {
	t.Parallel()

	vals := ApprovalPolicyDangerous.Values()
	assert.Len(t, vals, 4)
	assert.Contains(t, vals, ApprovalPolicyDangerous)
	assert.Contains(t, vals, ApprovalPolicyAll)
	assert.Contains(t, vals, ApprovalPolicyConfigured)
	assert.Contains(t, vals, ApprovalPolicyNone)
}

func TestValidate_SignerRPC_RequiresURL(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	cfg.Security.Signer.Provider = "rpc"
	cfg.Security.Signer.RPCUrl = ""
	err := Validate(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rpcUrl is required")
}

func TestValidate_SignerRPC_WithURL(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	cfg.Security.Signer.Provider = "rpc"
	cfg.Security.Signer.RPCUrl = "http://localhost:8080"
	assert.NoError(t, Validate(cfg))
}

func TestValidate_SignerAWSKMS_RequiresKeyID(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	cfg.Security.Signer.Provider = "aws-kms"
	cfg.Security.KMS.KeyID = ""
	err := Validate(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "keyId is required")
}

func TestValidate_SignerAzureKV_RequiresVaultURLAndKeyID(t *testing.T) {
	t.Parallel()

	t.Run("missing both", func(t *testing.T) {
		t.Parallel()
		cfg := DefaultConfig()
		cfg.Security.Signer.Provider = "azure-kv"
		err := Validate(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "vaultUrl is required")
		assert.Contains(t, err.Error(), "keyId is required")
	})

	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		cfg := DefaultConfig()
		cfg.Security.Signer.Provider = "azure-kv"
		cfg.Security.KMS.Azure.VaultURL = "https://vault.azure.net"
		cfg.Security.KMS.KeyID = "key-123"
		assert.NoError(t, Validate(cfg))
	})
}

func TestValidate_SignerPKCS11_RequiresModulePath(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	cfg.Security.Signer.Provider = "pkcs11"
	cfg.Security.KMS.PKCS11.ModulePath = ""
	err := Validate(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "modulePath is required")
}

func TestValidate_A2A_RequiresFields(t *testing.T) {
	t.Parallel()

	t.Run("missing baseUrl and agentName", func(t *testing.T) {
		t.Parallel()
		cfg := DefaultConfig()
		cfg.A2A.Enabled = true
		err := Validate(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "a2a.baseUrl")
		assert.Contains(t, err.Error(), "a2a.agentName")
	})

	t.Run("valid with both fields", func(t *testing.T) {
		t.Parallel()
		cfg := DefaultConfig()
		cfg.A2A.Enabled = true
		cfg.A2A.BaseURL = "http://localhost:8080"
		cfg.A2A.AgentName = "test"
		assert.NoError(t, Validate(cfg))
	})
}

func TestValidate_Payment_WalletProviders(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give    string
		wantErr bool
	}{
		{give: "local", wantErr: false},
		{give: "rpc", wantErr: false},
		{give: "composite", wantErr: false},
		{give: "ledger", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			cfg := DefaultConfig()
			cfg.Payment.Enabled = true
			cfg.Payment.Network.RPCURL = "https://rpc.example.com"
			cfg.Payment.WalletProvider = tt.give
			err := Validate(cfg)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "walletProvider")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidate_P2P_RequiresPayment(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	cfg.P2P.Enabled = true
	cfg.Payment.Enabled = false
	err := Validate(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "p2p requires payment.enabled")
}

func TestValidate_P2P_ZKPProvingScheme(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give    string
		wantErr bool
	}{
		{give: "plonk", wantErr: false},
		{give: "groth16", wantErr: false},
		{give: "", wantErr: false},
		{give: "marlin", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			cfg := DefaultConfig()
			cfg.P2P.Enabled = true
			cfg.Payment.Enabled = true
			cfg.Payment.Network.RPCURL = "https://rpc.example.com"
			cfg.P2P.ZKP.ProvingScheme = tt.give
			err := Validate(cfg)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "provingScheme")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidate_ContainerRuntime(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give    string
		wantErr bool
	}{
		{give: "auto", wantErr: false},
		{give: "docker", wantErr: false},
		{give: "gvisor", wantErr: false},
		{give: "native", wantErr: false},
		{give: "podman", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			cfg := DefaultConfig()
			cfg.P2P.ToolIsolation.Container.Enabled = true
			cfg.P2P.ToolIsolation.Container.Runtime = tt.give
			err := Validate(cfg)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "runtime")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidate_MultipleErrors(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	cfg.Server.Port = 0
	cfg.Logging.Level = "invalid"
	cfg.Logging.Format = "invalid"

	err := Validate(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid port")
	assert.Contains(t, err.Error(), "invalid log level")
	assert.Contains(t, err.Error(), "invalid log format")
}

func TestMCPServerConfig_IsEnabled(t *testing.T) {
	t.Parallel()

	t.Run("nil defaults to true", func(t *testing.T) {
		t.Parallel()
		cfg := MCPServerConfig{}
		assert.True(t, cfg.IsEnabled())
	})

	t.Run("explicit true", func(t *testing.T) {
		t.Parallel()
		b := true
		cfg := MCPServerConfig{Enabled: &b}
		assert.True(t, cfg.IsEnabled())
	})

	t.Run("explicit false", func(t *testing.T) {
		t.Parallel()
		b := false
		cfg := MCPServerConfig{Enabled: &b}
		assert.False(t, cfg.IsEnabled())
	})
}
