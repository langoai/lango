package app

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/contract"
	"github.com/langoai/lango/internal/economy/budget"
	"github.com/langoai/lango/internal/economy/escrow"
	"github.com/langoai/lango/internal/economy/escrow/hub"
	"github.com/langoai/lango/internal/economy/escrow/sentinel"
	"github.com/langoai/lango/internal/economy/negotiation"
	"github.com/langoai/lango/internal/economy/pricing"
	"github.com/langoai/lango/internal/economy/risk"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/lifecycle"
	p2pproto "github.com/langoai/lango/internal/p2p/protocol"
	"github.com/langoai/lango/internal/payment"
)

// economyComponents holds optional economy layer components.
type economyComponents struct {
	budgetEngine      *budget.Engine
	riskEngine        *risk.Engine
	pricingEngine     *pricing.Engine
	negotiationEngine *negotiation.Engine
	escrowEngine      *escrow.Engine
	escrowSettler     escrow.SettlementExecutor
	sentinelEngine    *sentinel.Engine
	eventMonitor      *hub.EventMonitor
	danglingDetector  *hub.DanglingDetector
}

// initEconomy creates the economy layer components if enabled.
func initEconomy(cfg *config.Config, p2pc *p2pComponents, pc *paymentComponents, bus *eventbus.Bus) *economyComponents {
	if !cfg.Economy.Enabled {
		logger().Info("economy layer disabled")
		return nil
	}

	ec := &economyComponents{}

	// 1. Budget Engine — collect options first, create engine after risk engine.
	budgetStore := budget.NewStore()
	var budgetOpts []budget.Option
	if bus != nil {
		budgetOpts = append(budgetOpts, budget.WithAlertCallback(func(taskID string, pct float64) {
			bus.Publish(eventbus.BudgetAlertEvent{TaskID: taskID, Threshold: pct})
		}))
	}

	// 2. Risk Engine — wire reputation querier from P2P if available.
	var reputationFn risk.ReputationQuerier
	if p2pc != nil && p2pc.reputation != nil {
		rep := p2pc.reputation
		reputationFn = func(ctx context.Context, peerDID string) (float64, error) {
			return rep.GetScore(ctx, peerDID)
		}
	} else {
		reputationFn = func(_ context.Context, _ string) (float64, error) {
			return 0.5, nil // neutral default
		}
	}
	riskEngine, err := risk.New(cfg.Economy.Risk, reputationFn)
	if err != nil {
		logger().Warnw("risk engine init", "error", err)
	} else {
		ec.riskEngine = riskEngine
		logger().Info("economy: risk engine initialized")
	}

	// Wire risk assessor into budget options before creating engine.
	if ec.riskEngine != nil {
		riskEng := ec.riskEngine
		budgetOpts = append(budgetOpts, budget.WithRiskAssessor(
			func(ctx context.Context, peerDID string, amount *big.Int) error {
				assessment, err := riskEng.Assess(ctx, peerDID, amount, risk.VerifiabilityMedium)
				if err != nil {
					return err
				}
				if assessment.RiskLevel == risk.RiskCritical {
					return budget.ErrBudgetExceeded
				}
				return nil
			},
		))
	}

	// Create budget engine with all collected options.
	budgetEngine, err := budget.NewEngine(budgetStore, cfg.Economy.Budget, budgetOpts...)
	if err != nil {
		logger().Warnw("budget engine init", "error", err)
	} else {
		ec.budgetEngine = budgetEngine
		logger().Info("economy: budget engine initialized")
	}

	// 3. Pricing Engine
	if cfg.Economy.Pricing.Enabled {
		pricingEngine, err := pricing.New(cfg.Economy.Pricing)
		if err != nil {
			logger().Warnw("pricing engine init", "error", err)
		} else {
			// Wire reputation into pricing for trust discounts.
			// pricing.ReputationQuerier has the same signature as risk.ReputationQuerier
			// but is a separate type; wrap to satisfy the pricing package's type.
			pricingEngine.SetReputation(func(ctx context.Context, peerDID string) (float64, error) {
				return reputationFn(ctx, peerDID)
			})
			ec.pricingEngine = pricingEngine

			// If P2P is active, adapt pricing engine into paygate PricingFunc.
			if p2pc != nil && p2pc.payGate != nil {
				p2pc.pricingFn = pricingEngine.AdaptToPricingFunc()
				logger().Info("economy: pricing engine wired to paygate")
			}
			logger().Info("economy: pricing engine initialized")
		}
	}

	// 4. Negotiation Engine
	if cfg.Economy.Negotiate.Enabled {
		negEngine := negotiation.New(cfg.Economy.Negotiate)
		ec.negotiationEngine = negEngine

		// Wire pricing into negotiation for auto-respond.
		if ec.pricingEngine != nil {
			pe := ec.pricingEngine
			negEngine.SetPricing(func(toolName string, peerDID string) (*big.Int, error) {
				quote, err := pe.Quote(context.Background(), toolName, peerDID)
				if err != nil {
					return nil, err
				}
				return quote.FinalPrice, nil
			})
		}

		// Wire negotiation events to event bus.
		if bus != nil {
			negEngine.SetEventCallback(func(sessionID string, phase negotiation.Phase) {
				switch phase {
				case negotiation.PhaseProposed:
					sess, err := negEngine.Get(sessionID)
					if err == nil {
						bus.Publish(eventbus.NegotiationStartedEvent{
							SessionID:    sessionID,
							InitiatorDID: sess.InitiatorDID,
							ResponderDID: sess.ResponderDID,
							ToolName:     sess.CurrentTerms.ToolName,
						})
					}
				case negotiation.PhaseAccepted:
					sess, err := negEngine.Get(sessionID)
					if err == nil {
						bus.Publish(eventbus.NegotiationCompletedEvent{
							SessionID:    sessionID,
							InitiatorDID: sess.InitiatorDID,
							ResponderDID: sess.ResponderDID,
							AgreedPrice:  sess.CurrentTerms.Price,
						})
					}
				case negotiation.PhaseRejected:
					bus.Publish(eventbus.NegotiationFailedEvent{SessionID: sessionID, Reason: "rejected"})
				case negotiation.PhaseExpired:
					bus.Publish(eventbus.NegotiationFailedEvent{SessionID: sessionID, Reason: "expired"})
				case negotiation.PhaseCancelled:
					bus.Publish(eventbus.NegotiationFailedEvent{SessionID: sessionID, Reason: "cancelled"})
				}
			})
		}

		// Wire negotiation handler into P2P protocol.
		if p2pc != nil && p2pc.handler != nil {
			ne := negEngine
			localDID := ""
			if p2pc.identity != nil {
				if did, err := p2pc.identity.DID(context.Background()); err == nil {
					localDID = did.ID
				}
			}
			p2pc.handler.SetNegotiator(func(ctx context.Context, peerDID string, payload p2pproto.NegotiatePayload) (map[string]interface{}, error) {
				return handleNegotiateProtocol(ctx, ne, localDID, peerDID, payload)
			})
		}

		logger().Info("economy: negotiation engine initialized")
	}

	// 5. Escrow Engine
	if cfg.Economy.Escrow.Enabled {
		escrowStore := escrow.NewMemoryStore()
		escrowCfg := escrow.EngineConfig{
			DefaultTimeout: cfg.Economy.Escrow.DefaultTimeout,
			MaxMilestones:  cfg.Economy.Escrow.MaxMilestones,
			AutoRelease:    cfg.Economy.Escrow.AutoRelease,
			DisputeWindow:  cfg.Economy.Escrow.DisputeWindow,
		}
		if escrowCfg.DefaultTimeout == 0 {
			escrowCfg.DefaultTimeout = escrow.DefaultEngineConfig().DefaultTimeout
		}
		if escrowCfg.MaxMilestones == 0 {
			escrowCfg.MaxMilestones = escrow.DefaultEngineConfig().MaxMilestones
		}
		if escrowCfg.DisputeWindow == 0 {
			escrowCfg.DisputeWindow = escrow.DefaultEngineConfig().DisputeWindow
		}

		// Select settlement mode based on config.
		settler := selectSettler(cfg, pc)
		ec.escrowSettler = settler

		escrowEngine := escrow.NewEngine(escrowStore, settler, escrowCfg)
		ec.escrowEngine = escrowEngine
		logger().Info("economy: escrow engine initialized")

		// 5a. Security Sentinel Engine
		sentinelCfg := sentinel.DefaultSentinelConfig()
		sentinelEngine := sentinel.New(bus, sentinelCfg)
		if err := sentinelEngine.Start(); err != nil {
			logger().Warnw("sentinel engine start", "error", err)
		} else {
			ec.sentinelEngine = sentinelEngine
			logger().Info("economy: sentinel engine initialized")
		}

		// 5b. On-chain event reconciliation (requires on-chain mode + RPC).
		oc := cfg.Economy.Escrow.OnChain
		if oc.Enabled && pc != nil && pc.rpcClient != nil {
			hubAddr := common.HexToAddress(oc.HubAddress)

			// EventMonitor: watches on-chain contract events.
			monitorOpts := []hub.MonitorOption{
				hub.WithMonitorLogger(logger()),
			}
			if oc.PollInterval > 0 {
				monitorOpts = append(monitorOpts, hub.WithPollInterval(oc.PollInterval))
			}
			confirmDepth := uint64(2) // Base L2 default
			if oc.ConfirmationDepth > 0 {
				confirmDepth = oc.ConfirmationDepth
			}
			monitorOpts = append(monitorOpts, hub.WithConfirmationDepth(confirmDepth))
			monitor, err := hub.NewEventMonitor(pc.rpcClient, bus, nil, hubAddr, monitorOpts...)
			if err != nil {
				logger().Warnw("event monitor init", "error", err)
			} else {
				ec.eventMonitor = monitor
				logger().Info("economy: event monitor initialized")
			}

			// DanglingDetector: expires stuck pending escrows.
			dd := hub.NewDanglingDetector(escrowStore, escrowEngine, bus,
				hub.WithDanglingLogger(logger()),
			)
			ec.danglingDetector = dd
			logger().Info("economy: dangling detector initialized")

			// Wire on-chain events to escrow engine state transitions.
			initOnChainEscrowBridge(bus, escrowEngine, logger())
		}
	}

	return ec
}

// selectSettler chooses the settlement executor based on config.
// Returns: USDCSettler (custodian), HubSettler, VaultSettler, or noopSettler.
func selectSettler(cfg *config.Config, pc *paymentComponents) escrow.SettlementExecutor {
	oc := cfg.Economy.Escrow.OnChain

	// On-chain mode requires payment components.
	if oc.Enabled && pc != nil {
		abiCache := contract.NewABICache()
		caller := contract.NewCaller(pc.rpcClient, pc.wallet, pc.chainID, abiCache)

		switch oc.Mode {
		case "hub":
			if oc.HubAddress != "" {
				hubAddr := common.HexToAddress(oc.HubAddress)
				tokenAddr := common.HexToAddress(oc.TokenAddress)
				settler := hub.NewHubSettler(caller, hubAddr, tokenAddr, pc.chainID)
				logger().Infow("economy: escrow using Hub settler",
					"hub", oc.HubAddress, "token", oc.TokenAddress)
				return settler
			}
			logger().Warn("economy: hub mode enabled but hubAddress not set, falling back to custodian")

		case "vault":
			if oc.VaultFactoryAddress != "" && oc.VaultImplementation != "" {
				factoryAddr := common.HexToAddress(oc.VaultFactoryAddress)
				implAddr := common.HexToAddress(oc.VaultImplementation)
				tokenAddr := common.HexToAddress(oc.TokenAddress)
				arbitrator := common.HexToAddress(oc.ArbitratorAddress)
				settler := hub.NewVaultSettler(caller, factoryAddr, implAddr, tokenAddr, arbitrator, pc.chainID)
				logger().Infow("economy: escrow using Vault settler",
					"factory", oc.VaultFactoryAddress, "token", oc.TokenAddress)
				return settler
			}
			logger().Warn("economy: vault mode enabled but addresses not set, falling back to custodian")
		}
	}

	// Default: custodian mode (existing USDCSettler).
	if pc != nil {
		settler := escrow.NewUSDCSettler(
			pc.wallet,
			payment.NewTxBuilder(pc.rpcClient, pc.chainID, cfg.Payment.Network.USDCContract),
			pc.rpcClient,
			pc.chainID,
			escrow.WithReceiptTimeout(cfg.Economy.Escrow.Settlement.ReceiptTimeout),
			escrow.WithMaxRetries(cfg.Economy.Escrow.Settlement.MaxRetries),
		)
		logger().Info("economy: escrow using USDC settler (custodian)")
		return settler
	}

	return escrow.NoopSettler{}
}

// handleNegotiateProtocol routes P2P negotiation messages to the negotiation engine.
func handleNegotiateProtocol(ctx context.Context, ne *negotiation.Engine, localDID, peerDID string, payload p2pproto.NegotiatePayload) (map[string]interface{}, error) {
	switch payload.Action {
	case string(negotiation.ActionPropose):
		price, ok := new(big.Int).SetString(payload.Price, 10)
		if !ok {
			price = new(big.Int)
		}
		terms := negotiation.Terms{
			ToolName: payload.ToolName,
			Price:    price,
		}
		sess, err := ne.Propose(ctx, peerDID, localDID, terms)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"sessionId": sess.ID,
			"phase":     string(sess.Phase),
		}, nil

	case string(negotiation.ActionCounter):
		price, ok := new(big.Int).SetString(payload.Price, 10)
		if !ok {
			price = new(big.Int)
		}
		terms := negotiation.Terms{
			ToolName: payload.ToolName,
			Price:    price,
		}
		sess, err := ne.Counter(ctx, payload.SessionID, localDID, terms, payload.Reason)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"sessionId": sess.ID,
			"phase":     string(sess.Phase),
			"round":     sess.Round,
		}, nil

	case string(negotiation.ActionAccept):
		sess, err := ne.Accept(ctx, payload.SessionID, localDID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"sessionId": sess.ID,
			"phase":     string(sess.Phase),
		}, nil

	case string(negotiation.ActionReject):
		sess, err := ne.Reject(ctx, payload.SessionID, localDID, payload.Reason)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"sessionId": sess.ID,
			"phase":     string(sess.Phase),
		}, nil

	default:
		return nil, negotiation.ErrSessionNotFound
	}
}

// registerEconomyLifecycle registers economy lifecycle components with the registry.
func registerEconomyLifecycle(reg *lifecycle.Registry, ec *economyComponents) {
	if ec == nil {
		return
	}
	if ec.eventMonitor != nil {
		reg.Register(ec.eventMonitor, lifecycle.PriorityNetwork)
	}
	if ec.danglingDetector != nil {
		reg.Register(ec.danglingDetector, lifecycle.PriorityAutomation)
	}
}
