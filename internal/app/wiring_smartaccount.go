package app

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/contract"
	"github.com/langoai/lango/internal/economy/budget"
	"github.com/langoai/lango/internal/economy/escrow/sentinel"
	"github.com/langoai/lango/internal/economy/risk"
	"github.com/langoai/lango/internal/eventbus"
	sa "github.com/langoai/lango/internal/smartaccount"
	"github.com/langoai/lango/internal/smartaccount/bindings"
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

// SessionManager returns the session key manager.
func (sac *smartAccountComponents) SessionManager() *sasession.Manager {
	return sac.sessionManager
}

// PolicyEngine returns the policy engine.
func (sac *smartAccountComponents) PolicyEngine() *policy.Engine {
	return sac.policyEngine
}

// OnChainTracker returns the on-chain spending tracker.
func (sac *smartAccountComponents) OnChainTracker() *budget.OnChainTracker {
	return sac.onChainTracker
}

// PaymasterProvider returns the paymaster provider, or nil if not configured.
func (sac *smartAccountComponents) PaymasterProvider() paymaster.PaymasterProvider {
	return sac.paymasterProvider
}

// ModuleRegistry returns the module registry.
func (sac *smartAccountComponents) ModuleRegistry() *module.Registry {
	return sac.moduleRegistry
}

// BundlerClient returns the bundler client.
func (sac *smartAccountComponents) BundlerClient() *bundler.Client {
	return sac.bundlerClient
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

	// Wire on-chain registration/revocation if SessionValidator is configured.
	if cfg.SmartAccount.Modules.SessionValidatorAddress != "" {
		svABICache := contract.NewABICache()
		svCaller := contract.NewCaller(pc.rpcClient, pc.wallet, pc.chainID, svABICache)
		svAddr := common.HexToAddress(cfg.SmartAccount.Modules.SessionValidatorAddress)
		svClient := bindings.NewSessionValidatorClient(svCaller, svAddr, pc.chainID)

		sessionOpts = append(sessionOpts,
			sasession.WithOnChainRegistration(func(ctx context.Context, addr common.Address, p sa.SessionPolicy) (string, error) {
				return svClient.RegisterSessionKey(ctx, addr, toOnChainPolicy(p))
			}),
			sasession.WithOnChainRevocation(func(ctx context.Context, addr common.Address) (string, error) {
				return svClient.RevokeSessionKey(ctx, addr)
			}),
		)
		logger().Info("smart account: session on-chain wiring configured", "validator", svAddr.Hex())
	}

	sac.sessionManager = sasession.NewManager(sessionStore, sessionOpts...)

	// 4. Policy engine
	sac.policyEngine = policy.New()

	// 5. Account manager + factory
	abiCache := contract.NewABICache()
	caller := contract.NewCaller(pc.rpcClient, pc.wallet, pc.chainID, abiCache)
	factory := sa.NewFactory(
		caller,
		pc.rpcClient,
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
		be := econc.budgetEngine
		sac.onChainTracker.SetCallback(func(sessionID string, spent *big.Int) {
			_ = be.Record(sessionID, budget.SpendEntry{
				Amount:    new(big.Int).Set(spent),
				Reason:    "on-chain spend sync",
				Timestamp: time.Now(),
			})
		})
		logger().Info("smart account: budget sync wired")
	}

	logger().Info("smart account subsystem initialized")
	return sac
}

// initPaymasterProvider creates a paymaster provider based on config.
// The provider is wrapped with RecoverableProvider for transient error retry
// and fallback behavior.
func initPaymasterProvider(cfg config.SmartAccountPaymasterConfig) paymaster.PaymasterProvider {
	if cfg.RPCURL == "" {
		logger().Warn("paymaster enabled but no rpcURL configured")
		return nil
	}
	var inner paymaster.PaymasterProvider
	switch cfg.Provider {
	case "circle":
		inner = paymaster.NewCircleProvider(cfg.RPCURL)
	case "pimlico":
		inner = paymaster.NewPimlicoProvider(cfg.RPCURL, cfg.PolicyID)
	case "alchemy":
		inner = paymaster.NewAlchemyProvider(cfg.RPCURL, cfg.PolicyID)
	default:
		logger().Warn("unknown paymaster provider", "provider", cfg.Provider)
		return nil
	}

	// Wrap with recovery (retry + fallback).
	rcfg := paymaster.DefaultRecoveryConfig()
	if cfg.FallbackMode == "direct" {
		rcfg.FallbackMode = paymaster.FallbackDirectGas
	}
	return paymaster.NewRecoverableProvider(inner, rcfg)
}

// toOnChainPolicy converts a Go SessionPolicy to the on-chain tuple format
// expected by LangoSessionValidator. Time values are converted to uint48
// timestamps, function selectors from hex strings to [4]byte arrays.
func toOnChainPolicy(p sa.SessionPolicy) interface{} {
	// Convert function selectors from hex strings to [4]byte.
	var funcSelectors [][4]byte
	for _, hexSel := range p.AllowedFunctions {
		sel := common.FromHex(hexSel)
		if len(sel) >= 4 {
			var s [4]byte
			copy(s[:], sel[:4])
			funcSelectors = append(funcSelectors, s)
		}
	}

	spendLimit := p.SpendLimit
	if spendLimit == nil {
		spendLimit = new(big.Int)
	}
	spentAmount := p.SpentAmount
	if spentAmount == nil {
		spentAmount = new(big.Int)
	}

	// Return as an anonymous struct matching the Solidity tuple.
	type onChainPolicy struct {
		AllowedTargets    []common.Address
		AllowedFunctions  [][4]byte
		SpendLimit        *big.Int
		SpentAmount       *big.Int
		ValidAfter        *big.Int // uint48
		ValidUntil        *big.Int // uint48
		Active            bool
		AllowedPaymasters []common.Address
	}

	return onChainPolicy{
		AllowedTargets:    p.AllowedTargets,
		AllowedFunctions:  funcSelectors,
		SpendLimit:        spendLimit,
		SpentAmount:       spentAmount,
		ValidAfter:        big.NewInt(p.ValidAfter.Unix()),
		ValidUntil:        big.NewInt(p.ValidUntil.Unix()),
		Active:            true,
		AllowedPaymasters: p.AllowedPaymasters,
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
