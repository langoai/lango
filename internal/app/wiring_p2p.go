package app

import (
	"context"
	"fmt"
	"time"

	"github.com/consensys/gnark/frontend"
	"github.com/ethereum/go-ethereum/common"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/p2p"
	"github.com/langoai/lango/internal/p2p/agentpool"
	"github.com/langoai/lango/internal/p2p/discovery"
	"github.com/langoai/lango/internal/p2p/firewall"
	"github.com/langoai/lango/internal/p2p/handshake"
	"github.com/langoai/lango/internal/p2p/identity"
	"github.com/langoai/lango/internal/p2p/paygate"
	p2pproto "github.com/langoai/lango/internal/p2p/protocol"
	"github.com/langoai/lango/internal/p2p/reputation"
	"github.com/langoai/lango/internal/p2p/settlement"
	"github.com/langoai/lango/internal/p2p/team"
	"github.com/langoai/lango/internal/p2p/zkp"
	"github.com/langoai/lango/internal/p2p/zkp/circuits"
	"github.com/langoai/lango/internal/payment/contracts"
	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/wallet"
	libp2pproto "github.com/libp2p/go-libp2p/core/protocol"
)

// p2pComponents holds optional P2P networking components.
type p2pComponents struct {
	node        *p2p.Node
	sessions    *handshake.SessionStore
	handshaker  *handshake.Handshaker
	fw          *firewall.Firewall
	gossip      *discovery.GossipService
	identity    *identity.WalletDIDProvider
	handler     *p2pproto.Handler
	payGate     *paygate.Gate
	reputation  *reputation.Store
	pricingCfg  config.P2PPricingConfig
	pricingFn   func(toolName string) (string, bool)
	agentPool   *agentpool.Pool
	selector    *agentpool.Selector
	coordinator *team.Coordinator
	provider    *agentpool.PoolProvider
}

// initP2P creates the P2P networking components if enabled.
func initP2P(cfg *config.Config, wp wallet.WalletProvider, pc *paymentComponents, dbClient *ent.Client, secrets *security.SecretsStore, bus *eventbus.Bus) *p2pComponents {
	if !cfg.P2P.Enabled {
		logger().Info("P2P networking disabled")
		return nil
	}

	if wp == nil {
		logger().Warn("P2P networking requires wallet provider, skipping")
		return nil
	}

	pLogger := logger()

	// Create P2P node with SecretsStore for encrypted key storage.
	node, err := p2p.NewNode(cfg.P2P, pLogger, secrets)
	if err != nil {
		pLogger.Warnw("P2P node creation failed, skipping", "error", err)
		return nil
	}

	// Create identity provider from wallet.
	idProvider := identity.NewProvider(wp, pLogger)

	// Create session store.
	sessionTTL := cfg.P2P.SessionTokenTTL
	if sessionTTL <= 0 {
		sessionTTL = 24 * time.Hour
	}
	sessions, err := handshake.NewSessionStore(sessionTTL)
	if err != nil {
		pLogger.Warnw("P2P session store creation failed, skipping", "error", err)
		return nil
	}

	// Initialize ZKP prover (optional).
	zkProver := initZKP(cfg)

	// Create nonce cache for replay protection (TTL = 2 * handshake timeout).
	nonceTTL := 2 * cfg.P2P.HandshakeTimeout
	if nonceTTL <= 0 {
		nonceTTL = 60 * time.Second
	}
	nonceCache := handshake.NewNonceCache(nonceTTL)
	nonceCache.Start()

	// Create handshaker.
	hsTimeout := cfg.P2P.HandshakeTimeout
	if hsTimeout <= 0 {
		hsTimeout = 30 * time.Second
	}
	hsCfg := handshake.Config{
		Wallet:                 wp,
		Sessions:               sessions,
		ZKEnabled:              cfg.P2P.ZKHandshake,
		Timeout:                hsTimeout,
		AutoApproveKnown:       cfg.P2P.AutoApproveKnownPeers,
		NonceCache:             nonceCache,
		RequireSignedChallenge: cfg.P2P.RequireSignedChallenge,
		Logger:                 pLogger,
	}

	// Wire ZK prover/verifier into handshake if available.
	if zkProver != nil && cfg.P2P.ZKHandshake {
		hsCfg.ZKProver = func(ctx context.Context, challenge []byte) ([]byte, error) {
			assignment := &circuits.WalletOwnershipCircuit{
				Challenge: challenge,
				Response:  challenge, // simplified: use challenge as witness in MVP
			}
			proof, err := zkProver.Prove(ctx, "wallet_ownership", assignment)
			if err != nil {
				return nil, err
			}
			return proof.Data, nil
		}
		hsCfg.ZKVerifier = func(ctx context.Context, proof, challenge, publicKey []byte) (bool, error) {
			p := &zkp.Proof{
				CircuitID: "wallet_ownership",
				Data:      proof,
				Scheme:    zkProver.Scheme(),
			}
			return zkProver.Verify(ctx, p, &circuits.WalletOwnershipCircuit{})
		}
		pLogger.Info("ZK handshake prover/verifier wired")
	}

	handshaker := handshake.NewHandshaker(hsCfg)

	// Create firewall.
	var aclRules []firewall.ACLRule
	for _, r := range cfg.P2P.FirewallRules {
		aclRules = append(aclRules, firewall.ACLRule{
			PeerDID:   r.PeerDID,
			Action:    firewall.ACLAction(r.Action),
			Tools:     r.Tools,
			RateLimit: r.RateLimit,
		})
	}
	fw := firewall.New(aclRules, pLogger)

	// Wire Owner Shield if configured.
	ownerCfg := cfg.P2P.OwnerProtection
	if ownerCfg.OwnerName != "" || ownerCfg.OwnerEmail != "" || ownerCfg.OwnerPhone != "" {
		blockConv := true
		if ownerCfg.BlockConversations != nil {
			blockConv = *ownerCfg.BlockConversations
		}
		shield := firewall.NewOwnerShield(firewall.OwnerProtectionConfig{
			OwnerName:          ownerCfg.OwnerName,
			OwnerEmail:         ownerCfg.OwnerEmail,
			OwnerPhone:         ownerCfg.OwnerPhone,
			ExtraTerms:         ownerCfg.ExtraTerms,
			BlockConversations: blockConv,
		}, pLogger)
		fw.SetOwnerShield(shield)
		pLogger.Info("P2P owner data shield enabled")
	}

	// Wire ZK attestation into firewall if available.
	if zkProver != nil && cfg.P2P.ZKAttestation {
		fw.SetZKAttestFunc(func(responseHash, agentDIDHash []byte) (*firewall.AttestationResult, error) {
			now := time.Now().Unix()
			assignment := &circuits.ResponseAttestationCircuit{
				ResponseHash: responseHash,
				AgentDIDHash: agentDIDHash,
				Timestamp:    now,
				MinTimestamp:  now - 300, // 5-minute window
				MaxTimestamp:  now + 30,  // 30-second future grace
			}
			proof, err := zkProver.Prove(context.Background(), "response_attestation", assignment)
			if err != nil {
				return nil, err
			}
			return &firewall.AttestationResult{
				Proof:        proof.Data,
				PublicInputs: proof.PublicInputs,
				CircuitID:    proof.CircuitID,
				Scheme:       string(proof.Scheme),
			}, nil
		})
		pLogger.Info("ZK response attestation wired to firewall")
	}

	// Wire reputation system if DB client is available.
	var repStore *reputation.Store
	if dbClient != nil {
		repStore = reputation.NewStore(dbClient, pLogger)
		minScore := cfg.P2P.MinTrustScore
		if minScore <= 0 {
			minScore = 0.3
		}
		fw.SetReputationChecker(func(ctx context.Context, peerDID string) (float64, error) {
			return repStore.GetScore(ctx, peerDID)
		}, minScore)
		pLogger.Infow("P2P reputation system enabled", "minTrustScore", minScore)
	}

	// Register handshake protocol handlers (v1.0 legacy + v1.1 signed challenge).
	node.Host().SetStreamHandler(libp2pproto.ID(handshake.ProtocolID), handshaker.StreamHandler())
	node.Host().SetStreamHandler(libp2pproto.ID(handshake.ProtocolIDv11), handshaker.StreamHandlerV11())

	// Get local DID for protocol handler.
	var localDID string
	ctx := context.Background()
	d, err := idProvider.DID(ctx)
	if err == nil && d != nil {
		localDID = d.ID
	}

	// Create A2A-over-P2P protocol handler.
	handler := p2pproto.NewHandler(p2pproto.HandlerConfig{
		Sessions: sessions,
		Firewall: fw,
		LocalDID: localDID,
		Logger:   pLogger,
	})
	node.Host().SetStreamHandler(libp2pproto.ID(p2pproto.ProtocolID), handler.StreamHandler())

	// Wire security event handler for auto-invalidation on repeated failures
	// or reputation drops.
	minTrust := cfg.P2P.MinTrustScore
	if minTrust <= 0 {
		minTrust = 0.3
	}
	secEvents := handshake.NewSecurityEventHandler(sessions, 5, minTrust, pLogger)
	handler.SetSecurityEvents(secEvents)
	if repStore != nil {
		repStore.SetOnChangeCallback(secEvents.OnReputationChange)
	}
	pLogger.Info("P2P security event handler wired")

	// Create gossip discovery service.
	var gossip *discovery.GossipService
	gossipInterval := cfg.P2P.GossipInterval
	if gossipInterval <= 0 {
		gossipInterval = 30 * time.Second
	}

	agentName := cfg.A2A.AgentName
	if agentName == "" {
		agentName = "lango"
	}
	// Wire payment gate if pricing is enabled.
	var pg *paygate.Gate
	if cfg.P2P.Pricing.Enabled && pc != nil {
		walletAddr := ""
		ctx2 := context.Background()
		if a, err := wp.Address(ctx2); err == nil {
			walletAddr = a
		}
		usdcAddr, _ := contracts.LookupUSDC(pc.chainID)

		pricingFn := func(toolName string) (string, bool) {
			if price, ok := cfg.P2P.Pricing.ToolPrices[toolName]; ok {
				return price, false
			}
			if cfg.P2P.Pricing.PerQuery != "" {
				return cfg.P2P.Pricing.PerQuery, false
			}
			return "", true // free by default
		}

		// Build trust config from P2P pricing thresholds.
		trustCfg := paygate.DefaultTrustConfig()
		if cfg.P2P.Pricing.TrustThresholds.PostPayMinScore > 0 {
			trustCfg.PostPayMinScore = cfg.P2P.Pricing.TrustThresholds.PostPayMinScore
		}

		// Wire reputation function for trust-based payment tiers.
		var reputationFn paygate.ReputationFunc
		if repStore != nil {
			reputationFn = func(ctx context.Context, peerDID string) (float64, error) {
				return repStore.GetScore(ctx, peerDID)
			}
		}

		pg = paygate.New(paygate.Config{
			PricingFn:    pricingFn,
			ReputationFn: reputationFn,
			TrustCfg:     trustCfg,
			LocalAddr:    walletAddr,
			ChainID:      pc.chainID,
			USDCAddr:     usdcAddr,
			RPCClient:    pc.rpcClient,
			Logger:       pLogger,
		})

		// Wire PayGate to handler via adapter.
		handler.SetPayGate(&payGateAdapter{gate: pg, chainID: pc.chainID, usdcAddr: usdcAddr})
		pLogger.Infow("P2P payment gate enabled",
			"perQuery", cfg.P2P.Pricing.PerQuery,
			"toolPrices", len(cfg.P2P.Pricing.ToolPrices),
			"postPayMinScore", trustCfg.PostPayMinScore,
		)

		// Wire settlement service for on-chain payment processing.
		if bus != nil && pc.rpcClient != nil && dbClient != nil {
			receiptTimeout := cfg.P2P.Pricing.Settlement.ReceiptTimeout
			if receiptTimeout <= 0 {
				receiptTimeout = 2 * time.Minute
			}
			maxRetries := cfg.P2P.Pricing.Settlement.MaxRetries
			if maxRetries <= 0 {
				maxRetries = 3
			}
			settleSvc := settlement.New(settlement.Config{
				Wallet:         wp,
				RPCClient:      pc.rpcClient,
				DBClient:       dbClient,
				ChainID:        pc.chainID,
				USDCAddr:       usdcAddr,
				ReceiptTimeout: receiptTimeout,
				MaxRetries:     maxRetries,
				Logger:         pLogger,
			})
			if repStore != nil {
				settleSvc.SetReputationRecorder(repStore)
			}
			settleSvc.Subscribe(bus)
			handler.SetEventBus(bus)
			pLogger.Info("P2P settlement service wired to event bus")
		}
	}

	localCard := &discovery.GossipCard{
		Name:   agentName,
		DID:    localDID,
		PeerID: node.PeerID().String(),
	}
	for _, a := range node.Multiaddrs() {
		localCard.Multiaddrs = append(localCard.Multiaddrs, a.String())
	}

	// Set pricing info on gossip card if pricing is enabled.
	if cfg.P2P.Pricing.Enabled {
		localCard.Pricing = &discovery.PricingInfo{
			Currency:   wallet.CurrencyUSDC,
			PerQuery:   cfg.P2P.Pricing.PerQuery,
			ToolPrices: cfg.P2P.Pricing.ToolPrices,
		}
	}

	gossip, err = discovery.NewGossipService(discovery.GossipConfig{
		Host:      node.Host(),
		LocalCard: localCard,
		Interval:  gossipInterval,
		Logger:    pLogger,
	})
	if err != nil {
		pLogger.Warnw("gossip service creation failed", "error", err)
	}

	// Set credential max age from config.
	if gossip != nil && cfg.P2P.ZKP.MaxCredentialAge != "" {
		if maxAge, err := time.ParseDuration(cfg.P2P.ZKP.MaxCredentialAge); err == nil {
			gossip.SetMaxCredentialAge(maxAge)
		}
	}

	pLogger.Infow("P2P networking initialized",
		"peerID", node.PeerID(),
		"did", localDID,
		"listenAddrs", cfg.P2P.ListenAddrs,
		"zkHandshake", cfg.P2P.ZKHandshake,
		"firewallRules", len(aclRules),
	)

	// Build a pricing function for external use (e.g., approval wiring).
	var extPricingFn func(string) (string, bool)
	if cfg.P2P.Pricing.Enabled {
		extPricingFn = func(toolName string) (string, bool) {
			if price, ok := cfg.P2P.Pricing.ToolPrices[toolName]; ok {
				return price, false
			}
			if cfg.P2P.Pricing.PerQuery != "" {
				return cfg.P2P.Pricing.PerQuery, false
			}
			return "", true
		}
	}

	// Create agent pool and selector for dynamic P2P agent management.
	pool := agentpool.New(pLogger)
	selector := agentpool.NewSelector(pool, agentpool.DefaultWeights())
	provider := agentpool.NewPoolProvider(pool, selector)

	// Create team coordinator for distributed agent collaboration.
	var coord *team.Coordinator
	invokeFn := func(ctx context.Context, peerID, toolName string, params map[string]interface{}) (map[string]interface{}, error) {
		// Default invoke function — can be overridden via handler wiring.
		return nil, fmt.Errorf("P2P team invoke not configured for peer %s", peerID)
	}
	coord = team.NewCoordinator(team.CoordinatorConfig{
		Pool:     pool,
		Selector: selector,
		InvokeFn: invokeFn,
		Bus:      bus,
		Logger:   pLogger,
	})

	pLogger.Infow("P2P agent pool and team coordinator initialized",
		"selectorWeights", "default",
	)

	return &p2pComponents{
		node:        node,
		sessions:    sessions,
		handshaker:  handshaker,
		fw:          fw,
		gossip:      gossip,
		identity:    idProvider,
		handler:     handler,
		payGate:     pg,
		reputation:  repStore,
		pricingCfg:  cfg.P2P.Pricing,
		pricingFn:   extPricingFn,
		agentPool:   pool,
		selector:    selector,
		coordinator: coord,
		provider:    provider,
	}
}

// payGateAdapter adapts paygate.Gate to protocol.PayGateChecker.
type payGateAdapter struct {
	gate     *paygate.Gate
	chainID  int64
	usdcAddr common.Address
}

func (a *payGateAdapter) Check(peerDID, toolName string, payload map[string]interface{}) (p2pproto.PayGateResult, error) {
	result, err := a.gate.Check(peerDID, toolName, payload)
	if err != nil {
		return p2pproto.PayGateResult{}, err
	}
	pgr := p2pproto.PayGateResult{
		Status:       string(result.Status),
		SettlementID: result.SettlementID,
	}
	if result.Auth != nil {
		pgr.Auth = result.Auth
	}
	if result.PriceQuote != nil {
		pgr.PriceQuote = map[string]interface{}{
			"toolName":     result.PriceQuote.ToolName,
			"price":        result.PriceQuote.Price,
			"currency":     result.PriceQuote.Currency,
			"usdcContract": result.PriceQuote.USDCContract,
			"chainId":      result.PriceQuote.ChainID,
			"sellerAddr":   result.PriceQuote.SellerAddr,
			"quoteExpiry":  result.PriceQuote.QuoteExpiry,
			"isFree":       false,
		}
	}
	return pgr, nil
}

// initZKP creates ZKP components if enabled.
func initZKP(cfg *config.Config) *zkp.ProverService {
	if !cfg.P2P.ZKHandshake && !cfg.P2P.ZKAttestation {
		return nil
	}

	prover, err := zkp.NewProverService(zkp.Config{
		CacheDir: cfg.P2P.ZKP.ProofCacheDir,
		Scheme:   zkp.ProofScheme(cfg.P2P.ZKP.ProvingScheme),
		SRSMode:  zkp.SRSMode(cfg.P2P.ZKP.SRSMode),
		SRSPath:  cfg.P2P.ZKP.SRSPath,
		Logger:   logger(),
	})
	if err != nil {
		logger().Warnw("ZKP prover init error, skipping", "error", err)
		return nil
	}

	// Compile all 4 circuits.
	circuitDefs := map[string]interface {
		Define(frontend.API) error
	}{
		"wallet_ownership":     &circuits.WalletOwnershipCircuit{},
		"response_attestation": &circuits.ResponseAttestationCircuit{},
		"balance_range":        &circuits.BalanceRangeCircuit{},
		"agent_capability":     &circuits.AgentCapabilityCircuit{},
	}

	for id, circuit := range circuitDefs {
		if err := prover.Compile(id, circuit); err != nil {
			logger().Warnw("ZKP circuit compile error", "circuitID", id, "error", err)
		}
	}

	logger().Infow("ZKP prover initialized",
		"scheme", prover.Scheme(),
		"circuits", len(circuitDefs),
	)
	return prover
}
