// Package settlement handles asynchronous on-chain settlement of P2P tool
// invocation payments. It subscribes to ToolExecutionPaidEvent from the event
// bus and submits transferWithAuthorization transactions to the USDC contract.
package settlement

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/ent/paymenttx"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/payment/eip3009"
	"github.com/langoai/lango/internal/wallet"
)

// ReputationRecorder records success/failure outcomes for reputation tracking.
type ReputationRecorder interface {
	RecordSuccess(ctx context.Context, peerDID string) error
	RecordFailure(ctx context.Context, peerDID string) error
}

// Config holds construction parameters for a settlement Service.
type Config struct {
	Wallet         wallet.WalletProvider
	RPCClient      *ethclient.Client
	DBClient       *ent.Client
	ChainID        int64
	USDCAddr       common.Address
	ReceiptTimeout time.Duration // default: 2m
	MaxRetries     int           // default: 3
	Logger         *zap.SugaredLogger
}

// Service processes on-chain settlement for paid tool invocations.
type Service struct {
	wallet     wallet.WalletProvider
	rpc        *ethclient.Client
	db         *ent.Client
	chainID    *big.Int
	usdcAddr   common.Address
	timeout    time.Duration
	maxRetries int
	reputation ReputationRecorder
	logger     *zap.SugaredLogger

	// nonceMu serializes transaction building to avoid nonce collisions.
	nonceMu sync.Mutex
	// wg tracks in-flight handleEvent goroutines for graceful shutdown.
	wg sync.WaitGroup
}

// New creates a settlement service with the given configuration.
func New(cfg Config) *Service {
	timeout := cfg.ReceiptTimeout
	if timeout <= 0 {
		timeout = 2 * time.Minute
	}
	maxRetries := cfg.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}

	return &Service{
		wallet:     cfg.Wallet,
		rpc:        cfg.RPCClient,
		db:         cfg.DBClient,
		chainID:    big.NewInt(cfg.ChainID),
		usdcAddr:   cfg.USDCAddr,
		timeout:    timeout,
		maxRetries: maxRetries,
		logger:     cfg.Logger,
	}
}

// SetReputationRecorder sets the reputation recorder for post-settlement
// success/failure tracking.
func (s *Service) SetReputationRecorder(r ReputationRecorder) {
	s.reputation = r
}

// Subscribe registers the settlement service as a subscriber to
// ToolExecutionPaidEvent on the given event bus.
func (s *Service) Subscribe(bus *eventbus.Bus) {
	eventbus.SubscribeTyped(bus, func(evt eventbus.ToolExecutionPaidEvent) {
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.handleEvent(evt)
		}()
	})
	s.logger.Info("settlement service subscribed to tool.execution.paid events")
}

// Close waits for all in-flight settlement goroutines to finish.
func (s *Service) Close() {
	s.wg.Wait()
}

// handleEvent processes a single paid tool execution event.
func (s *Service) handleEvent(evt eventbus.ToolExecutionPaidEvent) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout+30*time.Second)
	defer cancel()

	auth, ok := evt.Auth.(*eip3009.Authorization)
	if !ok || auth == nil {
		s.logger.Warnw("settlement event missing valid authorization",
			"peerDID", evt.PeerDID, "tool", evt.ToolName)
		return
	}

	if err := s.settle(ctx, auth, evt.PeerDID, evt.ToolName); err != nil {
		s.logger.Errorw("settlement failed",
			"peerDID", evt.PeerDID, "tool", evt.ToolName, "error", err)
		if s.reputation != nil {
			_ = s.reputation.RecordFailure(ctx, evt.PeerDID)
		}
		return
	}

	if s.reputation != nil {
		_ = s.reputation.RecordSuccess(ctx, evt.PeerDID)
	}
}

// settle executes the full settlement lifecycle:
// 1. Create DB record (pending)
// 2. Build transaction (calldata + EIP-1559)
// 3. Sign transaction via wallet
// 4. Submit with retry
// 5. Wait for on-chain confirmation
func (s *Service) settle(ctx context.Context, auth *eip3009.Authorization, peerDID, toolName string) error {
	if s.db == nil {
		return fmt.Errorf("db client not configured")
	}

	// 1. Create DB record.
	txRecord, err := s.db.PaymentTx.Create().
		SetID(uuid.New()).
		SetFromAddress(auth.From.Hex()).
		SetToAddress(auth.To.Hex()).
		SetAmount(auth.Value.String()).
		SetChainID(s.chainID.Int64()).
		SetStatus(paymenttx.StatusPending).
		SetPaymentMethod(paymenttx.PaymentMethodP2pSettlement).
		SetPurpose(fmt.Sprintf("p2p settlement: %s from %s", toolName, peerDID)).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create payment record: %w", err)
	}

	s.logger.Infow("settlement record created",
		"id", txRecord.ID, "tool", toolName, "peerDID", peerDID)

	// 2. Build settlement transaction.
	signedTx, err := s.buildAndSignTx(ctx, auth)
	if err != nil {
		s.updateStatus(ctx, txRecord.ID, paymenttx.StatusFailed, "", err.Error())
		return fmt.Errorf("build/sign tx: %w", err)
	}

	// 3. Submit with retry.
	txHash, err := s.submitWithRetry(ctx, signedTx)
	if err != nil {
		s.updateStatus(ctx, txRecord.ID, paymenttx.StatusFailed, "", err.Error())
		return fmt.Errorf("submit tx: %w", err)
	}

	// Update DB with submitted status.
	s.updateStatus(ctx, txRecord.ID, paymenttx.StatusSubmitted, txHash, "")
	s.logger.Infow("settlement tx submitted", "txHash", txHash, "id", txRecord.ID)

	// 4. Wait for confirmation.
	if err := s.waitForConfirmation(ctx, common.HexToHash(txHash)); err != nil {
		s.updateStatus(ctx, txRecord.ID, paymenttx.StatusFailed, txHash, err.Error())
		return fmt.Errorf("wait confirmation: %w", err)
	}

	s.updateStatus(ctx, txRecord.ID, paymenttx.StatusConfirmed, txHash, "")
	s.logger.Infow("settlement confirmed", "txHash", txHash, "id", txRecord.ID)
	return nil
}

// buildAndSignTx constructs the transferWithAuthorization calldata, builds an
// EIP-1559 transaction, and signs it with the wallet.
func (s *Service) buildAndSignTx(ctx context.Context, auth *eip3009.Authorization) (*types.Transaction, error) {
	s.nonceMu.Lock()
	defer s.nonceMu.Unlock()

	calldata := eip3009.EncodeCalldata(auth)

	fromAddr, err := s.wallet.Address(ctx)
	if err != nil {
		return nil, fmt.Errorf("get wallet address: %w", err)
	}
	from := common.HexToAddress(fromAddr)

	nonce, err := s.rpc.PendingNonceAt(ctx, from)
	if err != nil {
		return nil, fmt.Errorf("get nonce: %w", err)
	}

	gasLimit, err := s.rpc.EstimateGas(ctx, ethereum.CallMsg{
		From: from,
		To:   &s.usdcAddr,
		Data: calldata,
	})
	if err != nil {
		return nil, fmt.Errorf("estimate gas: %w", err)
	}

	header, err := s.rpc.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("get block header: %w", err)
	}

	baseFee := header.BaseFee
	if baseFee == nil {
		baseFee = big.NewInt(1_000_000_000) // 1 gwei fallback
	}

	maxPriorityFee := big.NewInt(1_500_000_000) // 1.5 gwei
	maxFee := new(big.Int).Add(
		new(big.Int).Mul(baseFee, big.NewInt(2)),
		maxPriorityFee,
	)

	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   s.chainID,
		Nonce:     nonce,
		GasFeeCap: maxFee,
		GasTipCap: maxPriorityFee,
		Gas:       gasLimit,
		To:        &s.usdcAddr,
		Value:     big.NewInt(0),
		Data:      calldata,
	})

	// Serialize unsigned tx for wallet signing.
	signer := types.LatestSignerForChainID(s.chainID)
	txHash := signer.Hash(tx)

	sig, err := s.wallet.SignTransaction(ctx, txHash.Bytes())
	if err != nil {
		return nil, fmt.Errorf("sign tx: %w", err)
	}

	signedTx, err := tx.WithSignature(signer, sig)
	if err != nil {
		return nil, fmt.Errorf("apply signature: %w", err)
	}

	return signedTx, nil
}

// submitWithRetry sends the signed transaction with exponential backoff.
func (s *Service) submitWithRetry(ctx context.Context, tx *types.Transaction) (string, error) {
	var lastErr error
	for attempt := 0; attempt < s.maxRetries; attempt++ {
		err := s.rpc.SendTransaction(ctx, tx)
		if err == nil {
			return tx.Hash().Hex(), nil
		}
		lastErr = err
		s.logger.Warnw("settlement tx submission failed, retrying",
			"attempt", attempt+1, "error", err)

		backoff := time.Duration(1<<uint(attempt)) * time.Second
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(backoff):
		}
	}
	return "", fmt.Errorf("submit after %d retries: %w", s.maxRetries, lastErr)
}

// waitForConfirmation polls for transaction receipt with exponential backoff.
func (s *Service) waitForConfirmation(ctx context.Context, txHash common.Hash) error {
	deadline := time.After(s.timeout)
	backoff := 1 * time.Second
	maxBackoff := 16 * time.Second

	for {
		receipt, err := s.rpc.TransactionReceipt(ctx, txHash)
		if err == nil {
			if receipt.Status == types.ReceiptStatusSuccessful {
				return nil
			}
			return fmt.Errorf("tx reverted: status=%d", receipt.Status)
		}

		select {
		case <-deadline:
			return fmt.Errorf("receipt timeout after %v", s.timeout)
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
		}

		if backoff < maxBackoff {
			backoff *= 2
		}
	}
}

// updateStatus updates the payment transaction record in the database.
func (s *Service) updateStatus(ctx context.Context, id uuid.UUID, status paymenttx.Status, txHash, errMsg string) {
	update := s.db.PaymentTx.UpdateOneID(id).SetStatus(status)
	if txHash != "" {
		update = update.SetTxHash(txHash)
	}
	if errMsg != "" {
		update = update.SetErrorMessage(errMsg)
	}
	if err := update.Exec(ctx); err != nil {
		s.logger.Warnw("update payment tx status",
			"id", id, "status", status, "error", err)
	}
}
