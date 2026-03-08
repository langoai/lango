package app

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/contract"
	"github.com/langoai/lango/internal/economy/budget"
	"github.com/langoai/lango/internal/economy/escrow/sentinel"
	"github.com/langoai/lango/internal/economy/risk"
	"github.com/langoai/lango/internal/eventbus"
	sa "github.com/langoai/lango/internal/smartaccount"
	"github.com/langoai/lango/internal/smartaccount/bundler"
	"github.com/langoai/lango/internal/smartaccount/module"
	"github.com/langoai/lango/internal/smartaccount/paymaster"
	"github.com/langoai/lango/internal/smartaccount/policy"
	sasession "github.com/langoai/lango/internal/smartaccount/session"
)

// smartAccountComponents holds optional smart account subsystem components.
type smartAccountComponents struct {
	manager           sa.AccountManager
	sessionManager    *sasession.Manager
	policyEngine      *policy.Engine
	moduleRegistry    *module.Registry
	bundlerClient     *bundler.Client
	onChainTracker    *budget.OnChainTracker
	sessionGuard      *sentinel.SessionGuard
	paymasterProvider paymaster.PaymasterProvider
}

// initSmartAccount creates the smart account subsystem if enabled.
func initSmartAccount(
	cfg *config.Config,
	pc *paymentComponents,
	econc *economyComponents,
	bus *eventbus.Bus,
) *smartAccountComponents {
	if !cfg.SmartAccount.Enabled {
		logger().Info("smart account disabled")
		return nil
	}
	if pc == nil {
		logger().Warn("smart account requires payment components")
		return nil
	}

	sac := &smartAccountComponents{}

	// 1. Bundler client
	entryPoint := common.HexToAddress(cfg.SmartAccount.EntryPointAddress)
	sac.bundlerClient = bundler.NewClient(cfg.SmartAccount.BundlerURL, entryPoint)

	// 2. Module registry — pre-register Lango modules
	sac.moduleRegistry = module.NewRegistry()
	registerDefaultModules(sac.moduleRegistry, cfg.SmartAccount.Modules)

	// 3. Session store + manager
	sessionStore := sasession.NewMemoryStore()
	var sessionOpts []sasession.ManagerOption
	if cfg.SmartAccount.Session.MaxDuration > 0 {
		sessionOpts = append(sessionOpts, sasession.WithMaxDuration(cfg.SmartAccount.Session.MaxDuration))
	}
	if cfg.SmartAccount.Session.MaxActiveKeys > 0 {
		sessionOpts = append(sessionOpts, sasession.WithMaxKeys(cfg.SmartAccount.Session.MaxActiveKeys))
	}
	sac.sessionManager = sasession.NewManager(sessionStore, sessionOpts...)

	// 4. Policy engine
	sac.policyEngine = policy.New()

	// 5. Account manager + factory
	abiCache := contract.NewABICache()
	caller := contract.NewCaller(pc.rpcClient, pc.wallet, pc.chainID, abiCache)
	factory := sa.NewFactory(
		caller,
		common.HexToAddress(cfg.SmartAccount.FactoryAddress),
		common.HexToAddress(cfg.SmartAccount.Safe7579Address),
		common.HexToAddress(cfg.SmartAccount.FallbackHandler),
		pc.chainID,
	)
	mgr := sa.NewManager(factory, sac.bundlerClient, caller, pc.wallet, pc.chainID, entryPoint)
	sac.manager = mgr

	// 5a. Paymaster provider (optional)
	if cfg.SmartAccount.Paymaster.Enabled {
		provider := initPaymasterProvider(cfg.SmartAccount.Paymaster)
		if provider != nil {
			sac.paymasterProvider = provider
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
					ChainID:    pc.chainID,
					Stub:       stub,
				}
				result, err := provider.SponsorUserOp(ctx, req)
				if err != nil {
					return nil, nil, err
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
			logger().Info("smart account: paymaster wired", "provider", provider.Type())
		}
	}

	// 6. Wire risk engine → policy engine (callback, no direct import)
	if econc != nil && econc.riskEngine != nil {
		fullBudget := big.NewInt(100_000_000) // 100 USDC default (6 decimals)
		adapter := risk.NewPolicyAdapter(econc.riskEngine, fullBudget)
		sac.policyEngine.SetRiskPolicy(func(ctx context.Context, peerDID string) (*policy.HarnessPolicy, error) {
			rec, err := adapter.Recommend(ctx, peerDID, fullBudget)
			if err != nil {
				return nil, err
			}
			return &policy.HarnessPolicy{
				MaxTxAmount:      rec.MaxSpendLimit,
				AutoApproveBelow: rec.MaxSpendLimit,
			}, nil
		})
		logger().Info("smart account: risk engine wired to policy")
	}

	// 7. Wire sentinel → session guard
	if econc != nil && econc.sentinelEngine != nil && bus != nil {
		guard := sentinel.NewSessionGuard(bus)
		sm := sac.sessionManager
		guard.SetRevokeFunc(func() error {
			return sm.RevokeAll(context.Background())
		})
		guard.Start()
		sac.sessionGuard = guard
		logger().Info("smart account: sentinel session guard wired")
	}

	// 8. On-chain spending tracker
	sac.onChainTracker = budget.NewOnChainTracker()
	if econc != nil && econc.budgetEngine != nil {
		logger().Info("smart account: budget sync wired")
	}

	logger().Info("smart account subsystem initialized")
	return sac
}

// initPaymasterProvider creates a paymaster provider based on config.
func initPaymasterProvider(cfg config.SmartAccountPaymasterConfig) paymaster.PaymasterProvider {
	if cfg.RPCURL == "" {
		logger().Warn("paymaster enabled but no rpcURL configured")
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
		logger().Warn("unknown paymaster provider", "provider", cfg.Provider)
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
