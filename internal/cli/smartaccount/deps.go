package smartaccount

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/contract"
	"github.com/langoai/lango/internal/security"
	sa "github.com/langoai/lango/internal/smartaccount"
	"github.com/langoai/lango/internal/smartaccount/bundler"
	"github.com/langoai/lango/internal/smartaccount/module"
	"github.com/langoai/lango/internal/smartaccount/paymaster"
	"github.com/langoai/lango/internal/smartaccount/policy"
	sasession "github.com/langoai/lango/internal/smartaccount/session"
	"github.com/langoai/lango/internal/wallet"
)

// smartAccountDeps holds lazily-initialized smart account dependencies for CLI.
type smartAccountDeps struct {
	manager        sa.AccountManager
	sessionManager *sasession.Manager
	policyEngine   *policy.Engine
	moduleRegistry *module.Registry
	bundlerClient  *bundler.Client
	paymasterProv  paymaster.PaymasterProvider
	cfg            config.SmartAccountConfig
	cleanup        func()
}

// initSmartAccountDeps creates smart account components from a bootstrap result.
// Unlike wiring_smartaccount.go which runs inside the full app, this builds
// only the components needed for CLI commands.
func initSmartAccountDeps(boot *bootstrap.Result) (*smartAccountDeps, error) {
	cfg := boot.Config
	if !cfg.SmartAccount.Enabled {
		return nil, fmt.Errorf("smart account not enabled (set smartAccount.enabled = true)")
	}

	if !cfg.Payment.Enabled {
		return nil, fmt.Errorf("smart account requires payment to be enabled (set payment.enabled = true)")
	}

	// Build secrets store for wallet key management.
	ctx := context.Background()
	registry := security.NewKeyRegistry(boot.DBClient)
	if _, err := registry.RegisterKey(ctx, "default", "local", security.KeyTypeEncryption); err != nil {
		return nil, fmt.Errorf("register default key: %w", err)
	}
	secrets := security.NewSecretsStore(boot.DBClient, registry, boot.Crypto)

	// Create RPC client for blockchain interaction.
	rpcClient, err := ethclient.Dial(cfg.Payment.Network.RPCURL)
	if err != nil {
		return nil, fmt.Errorf("connect to RPC %q: %w", cfg.Payment.Network.RPCURL, err)
	}

	// Create wallet provider.
	var wp wallet.WalletProvider
	switch cfg.Payment.WalletProvider {
	case "local":
		wp = wallet.NewLocalWallet(secrets, cfg.Payment.Network.RPCURL, cfg.Payment.Network.ChainID)
	case "rpc":
		wp = wallet.NewRPCWallet()
	case "composite":
		local := wallet.NewLocalWallet(secrets, cfg.Payment.Network.RPCURL, cfg.Payment.Network.ChainID)
		rpc := wallet.NewRPCWallet()
		wp = wallet.NewCompositeWallet(rpc, local, nil)
	default:
		wp = wallet.NewLocalWallet(secrets, cfg.Payment.Network.RPCURL, cfg.Payment.Network.ChainID)
	}

	chainID := cfg.Payment.Network.ChainID

	deps := &smartAccountDeps{
		cfg: cfg.SmartAccount,
		cleanup: func() {
			rpcClient.Close()
		},
	}

	// 1. Bundler client.
	entryPoint := common.HexToAddress(cfg.SmartAccount.EntryPointAddress)
	deps.bundlerClient = bundler.NewClient(cfg.SmartAccount.BundlerURL, entryPoint)

	// 2. Module registry with default modules.
	deps.moduleRegistry = module.NewRegistry()
	registerDefaultModules(deps.moduleRegistry, cfg.SmartAccount.Modules)

	// 3. Session store + manager.
	sessionStore := sasession.NewMemoryStore()
	var sessionOpts []sasession.ManagerOption
	if cfg.SmartAccount.Session.MaxDuration > 0 {
		sessionOpts = append(sessionOpts, sasession.WithMaxDuration(cfg.SmartAccount.Session.MaxDuration))
	}
	if cfg.SmartAccount.Session.MaxActiveKeys > 0 {
		sessionOpts = append(sessionOpts, sasession.WithMaxKeys(cfg.SmartAccount.Session.MaxActiveKeys))
	}
	// Provide entryPoint and chainID for correct UserOp hash computation.
	sessionOpts = append(sessionOpts,
		sasession.WithEntryPoint(entryPoint),
		sasession.WithChainID(chainID),
	)

	deps.sessionManager = sasession.NewManager(sessionStore, sessionOpts...)

	// 4. Policy engine.
	deps.policyEngine = policy.New()

	// 5. Account manager + factory.
	abiCache := contract.NewABICache()
	caller := contract.NewCaller(rpcClient, wp, chainID, abiCache)
	factory := sa.NewFactory(
		caller,
		rpcClient,
		common.HexToAddress(cfg.SmartAccount.FactoryAddress),
		common.HexToAddress(cfg.SmartAccount.Safe7579Address),
		common.HexToAddress(cfg.SmartAccount.FallbackHandler),
		chainID,
	)
	mgr := sa.NewManager(factory, deps.bundlerClient, caller, wp, chainID, entryPoint)
	deps.manager = mgr

	// 6. Paymaster provider (optional).
	if cfg.SmartAccount.Paymaster.Enabled {
		provider := initPaymasterProvider(cfg.SmartAccount.Paymaster, wp, rpcClient, chainID)
		if provider != nil {
			deps.paymasterProv = provider
			mgr.SetPaymasterFunc(func(ctx context.Context, op *sa.UserOperation, stub bool) ([]byte, *sa.PaymasterGasOverrides, error) {
				req := &paymaster.SponsorRequest{
					UserOp: &paymaster.UserOpData{
						Sender:               op.Sender,
						Nonce:                op.Nonce,
						InitCode:             op.InitCode,
						CallData:             op.CallData,
						CallGasLimit:         op.CallGasLimit,
						VerificationGasLimit: op.VerificationGasLimit,
						PreVerificationGas:   op.PreVerificationGas,
						MaxFeePerGas:         op.MaxFeePerGas,
						MaxPriorityFeePerGas: op.MaxPriorityFeePerGas,
						PaymasterAndData:     op.PaymasterAndData,
						Signature:            op.Signature,
					},
					EntryPoint: entryPoint,
					ChainID:    chainID,
					Stub:       stub,
				}
				result, sponsorErr := provider.SponsorUserOp(ctx, req)
				if sponsorErr != nil {
					return nil, nil, sponsorErr
				}
				var gasOverrides *sa.PaymasterGasOverrides
				if result.GasOverrides != nil {
					gasOverrides = &sa.PaymasterGasOverrides{
						CallGasLimit:         result.GasOverrides.CallGasLimit,
						VerificationGasLimit: result.GasOverrides.VerificationGasLimit,
						PreVerificationGas:   result.GasOverrides.PreVerificationGas,
					}
				}
				return result.PaymasterAndData, gasOverrides, nil
			})
		}
	}

	return deps, nil
}

// initPaymasterProvider creates a paymaster provider based on config.
// When mode="permit" and provider="circle", uses on-chain permit mode (no RPC URL needed).
func initPaymasterProvider(
	cfg config.SmartAccountPaymasterConfig,
	wp wallet.WalletProvider,
	rpcClient *ethclient.Client,
	chainID int64,
) paymaster.PaymasterProvider {
	// Permit mode: on-chain paymaster, no RPC URL required.
	if cfg.Mode == "permit" && cfg.Provider == "circle" {
		if wp == nil || rpcClient == nil {
			return nil
		}
		pmAddr := common.HexToAddress(cfg.PaymasterAddress)
		tokenAddr := common.HexToAddress(cfg.TokenAddress)
		return paymaster.NewCirclePermitProvider(pmAddr, tokenAddr, chainID, wp, rpcClient)
	}

	// RPC mode (default).
	if cfg.RPCURL == "" {
		return nil
	}
	switch cfg.Provider {
	case "circle":
		return paymaster.NewCircleProvider(cfg.RPCURL)
	case "pimlico":
		return paymaster.NewPimlicoProvider(cfg.RPCURL, cfg.PolicyID)
	case "alchemy":
		return paymaster.NewAlchemyProvider(cfg.RPCURL, cfg.PolicyID)
	default:
		return nil
	}
}

// registerDefaultModules registers well-known Lango module descriptors.
func registerDefaultModules(reg *module.Registry, cfg config.SmartAccountModulesConfig) {
	if cfg.SessionValidatorAddress != "" {
		_ = reg.Register(&module.ModuleDescriptor{
			Name:    "LangoSessionValidator",
			Address: common.HexToAddress(cfg.SessionValidatorAddress),
			Type:    sa.ModuleTypeValidator,
			Version: "1.0.0",
		})
	}
	if cfg.SpendingHookAddress != "" {
		_ = reg.Register(&module.ModuleDescriptor{
			Name:    "LangoSpendingHook",
			Address: common.HexToAddress(cfg.SpendingHookAddress),
			Type:    sa.ModuleTypeHook,
			Version: "1.0.0",
		})
	}
	if cfg.EscrowExecutorAddress != "" {
		_ = reg.Register(&module.ModuleDescriptor{
			Name:    "LangoEscrowExecutor",
			Address: common.HexToAddress(cfg.EscrowExecutorAddress),
			Type:    sa.ModuleTypeExecutor,
			Version: "1.0.0",
		})
	}
}
